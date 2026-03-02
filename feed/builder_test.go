package feed_test

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"audiodrive/feed"
	"audiodrive/internal/model"
)

func makeChannel() feed.Channel {
	return feed.Channel{
		Title:       "AudioDrive",
		BaseURL:     "https://host.example.com",
		Description: "Web pages converted to audio",
		Language:    "en",
	}
}

func TestBuild_EmptyItems(t *testing.T) {
	data, err := feed.Build(makeChannel(), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	var rss struct {
		XMLName xml.Name `xml:"rss"`
	}
	if err := xml.Unmarshal(data, &rss); err != nil {
		t.Fatalf("invalid XML: %v\n%s", err, data)
	}
	if strings.Contains(string(data), "<item>") {
		t.Error("expected no <item> elements")
	}
}

func TestBuild_TitleFallbackToURL(t *testing.T) {
	items := []feed.Item{
		{URL: model.URL{ID: 1, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), CreatedAt: time.Now()}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !strings.Contains(string(data), "https://a.com") {
		t.Error("expected URL as title fallback")
	}
}

func TestBuild_TitleField(t *testing.T) {
	title := "My Article"
	items := []feed.Item{
		{URL: model.URL{ID: 1, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), Title: &title, CreatedAt: time.Now()}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !strings.Contains(string(data), "My Article") {
		t.Errorf("expected title in output: %s", data)
	}
}

func TestBuild_SortedNewestFirst(t *testing.T) {
	oldTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	items := []feed.Item{
		{URL: model.URL{ID: 1, RawURL: "https://old.com", Status: model.StatusDone, AudioPath: ptr("/1.mp3"), CreatedAt: oldTime}},
		{URL: model.URL{ID: 2, RawURL: "https://new.com", Status: model.StatusDone, AudioPath: ptr("/2.mp3"), CreatedAt: newTime}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := string(data)
	oldIdx := strings.Index(s, "old.com")
	newIdx := strings.Index(s, "new.com")
	if newIdx == -1 || oldIdx == -1 {
		t.Fatalf("expected both URLs in output")
	}
	if newIdx > oldIdx {
		t.Error("expected newest item first")
	}
}

func TestBuild_EnclosureURL(t *testing.T) {
	items := []feed.Item{
		{URL: model.URL{ID: 42, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), CreatedAt: time.Now()}, AudioSizeBytes: 12345},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, `url="https://host.example.com/audio/42"`) {
		t.Errorf("expected enclosure url, got:\n%s", s)
	}
	if !strings.Contains(s, `length="12345"`) {
		t.Errorf("expected length=12345 in output:\n%s", s)
	}
	if !strings.Contains(s, `type="audio/mpeg"`) {
		t.Errorf("expected type=audio/mpeg:\n%s", s)
	}
}

func TestBuild_GUID(t *testing.T) {
	items := []feed.Item{
		{URL: model.URL{ID: 7, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), CreatedAt: time.Now()}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "audiodrive:7") {
		t.Errorf("expected guid audiodrive:7:\n%s", s)
	}
	if !strings.Contains(s, `isPermaLink="false"`) {
		t.Errorf("expected isPermaLink=false:\n%s", s)
	}
}

func TestBuild_PubDate(t *testing.T) {
	createdAt := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)
	items := []feed.Item{
		{URL: model.URL{ID: 1, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), CreatedAt: createdAt}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !strings.Contains(string(data), "Fri, 15 Mar 2024 12:00:00 +0000") {
		t.Errorf("expected pubDate in RFC1123Z format:\n%s", data)
	}
}

func TestBuild_LastBuildDate_PresentWhenItems(t *testing.T) {
	items := []feed.Item{
		{URL: model.URL{ID: 1, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), CreatedAt: time.Now()}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !strings.Contains(string(data), "<lastBuildDate>") {
		t.Error("expected lastBuildDate element")
	}
}

func TestBuild_LastBuildDate_AbsentWhenNoItems(t *testing.T) {
	data, err := feed.Build(makeChannel(), nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if strings.Contains(string(data), "<lastBuildDate>") {
		t.Error("expected no lastBuildDate when no items")
	}
}

func TestBuild_DescriptionCDATA(t *testing.T) {
	desc := "A <cool> article & more"
	items := []feed.Item{
		{URL: model.URL{ID: 1, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), Description: &desc, CreatedAt: time.Now()}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "<![CDATA[") {
		t.Errorf("expected CDATA in output:\n%s", s)
	}
	if !strings.Contains(s, "A <cool> article & more") {
		t.Errorf("expected description content verbatim:\n%s", s)
	}
}

func TestBuild_EmptyDescription_Omitted(t *testing.T) {
	empty := ""
	items := []feed.Item{
		{URL: model.URL{ID: 1, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: ptr("/a.mp3"), Description: &empty, CreatedAt: time.Now()}},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	// Only item descriptions use CDATA; channel description uses plain XML escaping.
	if strings.Contains(string(data), "<![CDATA[") {
		t.Errorf("expected no CDATA for empty description:\n%s", data)
	}
}

func TestBuild_ValidXML(t *testing.T) {
	title := "Test Article"
	desc := "Some description"
	items := []feed.Item{
		{URL: model.URL{
			ID:          1,
			RawURL:      "https://a.com",
			Status:      model.StatusDone,
			AudioPath:   ptr("/a.mp3"),
			Title:       &title,
			Description: &desc,
			CreatedAt:   time.Now(),
		}, AudioSizeBytes: 5000},
	}
	data, err := feed.Build(makeChannel(), items)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	var rss struct {
		XMLName xml.Name `xml:"rss"`
	}
	if err := xml.Unmarshal(data, &rss); err != nil {
		t.Fatalf("invalid XML: %v\n%s", err, data)
	}
}
