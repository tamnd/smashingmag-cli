// Package smashingmag is the library behind the smashingmag command: the HTTP
// client, request shaping, and typed data models for Smashing Magazine.
//
// The client fetches the public RSS 2.0 feed at
// https://www.smashingmagazine.com/feed/. No authentication is required. It
// sets a real User-Agent, paces requests, and retries transient 429/5xx errors
// with exponential backoff.
package smashingmag

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to Smashing Magazine.
const DefaultUserAgent = "smashingmag/dev (+https://github.com/tamnd/smashingmag-cli)"

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://www.smashingmagazine.com",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to the Smashing Magazine RSS feed.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// Articles fetches articles from the RSS feed. limit <= 0 means return all items.
func (c *Client) Articles(ctx context.Context, limit int) ([]Article, error) {
	url := c.cfg.BaseURL + "/feed/"
	raw, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}

	var root rssRoot
	if err := xml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("decode rss: %w", err)
	}

	items := root.Channel.Items
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	out := make([]Article, 0, len(items))
	for i, item := range items {
		out = append(out, wireToArticle(item, i+1))
	}
	return out, nil
}

// get fetches a URL with retry/backoff logic.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, url)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/rss+xml, text/xml, */*")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
