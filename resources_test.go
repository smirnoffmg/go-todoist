package todoist

import (
	"context"
	"net/http"
	"testing"
)

func TestProjectLifecycleMethodsAndPaths(t *testing.T) {
	type call struct {
		method string
		path   string
	}
	var got call
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = call{r.Method, r.URL.Path}
		_, _ = w.Write([]byte(`{"id":"p1","name":"Work"}`))
	})
	ctx := context.Background()

	cases := []struct {
		name       string
		run        func() error
		wantMethod string
		wantPath   string
	}{
		{"create", func() error { _, e := c.CreateProject(ctx, CreateProjectArgs{Name: "Work"}); return e }, http.MethodPost, "/projects"},
		{"get", func() error { _, e := c.GetProject(ctx, "p1"); return e }, http.MethodGet, "/projects/p1"},
		{"update", func() error { _, e := c.UpdateProject(ctx, "p1", UpdateProjectArgs{}); return e }, http.MethodPost, "/projects/p1"},
		{"archive", func() error { return c.ArchiveProject(ctx, "p1") }, http.MethodPost, "/projects/p1/archive"},
		{"unarchive", func() error { return c.UnarchiveProject(ctx, "p1") }, http.MethodPost, "/projects/p1/unarchive"},
		{"delete", func() error { return c.DeleteProject(ctx, "p1") }, http.MethodDelete, "/projects/p1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if got.method != tc.wantMethod || got.path != tc.wantPath {
				t.Errorf("got %s %s, want %s %s", got.method, got.path, tc.wantMethod, tc.wantPath)
			}
		})
	}
}

func TestGetCommentsRequiresParentFilter(t *testing.T) {
	var gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":[{"id":"c1","content":"hi"}],"next_cursor":""}`))
	})

	page, err := c.GetComments(context.Background(), &GetCommentsArgs{TaskID: "t1"})
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if gotQuery != "task_id=t1" {
		t.Errorf("query = %q, want task_id=t1", gotQuery)
	}
	if len(page.Results) != 1 || page.Results[0].ID != "c1" {
		t.Errorf("results = %+v", page.Results)
	}
}
