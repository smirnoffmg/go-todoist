package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// Label is a personal label.
type Label struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Order      int    `json:"item_order"`
	IsFavorite bool   `json:"is_favorite"`
}

// GetLabelsArgs controls listing and pagination of labels.
type GetLabelsArgs struct {
	Cursor string
	Limit  int
}

func (a *GetLabelsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateLabelArgs are the parameters for creating a label. Name is required.
type CreateLabelArgs struct {
	Name       string  `json:"name"`
	Color      *string `json:"color,omitempty"`
	Order      *int    `json:"order,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
}

// UpdateLabelArgs are the parameters for updating a label. Only non-nil fields
// are sent.
type UpdateLabelArgs struct {
	Name       *string `json:"name,omitempty"`
	Color      *string `json:"color,omitempty"`
	Order      *int    `json:"order,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
}

// GetLabels returns a single page of labels.
func (c *Client) GetLabels(ctx context.Context, args *GetLabelsArgs) (Page[Label], error) {
	return doList[Label](ctx, c, "/labels", args.query())
}

// Labels returns an iterator over all labels, following pagination.
func (c *Client) Labels(ctx context.Context, args *GetLabelsArgs) iter.Seq2[Label, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Label], error) {
		return doList[Label](ctx, c, "/labels", setCursor(base, cursor))
	})
}

// GetLabel returns a single label by ID.
func (c *Client) GetLabel(ctx context.Context, id string) (Label, error) {
	return doGet[Label](ctx, c, "/labels/"+id)
}

// CreateLabel creates a new personal label.
func (c *Client) CreateLabel(ctx context.Context, args CreateLabelArgs) (Label, error) {
	return doPost[Label](ctx, c, "/labels", args)
}

// UpdateLabel updates an existing label.
func (c *Client) UpdateLabel(ctx context.Context, id string, args UpdateLabelArgs) (Label, error) {
	return doPost[Label](ctx, c, "/labels/"+id, args)
}

// DeleteLabel deletes a label.
func (c *Client) DeleteLabel(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/labels/"+id)
}

// GetSharedLabelsArgs controls listing of shared labels. When OmitPersonal is
// true, labels that also exist as personal labels are excluded.
type GetSharedLabelsArgs struct {
	OmitPersonal bool
	Cursor       string
	Limit        int
}

func (a *GetSharedLabelsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.OmitPersonal {
		q.Set("omit_personal", "true")
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// GetSharedLabels returns a page of shared label names (labels used on tasks but
// not saved as personal labels).
func (c *Client) GetSharedLabels(ctx context.Context, args *GetSharedLabelsArgs) (Page[string], error) {
	return doList[string](ctx, c, "/labels/shared", args.query())
}

// SharedLabels returns an iterator over all shared label names, following
// pagination.
func (c *Client) SharedLabels(ctx context.Context, args *GetSharedLabelsArgs) iter.Seq2[string, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[string], error) {
		return doList[string](ctx, c, "/labels/shared", setCursor(base, cursor))
	})
}

// RenameSharedLabel renames a shared label across all tasks that use it.
func (c *Client) RenameSharedLabel(ctx context.Context, name, newName string) error {
	return doActionBody(ctx, c, "/labels/shared/rename", map[string]string{"name": name, "new_name": newName})
}

// RemoveSharedLabel removes a shared label from all tasks that use it.
func (c *Client) RemoveSharedLabel(ctx context.Context, name string) error {
	return doActionBody(ctx, c, "/labels/shared/remove", map[string]string{"name": name})
}
