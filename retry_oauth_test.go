package todoist

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetrySucceedsAfter429(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"id":"1","content":"ok"}`))
	})
	c.maxRetries = 3
	c.retryBaseDelay = time.Millisecond

	task, err := c.GetTask(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.ID != "1" {
		t.Errorf("task = %+v", task)
	}
	if got := calls.Load(); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestRetryExhaustedReturnsError(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	c.maxRetries = 2
	c.retryBaseDelay = time.Millisecond

	_, err := c.GetTask(context.Background(), "1")
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("err = %v", err)
	}
	if got := calls.Load(); got != 3 { // initial + 2 retries
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestRetryReplaysRequestBody(t *testing.T) {
	var calls atomic.Int32
	var lastBody string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		lastBody = string(buf)
		if calls.Add(1) < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"id":"1"}`))
	})
	c.maxRetries = 2
	c.retryBaseDelay = time.Millisecond

	_, err := c.CreateTask(context.Background(), CreateTaskArgs{Content: "retry me"})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if lastBody == "" || lastBody[0] != '{' {
		t.Errorf("body not replayed on retry: %q", lastBody)
	}
}

func TestRetryRespectsContextCancellation(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	c.maxRetries = 5
	c.retryBaseDelay = time.Hour // force the wait so the ctx deadline wins in sleepCtx

	// The first request succeeds (429); cancellation must interrupt the backoff.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := c.GetTask(ctx, "1")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestNoRetryByDefault(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, _ = c.GetTask(context.Background(), "1")
	if got := calls.Load(); got != 1 {
		t.Errorf("calls = %d, want 1 (no retry by default)", got)
	}
}

func TestWithRetryOption(t *testing.T) {
	c := New("t", WithRetry(3))
	if c.maxRetries != 3 || c.retryBaseDelay != 500*time.Millisecond {
		t.Errorf("maxRetries=%d base=%v", c.maxRetries, c.retryBaseDelay)
	}
}

func TestRevokeToken(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
	})

	err := c.RevokeToken(context.Background(), "cid", "csecret", "atoken")
	if err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/access_tokens" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}
	q, _ := url.ParseQuery(gotQuery)
	if q.Get("client_id") != "cid" || q.Get("client_secret") != "csecret" || q.Get("access_token") != "atoken" {
		t.Errorf("query = %q", gotQuery)
	}
}

func TestMigratePersonalToken(t *testing.T) {
	var gotPath string
	var body map[string]string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"access_token":"newtok","token_type":"Bearer"}`))
	})

	tok, err := c.MigratePersonalToken(context.Background(), "cid", "csecret", "ptoken", "data:read,data:delete")
	if err != nil {
		t.Fatalf("MigratePersonalToken: %v", err)
	}
	if gotPath != "/access_tokens/migrate_personal_token" {
		t.Errorf("path = %q", gotPath)
	}
	if body["personal_token"] != "ptoken" || body["scope"] != "data:read,data:delete" {
		t.Errorf("body = %+v", body)
	}
	if tok.AccessToken != "newtok" || tok.TokenType != "Bearer" {
		t.Errorf("token = %+v", tok)
	}
}
