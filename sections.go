package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// Section is a section within a project.
type Section struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Order       int    `json:"section_order"`
	IsArchived  bool   `json:"is_archived"`
	WorkspaceID string `json:"workspace_id"`
}

// GetSectionsArgs controls listing and pagination of sections.
type GetSectionsArgs struct {
	ProjectID string
	Cursor    string
	Limit     int
}

func (a *GetSectionsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.ProjectID != "" {
		q.Set("project_id", a.ProjectID)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateSectionArgs are the parameters for creating a section. Name and
// ProjectID are required.
type CreateSectionArgs struct {
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
	Order     *int   `json:"order,omitempty"`
}

// UpdateSectionArgs are the parameters for updating a section.
type UpdateSectionArgs struct {
	Name string `json:"name"`
}

// GetSections returns a single page of sections.
func (c *Client) GetSections(ctx context.Context, args *GetSectionsArgs) (Page[Section], error) {
	return doList[Section](ctx, c, "/sections", args.query())
}

// Sections returns an iterator over all sections, following pagination.
func (c *Client) Sections(ctx context.Context, args *GetSectionsArgs) iter.Seq2[Section, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Section], error) {
		return doList[Section](ctx, c, "/sections", setCursor(base, cursor))
	})
}

// GetSection returns a single section by ID.
func (c *Client) GetSection(ctx context.Context, id string) (Section, error) {
	return doGet[Section](ctx, c, "/sections/"+id)
}

// CreateSection creates a new section.
func (c *Client) CreateSection(ctx context.Context, args CreateSectionArgs) (Section, error) {
	return doPost[Section](ctx, c, "/sections", args)
}

// UpdateSection updates an existing section.
func (c *Client) UpdateSection(ctx context.Context, id string, args UpdateSectionArgs) (Section, error) {
	return doPost[Section](ctx, c, "/sections/"+id, args)
}

// DeleteSection deletes a section and all of its tasks.
func (c *Client) DeleteSection(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/sections/"+id)
}
