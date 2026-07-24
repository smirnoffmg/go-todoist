package todoist

import (
	"context"
	"net/http"
	"testing"
)

func TestSearchProjects(t *testing.T) {
	var gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":[{"id":"p1","name":"Work"}],"next_cursor":""}`))
	})

	page, err := c.SearchProjects(context.Background(), &SearchProjectsArgs{Query: "work", Limit: 10})
	if err != nil {
		t.Fatalf("SearchProjects: %v", err)
	}
	if gotPath != "/projects/search" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "limit=10&query=work" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(page.Results) != 1 || page.Results[0].ID != "p1" {
		t.Errorf("results = %+v", page.Results)
	}
}

func TestSearchSectionsWithProjectFilter(t *testing.T) {
	var gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":[{"id":"s1"}],"next_cursor":""}`))
	})

	if _, err := c.SearchSections(context.Background(), &SearchSectionsArgs{Query: "phase", ProjectID: "p1", Cursor: "c"}); err != nil {
		t.Fatalf("SearchSections: %v", err)
	}
	if gotPath != "/sections/search" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "cursor=c&project_id=p1&query=phase" {
		t.Errorf("query = %q", gotQuery)
	}
}

func TestSearchLabelsIterator(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/labels/search" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"l1","name":"urgent"}],"next_cursor":""}`))
	})

	var count int
	for label, err := range c.SearchLabelsSeq(context.Background(), &SearchLabelsArgs{Query: "urg"}) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		if label.Name != "urgent" {
			t.Errorf("label = %+v", label)
		}
		count++
	}
	if count != 1 {
		t.Errorf("count = %d", count)
	}
}

func TestSearchNilArgsAndProjectIterator(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"p1"}],"next_cursor":""}`))
	})
	ctx := context.Background()

	// nil-args query branch for each search type
	var sp *SearchProjectsArgs
	var ss *SearchSectionsArgs
	var sl *SearchLabelsArgs
	if len(sp.query()) != 0 || len(ss.query()) != 0 || len(sl.query()) != 0 {
		t.Error("nil args should produce empty query")
	}

	var count int
	for _, err := range c.SearchProjectsSeq(ctx, nil) {
		if err != nil {
			t.Fatalf("iter: %v", err)
		}
		count++
	}
	for _, err := range c.SearchSectionsSeq(ctx, nil) {
		if err != nil {
			t.Fatalf("iter: %v", err)
		}
	}
	if count != 1 {
		t.Errorf("count = %d", count)
	}
}
