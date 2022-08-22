package feed

import (
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

type Storage interface {
	GetSource(id int64) (Source, error)
	ListSource() []Source
	CreateSource(source *Source) error
	UpdateSource(source *Source) error
	DeleteSource(id int64) error
	CreateItem(item *Item) error
	UpdateItem(item *Item) error
	DeleteItem(item *Item) error
	GetSourceItems(sourceID int64, offset int, limit int) ([]Item, error)
	UpsertSourceItem(item *Item) (int64, error)
	UpsertSourceItems(sourceID int64, items []Item) error
	DeleteItemsBySource(sourceID int64) (int64, error)
	GetItemsForCustomFeed(offset int, limit int) ([]Item, error)
}
