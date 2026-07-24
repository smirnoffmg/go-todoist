package todoist

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// UploadResult describes a file uploaded to Todoist, suitable for attaching to a
// comment.
type UploadResult struct {
	FileURL      string `json:"file_url"`
	FileName     string `json:"file_name"`
	FileType     string `json:"file_type"`
	FileSize     int    `json:"file_size"`
	ResourceType string `json:"resource_type"`
	UploadState  string `json:"upload_state"`
	Image        string `json:"image"`
	ImageWidth   int    `json:"image_width"`
	ImageHeight  int    `json:"image_height"`
}

// Attachment converts an upload into the Attachment form accepted by
// CreateComment.
func (u UploadResult) Attachment() Attachment {
	return Attachment{
		ResourceType: u.ResourceType,
		FileName:     u.FileName,
		FileType:     u.FileType,
		FileURL:      u.FileURL,
		FileSize:     u.FileSize,
		UploadState:  u.UploadState,
	}
}

// UploadFile uploads a file and returns its metadata. When projectID is non-nil
// the upload is associated with that project.
func (c *Client) UploadFile(ctx context.Context, file io.Reader, fileName string, projectID *string) (UploadResult, error) {
	fields := map[string]string{"file_name": fileName}
	if projectID != nil {
		fields["project_id"] = *projectID
	}
	var out UploadResult
	err := c.doMultipart(ctx, "/uploads", fields, "file", fileName, file, &out)
	return out, err
}

// DeleteUpload deletes a previously uploaded file by its URL.
func (c *Client) DeleteUpload(ctx context.Context, fileURL string) error {
	q := url.Values{}
	q.Set("file_url", fileURL)
	return c.do(ctx, http.MethodDelete, "/uploads", q, nil, nil)
}
