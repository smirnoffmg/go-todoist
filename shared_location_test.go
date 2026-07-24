package todoist

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetSharedLabels(t *testing.T) {
	var gotPath, gotQuery string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"results":["waiting","follow-up"],"next_cursor":""}`))
	})

	page, err := c.GetSharedLabels(context.Background(), &GetSharedLabelsArgs{OmitPersonal: true, Cursor: "cur", Limit: 20})
	if err != nil {
		t.Fatalf("GetSharedLabels: %v", err)
	}
	if gotPath != "/labels/shared" {
		t.Errorf("path = %q", gotPath)
	}
	if gotQuery != "cursor=cur&limit=20&omit_personal=true" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(page.Results) != 2 || page.Results[0] != "waiting" {
		t.Errorf("results = %+v", page.Results)
	}

	var count int
	for _, err := range c.SharedLabels(context.Background(), nil) {
		if err != nil {
			t.Fatalf("iteration: %v", err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("count = %d", count)
	}
}

func TestRenameAndRemoveSharedLabel(t *testing.T) {
	var gotPath string
	var body map[string]string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body = map[string]string{}
		_ = json.NewDecoder(r.Body).Decode(&body)
	})

	if err := c.RenameSharedLabel(context.Background(), "waiting", "follow-up"); err != nil {
		t.Fatalf("RenameSharedLabel: %v", err)
	}
	if gotPath != "/labels/shared/rename" || body["name"] != "waiting" || body["new_name"] != "follow-up" {
		t.Errorf("rename: path=%q body=%+v", gotPath, body)
	}

	if err := c.RemoveSharedLabel(context.Background(), "waiting"); err != nil {
		t.Fatalf("RemoveSharedLabel: %v", err)
	}
	if gotPath != "/labels/shared/remove" || body["name"] != "waiting" {
		t.Errorf("remove: path=%q body=%+v", gotPath, body)
	}
}

func TestLocationReminderLifecycle(t *testing.T) {
	type call struct{ method, path string }
	var got call
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = call{r.Method, r.URL.Path}
		_, _ = w.Write([]byte(`{"id":"lr1","name":"Home"}`))
	})
	ctx := context.Background()

	cases := []struct {
		name             string
		run              func() error
		method, wantPath string
	}{
		{"list", func() error { _, e := c.GetLocationReminders(ctx, &GetLocationRemindersArgs{TaskID: "t1"}); return e }, http.MethodGet, "/location_reminders"},
		{"get", func() error { _, e := c.GetLocationReminder(ctx, "lr1"); return e }, http.MethodGet, "/location_reminders/lr1"},
		{"create", func() error {
			_, e := c.CreateLocationReminder(ctx, CreateLocationReminderArgs{TaskID: "t1", Name: "Home", LocLat: "1.0", LocLong: "2.0", LocTrigger: "on_enter"})
			return e
		}, http.MethodPost, "/location_reminders"},
		{"update", func() error {
			_, e := c.UpdateLocationReminder(ctx, "lr1", UpdateLocationReminderArgs{Name: Ptr("Work")})
			return e
		}, http.MethodPost, "/location_reminders/lr1"},
		{"delete", func() error { return c.DeleteLocationReminder(ctx, "lr1") }, http.MethodDelete, "/location_reminders/lr1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if got.method != tc.method || got.path != tc.wantPath {
				t.Errorf("got %s %s, want %s %s", got.method, got.path, tc.method, tc.wantPath)
			}
		})
	}

	// nil-args query branch
	if _, err := c.GetLocationReminders(ctx, nil); err != nil {
		t.Fatalf("GetLocationReminders(nil): %v", err)
	}
}
