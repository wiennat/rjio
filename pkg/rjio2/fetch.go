package rjio2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/antchfx/xmlquery"

	"gopkg.in/yaml.v2"
)

var netClient = &http.Client{
	Timeout: time.Second * 30,
}

var defaultLocation = time.FixedZone("GMT", 0)
var defaultDate = time.Date(2001, time.January, 13, 18, 12, 23, 0, defaultLocation)

type FetchOption struct {
	SourcePath     string
	TemplatePath   string
	OutputPath     string
	TrackingPrefix string
}

type FeedSourceConfig struct {
	Sources []FeedSourceConfigItem `yaml:"sources"`
}

type FeedSourceConfigItem struct {
	Href string `yaml:"href" json:"url"`
	Slug string `yaml:"slug" json:"slug"`
}

func Execute(option *FetchOption) { // config string, templatePath string, outPath string) {
	storage := NewFileStorage(".")
	// read
	yamlFile, err := ioutil.ReadFile(option.SourcePath)
	if err != nil {
		log.Error().Msgf("yamlFile.Get err #%v ", err)
	}

	var c FeedSourceConfig

	if strings.HasSuffix(option.SourcePath, ".yaml") || strings.HasSuffix(option.SourcePath, ".yml") {
		err = yaml.Unmarshal(fileContent, &c)
		if err != nil {
			log.Error().Msgf("Unmarshal: %v", err)
		}
	}
	
	if strings.HasSuffix(option.SourcePath, ".json") {
		var arr []FeedSourceConfigItem
		err = json.Unmarshal(fileContent, &arr)
		if err != nil {
			log.Error().Msgf("Unmarshal: %v", err)
		}
		c.Sources = arr
	}
		

	allitems := make([]Item, 0)
	for _, v := range c.Sources {
		items, err := doFetch(v.Href)
		if err != nil {
			log.Error().AnErr("error", err)
		} else {
			allitems = append(allitems, *items...)
		}
	}
	sort.SliceStable(allitems, func(i, j int) bool {
		return allitems[i].PubDate.After(allitems[j].PubDate)
	})

	log.Debug().Msg("Finish fetching loop")
	log.Debug().Msg("------")

	// render
	rss, err := RenderRss(allitems, &Config{
		TrackingPrefix: option.TrackingPrefix,
		TemplatePath:   option.TemplatePath,
	})

	if err != nil {
		log.Fatal().AnErr("error", err)
	}
	storage.StoreRSS(option.OutputPath, rss)
}

func doFetch(sourceURL string) (*[]Item, error) {
	log.Info().Str("source", sourceURL).Msg("Updating feed")
	response, err := netClient.Get(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("Error during fetching for %s, err=%v", sourceURL, err)
	}
	defer response.Body.Close()
	log.Info().Str("source", sourceURL).Msg("Reading fetched rss")
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error during reading body for %s, err=%v", sourceURL, err)
	}

	log.Debug().Str("source", sourceURL).Msg("Parsing rss")
	doc, err := xmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("Error during parsing feed for %s, err=%v", sourceURL, err)
	}
	log.Debug().Str("source", sourceURL).Msg("Acquiring item list")

	list, err := xmlquery.QueryAll(doc, "//item")
	if err != nil {
		return nil, fmt.Errorf("Error during querying feed items for %s, err=%v", sourceURL, err)
	}

	items := make([]Item, 0)
	for _, it := range list {
		item, err := parseItem(it)
		if err != nil {
			log.Error().Msgf("error, it=%s, err=%v\n", it, err)
			continue
		}
		log.Debug().Str("title", item.Title).Msg("process item done")
		items = append(items, *item)
	}
	return &items, nil
}

func parseItem(it *xmlquery.Node) (*Item, error) {
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
				log.Debug().Str("enclosureUrl", enclosureUrl).Msg("extract enclosure url")
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
		FeedID:       0,
		PubDate:      pubDateTime,
		Title:        title,
		Description:  description,
		Raw:          raw,
		Entry:        entry,
		EnclosureUrl: enclosureUrl,
	}
	return &item, nil
}
