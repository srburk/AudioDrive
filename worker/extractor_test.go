package worker_test

import (
	"strings"
	"testing"

	"audiodrive/worker"
)

func TestExtractText_Basic(t *testing.T) {
	html := `<html><body><p>Hello world</p></body></html>`
	got := worker.ExtractText(html)
	if got != "Hello world" {
		t.Errorf("ExtractText = %q, want %q", got, "Hello world")
	}
}

func TestExtractText_StripsScript(t *testing.T) {
	html := `<html><body><p>Visible</p><script>hidden()</script></body></html>`
	got := worker.ExtractText(html)
	if strings.Contains(got, "hidden") {
		t.Errorf("ExtractText should strip script content, got: %q", got)
	}
	if !strings.Contains(got, "Visible") {
		t.Errorf("ExtractText should keep visible text, got: %q", got)
	}
}

func TestExtractText_StripsStyle(t *testing.T) {
	html := `<html><head><style>body { color: red; }</style></head><body><p>Text</p></body></html>`
	got := worker.ExtractText(html)
	if strings.Contains(got, "color") {
		t.Errorf("ExtractText should strip style content, got: %q", got)
	}
}

func TestExtractText_StripsHead(t *testing.T) {
	html := `<html><head><title>Page Title</title></head><body><p>Body</p></body></html>`
	got := worker.ExtractText(html)
	if strings.Contains(got, "Page Title") {
		t.Errorf("ExtractText should strip head content, got: %q", got)
	}
	if !strings.Contains(got, "Body") {
		t.Errorf("ExtractText should keep body text, got: %q", got)
	}
}

func TestExtractText_StripsNav(t *testing.T) {
	html := `<html><body><nav>Menu items</nav><p>Article</p></body></html>`
	got := worker.ExtractText(html)
	if strings.Contains(got, "Menu items") {
		t.Errorf("ExtractText should strip nav content, got: %q", got)
	}
}

func TestExtractText_StripsFooter(t *testing.T) {
	html := `<html><body><p>Content</p><footer>Footer text</footer></body></html>`
	got := worker.ExtractText(html)
	if strings.Contains(got, "Footer text") {
		t.Errorf("ExtractText should strip footer content, got: %q", got)
	}
}

func TestExtractText_CollapsesWhitespace(t *testing.T) {
	html := "<html><body><p>  lots   of   space  </p>\n<p>and\nnewlines</p></body></html>"
	got := worker.ExtractText(html)
	if strings.Contains(got, "  ") {
		t.Errorf("ExtractText should collapse whitespace, got: %q", got)
	}
	if got != "lots of space and newlines" {
		t.Errorf("ExtractText = %q, want %q", got, "lots of space and newlines")
	}
}

func TestExtractText_EmptyInput(t *testing.T) {
	got := worker.ExtractText("")
	if got != "" {
		t.Errorf("ExtractText of empty = %q, want empty", got)
	}
}
