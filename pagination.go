package todoist

import (
	"context"
	"iter"
	"maps"
	"net/http"
	"net/url"
)

// Page is a single page of a cursor-paginated list response. NextCursor is
// empty when there are no further pages.
type Page[T any] struct {
	Results    []T    `json:"results"`
	NextCursor string `json:"next_cursor"`
}

func doList[T any](ctx context.Context, c *Client, path string, q url.Values) (Page[T], error) {
	var page Page[T]
	err := c.do(ctx, http.MethodGet, path, q, nil, &page)
	return page, err
}

func doGet[T any](ctx context.Context, c *Client, path string) (T, error) {
	var out T
	err := c.do(ctx, http.MethodGet, path, nil, nil, &out)
	return out, err
}

func doPost[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	var out T
	err := c.do(ctx, http.MethodPost, path, nil, body, &out)
	return out, err
}

// doAction performs a POST with no request or response body (e.g. close,
// reopen, archive).
func doAction(ctx context.Context, c *Client, path string) error {
	return c.do(ctx, http.MethodPost, path, nil, nil, nil)
}

func doDelete(ctx context.Context, c *Client, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}

// paginate turns a page-fetching function into an iterator over individual
// items. Iteration stops after yielding an error or when a page reports no next
// cursor. The fetch function receives the cursor for the page to load ("" for
// the first page).
func paginate[T any](fetch func(cursor string) (Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		cursor := ""
		for {
			page, err := fetch(cursor)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for _, item := range page.Results {
				if !yield(item, nil) {
					return
				}
			}
			if page.NextCursor == "" {
				return
			}
			cursor = page.NextCursor
		}
	}
}

// setCursor returns a copy of q with the pagination cursor set (or the cursor
// key removed when empty).
func setCursor(q url.Values, cursor string) url.Values {
	out := url.Values{}
	maps.Copy(out, q)
	if cursor == "" {
		out.Del("cursor")
	} else {
		out.Set("cursor", cursor)
	}
	return out
}
