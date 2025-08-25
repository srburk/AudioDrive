package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "encoding/xml"
    "audiodrive/rss"
)

const DEFAULT_PORT int = 8080
const DEFAULT_FOLDER string = "./audio"
const DEFAULT_IMAGE string = "./image.png"
const TOKEN_LENGTH int = 32

func imageHandler(imagePath string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    
        log.Printf("Client IP: %s, User-Agent: %s, Method: %s, URL: %s",
        r.RemoteAddr, r.UserAgent(), r.Method, r.URL.Path)
    
        http.ServeFile(w, r, imagePath)
    }
}

func rssHandler(folder string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    
        log.Printf("Client IP: %s, User-Agent: %s, Method: %s, URL: %s",
        r.RemoteAddr, r.UserAgent(), r.Method, r.URL.Path)
    
        baseURL := fmt.Sprintf("https://%s/", r.Host)
        rss, err := rss.GenerateRSS(folder, baseURL)
        if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
        output, _ := rss.ToXML()
        w.Header().Set("Content-Type", "application/rss+xml")
        w.Write([]byte(xml.Header + string(output)))
    }
}

var port int

func main() {

    flag.IntVar(&port, "p", DEFAULT_PORT, "port")
    flag.IntVar(&port, "port", DEFAULT_PORT, "port")

    folderPtr := flag.String("folder", DEFAULT_FOLDER, "directory for files")
    imagePtr := flag.String("image", DEFAULT_IMAGE, "Path to image file for podcast clients")
    flag.Parse()
        
    http.HandleFunc("/rss.xml", rssHandler(*folderPtr))
    http.HandleFunc("/image.png", imageHandler(*imagePtr))
    http.Handle("/", http.FileServer(http.Dir(*folderPtr)))
    
    fmt.Printf("Listening on http://127.0.0.1:%d\n", port)
    addr := fmt.Sprintf(":%d", port)
    log.Fatal(http.ListenAndServe(addr, nil))
}