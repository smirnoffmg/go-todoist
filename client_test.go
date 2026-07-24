package todoist

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testClient spins up an httptest server with the given handler and returns a
// Client pointed at it.
func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("test-token", WithBaseURL(srv.URL))
}

func TestAuthorizationHeader(t *testing.T) {
	var gotAuth, gotUA string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"id":"1","content":"x"}`))
	})

	if _, err := c.GetTask(context.Background(), "1"); err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-token")
	}
	if gotUA != defaultUserAgent {
		t.Errorf("User-Agent = %q, want %q", gotUA, defaultUserAgent)
	}
}

func TestErrorMapping(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "42")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	})

	_, err := c.GetTask(context.Background(), "1")
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *Error", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want 429", apiErr.StatusCode)
	}
	if apiErr.RetryAfter.Seconds() != 42 {
		t.Errorf("RetryAfter = %v, want 42s", apiErr.RetryAfter)
	}
	if !apiErr.Temporary() {
		t.Error("Temporary() = false, want true")
	}
}

func TestGetTasksSendsFilters(t *testing.T) {
	var gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
	})

	_, err := c.GetTasks(context.Background(), &GetTasksArgs{ProjectID: "p1", Limit: 50})
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	if gotQuery != "limit=50&project_id=p1" {
		t.Errorf("query = %q", gotQuery)
	}
}

func TestTasksIteratorFollowsCursor(t *testing.T) {
	pages := map[string]string{
		"":      `{"results":[{"id":"1"},{"id":"2"}],"next_cursor":"page2"}`,
		"page2": `{"results":[{"id":"3"}],"next_cursor":""}`,
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, ok := pages[r.URL.Query().Get("cursor")]
		if !ok {
			t.Errorf("unexpected cursor %q", r.URL.Query().Get("cursor"))
			http.Error(w, "bad cursor", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(body))
	})

	var ids []string
	for task, err := range c.Tasks(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iteration error: %v", err)
		}
		ids = append(ids, task.ID)
	}
	if got, want := len(ids), 3; got != want {
		t.Fatalf("collected %d tasks, want %d", got, want)
	}
	if ids[0] != "1" || ids[2] != "3" {
		t.Errorf("ids = %v", ids)
	}
}

func TestTasksIteratorStopsOnError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	var sawErr bool
	for _, err := range c.Tasks(context.Background(), nil) {
		if err != nil {
			sawErr = true
			break
		}
	}
	if !sawErr {
		t.Error("expected an error during iteration")
	}
}

func TestCreateTaskOmitsUnsetFields(t *testing.T) {
	var body map[string]any
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"id":"1","content":"buy milk"}`))
	})

	_, err := c.CreateTask(context.Background(), CreateTaskArgs{Content: "buy milk"})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, ok := body["priority"]; ok {
		t.Error("priority should be omitted when unset")
	}
	if body["content"] != "buy milk" {
		t.Errorf("content = %v", body["content"])
	}
}
