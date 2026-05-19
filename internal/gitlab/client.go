// Package gitlab wraps the GitLab REST API endpoints qctx uses.
package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/RandomCodeSpace/qctx/internal/httpclient"
)

// Options configures a GitLab API client.
type Options struct {
	BaseURL string
	Token   string
	HTTP    *httpclient.Client
}

// Client is a thin wrapper over a subset of the GitLab REST API (v4).
type Client struct {
	baseURL string
	token   string
	http    *httpclient.Client
}

// New constructs a GitLab Client. BaseURL is required.
func New(o Options) (*Client, error) {
	if strings.TrimSpace(o.BaseURL) == "" {
		return nil, errors.New("gitlab: base URL is required")
	}
	if o.HTTP == nil {
		return nil, errors.New("gitlab: HTTP client is required")
	}
	return &Client{
		baseURL: strings.TrimRight(o.BaseURL, "/"),
		token:   o.Token,
		http:    o.HTTP,
	}, nil
}

// BaseURL exposes the configured host (used by mrurl parser tests).
func (c *Client) BaseURL() string { return c.baseURL }

// GetJSON performs GET path?q and decodes JSON into out. Path may be absolute.
func (c *Client) GetJSON(path string, q map[string]string, out any) error {
	resp, err := c.do(http.MethodGet, path, q, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return decodeJSON(resp, out)
}

// GetRaw performs GET and returns the raw response (used for traces).
func (c *Client) GetRaw(path string, q map[string]string, headers http.Header) (*http.Response, error) {
	return c.do(http.MethodGet, path, q, headers)
}

// ListJSON repeatedly calls GET path while following the X-Next-Page header.
// Each page is decoded into a value produced by newPage and passed to onPage.
func (c *Client) ListJSON(path string, q map[string]string, newPage func() any, onPage func(any) error) error {
	page := "1"
	per := "100"
	for {
		params := map[string]string{"page": page, "per_page": per}
		for k, v := range q {
			params[k] = v
		}
		resp, err := c.do(http.MethodGet, path, params, nil)
		if err != nil {
			return err
		}
		v := newPage()
		if err := decodeJSON(resp, v); err != nil {
			_ = resp.Body.Close()
			return err
		}
		next := resp.Header.Get("X-Next-Page")
		_ = resp.Body.Close()
		if err := onPage(v); err != nil {
			return err
		}
		if next == "" {
			return nil
		}
		page = next
	}
}

func (c *Client) do(method, path string, q map[string]string, headers http.Header) (*http.Response, error) {
	full := path
	if !strings.HasPrefix(path, "http") {
		full = c.baseURL + path
	}
	u, err := url.Parse(full)
	if err != nil {
		return nil, fmt.Errorf("gitlab: parse url: %w", err)
	}
	qs := u.Query()
	for k, v := range q {
		qs.Set(k, v)
	}
	u.RawQuery = qs.Encode()
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("gitlab: build req: %w", err)
	}
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitlab: do: %w", err)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("gitlab: %d %s: %s", resp.StatusCode, resp.Status, strings.TrimSpace(string(body)))
	}
	return resp, nil
}

func decodeJSON(resp *http.Response, out any) error {
	if out == nil {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(out); err != nil && err != io.EOF {
		return fmt.Errorf("gitlab: decode: %w", err)
	}
	return nil
}
