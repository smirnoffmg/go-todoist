package todoist

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
)

// randReader is the source of randomness for NewUUID. It is a package variable
// so tests can substitute a failing reader.
var randReader io.Reader = rand.Reader

// Command is a single write operation submitted to the Sync API. Multiple
// commands can be batched in one Sync call and may reference each other's
// temp_id values.
type Command struct {
	Type   string `json:"type"`
	UUID   string `json:"uuid"`
	TempID string `json:"temp_id,omitempty"`
	Args   any    `json:"args"`
}

// NewCommand builds a Command with a freshly generated UUID. Pass a non-empty
// tempID when the created object needs to be referenced by later commands in the
// same batch.
func NewCommand(commandType, tempID string, args any) Command {
	return Command{Type: commandType, UUID: NewUUID(), TempID: tempID, Args: args}
}

// SyncRequest is the payload for a call to the Sync API. An empty SyncToken is
// sent as "*", requesting a full sync.
type SyncRequest struct {
	SyncToken     string
	ResourceTypes []string
	Commands      []Command
}

// SyncResponse is the result of a Sync API call. Only the resource slices
// relevant to the requested ResourceTypes are populated.
type SyncResponse struct {
	SyncToken     string                     `json:"sync_token"`
	FullSync      bool                       `json:"full_sync"`
	SyncStatus    map[string]json.RawMessage `json:"sync_status"`
	TempIDMapping map[string]string          `json:"temp_id_mapping"`

	Projects  []Project  `json:"projects"`
	Items     []Task     `json:"items"`
	Sections  []Section  `json:"sections"`
	Labels    []Label    `json:"labels"`
	Notes     []Comment  `json:"notes"`
	Reminders []Reminder `json:"reminders"`
	Filters   []Filter   `json:"filters"`
	User      *User      `json:"user"`
}

// Sync performs a synchronization request against the /sync endpoint, batching
// any commands in req. Callers should persist the returned SyncToken to make
// subsequent incremental syncs.
func (c *Client) Sync(ctx context.Context, req SyncRequest) (*SyncResponse, error) {
	syncToken := req.SyncToken
	if syncToken == "" {
		syncToken = "*"
	}

	resourceTypes := req.ResourceTypes
	if resourceTypes == nil {
		resourceTypes = []string{"all"}
	}
	resourceJSON, _ := json.Marshal(resourceTypes) //nolint:errchkjson // marshaling a []string never fails

	form := url.Values{}
	form.Set("sync_token", syncToken)
	form.Set("resource_types", string(resourceJSON))

	if len(req.Commands) > 0 {
		commandsJSON, err := json.Marshal(req.Commands)
		if err != nil {
			return nil, err
		}
		form.Set("commands", string(commandsJSON))
	}

	var out SyncResponse
	if err := c.doForm(ctx, "/sync", form, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Err aggregates any per-command failures reported in SyncStatus. It returns nil
// when every command succeeded (status "ok") or when no commands were sent.
func (r *SyncResponse) Err() error {
	var failures []string
	for uuid, raw := range r.SyncStatus {
		var status string
		if err := json.Unmarshal(raw, &status); err == nil {
			if status == "ok" {
				continue
			}
		}
		failures = append(failures, fmt.Sprintf("%s: %s", uuid, strings.TrimSpace(string(raw))))
	}
	if len(failures) == 0 {
		return nil
	}
	sort.Strings(failures)
	return fmt.Errorf("todoist: sync command(s) failed: %s", strings.Join(failures, "; "))
}

// NewUUID returns a random RFC 4122 version 4 UUID string, suitable for command
// and temp IDs.
func NewUUID() string {
	var b [16]byte
	if _, err := io.ReadFull(randReader, b[:]); err != nil {
		panic(fmt.Errorf("todoist: failed to read random bytes for UUID: %w", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
