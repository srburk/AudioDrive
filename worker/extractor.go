package worker

import (
	"strings"

	"golang.org/x/net/html"
)

// noiseTags are elements whose entire subtree should be skipped —
// they contain chrome, metadata, or non-prose content.
var noiseTags = map[string]bool{
	// code & metadata
	"script": true, "style": true, "head": true, "noscript": true,
	// site chrome
	"nav": true, "header": true, "footer": true, "aside": true, "menu": true,
	// forms & interactive controls
	"form": true, "button": true, "input": true, "select": true,
	"textarea": true, "label": true,
	// embedded / visual media
	"iframe": true, "embed": true, "object": true, "svg": true,
	"canvas": true, "figure": true,
}

// noiseRoles are ARIA landmark roles that identify non-article regions.
var noiseRoles = map[string]bool{
	"navigation": true, "banner": true, "complementary": true,
	"search": true, "form": true,
}

// ExtractText parses htmlBody and returns the readable prose content.
// It prefers an <article> or <main> element when present, falling back to
// the full document. Noise tags and aria-hidden subtrees are skipped.
func ExtractText(htmlBody string) string {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return ""
	}

	root := findContentRoot(doc)
	if root == nil {
		root = doc
	}

	var sb strings.Builder
	walkText(root, &sb)
	return strings.Join(strings.Fields(sb.String()), " ")
}

// findContentRoot returns the first <article> in the tree, then <main>,
// or nil if neither exists.
func findContentRoot(n *html.Node) *html.Node {
	if n.Type == html.ElementNode {
		if n.Data == "article" {
			return n
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findContentRoot(c); found != nil {
			return found
		}
	}
	// Second pass for <main> — only reached if no <article> found
	return findMain(n)
}

func findMain(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "main" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findMain(c); found != nil {
			return found
		}
	}
	return nil
}

func walkText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.ElementNode {
		if noiseTags[n.Data] {
			return
		}
		if isNoiseByAttr(n) {
			return
		}
	}
	if n.Type == html.TextNode {
		t := strings.TrimSpace(n.Data)
		if t != "" {
			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(t)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkText(c, sb)
	}
}

func isNoiseByAttr(n *html.Node) bool {
	for _, a := range n.Attr {
		switch a.Key {
		case "aria-hidden":
			if a.Val == "true" {
				return true
			}
		case "role":
			if noiseRoles[a.Val] {
				return true
			}
		}
	}
	return false
}
