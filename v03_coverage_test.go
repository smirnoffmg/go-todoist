package todoist

import (
	"context"
	"net/http"
	"testing"
)

// Exercises the remaining query() branches and page methods for v0.3.0.
func TestV03QueryBranches(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
	})
	ctx := context.Background()

	calls := []func() error{
		func() error {
			_, e := c.SearchProjects(ctx, &SearchProjectsArgs{Query: "q", Cursor: "c", Limit: 1})
			return e
		},
		func() error {
			_, e := c.SearchSections(ctx, &SearchSectionsArgs{Query: "q", ProjectID: "p", Cursor: "c", Limit: 1})
			return e
		},
		func() error {
			_, e := c.SearchLabels(ctx, &SearchLabelsArgs{Query: "q", Cursor: "c", Limit: 1})
			return e
		},
		func() error { _, e := c.GetFolders(ctx, "ws1", &GetFoldersArgs{Cursor: "c", Limit: 1}); return e },
	}
	for i, fn := range calls {
		if err := fn(); err != nil {
			t.Errorf("call %d: %v", i, err)
		}
	}

	drain(t, c.SearchLabelsSeq(ctx, nil))
	drain(t, c.Folders(ctx, "ws1", nil))
}
