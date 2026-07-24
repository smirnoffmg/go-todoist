package todoist

// Ptr returns a pointer to v. It is a convenience for setting the optional
// pointer fields on request argument structs inline.
func Ptr[T any](v T) *T { return &v }

// Priority levels for tasks. The API uses 1 (natural, lowest) through 4
// (urgent, highest); note this is the reverse of what the Todoist UI shows.
const (
	PriorityNormal = 1
	PriorityMedium = 2
	PriorityHigh   = 3
	PriorityUrgent = 4
)

// Due represents the due date of a task.
type Due struct {
	Date        string `json:"date"`
	String      string `json:"string"`
	Lang        string `json:"lang"`
	IsRecurring bool   `json:"is_recurring"`
	Timezone    string `json:"timezone,omitempty"`
}

// Duration represents the estimated duration of a task.
type Duration struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

// Deadline represents a task deadline (a full-day date).
type Deadline struct {
	Date string `json:"date"`
	Lang string `json:"lang,omitempty"`
}

// Attachment describes a file attached to a comment.
type Attachment struct {
	ResourceType string `json:"resource_type,omitempty"`
	FileName     string `json:"file_name,omitempty"`
	FileType     string `json:"file_type,omitempty"`
	FileURL      string `json:"file_url,omitempty"`
	FileSize     int    `json:"file_size,omitempty"`
	UploadState  string `json:"upload_state,omitempty"`
}
