package todoist

import (
	"context"
	"iter"
	"net/url"
	"strconv"
)

// Reminder is a reminder attached to a task.
type Reminder struct {
	ID         string `json:"id"`
	ItemID     string `json:"item_id"`
	Type       string `json:"type"`
	Due        *Due   `json:"due"`
	MinuteBias int    `json:"minute_offset"`
	Name       string `json:"name"`
	LocLat     string `json:"loc_lat"`
	LocLong    string `json:"loc_long"`
	LocTrigger string `json:"loc_trigger"`
	Radius     int    `json:"radius"`
	IsDeleted  bool   `json:"is_deleted"`
}

// GetRemindersArgs controls listing and pagination of reminders.
type GetRemindersArgs struct {
	Cursor string
	Limit  int
}

func (a *GetRemindersArgs) query() url.Values {
	q := url.Values{}
	if a == nil {
		return q
	}
	if a.Cursor != "" {
		q.Set("cursor", a.Cursor)
	}
	if a.Limit > 0 {
		q.Set("limit", strconv.Itoa(a.Limit))
	}
	return q
}

// CreateReminderArgs are the parameters for creating a reminder. ItemID and
// Type are required.
type CreateReminderArgs struct {
	ItemID      string  `json:"item_id"`
	Type        string  `json:"type"`
	DueString   *string `json:"due_string,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	DueDatetime *string `json:"due_datetime,omitempty"`
	MinuteBias  *int    `json:"minute_offset,omitempty"`
	Name        *string `json:"name,omitempty"`
	LocLat      *string `json:"loc_lat,omitempty"`
	LocLong     *string `json:"loc_long,omitempty"`
	LocTrigger  *string `json:"loc_trigger,omitempty"`
	Radius      *int    `json:"radius,omitempty"`
}

// UpdateReminderArgs are the parameters for updating a reminder.
type UpdateReminderArgs struct {
	Type        *string `json:"type,omitempty"`
	DueString   *string `json:"due_string,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	DueDatetime *string `json:"due_datetime,omitempty"`
	MinuteBias  *int    `json:"minute_offset,omitempty"`
	Name        *string `json:"name,omitempty"`
}

// GetReminders returns a single page of reminders.
func (c *Client) GetReminders(ctx context.Context, args *GetRemindersArgs) (Page[Reminder], error) {
	return doList[Reminder](ctx, c, "/reminders", args.query())
}

// Reminders returns an iterator over all reminders, following pagination.
func (c *Client) Reminders(ctx context.Context, args *GetRemindersArgs) iter.Seq2[Reminder, error] {
	base := args.query()
	return paginate(func(cursor string) (Page[Reminder], error) {
		return doList[Reminder](ctx, c, "/reminders", setCursor(base, cursor))
	})
}

// CreateReminder creates a new reminder.
func (c *Client) CreateReminder(ctx context.Context, args CreateReminderArgs) (Reminder, error) {
	return doPost[Reminder](ctx, c, "/reminders", args)
}

// UpdateReminder updates an existing reminder.
func (c *Client) UpdateReminder(ctx context.Context, id string, args UpdateReminderArgs) (Reminder, error) {
	return doPost[Reminder](ctx, c, "/reminders/"+id, args)
}

// DeleteReminder deletes a reminder.
func (c *Client) DeleteReminder(ctx context.Context, id string) error {
	return doDelete(ctx, c, "/reminders/"+id)
}
