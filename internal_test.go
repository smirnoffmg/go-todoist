package todoist

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// errTransport fails every request with a fixed error.
type errTransport struct{ err error }

func (t errTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, t.err }

// errBody is an io.ReadCloser whose Read always fails.
type errBody struct{}

func (errBody) Read([]byte) (int, error) { return 0, errors.New("read boom") }
func (errBody) Close() error             { return nil }

// readErrTransport returns a 200 response whose body fails to read.
type readErrTransport struct{}

func (readErrTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: errBody{}, Header: make(http.Header)}, nil
}

func TestOptions(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	t.Cleanup(srv.Close)

	c := New("t",
		WithBaseURL(srv.URL),
		WithUserAgent("custom-agent"),
		WithHTTPClient(&http.Client{}),
		WithHTTPClient(nil), // nil is ignored
		WithBaseURL(""),     // empty is ignored
		WithUserAgent(""),   // empty is ignored
	)
	if _, err := c.GetTask(context.Background(), "1"); err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if gotUA != "custom-agent" {
		t.Errorf("User-Agent = %q, want custom-agent", gotUA)
	}
}

func TestErrorMessage(t *testing.T) {
	withBody := &Error{Status: "400 Bad Request", Body: `{"error":"nope"}`}
	if got := withBody.Error(); got != `todoist: 400 Bad Request: {"error":"nope"}` {
		t.Errorf("Error() = %q", got)
	}
	empty := &Error{Status: "500 Internal Server Error"}
	if got := empty.Error(); got != "todoist: 500 Internal Server Error" {
		t.Errorf("Error() = %q", got)
	}
}

func TestParseRetryAfterNonNumeric(t *testing.T) {
	if d := parseRetryAfter("not-a-number"); d != 0 {
		t.Errorf("parseRetryAfter = %v, want 0", d)
	}
	if d := parseRetryAfter(""); d != 0 {
		t.Errorf("parseRetryAfter empty = %v, want 0", d)
	}
}

func TestPtr(t *testing.T) {
	p := Ptr(42)
	if p == nil || *p != 42 {
		t.Fatalf("Ptr(42) = %v", p)
	}
	s := Ptr("hi")
	if *s != "hi" {
		t.Errorf("Ptr(hi) = %q", *s)
	}
}

func TestDoMarshalError(t *testing.T) {
	c := New("t")
	err := c.do(context.Background(), http.MethodPost, "/x", nil, map[string]any{"bad": make(chan int)}, nil)
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestDoRequestError(t *testing.T) {
	c := New("t", WithBaseURL("http://%zz"))
	err := c.do(context.Background(), http.MethodGet, "/x", nil, nil, nil)
	if err == nil {
		t.Fatal("expected request construction error")
	}
}

func TestDoFormRequestError(t *testing.T) {
	c := New("t", WithBaseURL("http://%zz"))
	err := c.doForm(context.Background(), "/sync", url.Values{}, nil)
	if err == nil {
		t.Fatal("expected request construction error")
	}
}

func TestSendTransportError(t *testing.T) {
	c := New("t", WithHTTPClient(&http.Client{Transport: errTransport{err: errors.New("dial boom")}}))
	_, err := c.GetTask(context.Background(), "1")
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestSendReadBodyError(t *testing.T) {
	c := New("t", WithHTTPClient(&http.Client{Transport: readErrTransport{}}))
	_, err := c.GetTask(context.Background(), "1")
	if err == nil {
		t.Fatal("expected body read error")
	}
}

func TestSendEmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK) // empty body, out is non-nil
	}))
	t.Cleanup(srv.Close)
	c := New("t", WithBaseURL(srv.URL))
	task, err := c.GetTask(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.ID != "" {
		t.Errorf("task = %+v, want zero value", task)
	}
}

func TestPaginateEarlyBreak(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"1"},{"id":"2"}],"next_cursor":"more"}`))
	}))
	t.Cleanup(srv.Close)
	c := New("t", WithBaseURL(srv.URL))

	var count int
	for _, err := range c.Tasks(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		count++
		break // exercises the yield-returns-false path
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestSyncWithTokenAndResourceTypes(t *testing.T) {
	var form url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(`{"sync_token":"t2"}`))
	}))
	t.Cleanup(srv.Close)
	c := New("t", WithBaseURL(srv.URL))

	_, err := c.Sync(context.Background(), SyncRequest{
		SyncToken:     "prev-token",
		ResourceTypes: []string{"projects", "items"},
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if form.Get("sync_token") != "prev-token" {
		t.Errorf("sync_token = %q", form.Get("sync_token"))
	}
	if form.Get("resource_types") != `["projects","items"]` {
		t.Errorf("resource_types = %q", form.Get("resource_types"))
	}
	if _, ok := form["commands"]; ok {
		t.Error("commands should be absent when none are provided")
	}
}

func TestSyncCommandMarshalError(t *testing.T) {
	c := New("t")
	_, err := c.Sync(context.Background(), SyncRequest{
		Commands: []Command{{Type: "item_add", UUID: "u", Args: make(chan int)}},
	})
	if err == nil {
		t.Fatal("expected commands marshal error")
	}
}

func TestSyncTransportError(t *testing.T) {
	c := New("t", WithHTTPClient(&http.Client{Transport: errTransport{err: errors.New("boom")}}))
	_, err := c.Sync(context.Background(), SyncRequest{})
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestNewUUIDReadError(t *testing.T) {
	orig := randReader
	randReader = errBody{}
	defer func() { randReader = orig }()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on random read failure")
		}
	}()
	_ = NewUUID()
}
