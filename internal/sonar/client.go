// Package sonar wraps the subset of the SonarQube REST API qctx needs.
package sonar

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
)

// Options configures a Sonar API client.
type Options struct {
	BaseURL   string
	Token     string
	BasicAuth bool
	HTTP      *httpclient.Client
}

// Client is a thin wrapper over the SonarQube web API.
type Client struct {
	baseURL    string
	auth       string
	http       *httpclient.Client
	rulesMu    sync.Mutex
	rulesCache map[string]string
}

// New constructs a Sonar Client. BaseURL is required.
func New(o Options) (*Client, error) {
	if strings.TrimSpace(o.BaseURL) == "" {
		return nil, errors.New("sonar: base URL is required")
	}
	if o.HTTP == nil {
		return nil, errors.New("sonar: HTTP client is required")
	}
	c := &Client{baseURL: strings.TrimRight(o.BaseURL, "/"), http: o.HTTP}
	if o.Token != "" {
		if o.BasicAuth {
			c.auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(o.Token+":"))
		} else {
			c.auth = "Bearer " + o.Token
		}
	}
	return c, nil
}

// GetJSON performs GET path?q and decodes the JSON body into out.
func (c *Client) GetJSON(path string, q map[string]string, out any) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("sonar: parse url: %w", err)
	}
	qs := u.Query()
	for k, v := range q {
		qs.Set(k, v)
	}
	u.RawQuery = qs.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Errorf("sonar: build request: %w", err)
	}
	if c.auth != "" {
		req.Header.Set("Authorization", c.auth)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("sonar: do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("sonar: %d %s: %s", resp.StatusCode, resp.Status, strings.TrimSpace(string(body)))
	}
	if out == nil {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(out); err != nil && err != io.EOF {
		return fmt.Errorf("sonar: decode: %w", err)
	}
	return nil
}
