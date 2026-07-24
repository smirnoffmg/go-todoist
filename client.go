package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the root of the Todoist API v1.
const DefaultBaseURL = "https://api.todoist.com/api/v1"

const defaultUserAgent = "go-todoist"

// Client is a Todoist API v1 client. It is safe for concurrent use as long as
// its configuration is not mutated after construction.
type Client struct {
	token      string
	baseURL    string
	userAgent  string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets the underlying HTTP client used for requests.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithBaseURL overrides the API base URL. Useful for testing against a stub
// server. The trailing slash, if any, is trimmed.
func WithBaseURL(raw string) Option {
	return func(c *Client) {
		if raw != "" {
			c.baseURL = strings.TrimRight(raw, "/")
		}
	}
}

// WithUserAgent sets the User-Agent header sent with each request.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// New returns a Client authenticated with the given personal or OAuth2 token.
func New(token string, opts ...Option) *Client {
	c := &Client{
		token:      token,
		baseURL:    DefaultBaseURL,
		userAgent:  defaultUserAgent,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// do performs an HTTP request against the API. When body is non-nil it is
// JSON-encoded. On a 2xx response, out (if non-nil) is JSON-decoded from the
// body; otherwise an *Error is returned.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return err
	}
	c.setHeaders(req, body != nil)

	return c.send(req, out)
}

// doForm performs an application/x-www-form-urlencoded POST, used by the Sync
// endpoint. On a 2xx response, out (if non-nil) is JSON-decoded from the body.
func (c *Client) doForm(ctx context.Context, path string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	c.setHeaders(req, false)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.send(req, out)
}

func (c *Client) setHeaders(req *http.Request, hasJSONBody bool) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if hasJSONBody {
		req.Header.Set("Content-Type", "application/json")
	}
}

func (c *Client) send(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(data),
			RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
		}
	}

	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 0
}
