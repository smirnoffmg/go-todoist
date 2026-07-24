package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// LocationReminder is a reminder that triggers on arriving at or leaving a
// geographic location.
type LocationReminder struct {
	ID         string `json:"id"`
	ItemID     string `json:"item_id"`
	ProjectID  string `json:"project_id"`
	NotifyUID  string `json:"notify_uid"`
	Name       string `json:"name"`
	LocLat     string `json:"loc_lat"`
	LocLong    string `json:"loc_long"`
	LocTrigger string `json:"loc_trigger"`
	Radius     int    `json:"radius"`
	Type       string `json:"type"`
	IsDeleted  bool   `json:"is_deleted"`
}

// GetLocationRemindersArgs controls listing and pagination of location
// reminders.
type GetLocationRemindersArgs struct {
	TaskID string
	Cursor string
	Limit  int
}

func (a *GetLocationRemindersArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.TaskID != "" {
		q.Set("task_id", a.TaskID)
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateLocationReminderArgs are the parameters for creating a location
// reminder. TaskID, Name, LocLat, LocLong and LocTrigger are required;
// LocTrigger is "on_enter" or "on_leave".
type CreateLocationReminderArgs struct {
	TaskID     string `json:"task_id"`
	Name       string `json:"name"`
	LocLat     string `json:"loc_lat"`
	LocLong    string `json:"loc_long"`
	LocTrigger string `json:"loc_trigger"`
	Radius     *int   `json:"radius,omitempty"`
}

// UpdateLocationReminderArgs are the parameters for updating a location
// reminder. Only non-nil fields are sent.
type UpdateLocationReminderArgs struct {
	Name       *string `json:"name,omitempty"`
	LocLat     *string `json:"loc_lat,omitempty"`
	LocLong    *string `json:"loc_long,omitempty"`
	LocTrigger *string `json:"loc_trigger,omitempty"`
	Radius     *int    `json:"radius,omitempty"`
}

// GetLocationReminders returns a single page of location reminders.
func (c *Client) GetLocationReminders(ctx context.Context, args *GetLocationRemindersArgs) (Page[LocationReminder], error) {
	return doList[LocationReminder](ctx, c, "/location_reminders", args.query())
}

// LocationReminders returns an iterator over all location reminders, following
// pagination.
func (c *Client) LocationReminders(ctx context.Context, args *GetLocationRemindersArgs) iter.Seq2[LocationReminder, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[LocationReminder], error) {
		return doList[LocationReminder](ctx, c, "/location_reminders", setCursor(base, cursor))
	})
}

// GetLocationReminder returns a single location reminder by ID.
func (c *Client) GetLocationReminder(ctx context.Context, id string) (LocationReminder, error) {
	return doGet[LocationReminder](ctx, c, "/location_reminders/"+id)
}

// CreateLocationReminder creates a new location reminder.
func (c *Client) CreateLocationReminder(ctx context.Context, args CreateLocationReminderArgs) (LocationReminder, error) {
	return doPost[LocationReminder](ctx, c, "/location_reminders", args)
}

// UpdateLocationReminder updates an existing location reminder.
func (c *Client) UpdateLocationReminder(ctx context.Context, id string, args UpdateLocationReminderArgs) (LocationReminder, error) {
	return doPost[LocationReminder](ctx, c, "/location_reminders/"+id, args)
}

// DeleteLocationReminder deletes a location reminder.
func (c *Client) DeleteLocationReminder(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/location_reminders/"+id)
}
