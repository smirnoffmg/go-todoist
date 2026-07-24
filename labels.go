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
