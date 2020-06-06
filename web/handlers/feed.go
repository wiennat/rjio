package handlers

import (
	"context"
	_ "errors"
	"fmt"
	"log"
	"strconv"

	"html/template"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/wiennat/rjio/feed"
)

type FeedHandler struct {
	service *feed.Service
	session *sessions.CookieStore
	fetcher *feed.Fetcher
}

func NewFeedHandler(service *feed.Service, session *sessions.CookieStore, fetcher *feed.Fetcher) *FeedHandler {
	return &FeedHandler{service: service, session: session, fetcher: fetcher}
}

func (h *FeedHandler) FeedSourceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceID := chi.URLParam(r, "sourceID")
		sid, err := strconv.ParseInt(sourceID, 10, 64)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		source, err := h.service.GetSource(sid)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "source", source)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *FeedHandler) ListSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	flash := ctx.Value("flash")

	sources := h.service.ListSource()
	err := h.renderTemplate(w, "list.html", map[string]interface{}{
		"sources": sources,
		"message": flash,
	})

	if err != nil {
		log.Printf("\nRender Error: %v\n", err)
		return
	}
}

func (h *FeedHandler) CreateSource(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	slug := r.Form.Get("slug")
	if slug == "" {
		w.Write([]byte(fmt.Sprintf("slug is required")))
		return
	}

	name := r.Form.Get("name")
	if name == "" {
		w.Write([]byte(fmt.Sprintf("name is required")))
		return
	}

	url := r.Form.Get("url")
	if url == "" {
		w.Write([]byte(fmt.Sprintf("url is required")))
		return
	}
	source := feed.Source{
		Slug: slug,
		Name: name,
		URL:  url,
	}
	err := h.service.CreateSource(&source)
	if err != nil {
		h.SaveFlash(w, r, fmt.Sprintf("cannot add feed source, err=%v", err))
		http.Redirect(w, r, "/feeds", http.StatusSeeOther)
		return
	}

	err = h.SaveFlash(w, r, fmt.Sprintf("new feed source added, id=%s", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		log.Printf("updating feed items. source=%d, slug=%s", source.ID, source.Slug)
		h.fetcher.UpdateFeed(&source)
	}()
	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}

func (h *FeedHandler) GetUpdateSourceForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(feed.Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err := h.renderTemplate(w, "edit_source.html", map[string]interface{}{
		"source": source,
	})

	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func (h *FeedHandler) UpdateSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(feed.Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	r.ParseForm()
	slug := r.Form.Get("slug")
	if slug == "" {
		w.Write([]byte(fmt.Sprintf("slug is required")))
		return
	}

	name := r.Form.Get("name")
	if name == "" {
		w.Write([]byte(fmt.Sprintf("name is required")))
		return
	}

	url := r.Form.Get("url")
	if url == "" {
		w.Write([]byte(fmt.Sprintf("url is required")))
		return
	}

	newSource := feed.Source{
		ID:   source.ID,
		Slug: slug,
		Name: name,
		URL:  url,
	}

	err := h.service.UpdateSource(&newSource)
	if err != nil {
		log.Printf("cannot update source, err=%s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = h.SaveFlash(w, r, fmt.Sprintf("source id: %d updated", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		source, err = h.service.GetSource(source.ID)
		if err != nil {
			log.Printf("Cannot fetch source, ID=%d, err=%v", source.ID, err)
			return
		}
		err := h.fetcher.UpdateFeed(&source)
		if err != nil {
			log.Printf("Cannot update items, ID=%d, err=%v", source.ID, err)
			return
		}
	}()
	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}

func (h *FeedHandler) GetDeleteSourceForm(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	source, ok := ctx.Value("source").(feed.Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err := h.renderTemplate(w, "delete_source.html", map[string]interface{}{
		"source": source,
	})
	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func (h *FeedHandler) DeleteSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(feed.Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err := h.service.DeleteSource(source.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	_, err = h.service.DeleteItemsBySource(source.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = h.SaveFlash(w, r, fmt.Sprintf("source id: %d deleted", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}

func (h *FeedHandler) ListFeedItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(feed.Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	items, err := h.service.GetItems(source.ID, 0, 50)
	if err != nil {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err = h.renderTemplate(w, "view_feed_items.html", map[string]interface{}{
		"source": source,
		"items":  items,
	})
	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func (h *FeedHandler) RefreshFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(feed.Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	log.Printf("updating feed items. source=%d, slug=%s", source.ID, source.Slug)
	err := h.fetcher.UpdateFeed(&source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.SaveFlash(w, r, fmt.Sprintf("source id: %d refreshed", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/feeds/%d/items", source.ID), http.StatusSeeOther)
}

func (h *FeedHandler) renderTemplate(w http.ResponseWriter, tmpl string, param map[string]interface{}) error {
	templateBox, err := rice.FindBox("../../templates")
	if err != nil {
		log.Fatal(err)
	}

	// get file contents as string
	templateString, err := templateBox.String(tmpl)
	if err != nil {
		log.Fatal(err)
	}

	// parse and execute the template
	tmplMessage, err := template.New(tmpl).Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	return tmplMessage.Execute(w, param)
}

func (h *FeedHandler) SaveFlash(w http.ResponseWriter, r *http.Request, message string) error {
	session, err := h.session.Get(r, "rjiosession")
	if err != nil {
		return err
	}

	session.Values["flash"] = message
	err = session.Save(r, w)
	if err != nil {
		return err
	}
	return nil
}
