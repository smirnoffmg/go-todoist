package todoist

// Filter is a saved filter query.
//
// The Todoist API v1 does not expose filters as REST endpoints; they are managed
// through the Sync API using the "filter_add", "filter_update" and
// "filter_delete" commands, and returned in SyncResponse.Filters when
// "filters" is among the requested resource types. See Sync and NewCommand.
type Filter struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Query      string `json:"query"`
	Color      string `json:"color"`
	Order      int    `json:"item_order"`
	IsFavorite bool   `json:"is_favorite"`
}
