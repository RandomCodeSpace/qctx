// Package httpclient builds a *http.Client with retry, custom CA, proxy, and per-request extra headers.
package httpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

type Options struct {
	CACertPath   string
	Insecure     bool
	ExtraHeaders map[string]string
	Timeout      time.Duration
	MaxRetries   int
	RetryWait    time.Duration
}

type Client struct {
	inner   *retryablehttp.Client
	headers map[string]string
}

func New(opts Options) (*Client, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}
	if opts.RetryWait == 0 {
		opts.RetryWait = 500 * time.Millisecond
	}

	tlsCfg, err := buildTLSConfig(opts.CACertPath, opts.Insecure)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSClientConfig:       tlsCfg,
		MaxIdleConns:          50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	rc := retryablehttp.NewClient()
	rc.HTTPClient = &http.Client{Transport: transport, Timeout: opts.Timeout}
	rc.RetryMax = opts.MaxRetries
	rc.RetryWaitMin = opts.RetryWait
	rc.RetryWaitMax = 8 * opts.RetryWait
	rc.CheckRetry = retryPolicy
	rc.Backoff = retryablehttp.DefaultBackoff
	rc.Logger = nil

	if opts.Insecure {
		fmt.Fprintln(os.Stderr, "qctx: WARNING --insecure / QCTX_INSECURE set — TLS verification disabled")
	}
	return &Client{inner: rc, headers: opts.ExtraHeaders}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for k, v := range c.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	rr, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.inner.Do(rr)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("method", req.Method).Str("url", req.URL.String()).Int("status", resp.StatusCode).Msg("http")
	return resp, nil
}

func buildTLSConfig(caPath string, insecure bool) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecure}
	if caPath == "" {
		return cfg, nil
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	ca, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert %q: %w", caPath, err)
	}
	if !pool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("no certificates parsed from %q", caPath)
	}
	cfg.RootCAs = pool
	return cfg, nil
}

func retryPolicy(_ context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil || resp == nil {
		return true, nil
	}
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		return true, nil
	}
	return false, nil
}
