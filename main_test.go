package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type feed struct {
	Channel struct {
		Image struct {
			Href string `xml:"href,attr"`
		} `xml:"image"`
		Items []struct {
			Enclosure struct {
				URL string `xml:"url,attr"`
			} `xml:"enclosure"`
		} `xml:"item"`
	} `xml:"channel"`
}

func setup(t *testing.T) (http.Handler, string) {
	t.Helper()
	dir := t.TempDir()
	audio := filepath.Join(dir, "audio")
	if err := os.Mkdir(audio, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(audio, "episode.mp3"), []byte("fake-mp3"), 0644); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(dir, "image.png")
	if err := os.WriteFile(image, []byte("png"), 0644); err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("a", TOKEN_LENGTH)
	return newHandler(audio, image, token), token
}

func TestMissingTokenIsDenied(t *testing.T) {
	h, _ := setup(t)

	for _, path := range []string{"/rss.xml", "/image.png", "/episode.mp3", "/"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s: got %d, want %d", path, rec.Code, http.StatusForbidden)
		}
	}
}

func TestWrongTokenIsDenied(t *testing.T) {
	h, token := setup(t)
	wrong := "/" + strings.Repeat("b", len(token))

	for _, path := range []string{wrong + "/rss.xml", wrong + "/image.png", wrong + "/episode.mp3"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s: got %d, want %d", path, rec.Code, http.StatusForbidden)
		}
	}
}

func TestCorrectTokenServesFeed(t *testing.T) {
	h, token := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/"+token+"/rss.xml", nil)
	req.Host = "podcast.example"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(rec.Header().Get("Content-Type"), "rss") {
		t.Errorf("content-type %q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(body, "episode.mp3") {
		t.Fatalf("feed missing episode: %s", body)
	}
	wantEnclosure := "https://podcast.example/" + token + "/episode.mp3"
	if !strings.Contains(body, wantEnclosure) {
		t.Errorf("enclosure URL missing token; want %s in:\n%s", wantEnclosure, body)
	}
	wantImage := "https://podcast.example/" + token + "/image.png"
	if !strings.Contains(body, wantImage) {
		t.Errorf("image URL missing token; want %s", wantImage)
	}
}

func TestEnclosuresNotFetchableWithoutToken(t *testing.T) {
	h, token := setup(t)

	req := httptest.NewRequest(http.MethodGet, "/"+token+"/rss.xml", nil)
	req.Host = "podcast.example"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("feed status %d", rec.Code)
	}

	var f feed
	if err := xml.Unmarshal(rec.Body.Bytes(), &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Channel.Items) != 1 {
		t.Fatalf("items: %d", len(f.Channel.Items))
	}
	enc := f.Channel.Items[0].Enclosure.URL
	if !strings.Contains(enc, "/"+token+"/") {
		t.Fatalf("enclosure missing token: %s", enc)
	}

	bare := httptest.NewRequest(http.MethodGet, "/episode.mp3", nil)
	bareRec := httptest.NewRecorder()
	h.ServeHTTP(bareRec, bare)
	if bareRec.Code != http.StatusForbidden {
		t.Errorf("bare mp3: got %d, want %d", bareRec.Code, http.StatusForbidden)
	}

	withToken := httptest.NewRequest(http.MethodGet, "/"+token+"/episode.mp3", nil)
	okRec := httptest.NewRecorder()
	h.ServeHTTP(okRec, withToken)
	if okRec.Code != http.StatusOK {
		t.Fatalf("token mp3: got %d", okRec.Code)
	}
	body, _ := io.ReadAll(okRec.Body)
	if string(body) != "fake-mp3" {
		t.Errorf("body %q", body)
	}
}

func TestDirectoryListingNotPublic(t *testing.T) {
	h, token := setup(t)

	req := httptest.NewRequest(http.MethodGet, "/"+token+"/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("token root listed files: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "episode.mp3") {
		t.Errorf("directory listing leaked filenames")
	}
}

func TestLoadOrCreateTokenPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token")

	first, err := loadOrCreateToken(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != TOKEN_LENGTH {
		t.Fatalf("len %d, want %d", len(first), TOKEN_LENGTH)
	}

	second, err := loadOrCreateToken(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("token changed across loads: %s vs %s", first, second)
	}

	rotated := strings.Repeat("c", TOKEN_LENGTH)
	got, err := loadOrCreateToken(path, rotated)
	if err != nil {
		t.Fatal(err)
	}
	if got != rotated {
		t.Fatalf("rotate: got %s", got)
	}
	reloaded, err := loadOrCreateToken(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if reloaded != rotated {
		t.Fatalf("persist after rotate: %s", reloaded)
	}
}
