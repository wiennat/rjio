package feed

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

type Service struct {
	Engine *xorm.Engine
}

func NewService(Engine *xorm.Engine) (*Service, error) {
	err := Engine.Sync2(new(Source))
	if err != nil {
		return nil, err
	}

	err = Engine.Sync2(new(Item))
	if err != nil {
		return nil, err
	}
	return &Service{Engine: Engine}, nil
}

func (s *Service) GetSource(id int64) (Source, error) {
	var feed = Source{ID: id}
	_, err := s.Engine.Get(&feed)

	return feed, err
}

func (s *Service) ListSource() []Source {
	var sources []Source
	err := s.Engine.Find(&sources)
	if err != nil {
		log.Fatalf("error getting data, %s", err)
	}

	return sources
}

func (s *Service) CreateSource(source *Source) error {
	_, err := s.Engine.Insert(source)
	return err
}

func (s *Service) UpdateSource(source *Source) error {
	_, err := s.Engine.Id(source.ID).Update(source)
	return err
}

func (s *Service) DeleteSource(id int64) error {
	_, err := s.Engine.Id(id).Delete(&Source{})
	return err
}

func (s *Service) CreateItem(item *Item) error {
	_, err := s.Engine.Insert(item)
	return err
}

func (s *Service) UpdateItem(item *Item) error {
	_, err := s.Engine.Id(item.ID).Update(item)
	return err
}

func (s *Service) DeleteItem(item *Item) error {
	_, err := s.Engine.Id(item.ID).Delete(item)
	return err
}

func (s *Service) GetItems(sourceID int64, offset int, limit int) ([]Item, error) {
	var items []Item
	err := s.Engine.Where("feed_id = ?", sourceID).Limit(limit, offset).Find(&items)
	return items, err
}

func (s *Service) UpsertItem(item *Item) (int64, error) {
	// find by feed id and guid
	old := Item{GUID: item.GUID, FeedID: item.FeedID}
	found, err := s.Engine.Get(&old)
	if err != nil {
		log.Fatalf("error finding feed item, %s", err)
	}
	if found {
		// update
		return s.Engine.Id(old.ID).Update(item)
	}

	return s.Engine.Insert(item)
}

func (s *Service) DeleteItemsBySource(sourceID int64) (int64, error) {
	return s.Engine.Where("feed_id = ?", sourceID).Delete(&Item{})
}

func (s *Service) GetChannelItems(offset int, limit int) ([]Item, error) {
	// find by feed id and guid
	var items []Item
	err := s.Engine.OrderBy("pub_date DESC").Limit(limit, offset).Find(&items)

	return items, err
}
