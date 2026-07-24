package todoist

import "context"

// User is the authenticated user's profile.
type User struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	FullName     string  `json:"full_name"`
	InboxID      string  `json:"inbox_project_id"`
	TZInfo       TZInfo  `json:"tz_info"`
	Lang         string  `json:"lang"`
	DateFormat   int     `json:"date_format"`
	TimeFormat   int     `json:"time_format"`
	StartDay     int     `json:"start_day"`
	Karma        float64 `json:"karma"`
	IsPremium    bool    `json:"is_premium"`
	PremiumUntil string  `json:"premium_until"`
	AvatarBig    string  `json:"avatar_big"`
}

// TZInfo is the user's timezone configuration.
type TZInfo struct {
	Timezone  string `json:"timezone"`
	GMTString string `json:"gmt_string"`
	Hours     int    `json:"hours"`
	Minutes   int    `json:"minutes"`
	IsDST     int    `json:"is_dst"`
}

// ProductivityStats holds the user's karma and completion statistics.
type ProductivityStats struct {
	KarmaLastUpdate float64        `json:"karma_last_update"`
	KarmaTrend      string         `json:"karma_trend"`
	Karma           float64        `json:"karma"`
	CompletedCount  int            `json:"completed_count"`
	DaysItems       []DayStats     `json:"days_items"`
	WeekItems       []WeekStats    `json:"week_items"`
	Goals           map[string]any `json:"goals"`
}

// DayStats holds completion counts for a single day.
type DayStats struct {
	Date           string `json:"date"`
	TotalCompleted int    `json:"total_completed"`
}

// WeekStats holds completion counts for a single week.
type WeekStats struct {
	From           string `json:"from"`
	To             string `json:"to"`
	TotalCompleted int    `json:"total_completed"`
}

// GetUser returns the authenticated user's profile.
func (c *Client) GetUser(ctx context.Context) (User, error) {
	return doGet[User](ctx, c, "/user")
}

// GetProductivityStats returns the authenticated user's completed-task and karma
// statistics.
func (c *Client) GetProductivityStats(ctx context.Context) (ProductivityStats, error) {
	return doGet[ProductivityStats](ctx, c, "/tasks/completed/stats")
}
