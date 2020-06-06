package feed

import (
	"context"
	_ "errors"
	"fmt"
	"html/template"
	"log"

	"net/http"
	"net/http/pprof"
	"strconv"
	text "text/template"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
)

const (
	SESSION_NAME = "session"
)

// Config stores all configuration
type Config struct {
	Channel  ChannelConfig  `yaml:"channel"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	Fetcher  FetcherConfig  `yaml:"fetcher"`
}

type ServerConfig struct {
	SessionKey string `yaml:"session-key"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

// ChannelConfig represents program configuration
type ChannelConfig struct {
	Title          string `yaml:"title"`
	Description    string `yaml:"description"`
	Category       string `yaml:"category"`
	Link           string `yaml:"link"`
	Author         string `yaml:"author"`
	Email          string `yaml:"email"`
	Language       string `yaml:"language"`
	PermaLink      string `yaml:"permalink"`
	CoverURL       string `yaml:"cover-url"`
	Explicit       string `yaml:"explicit"`
	TrackingPrefix string `yaml:"tracking-prefix"`
}

type DatabaseConfig struct {
	Driver   string
	Filename string
}

var cfg *Config
var store *sessions.CookieStore
var fetcher *Fetcher

func SetupFetcher(c *Config) *Fetcher {
	cfg = c
	fetcher = NewFetcher(cfg)
	return fetcher
}

func SetupHandler(c *Config) *chi.Mux {
	cfg = c
	store = sessions.NewCookieStore([]byte(cfg.Server.SessionKey))

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(FlashMiddleware)
	r.Get("/", indexHandler)
	r.Get("/rss", customFeedHandler)
	r.Route("/feeds", func(r chi.Router) {
		r.Get("/", listSourcesHandler)
		r.Post("/", createSourceHandler)

		r.Route("/{sourceID}", func(r chi.Router) {
			r.Use(FeedSourceCtx)
			r.Get("/items", getFeedItemsHandler)
			r.Get("/edit", updateSourceFormHandler)
			r.Post("/edit", updateSourceHandler)
			r.Delete("/", deleteSourceHandler)
			r.Get("/delete", confirmDeleteSourceHandler)
			r.Post("/delete", deleteSourceHandler)
			r.Post("/refresh", refreshFeedItemsHandler)
		})
	})

	r.Get("/debug/pprof/", pprof.Index)
	r.Get("/debug/pprof/cmdline", pprof.Cmdline)
	r.Get("/debug/pprof/profile", pprof.Profile)
	r.Get("/debug/pprof/symbol", pprof.Symbol)

	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/debug/pprof/block", pprof.Handler("block"))
	return r
}

func FlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, SESSION_NAME)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Set some session values.
		flash, ok := session.Values["flash"]

		if ok {
			delete(session.Values, "flash")
		}

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), "flash", flash)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/rss", http.StatusSeeOther)
}

func customFeedHandler(w http.ResponseWriter, r *http.Request) {
	d, err := DbGetItemsForCustomFeed(0, 9999)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// add prefix enclosure
	d, err = ApplyEnclosurePrefix(d, cfg.Channel.TrackingPrefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = renderText(w, "rss_raw.xml", map[string]interface{}{
		"Entries": d,
		"Config":  cfg.Channel,
	})

	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func listSourcesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	flash := ctx.Value("flash")

	sources := DbListSource()
	err := renderTemplate(w, "list.html", map[string]interface{}{
		"sources": sources,
		"message": flash,
	})

	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func FeedSourceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceID := chi.URLParam(r, "sourceID")
		sid, err := strconv.ParseInt(sourceID, 10, 64)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		source, err := DbGetSource(sid)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "source", source)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type SourceRequest struct {
	Slug string
	Name string
	URL  string
}

func createSourceHandler(w http.ResponseWriter, r *http.Request) {
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
	source := Source{
		Slug: slug,
		Name: name,
		URL:  url,
	}
	DbCreateSource(&source)
	err := saveFlash(w, r, fmt.Sprintf("new feed source added, id=%s", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		log.Printf("updating feed items. source=%d, slug=%s", source.ID, source.Slug)
		fetcher.UpdateFeed(&source)
	}()
	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}

func updateSourceFormHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	source, ok := ctx.Value("source").(Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err := renderTemplate(w, "edit_source.html", map[string]interface{}{
		"source": source,
	})

	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func updateSourceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(Source)

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

	newSource := Source{
		ID:   source.ID,
		Slug: slug,
		Name: name,
		URL:  url,
	}

	err := DbUpdateSource(&newSource)
	if err != nil {
		log.Printf("cannot update source, err=%s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = saveFlash(w, r, fmt.Sprintf("source id: %d updated", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		source, err = DbGetSource(source.ID)
		if err != nil {
			log.Printf("Cannot fetch source, ID=%d, err=%v", source.ID, err)
			return
		}
		err := fetcher.UpdateFeed(&source)
		if err != nil {
			log.Printf("Cannot update items, err=%v", source.ID, err)
			return
		}
	}()
	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}

func confirmDeleteSourceHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	source, ok := ctx.Value("source").(Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err := renderTemplate(w, "delete_source.html", map[string]interface{}{
		"source": source,
	})
	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func deleteSourceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err := DbDeleteSource(source.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	_, err = DbDeleteItemsBySource(source.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	err = saveFlash(w, r, fmt.Sprintf("source id: %d deleted", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}

func getFeedItemsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	items, err := DbGetSourceItems(source.ID, 0, 50)
	if err != nil {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err = renderTemplate(w, "view_feed_items.html", map[string]interface{}{
		"source": source,
		"items":  items,
	})
	if err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
}

func refreshFeedItemsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source, ok := ctx.Value("source").(Source)

	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	log.Printf("updating feed items. source=%d, slug=%s", source.ID, source.Slug)
	err := fetcher.UpdateFeed(&source)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = saveFlash(w, r, fmt.Sprintf("source id: %d refreshed", source.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/feeds/%d/items", source.ID), http.StatusSeeOther)
}

func renderText(w http.ResponseWriter, tmpl string, param map[string]interface{}) error {
	templateBox, err := rice.FindBox("../templates")
	if err != nil {
		log.Fatal(err)
	}

	// get file contents as string
	templateString, err := templateBox.String(tmpl)
	if err != nil {
		log.Fatal(err)
	}

	// parse and execute the template
	tmplMessage, err := text.New(tmpl).Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	return tmplMessage.Execute(w, param)
}

func renderTemplate(w http.ResponseWriter, tmpl string, param map[string]interface{}) error {
	templateBox, err := rice.FindBox("../templates")
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

func saveFlash(w http.ResponseWriter, r *http.Request, message string) error {
	session, err := store.Get(r, SESSION_NAME)
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
