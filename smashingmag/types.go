package smashingmag

import (
	"regexp"
	"strings"
	"time"
)

// Article is the record emitted for the articles command.
type Article struct {
	Rank        int    `json:"rank"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Category    string `json:"category"`
	Description string `json:"description"`
	PubDate     string `json:"pub_date"`
	URL         string `json:"url"`
}

// ─── RSS wire types ───────────────────────────────────────────────────────────

type rssRoot struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Creator     string `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Category    string `xml:"category"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

var htmlTagRE = regexp.MustCompile(`<[^>]+>`)

func stripHTML(s string) string {
	s = htmlTagRE.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.TrimSpace(s)
	return s
}

func truncate120(s string) string {
	rs := []rune(s)
	if len(rs) <= 120 {
		return s
	}
	return string(rs[:119]) + "…"
}

var pubDateFormats = []string{
	time.RFC1123Z,
	time.RFC1123,
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 MST",
}

func normDate(s string) string {
	for _, layout := range pubDateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return s
}

func wireToArticle(item rssItem, rank int) Article {
	desc := stripHTML(item.Description)
	desc = truncate120(desc)
	return Article{
		Rank:        rank,
		Title:       strings.TrimSpace(item.Title),
		Author:      strings.TrimSpace(item.Creator),
		Category:    strings.TrimSpace(item.Category),
		Description: desc,
		PubDate:     normDate(strings.TrimSpace(item.PubDate)),
		URL:         strings.TrimSpace(item.Link),
	}
}
