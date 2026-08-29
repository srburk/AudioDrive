package main

import (
	"audiodrive/rss"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

const DEFAULT_PORT int = 8080
const DEFAULT_FOLDER string = "./audio"
const DEFAULT_IMAGE string = "./image.png"
const TOKEN_LENGTH int = 32
const DEFAULT_TOKEN_FILE string = "token"

func logRequest(r *http.Request) {
	log.Printf("Client IP: %s, User-Agent: %s, Method: %s, URL: %s",
		r.RemoteAddr, r.UserAgent(), r.Method, r.URL.Path)
}

func imageHandler(imagePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		http.ServeFile(w, r, imagePath)
	}
}

func rssHandler(folder, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		baseURL := fmt.Sprintf("https://%s/%s/", r.Host, token)
		feed, err := rss.GenerateRSS(folder, baseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		output, _ := feed.ToXML()
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(xml.Header + string(output)))
	}
}

func audioHandler(folder, token string) http.HandlerFunc {
	prefix := "/" + token + "/"
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		name := strings.TrimPrefix(r.URL.Path, prefix)
		if name == "" || name != path.Base(name) {
			http.NotFound(w, r)
			return
		}
		decoded, err := url.PathUnescape(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if !rss.IsAudioFile(decoded) {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(folder, decoded))
	}
}

func authorized(path, token string) bool {
	prefix := "/" + token
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func newHandler(folder, imagePath, token string) http.Handler {
	mux := http.NewServeMux()
	prefix := "/" + token
	mux.HandleFunc(prefix+"/rss.xml", rssHandler(folder, token))
	mux.HandleFunc(prefix+"/image.png", imageHandler(imagePath))
	mux.HandleFunc(prefix+"/", audioHandler(folder, token))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r.URL.Path, token) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

var port int

func main() {
	flag.IntVar(&port, "p", DEFAULT_PORT, "port")
	flag.IntVar(&port, "port", DEFAULT_PORT, "port")

	folderPtr := flag.String("folder", DEFAULT_FOLDER, "directory for files")
	imagePtr := flag.String("image", DEFAULT_IMAGE, "Path to image file for podcast clients")
	tokenFilePtr := flag.String("token-file", DEFAULT_TOKEN_FILE, "file that stores the subscribe token")
	tokenPtr := flag.String("token", "", "subscribe token (saved to token-file; generated on first start if empty)")
	flag.Parse()

	token, err := loadOrCreateToken(*tokenFilePtr, *tokenPtr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening on http://127.0.0.1:%d\n", port)
	fmt.Printf("Subscribe URL: http://127.0.0.1:%d/%s/rss.xml\n", port, token)
	fmt.Println("Treat this URL as a password. Anyone with it can fetch the feed and episodes.")

	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, newHandler(*folderPtr, *imagePtr, token)))
}
