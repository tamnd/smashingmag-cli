package smashingmag_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/smashingmag-cli/smashingmag"
)

const mockRSSFeed = `<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
    <title>Articles on Smashing Magazine</title>
    <link>https://www.smashingmagazine.com/</link>
    <item>
      <title>Building Better CSS Variables</title>
      <link>https://www.smashingmagazine.com/2024/03/building-better-css-variables/</link>
      <description><![CDATA[<p>CSS custom properties, or variables, have become <strong>essential</strong> tools for modern web design. This article covers best practices.</p>]]></description>
      <pubDate>Fri, 15 Mar 2024 10:00:00 +0000</pubDate>
      <dc:creator>Jane Smith</dc:creator>
      <category>CSS</category>
    </item>
    <item>
      <title>Advanced TypeScript Patterns</title>
      <link>https://www.smashingmagazine.com/2024/03/advanced-typescript-patterns/</link>
      <description><![CDATA[TypeScript patterns for large-scale applications.]]></description>
      <pubDate>Mon, 11 Mar 2024 08:30:00 +0000</pubDate>
      <dc:creator>John Doe</dc:creator>
      <category>JavaScript</category>
    </item>
  </channel>
</rss>`

func newTestClient(ts *httptest.Server) *smashingmag.Client {
	cfg := smashingmag.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return smashingmag.NewClient(cfg)
}

func TestArticlesSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockRSSFeed))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Articles(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
}

func TestArticlesParsesResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSSFeed))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	articles, err := c.Articles(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("got %d articles, want 2", len(articles))
	}

	a := articles[0]
	if a.Rank != 1 {
		t.Errorf("rank = %d, want 1", a.Rank)
	}
	if a.Title != "Building Better CSS Variables" {
		t.Errorf("title = %q", a.Title)
	}
	if a.Author != "Jane Smith" {
		t.Errorf("author = %q, want dc:creator", a.Author)
	}
	if a.Category != "CSS" {
		t.Errorf("category = %q, want CSS", a.Category)
	}
	if a.PubDate != "2024-03-15" {
		t.Errorf("pub_date = %q, want 2024-03-15", a.PubDate)
	}
	if a.URL != "https://www.smashingmagazine.com/2024/03/building-better-css-variables/" {
		t.Errorf("url = %q", a.URL)
	}
	// Description should have HTML stripped
	for _, tag := range []string{"<p>", "</p>", "<strong>", "</strong>"} {
		if contains(a.Description, tag) {
			t.Errorf("description still contains HTML tag %q: %q", tag, a.Description)
		}
	}
}

func TestArticlesLimitRespected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSSFeed))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	articles, err := c.Articles(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 {
		t.Fatalf("got %d articles with limit=1, want 1", len(articles))
	}
}

func TestArticlesRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockRSSFeed))
	}))
	defer srv.Close()

	cfg := smashingmag.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := smashingmag.NewClient(cfg)

	start := time.Now()
	_, err := c.Articles(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestArticlesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Articles(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsRaw(s, sub))
}

func containsRaw(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
