package feed

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
)

var defaultLocation = time.FixedZone("GMT", 0)
var defaultDate = time.Date(2001, time.January, 13, 18, 12, 23, 0, defaultLocation)
var netClient = &http.Client{
	Timeout: time.Second * 30,
}

type FetcherConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
}
type Fetcher struct {
	Config *Config
}

func NewFetcher(config *Config) *Fetcher {
	return &Fetcher{Config: config}
}

func (f Fetcher) Start() {
	go func() {
		for range time.Tick(f.Config.Fetcher.Interval) {
			if cfg.Fetcher.Enabled {
				log.Println("Fetching feed")
				sources := DbListSource()
				for _, v := range sources {
					err := f.UpdateFeed(&v)
					if err != nil {
						log.Println(err)
					}
				}
				log.Println("Finish fetching loop")
			}
		}
	}()
}

func (f Fetcher) UpdateFeed(source *Source) error {
	log.Printf("Updating feed, source=%s", source)
	response, err := netClient.Get(source.URL)
	if err != nil {
		return fmt.Errorf("Error during fetching for %s, err=%v", source, err)
	}
	defer response.Body.Close()

	log.Printf("Reading fetched rss, source=%s", source)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error during reading body for %s, err=%v", source, err)
	}

	log.Printf("Parsing rss, source=%s", source)
	doc, err := xmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("Error during parsing feed for %s, err=%v", source, err)
	}

	log.Printf("Acquiring item list, source=%s", source)
	list, err := xmlquery.QueryAll(doc, "//item")
	if err != nil {
		return fmt.Errorf("Error during querying feed items for %s, err=%v", source, err)
	}

	log.Printf("Found %d items", len(list))
	var items []Item
	for i, it := range list {
		log.Printf("Parsing #%d item", i)
		item, err := f.parseItem(it, source)
		if err != nil {
			log.Printf("error, source=%s, it=%s, err=%v\n", source, it, err)
			continue
		}
		items = append(items, *item)
	}
	err = DbUpsertSourceItems(source.ID, items)
	if err != nil {
		return err
	}
	return nil
}

func (f Fetcher) parseItem(it *xmlquery.Node, source *Source) (*Item, error) {
	guidNode := it.SelectElement("guid")
	if guidNode == nil {
		return nil, fmt.Errorf("cannot parse guid")
	}

	guid := guidNode.InnerText()

	pubdateNode := it.SelectElement("pubDate")
	if pubdateNode == nil {
		return nil, fmt.Errorf("cannot find pubDate")
	}

	pubDate := pubdateNode.InnerText()
	pubDateTime, err := time.Parse(time.RFC1123, pubDate)
	if err != nil {
		pubDateTime, err = time.Parse(time.RFC1123Z, pubDate)
		if err != nil {
			// assign default pubdate
			pubDateTime = time.Date(2001, time.January, 13, 18, 12, 23, 0, defaultLocation)
		}
	}

	var title string
	if n := it.SelectElement("title"); n != nil {
		title = n.InnerText()
	}
	var description string
	if n := it.SelectElement("description"); n != nil {
		description = n.InnerText()
	}

	var enclosureUrl string
	if n := it.SelectElement("enclosure"); n != nil {
		for _, attr := range n.Attr {
			if attr.Name.Local == "url" {
				enclosureUrl = attr.Value
				fmt.Printf("enclosure URL: %s\n", enclosureUrl)
			}
		}
	}
	raw := it.OutputXML(true)

	// remove blacklist node
	if n := it.SelectElement("itunes:season"); n != nil {
		prev := n.PrevSibling
		next := n.NextSibling
		if prev != nil {
			prev.NextSibling = next
		}

		if next != nil {
			next.PrevSibling = prev
		}
	}

	entry := it.OutputXML(true)
	item := Item{
		GUID:         guid,
		FeedID:       source.ID,
		PubDate:      pubDateTime,
		Title:        title,
		Description:  description,
		Raw:          raw,
		Entry:        entry,
		EnclosureUrl: enclosureUrl,
	}
	return &item, nil
}
