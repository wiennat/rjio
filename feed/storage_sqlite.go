package feed

import (
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
	"xorm.io/core"
	"xorm.io/xorm"
)

type SqlStorage struct {
	engine *xorm.Engine
	dbConf *DatabaseConfig
}

func SetupSqlStorage(config *Config) *SqlStorage {
	dbConf := &config.Database
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
	return &SqlStorage{
		engine: engine,
		dbConf: dbConf,
	}
}

func (s *SqlStorage) GetSource(id int64) (Source, error) {
	var feed = Source{ID: id}
	_, err := s.engine.Get(&feed)
	return feed, err
}

func (s *SqlStorage) ListSource() []Source {
	var sources []Source
	err := s.engine.Find(&sources)
	if err != nil {
		log.Fatalf("error getting data, %s", err)
	}
	return sources
}

func (s *SqlStorage) CreateSource(source *Source) error {
	_, err := s.engine.Insert(source)
	return err
}

func (s *SqlStorage) UpdateSource(source *Source) error {
	_, err := s.engine.Id(source.ID).Update(source)
	return err
}

func (s *SqlStorage) DeleteSource(id int64) error {
	_, err := s.engine.Id(id).Delete(&Source{})
	return err
}

func (s *SqlStorage) CreateItem(item *Item) error {
	_, err := s.engine.Insert(item)
	return err
}

func (s *SqlStorage) UpdateItem(item *Item) error {
	_, err := s.engine.Id(item.ID).Update(item)
	return err
}

func (s *SqlStorage) DeleteItem(item *Item) error {
	_, err := s.engine.Id(item.ID).Delete(item)
	return err
}

func (s *SqlStorage) GetSourceItems(sourceID int64, offset int, limit int) ([]Item, error) {
	var items []Item
	err := s.engine.Where("feed_id = ?", sourceID).Limit(limit, offset).Find(&items)
	return items, err

}

func (s *SqlStorage) UpsertSourceItem(item *Item) (int64, error) {
	// find by feed id and guid
	old := Item{GUID: item.GUID, FeedID: item.FeedID}
	found, err := s.engine.Get(&old)
	if err != nil {
		log.Fatalf("error finding feed item, %s", err)
	}
	if found {
		// update
		log.Printf("update item(%d), %s, %s\n", item.ID, item.FeedID, item.EnclosureUrl)
		return s.engine.Id(old.ID).Update(item)
	}

	return s.engine.Insert(item)
}

func (s *SqlStorage) UpsertSourceItems(sourceID int64, items []Item) error {
	for _, e := range items {
		_, err := s.UpsertSourceItem(&e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SqlStorage) DeleteItemsBySource(sourceID int64) (int64, error) {
	return s.engine.Where("feed_id = ?", sourceID).Delete(&Item{})

}

func (s *SqlStorage) GetItemsForCustomFeed(offset int, limit int) ([]Item, error) {
	// find by feed id and guid
	var items []Item
	err := s.engine.OrderBy("pub_date DESC").Limit(limit, offset).Find(&items)

	return items, err
}
