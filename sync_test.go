package todoist

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

func TestSyncFormEncoding(t *testing.T) {
	var form url.Values
	var contentType string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		form, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(`{"sync_token":"newtoken","full_sync":true,"sync_status":{}}`))
	})

	cmd := NewCommand("item_add", "tmp1", map[string]any{"content": "write tests"})
	resp, err := c.Sync(context.Background(), SyncRequest{
		Commands: []Command{cmd},
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q", contentType)
	}
	if form.Get("sync_token") != "*" {
		t.Errorf("sync_token = %q, want *", form.Get("sync_token"))
	}
	if form.Get("resource_types") != `["all"]` {
		t.Errorf("resource_types = %q", form.Get("resource_types"))
	}

	var cmds []Command
	if err := json.Unmarshal([]byte(form.Get("commands")), &cmds); err != nil {
		t.Fatalf("commands not valid JSON: %v", err)
	}
	if len(cmds) != 1 || cmds[0].Type != "item_add" || cmds[0].TempID != "tmp1" {
		t.Errorf("commands = %+v", cmds)
	}
	if resp.SyncToken != "newtoken" || !resp.FullSync {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSyncResponseErrAggregatesFailures(t *testing.T) {
	resp := &SyncResponse{
		SyncStatus: map[string]json.RawMessage{
			"ok-uuid":  json.RawMessage(`"ok"`),
			"bad-uuid": json.RawMessage(`{"error_code":15,"error":"nope"}`),
		},
	}
	err := resp.Err()
	if err == nil {
		t.Fatal("expected error for failed command")
	}

	ok := &SyncResponse{SyncStatus: map[string]json.RawMessage{"u": json.RawMessage(`"ok"`)}}
	if ok.Err() != nil {
		t.Errorf("Err() = %v, want nil for all-ok", ok.Err())
	}
}

func TestNewUUIDFormat(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	seen := map[string]bool{}
	for range 100 {
		u := NewUUID()
		if !re.MatchString(u) {
			t.Fatalf("UUID %q not a valid v4", u)
		}
		if seen[u] {
			t.Fatalf("duplicate UUID %q", u)
		}
		seen[u] = true
	}
}
