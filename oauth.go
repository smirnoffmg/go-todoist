package todoist

import (
	"context"
	"net/http"
	"net/url"
)

// AccessToken is an OAuth2 access token returned by MigratePersonalToken.
type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// RevokeToken revokes an OAuth2 access token. It authenticates with the
// application's client credentials rather than the client's own token.
func (c *Client) RevokeToken(ctx context.Context, clientID, clientSecret, accessToken string) error {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("client_secret", clientSecret)
	q.Set("access_token", accessToken)
	return c.do(ctx, http.MethodDelete, "/access_tokens", q, nil, nil)
}

// MigratePersonalToken exchanges a personal API token for an OAuth2 access token
// scoped to the given application. Scope is a comma-separated list of OAuth
// scopes.
func (c *Client) MigratePersonalToken(ctx context.Context, clientID, clientSecret, personalToken, scope string) (AccessToken, error) {
	body := map[string]string{
		"client_id":      clientID,
		"client_secret":  clientSecret,
		"personal_token": personalToken,
		"scope":          scope,
	}
	return doPost[AccessToken](ctx, c, "/access_tokens/migrate_personal_token", body)
}
