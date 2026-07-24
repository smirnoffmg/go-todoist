package todoist

import (
	"fmt"
	"net/http"
	"time"
)

// Error is returned for any non-2xx response from the Todoist API. It carries
// the HTTP status and the raw response body so callers can inspect the failure.
type Error struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Status is the HTTP status text of the response.
	Status string
	// Body is the raw response body, typically a JSON error payload.
	Body string
	// RetryAfter is populated from the Retry-After header on 429 responses.
	RetryAfter time.Duration
}

func (e *Error) Error() string {
	if e.Body == "" {
		return "todoist: " + e.Status
	}
	return fmt.Sprintf("todoist: %s: %s", e.Status, e.Body)
}

// Temporary reports whether the error is likely transient (rate limiting or a
// server-side failure) and the request may succeed if retried.
func (e *Error) Temporary() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}
