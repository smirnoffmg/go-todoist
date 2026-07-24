package todoist

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestV02MethodsAndIterators(t *testing.T) {
	c := allMethodsClient(t)
	ctx := context.Background()

	// Fully-populated args exercise every query() branch of the new endpoints.
	completed := &GetCompletedTasksArgs{
		Since: time.Unix(1, 0), Until: time.Unix(2, 0),
		WorkspaceID: "w", ProjectID: "p", SectionID: "s", ParentID: "pa",
		FilterQuery: "today", FilterLang: "en", Cursor: "c", Limit: 1,
	}
	filter := &GetTasksByFilterArgs{Query: "today", Lang: "en", Cursor: "c", Limit: 1}
	loc := &GetLocationRemindersArgs{TaskID: "t", Cursor: "c", Limit: 1}

	calls := []func() error{
		func() error { _, e := c.GetCompletedByCompletionDate(ctx, completed); return e },
		func() error { _, e := c.GetCompletedByDueDate(ctx, completed); return e },
		func() error { _, e := c.GetTasksByFilter(ctx, filter); return e },
		func() error { _, e := c.MoveTask(ctx, "1", MoveTaskArgs{ProjectID: Ptr("p2")}); return e },
		func() error { _, e := c.GetArchivedProjects(ctx, &GetProjectsArgs{Cursor: "c", Limit: 1}); return e },
		func() error { return c.ArchiveSection(ctx, "s1") },
		func() error { return c.UnarchiveSection(ctx, "s1") },
		func() error { return c.JoinProject(ctx, "p1") },
		func() error { return c.RenameSharedLabel(ctx, "a", "b") },
		func() error { return c.RemoveSharedLabel(ctx, "a") },
		func() error { _, e := c.GetLocationReminders(ctx, loc); return e },
		func() error { _, e := c.GetLocationReminder(ctx, "1"); return e },
		func() error {
			_, e := c.CreateLocationReminder(ctx, CreateLocationReminderArgs{TaskID: "t", Name: "H", LocLat: "1", LocLong: "2", LocTrigger: "on_enter", Radius: Ptr(100)})
			return e
		},
		func() error {
			_, e := c.UpdateLocationReminder(ctx, "1", UpdateLocationReminderArgs{Name: Ptr("W")})
			return e
		},
		func() error { return c.DeleteLocationReminder(ctx, "1") },
		func() error { return c.RevokeToken(ctx, "cid", "cs", "tok") },
		func() error { _, e := c.MigratePersonalToken(ctx, "cid", "cs", "pt", "data:read"); return e },
	}
	for i, fn := range calls {
		if err := fn(); err != nil {
			t.Errorf("call %d: %v", i, err)
		}
	}

	drain(t, c.CompletedByCompletionDate(ctx, nil))
	drain(t, c.CompletedByDueDate(ctx, nil))
	drain(t, c.TasksByFilter(ctx, nil))
	drain(t, c.ArchivedProjects(ctx, nil))
	drain(t, c.LocationReminders(ctx, nil))
}

func TestRetryHonorsRetryAfterHeader(t *testing.T) {
	var calls atomic.Int32
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 2 {
			w.Header().Set("Retry-After", "0") // parsed to 0 => sleepCtx no-op, but exercises the retryAfter path
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"id":"1"}`))
	})
	c.maxRetries = 2
	c.retryBaseDelay = time.Millisecond

	if _, err := c.GetTask(context.Background(), "1"); err != nil {
		t.Fatalf("GetTask: %v", err)
	}
}

func TestRetryDelayUsesRetryAfter(t *testing.T) {
	c := New("t", WithRetry(1))
	if got := c.retryDelay(0, 3*time.Second); got != 3*time.Second {
		t.Errorf("retryDelay with Retry-After = %v, want 3s", got)
	}
}

func TestSleepCtxZeroDurationReturnsImmediately(t *testing.T) {
	if err := sleepCtx(context.Background(), 0); err != nil {
		t.Errorf("sleepCtx(0) = %v", err)
	}
}

func TestSendGetBodyErrorOnRetry(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	c.maxRetries = 2
	c.retryBaseDelay = time.Millisecond

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/x", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	req.GetBody = func() (io.ReadCloser, error) { return nil, errors.New("getbody boom") }

	if err := c.send(req, nil); err == nil || err.Error() != "getbody boom" {
		t.Fatalf("err = %v, want getbody boom", err)
	}
}
