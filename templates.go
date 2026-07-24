package todoist

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// ProjectImportResponse is returned when a template is imported into an existing
// project.
type ProjectImportResponse struct {
	Status       string    `json:"status"`
	TemplateType string    `json:"template_type"`
	Projects     []Project `json:"projects"`
	Sections     []Section `json:"sections"`
	Tasks        []Task    `json:"tasks"`
	Comments     []Comment `json:"comments"`
	ProjectNotes []Comment `json:"project_notes"`
}

// ProjectImportCreateResponse is returned when a template creates a new project.
type ProjectImportCreateResponse struct {
	ProjectImportResponse
	ProjectID string `json:"project_id"`
}

// FileURLResponse points to a generated template file.
type FileURLResponse struct {
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
}

// CreateProjectFromFile creates a new project from a CSV template file. When
// workspaceID is non-nil the project is created in that workspace.
func (c *Client) CreateProjectFromFile(ctx context.Context, name string, workspaceID *string, file io.Reader, fileName string) (ProjectImportCreateResponse, error) {
	fields := map[string]string{"name": name}
	if workspaceID != nil {
		fields["workspace_id"] = *workspaceID
	}
	var out ProjectImportCreateResponse
	err := c.doMultipart(ctx, "/templates/create_project_from_file", fields, "file", fileName, file, &out)
	return out, err
}

// ImportIntoProjectFromFile imports a CSV template file into an existing project.
func (c *Client) ImportIntoProjectFromFile(ctx context.Context, projectID string, file io.Reader, fileName string) (ProjectImportResponse, error) {
	fields := map[string]string{"project_id": projectID}
	var out ProjectImportResponse
	err := c.doMultipart(ctx, "/templates/import_into_project_from_file", fields, "file", fileName, file, &out)
	return out, err
}

// ImportIntoProjectFromTemplateID imports a shared template into an existing
// project. Locale is optional (e.g. "en").
func (c *Client) ImportIntoProjectFromTemplateID(ctx context.Context, projectID, templateID string, locale *string) (ProjectImportResponse, error) {
	body := map[string]string{"project_id": projectID, "template_id": templateID}
	if locale != nil {
		body["locale"] = *locale
	}
	return doPost[ProjectImportResponse](ctx, c, "/templates/import_into_project_from_template_id", body)
}

// GetTemplateFile returns a project exported as a CSV template, as raw text.
func (c *Client) GetTemplateFile(ctx context.Context, projectID string, useRelativeDates bool) (string, error) {
	q := templateQuery(projectID, useRelativeDates)
	var raw []byte
	if err := c.do(ctx, http.MethodGet, "/templates/file", q, nil, &raw); err != nil {
		return "", err
	}
	return string(raw), nil
}

// GetTemplateURL returns a URL to a project exported as a template file.
func (c *Client) GetTemplateURL(ctx context.Context, projectID string, useRelativeDates bool) (FileURLResponse, error) {
	var out FileURLResponse
	err := c.do(ctx, http.MethodGet, "/templates/url", templateQuery(projectID, useRelativeDates), nil, &out)
	return out, err
}

func templateQuery(projectID string, useRelativeDates bool) url.Values {
	q := url.Values{}
	q.Set("project_id", projectID)
	if useRelativeDates {
		q.Set("use_relative_dates", strconv.FormatBool(useRelativeDates))
	}
	return q
}
