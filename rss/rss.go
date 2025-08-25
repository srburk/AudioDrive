package rss

import (
    "fmt"
    "encoding/xml"
    "net/url"
    "os"
    "path/filepath"
    "sort"
    "time"
)

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

var audioExtensions = map[string]struct{}{
	".mp3": {},
	".wav": {},
	".flac": {},
	".aac": {},
	".ogg": {},
}

func isAudioFile(filename string) bool {
	ext := filepath.Ext(filename)
	_, exists := audioExtensions[ext]
	return exists
}

func GenerateRSS(folder string, baseURL string) (*RSS, error) {
    files, err := os.ReadDir(folder)
    if err != nil {
        return nil, fmt.Errorf("Unable to read folder: %w", err)
    }

    sort.Slice(files, func(i, j int) bool {
        fi, _ := files[i].Info()
        fj, _ := files[j].Info()
        return fi.ModTime().After(fj.ModTime())
    })

    items := []Item{}

    for _, f := range files {
        if f.IsDir() || !isAudioFile(f.Name()) {
            continue
        }

        info, _ := f.Info()
        pub := info.ModTime().Format(time.RFC1123Z)
        size := info.Size()
        guid := fmt.Sprintf("%s-%d", f.Name(), size)

        encodedName := url.PathEscape(f.Name())

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

    return &RSS{
        Version: "2.0",
        Itunes:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
        Atom:    "http://www.w3.org/2005/Atom",
        Channel: Channel{
            Title:       "AudioDrive",
            Link:        baseURL,
            Description: "A folder full of MP3s",
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
    }, nil
}

func (r *RSS) ToXML() ([]byte, error) {
	return xml.MarshalIndent(r, "", "  ")
}