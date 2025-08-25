package main

import (
    "fmt"
    "log"
    "encoding/xml"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "sort"
    "time"
    "flag"
)

const DEFAULT_PORT int = 8080
const DEFAULT_FOLDER string = "./audio"
const DEFAULT_IMAGE string = "./image.png"

type RSS struct {
    XMLName xml.Name `xml:"rss"`
    Version string   `xml:"version,attr"`
    Itunes  string   `xml:"xmlns:itunes,attr"`
    Atom    string   `xml:"xmlns:atom,attr"`
    Channel Channel  `xml:"channel"`
}

type AtomLink struct {
    Href string `xml:"href,attr"`
    Rel  string `xml:"rel,attr"`
    Type string `xml:"type,attr"`
}

type iTunesImage struct {
    Href string `xml:"href,attr"`
}

type Channel struct {
    Title       string  `xml:"title"`
    Link        string  `xml:"link"`
    Description string  `xml:"description"`
    Language    string  `xml:"language"`
    Author      string  `xml:"itunes:author"`
    Explicit    string  `xml:"itunes:explicit"`
    Image       iTunesImage  `xml:"itunes:image"`
    SelfLink    AtomLink     `xml:"atom:link"`
    Items       []Item  `xml:"item"`
}

type Item struct {
    Title       string `xml:"title"`
    Description string `xml:"description"`
    PubDate     string `xml:"pubDate"`
    GUID        string `xml:"guid"`
    Enclosure   Enclosure `xml:"enclosure"`
}

type Enclosure struct {
    URL    string `xml:"url,attr"`
    Type   string `xml:"type,attr"`
    Length string `xml:"length,attr"`
}

func rssHandler(folder string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        files, err := os.ReadDir(folder)
        if err != nil {
            http.Error(w, "Unable to read folder", http.StatusInternalServerError)
            return
        }

        // Sort newest first
        sort.Slice(files, func(i, j int) bool {
            fi, _ := files[i].Info()
            fj, _ := files[j].Info()
            return fi.ModTime().After(fj.ModTime())
        })

        items := []Item{}
        baseURL := fmt.Sprintf("https://%s/", r.Host)

        for _, f := range files {
            if f.IsDir() || filepath.Ext(f.Name()) != ".mp3" {
                continue
            }

            info, _ := f.Info()
            pub := info.ModTime().Format(time.RFC1123Z)
            size := info.Size()
            guid := fmt.Sprintf("%s-%d", f.Name(), size)

            encodedName := url.PathEscape(f.Name()) // Overcast-safe URL

            items = append(items, Item{
                Title:       f.Name(),
                Description: f.Name(),
                PubDate:     pub,
                GUID:        guid,
                Enclosure: Enclosure{
                    URL:    baseURL + encodedName,
                    Type:   "audio/mpeg",
                    Length: fmt.Sprintf("%d", size),
                },
            })
        }

        rss := RSS{
            Version: "2.0",
            Itunes:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
            Atom:    "http://www.w3.org/2005/Atom",
            Channel: Channel{
                Title:       "AudioDrive",
                Link:        baseURL,
                Description: "A folder full of mp3s",
                Language:    "en-us",
                Author:      "AudioDrive Folder",
                Explicit:    "false",
                Image:       iTunesImage{Href: baseURL + "image.png"},
                SelfLink:    AtomLink{
                Href: "https://" + baseURL + "/rss.xml",
                Rel:  "self",
                Type: "application/rss+xml",
            },
                Items:       items,
            },
        }

        w.Header().Set("Content-Type", "application/rss+xml")
        output, _ := xml.MarshalIndent(rss, "", "  ")
        w.Write([]byte(xml.Header + string(output)))
    }
}

func imageHandler(imagePath string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, imagePath)
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