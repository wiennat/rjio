package feed

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
)

type FirebaseStorage struct {
	app    *firebase.App
	client *firestore.Client
	ctx    context.Context
}

type SourceDocument struct {
	ID    int64  `json:"id" firestore:"id,omitempty"`
	URL   string `json:"url"  firestore:"url,omitempty"`
	Slug  string `json:"slug"  firestore:"slug,omitempty"`
	Name  string `json:"name"  firestore:"name,omitempty"`
	Items []Item `json:"items"  firestore:"items,omitempty"`
}

type ItemDocument struct {
	Item
	Hash string `firestore:"hash,omitempty"`
}

func SetupFirebaseStorage() (*FirebaseStorage, error) {
	ctx := context.Background()
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	return &FirebaseStorage{
		app:    app,
		client: client,
		ctx:    ctx,
	}, nil
}
func sourceDocumentToSource(sd *SourceDocument) (*Source, []Item) {
	return &Source{
		ID:   sd.ID,
		URL:  sd.URL,
		Slug: sd.Slug,
		Name: sd.Name,
	}, sd.Items
}

func sourceToSourceDocument(s *Source) *SourceDocument {
	return &SourceDocument{
		ID:    s.ID,
		URL:   s.URL,
		Slug:  s.Slug,
		Name:  s.Name,
		Items: make([]Item, 0),
	}
}

func (fb *FirebaseStorage) GetSource(id int64) (Source, error) {
	strId := strconv.FormatInt(id, 10)
	dsnap, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Get(fb.ctx)
	var src Source

	err = dsnap.DataTo(&src)
	if err != nil {
		log.Fatalf("Failed to get source: %v", err)
	}
	return src, nil
}

func (fb *FirebaseStorage) ListSource() []Source {
	var sources []Source
	iter := fb.client.Collection("feeds/samcoke/sources").Documents(fb.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		var src Source
		doc.DataTo(&src)
		sources = append(sources, src)
	}

	return sources
}

func (fb *FirebaseStorage) CreateSource(source *Source) error {
	strId := strconv.FormatInt(source.ID, 10)
	sourceDocument := sourceToSourceDocument(source)
	sourceDocument.ID = rand.Int63n(math.MaxInt64)

	_, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Set(fb.ctx, sourceDocument)
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred: %s", err)
	}
	return err
}

func (fb *FirebaseStorage) UpdateSource(source *Source) error {
	strId := strconv.FormatInt(source.ID, 10)

	_, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Update(fb.ctx, []firestore.Update{
		{Path: "id", Value: source.ID},
		{Path: "url", Value: source.URL},
		{Path: "slug", Value: source.Slug},
		{Path: "name", Value: source.Name},
	})
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred: %s", err)
	}
	return err
}

func (fb *FirebaseStorage) DeleteSource(id int64) error {
	strId := strconv.FormatInt(id, 10)

	_, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Delete(fb.ctx)
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred: %s", err)
	}

	fb.DeleteItemsBySource(id)
	return err
}

func (fb *FirebaseStorage) CreateItem(item *Item) error {
	strId := strconv.FormatInt(item.FeedID, 10)
	dsnap, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Get(fb.ctx)
	var src SourceDocument

	err = dsnap.DataTo(&src)
	if err != nil {
		log.Fatalf("Failed to get source: %v", err)
		return err
	}
	src.Items = append(src.Items, *item)

	_, err = fb.client.Collection("feeds/samcoke/sources").Doc(strId).Update(fb.ctx, []firestore.Update{
		{Path: "items", Value: src.Items},
	})
	if err != nil {
		log.Fatalf("Failed to save item after delete: %v", err)
		return err
	}
	return nil
}

func (fb *FirebaseStorage) UpdateItem(item *Item) error {
	strId := strconv.FormatInt(item.FeedID, 10)
	dsnap, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Get(fb.ctx)
	var src SourceDocument

	err = dsnap.DataTo(&src)
	if err != nil {
		log.Fatalf("Failed to get source: %v", err)
	}
	var newItem []Item
	for _, el := range src.Items {
		if el.GUID == item.GUID {
			newItem = append(newItem, *item)
		}
	}
	_, err = fb.client.Collection("feeds/samcoke/sources").Doc(strId).Update(fb.ctx, []firestore.Update{
		{Path: "items", Value: newItem},
	})
	if err != nil {
		log.Fatalf("Failed to save item after delete: %v", err)
	}
	return nil
}

func (fb *FirebaseStorage) DeleteItem(item *Item) error {
	strId := strconv.FormatInt(item.FeedID, 10)
	dsnap, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Get(fb.ctx)
	var src SourceDocument

	err = dsnap.DataTo(&src)
	if err != nil {
		log.Fatalf("Failed to get source: %v", err)
	}
	var newItem []Item
	for _, el := range src.Items {
		if el.GUID != item.GUID {
			newItem = append(newItem, el)
		}
	}
	_, err = fb.client.Collection("feeds/samcoke/sources").Doc(strId).Update(fb.ctx, []firestore.Update{
		{Path: "items", Value: newItem},
	})
	if err != nil {
		log.Fatalf("Failed to save item after delete: %v", err)
	}
	return nil
}

func (fb *FirebaseStorage) GetSourceItems(sourceID int64, offset int, limit int) ([]Item, error) {
	strId := strconv.FormatInt(sourceID, 10)
	dsnap, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Get(fb.ctx)
	var src SourceDocument

	err = dsnap.DataTo(&src)
	if err != nil {
		log.Fatalf("Failed to get source: %v", err)
	}
	return src.Items, nil
}

func (fb *FirebaseStorage) UpsertSourceItem(item *Item) (int64, error) {
	// find by feed id and guid
	err := fb.UpdateItem(item)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (fb *FirebaseStorage) UpsertSourceItems(sourceID int64, items []Item) error {
	strId := strconv.FormatInt(sourceID, 10)
	dsnap, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Get(fb.ctx)
	var src SourceDocument

	err = dsnap.DataTo(&src)
	if err != nil {
		log.Fatalf("Failed to get source: %v", err)
	}

	_, err = fb.client.Collection("feeds/samcoke/sources").Doc(strId).Update(fb.ctx, []firestore.Update{
		{Path: "items", Value: items},
	})
	if err != nil {
		log.Fatalf("Failed to update items: %v", err)
	}
	return nil
}

func (fb *FirebaseStorage) DeleteItemsBySource(sourceID int64) (int64, error) {
	strId := strconv.FormatInt(sourceID, 10)
	dsnap, err := fb.client.Collection("feeds/samcoke/sources").Doc(strId).Get(fb.ctx)
	var src SourceDocument

	err = dsnap.DataTo(&src)
	if err != nil {
		log.Fatalf("Failed to get source: %v", err)
	}
	numDeleted := len(src.Items)
	items := make([]Item, 0)
	fb.client.Collection("feeds/samcoke/sources").Doc(strId).Update(fb.ctx, []firestore.Update{
		{Path: "items", Value: items},
	})
	return int64(numDeleted), nil
}

func (fb *FirebaseStorage) GetItemsForCustomFeed(offset int, limit int) ([]Item, error) {
	// find by feed id and guid
	var items []Item

	sources := fb.ListSource()
	for _, source := range sources {
		srcItems, err := fb.GetSourceItems(source.ID, offset, limit)
		if err != nil {
			log.Fatalf("An error has occurred: %s", err)
		}
		items = append(items, srcItems...)
	}

	return items, nil
}
