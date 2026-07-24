package todoist

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestGetCompletedByCompletionDate(t *testing.T) {
	var gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		// Note: the completed endpoints wrap results in "items".
		_, _ = w.Write([]byte(`{"items":[{"id":"1"},{"id":"2"}],"next_cursor":""}`))
	})

	page, err := c.GetCompletedByCompletionDate(context.Background(), &GetCompletedTasksArgs{
		Since:     time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		Until:     time.Date(2025, 2, 28, 23, 59, 59, 0, time.UTC),
		ProjectID: "p1",
		Limit:     50,
	})
	if err != nil {
		t.Fatalf("GetCompletedByCompletionDate: %v", err)
	}
	if gotPath != "/tasks/completed/by_completion_date" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "limit=50&project_id=p1&since=2025-02-01T00%3A00%3A00Z&until=2025-02-28T23%3A59%3A59Z" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(page.Results) != 2 || page.Results[0].ID != "1" {
		t.Errorf("results = %+v", page.Results)
	}
}

func TestCompletedByDueDateIterator(t *testing.T) {
	pages := map[string]string{
		"":  `{"items":[{"id":"1"}],"next_cursor":"n"}`,
		"n": `{"items":[{"id":"2"}],"next_cursor":""}`,
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks/completed/by_due_date" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(pages[r.URL.Query().Get("cursor")]))
	})

	var ids []string
	for task, err := range c.CompletedByDueDate(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		ids = append(ids, task.ID)
	}
	if len(ids) != 2 || ids[1] != "2" {
		t.Errorf("ids = %v", ids)
	}
}

func TestGetTasksByFilter(t *testing.T) {
	var gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":[{"id":"1","content":"today task"}],"next_cursor":""}`))
	})

	page, err := c.GetTasksByFilter(context.Background(), &GetTasksByFilterArgs{Query: "today & @work", Lang: "en"})
	if err != nil {
		t.Fatalf("GetTasksByFilter: %v", err)
	}
	if gotPath != "/tasks/filter" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "lang=en&query=today+%26+%40work" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(page.Results) != 1 {
		t.Errorf("results = %+v", page.Results)
	}
}

func TestTasksByFilterIterator(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"1"}],"next_cursor":""}`))
	})
	var count int
	for _, err := range c.TasksByFilter(context.Background(), &GetTasksByFilterArgs{Query: "today"}) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("count = %d", count)
	}
}

func TestCompletedTasksArgsNilQuery(t *testing.T) {
	var a *GetCompletedTasksArgs
	if len(a.query()) != 0 {
		t.Error("nil args should produce empty query")
	}
	var f *GetTasksByFilterArgs
	if len(f.query()) != 0 {
		t.Error("nil filter args should produce empty query")
	}
}
