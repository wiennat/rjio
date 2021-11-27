package rjio2

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
)

// Item represents an item in a feed
type Item struct {
	ID           int64     `json:"id"`
	FeedID       int64     `json:"feedId"`
	GUID         string    `json:"guid"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	PubDate      time.Time `json:"pubdate"`
	Raw          string    `json:"raw"`
	EnclosureUrl string    `json:"enclosureUrl"`
	Entry        string    `json:"entry"`
}

type Feed struct {
	Name  string
	URL   string
	Items []Item
}

type Config struct {
	TrackingPrefix string
	TemplatePath   string
}

func RenderRss(items []Item, config *Config) (string, error) {
	// add prefix enclosure

	d, err := ApplyEnclosurePrefix(items, config.TrackingPrefix)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	var w strings.Builder
	err = renderText(&w, config.TemplatePath, map[string]interface{}{
		"Entries": d,
	})

	if err != nil {
		log.Error().Msgf("\nRender Error: %v\n", err)
		return "", err
	}
	return w.String(), nil
}

// apply prefix to enclosure url
func ApplyEnclosurePrefix(items []Item, prefix string) ([]Item, error) {
	// find enclosure url in Entry and replace with prefix one
	if prefix == "" {
		return items, nil
	}

	for i, item := range items {
		if strings.Contains(item.EnclosureUrl, "https://anchor.fm/") {
			// handle anchor.fm url
			enclosureURLAttr := fmt.Sprintf("url=\"%s\"", item.EnclosureUrl)
			colonPosition := strings.Index(item.EnclosureUrl, "://")
			protocol := item.EnclosureUrl[0:colonPosition]
			escapedURL := item.EnclosureUrl[(colonPosition + 3):]
			unescaped, err := url.PathUnescape(escapedURL)
			if err != nil {
				return nil, err
			}
			newEnclosureURLAttr := fmt.Sprintf("url=\"%s://%s%s\"", protocol, prefix, unescaped)
			items[i].Entry = strings.Replace(item.Entry, enclosureURLAttr, newEnclosureURLAttr, 1)
		} else {

			enclosureURLAttr := fmt.Sprintf("url=\"%s\"", item.EnclosureUrl)
			colonPosition := strings.Index(item.EnclosureUrl, "://")
			protocol := item.EnclosureUrl[0:colonPosition]
			url := item.EnclosureUrl[(colonPosition + 3):]
			newEnclosureURLAttr := fmt.Sprintf("url=\"%s://%s%s\"", protocol, prefix, url)
			items[i].Entry = strings.Replace(item.Entry, enclosureURLAttr, newEnclosureURLAttr, 1)
		}
	}

	return items, nil
}

func renderText(w io.Writer, tmplFile string, param map[string]interface{}) error {
	tmpl := template.Must(template.ParseFiles(tmplFile))

	err := tmpl.Execute(w, param)
	if err != nil {
		return err
	}
	return nil
}
