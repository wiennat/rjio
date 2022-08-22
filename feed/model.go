package feed

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// CustomFeed represents template variables used for rendering RSS feed
type CustomFeed struct {
	Title       string
	Description string
	Permalink   string
	CoverURL    string
	Language    string
	Date        time.Time
	Entries     []Item
	Author      string
}

type Source struct {
	ID   int64  `json:"id" firestore:"id,omitempty"`
	URL  string `xorm:" varchar(200) not null" json:"url"  firestore:"url,omitempty"`
	Slug string `xorm:" varchar(200) not null" json:"slug"  firestore:"slug,omitempty"`
	Name string `xorm:" varchar(200) not null" json:"name"  firestore:"name,omitempty"`
}

// Item represents an item in a feed
type Item struct {
	ID           int64     `json:"id"  firestore:"id,omitempty"`
	FeedID       int64     `json:"feedId"  firestore:"feedId,omitempty"`
	GUID         string    `xorm:" varchar(200) not null" json:"guid" firestore:"guid,omitempty"`
	Title        string    `xorm:" varchar(200) null" json:"title" firestore:"title,omitempty"`
	Description  string    `json:"description" firestore:"description,omitempty"`
	PubDate      time.Time `json:"pubdate"  firestore:"pubDate,omitempty"`
	Raw          string    `json:"raw"  firestore:"raw,omitempty"`
	EnclosureUrl string    `xorm:" varchar(200) null" json:"enclosureUrl"  firestore:"enclosure,omitempty"`
	Entry        string    `json:"entry" firestore:"entry,omitempty"`
}

var storage Storage

func SetupStorage(s Storage) {
	storage = s
}

func DbGetSource(id int64) (Source, error) {
	return storage.GetSource(id)
}

func DbListSource() []Source {
	return storage.ListSource()
}

func DbCreateSource(source *Source) error {
	return storage.CreateSource(source)
}

func DbUpdateSource(source *Source) error {
	return storage.UpdateSource(source)
}

func DbDeleteSource(id int64) error {
	return storage.DeleteSource(id)
}

func DbCreateItem(item *Item) error {
	return storage.CreateItem(item)
}

func DbUpdateItem(item *Item) error {
	return storage.UpdateItem(item)
}

func DbDeleteItem(item *Item) error {
	return storage.DeleteItem(item)
}

func DbGetSourceItems(sourceID int64, offset int, limit int) ([]Item, error) {
	return storage.GetSourceItems(sourceID, offset, limit)
}

func DbUpsertSourceItem(item *Item) (int64, error) {
	return storage.UpsertSourceItem(item)
}

func DbUpsertSourceItems(sourceID int64, items []Item) error {
	return storage.UpsertSourceItems(sourceID, items)
}

func DbDeleteItemsBySource(sourceID int64) (int64, error) {
	return storage.DeleteItemsBySource(sourceID)
}

func DbGetItemsForCustomFeed(offset int, limit int) ([]Item, error) {
	return storage.GetItemsForCustomFeed(offset, limit)
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
