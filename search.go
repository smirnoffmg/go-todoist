package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// SearchProjectsArgs are the parameters for searching projects. Query is
// required.
type SearchProjectsArgs struct {
	Query  string
	Cursor string
	Limit  int
}

func (a *SearchProjectsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.Query != "" {
		q.Set("query", a.Query)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// SearchProjects returns a page of projects matching the query.
func (c *Client) SearchProjects(ctx context.Context, args *SearchProjectsArgs) (Page[Project], error) {
	return doList[Project](ctx, c, "/projects/search", args.query())
}

// SearchProjectsSeq iterates over all projects matching the query, following
// pagination.
func (c *Client) SearchProjectsSeq(ctx context.Context, args *SearchProjectsArgs) iter.Seq2[Project, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Project], error) {
		return doList[Project](ctx, c, "/projects/search", setCursor(base, cursor))
	})
}

// SearchSectionsArgs are the parameters for searching sections. Query is
// required.
type SearchSectionsArgs struct {
	Query     string
	ProjectID string
	Cursor    string
	Limit     int
}

func (a *SearchSectionsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.Query != "" {
		q.Set("query", a.Query)
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

// SearchSections returns a page of sections matching the query.
func (c *Client) SearchSections(ctx context.Context, args *SearchSectionsArgs) (Page[Section], error) {
	return doList[Section](ctx, c, "/sections/search", args.query())
}

// SearchSectionsSeq iterates over all sections matching the query, following
// pagination.
func (c *Client) SearchSectionsSeq(ctx context.Context, args *SearchSectionsArgs) iter.Seq2[Section, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Section], error) {
		return doList[Section](ctx, c, "/sections/search", setCursor(base, cursor))
	})
}

// SearchLabelsArgs are the parameters for searching labels. Query is required.
type SearchLabelsArgs struct {
	Query  string
	Cursor string
	Limit  int
}

func (a *SearchLabelsArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.Query != "" {
		q.Set("query", a.Query)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// SearchLabels returns a page of labels matching the query.
func (c *Client) SearchLabels(ctx context.Context, args *SearchLabelsArgs) (Page[Label], error) {
	return doList[Label](ctx, c, "/labels/search", args.query())
}

// SearchLabelsSeq iterates over all labels matching the query, following
// pagination.
func (c *Client) SearchLabelsSeq(ctx context.Context, args *SearchLabelsArgs) iter.Seq2[Label, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Label], error) {
		return doList[Label](ctx, c, "/labels/search", setCursor(base, cursor))
	})
}
