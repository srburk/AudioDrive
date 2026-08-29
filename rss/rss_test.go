package rss

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRSSUsesBaseURLForEnclosures(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "show.mp3"), []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}

	base := "https://host.example/tokentokentokentokentokentoken/"
	feed, err := GenerateRSS(dir, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Channel.Items) != 1 {
		t.Fatalf("items: %d", len(feed.Channel.Items))
	}

	enc := feed.Channel.Items[0].Enclosure.URL
	if enc != base+"show.mp3" {
		t.Errorf("enclosure %s", enc)
	}
	if feed.Channel.Image.Href != base+"image.png" {
		t.Errorf("image %s", feed.Channel.Image.Href)
	}
	if feed.Channel.SelfLink.Href != base+"rss.xml" {
		t.Errorf("self %s", feed.Channel.SelfLink.Href)
	}
	if strings.Contains(feed.Channel.SelfLink.Href, "https://https://") {
		t.Errorf("double scheme in self link: %s", feed.Channel.SelfLink.Href)
	}
}

func TestGenerateRSSSkipsNonAudio(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("nope"), 0644); err != nil {
		t.Fatal(err)
	}
	feed, err := GenerateRSS(dir, "https://host.example/t/")
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Channel.Items) != 0 {
		t.Fatalf("items: %d", len(feed.Channel.Items))
	}
}
