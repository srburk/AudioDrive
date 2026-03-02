package feed

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
)

// Channel holds the feed-level metadata.
type Channel struct {
	Title       string
	BaseURL     string // e.g. "https://audiodrive.example.com"
	Description string
	Language    string
}

// mimeTypes maps audio file extensions to MIME types.
var mimeTypes = map[string]string{
	".mp3":  "audio/mpeg",
	".m4a":  "audio/mp4",
	".mp4":  "audio/mp4",
	".ogg":  "audio/ogg",
	".wav":  "audio/wav",
	".aac":  "audio/aac",
	".flac": "audio/flac",
}

func audioMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if m, ok := mimeTypes[ext]; ok {
		return m
	}
	return "audio/mpeg"
}

// xmlEsc XML-escapes s for use in element content or attribute values.
func xmlEsc(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

type rssItem struct {
	Title         string
	Link          string
	Description   string // empty means omit element
	EnclosureURL  string
	EnclosureLen  int64
	EnclosureType string
	GUID          string
	PubDate       string
}

type rssData struct {
	Channel       Channel
	LastBuildDate string // empty means omit element
	Items         []rssItem
}

const feedTmplText = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>{{xmlesc .Channel.Title}}</title>
    <link>{{xmlesc .Channel.BaseURL}}</link>
    <description>{{xmlesc .Channel.Description}}</description>
    <language>{{.Channel.Language}}</language>{{if .LastBuildDate}}
    <lastBuildDate>{{.LastBuildDate}}</lastBuildDate>{{end}}
    <generator>AudioDrive</generator>{{range .Items}}
    <item>
      <title>{{xmlesc .Title}}</title>
      <link>{{xmlesc .Link}}</link>{{if .Description}}
      <description><![CDATA[{{.Description}}]]></description>{{end}}
      <enclosure url="{{xmlesc .EnclosureURL}}" length="{{.EnclosureLen}}" type="{{.EnclosureType}}"/>
      <guid isPermaLink="false">{{.GUID}}</guid>
      <pubDate>{{.PubDate}}</pubDate>
    </item>{{end}}
  </channel>
</rss>`

var feedTmpl = template.Must(
	template.New("feed").Funcs(template.FuncMap{"xmlesc": xmlEsc}).Parse(feedTmplText),
)

// Build renders RSS 2.0 XML bytes. Items are sorted newest-first by CreatedAt.
func Build(ch Channel, items []Item) ([]byte, error) {
	sorted := make([]Item, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].URL.CreatedAt.After(sorted[j].URL.CreatedAt)
	})

	rssItems := make([]rssItem, 0, len(sorted))
	for _, item := range sorted {
		rssItems = append(rssItems, toRSSItem(ch, item))
	}

	lastBuildDate := ""
	if len(sorted) > 0 {
		lastBuildDate = sorted[0].URL.CreatedAt.UTC().Format(time.RFC1123Z)
	}

	data := rssData{
		Channel:       ch,
		LastBuildDate: lastBuildDate,
		Items:         rssItems,
	}

	var buf bytes.Buffer
	if err := feedTmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func toRSSItem(ch Channel, item Item) rssItem {
	u := item.URL

	title := u.RawURL
	if u.Title != nil && *u.Title != "" {
		title = *u.Title
	}

	encType := "audio/mpeg"
	if u.AudioPath != nil {
		encType = audioMIME(*u.AudioPath)
	}

	desc := ""
	if u.Description != nil {
		desc = *u.Description
	}

	return rssItem{
		Title:         title,
		Link:          u.RawURL,
		Description:   desc,
		EnclosureURL:  fmt.Sprintf("%s/audio/%d", ch.BaseURL, u.ID),
		EnclosureLen:  item.AudioSizeBytes,
		EnclosureType: encType,
		GUID:          fmt.Sprintf("audiodrive:%d", u.ID),
		PubDate:       u.CreatedAt.UTC().Format(time.RFC1123Z),
	}
}
