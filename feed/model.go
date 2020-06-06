package feed

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/core"
	"xorm.io/xorm"
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
	ID   int64
	URL  string `xorm:" varchar(200) not null"`
	Slug string `xorm:" varchar(200) not null"`
	Name string `xorm:" varchar(200) not null"`
}

// Item represents an item in a feed
type Item struct {
	ID           int64
	FeedID       int64
	GUID         string `xorm:" varchar(200) not null"`
	Title        string `xorm:" varchar(200) null"`
	Description  string
	PubDate      time.Time
	Raw          string
	EnclosureUrl string `xorm:" varchar(200) null"`
	Entry        string
}

var engine *xorm.Engine
var dbConf *DatabaseConfig

func SetupDb(config *Config) {
	dbConf = &config.Database
	engine, err := xorm.NewEngine(dbConf.Driver, dbConf.Filename)
	engine.SetMapper(core.GonicMapper{})
	if err != nil {
		log.Fatal("cannot start db")
		os.Exit(1)
	}
	err = engine.Sync2(new(Source))
	if err != nil {
		log.Fatalf("cannot sync db: %s", err)
		os.Exit(1)
	}
	err = engine.Sync2(new(Item))
	if err != nil {
		log.Fatalf("cannot sync db: %s", err)
		os.Exit(1)
	}
}

func getEngine() (*xorm.Engine, error) {
	engine, err := xorm.NewEngine(dbConf.Driver, dbConf.Filename)
	engine.SetMapper(core.GonicMapper{})
	return engine, err
}

func DbGetSource(id int64) (Source, error) {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}
	var feed = Source{ID: id}
	_, err = engine.Get(&feed)

	return feed, err
}

func DbListSource() []Source {
	var sources []Source
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	err = engine.Find(&sources)
	if err != nil {
		log.Fatalf("error getting data, %s", err)
	}

	return sources
}

func DbCreateSource(source *Source) error {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	_, err = engine.Insert(source)
	return err
}

func DbUpdateSource(source *Source) error {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	_, err = engine.Id(source.ID).Update(source)
	return err
}

func DbDeleteSource(id int64) error {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	_, err = engine.Id(id).Delete(&Source{})
	return err
}

func DbCreateItem(item *Item) error {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	_, err = engine.Insert(item)
	return err
}

func DbUpdateItem(item *Item) error {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	_, err = engine.Id(item.ID).Update(item)
	return err
}

func DbDeleteItem(item *Item) error {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	_, err = engine.Id(item.ID).Delete(item)
	return err
}

func DbGetSourceItems(sourceID int64, offset int, limit int) ([]Item, error) {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	var items []Item
	err = engine.Where("feed_id = ?", sourceID).Limit(limit, offset).Find(&items)
	return items, err
}

func DbUpsertSourceItem(item *Item) (int64, error) {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	// find by feed id and guid
	old := Item{GUID: item.GUID, FeedID: item.FeedID}
	found, err := engine.Get(&old)
	if err != nil {
		log.Fatalf("error finding feed item, %s", err)
	}
	if found {
		// update
		log.Printf("update item(%d), %s, %s\n", item.ID, item.FeedID, item.EnclosureUrl)
		return engine.Id(old.ID).Update(item)
	}

	return engine.Insert(item)
}

func DbDeleteItemsBySource(sourceID int64) (int64, error) {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	return engine.Where("feed_id = ?", sourceID).Delete(&Item{})
}

func DbGetItemsForCustomFeed(offset int, limit int) ([]Item, error) {
	engine, err := getEngine()
	if err != nil {
		log.Fatalf("error getting engine, %s", err)
	}

	// find by feed id and guid
	var items []Item
	err = engine.OrderBy("pub_date DESC").Limit(limit, offset).Find(&items)

	return items, err
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
