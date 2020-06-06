package web

import (
	"log"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
	"xorm.io/xorm"

	"github.com/wiennat/rjio/feed"
	"github.com/wiennat/rjio/web/handlers"
)

const (
	SESSION_NAME = "rjiosession"
)

// Config stores all configuration
type Config struct {
	Channel  feed.ChannelConfig  `yaml:"channel"`
	Database feed.DatabaseConfig `yaml:"database"`
	Server   feed.ServerConfig   `yaml:"server"`
	Fetcher  feed.FetcherConfig  `yaml:"fetcher"`
}

type ServerConfig struct {
	SessionKey string `yaml:"session-key"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

type Server struct {
	Engine  *xorm.Engine
	Config  *feed.Config
	Fetcher *feed.Fetcher
	Router  chi.Router
	Session *sessions.CookieStore
	Port    string
}

func NewServer() *Server {
	server := &Server{}
	return server
}

func (s *Server) SetupSession(sessionKey string) {
	s.Session = sessions.NewCookieStore([]byte(sessionKey))
}

func (s *Server) Routes() {
	r := chi.NewRouter()
	r.Use()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(6, "gzip"))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(s.FlashMiddleware)
	r.Get("/healthz", s.HealthCheck)

	fs, err := feed.NewService(s.Engine)
	if err != nil {
		log.Fatalf("cannot create service, err=%v", err)
	}
	ch := handlers.NewChannelHandler(fs, &s.Config.Channel)

	r.Get("/rss", ch.GetChannelCustomFeed)
	r.Head("/rss", ch.GetChannelCustomFeed)

	r.Route("/feeds", func(r chi.Router) {
		fh := handlers.NewFeedHandler(fs, s.Session, s.Fetcher)
		r.Get("/", fh.ListSource)
		r.Post("/", fh.CreateSource)
		r.Route("/{sourceID}", func(r chi.Router) {
			r.Use(fh.FeedSourceCtx)
			r.Get("/items", fh.ListFeedItems)
			r.Get("/edit", fh.GetUpdateSourceForm)
			r.Post("/edit", fh.UpdateSource)
			r.Delete("/", fh.DeleteSource)
			r.Get("/delete", fh.GetDeleteSourceForm)
			r.Post("/delete", fh.DeleteSource)
			r.Post("/refresh", fh.RefreshFeed)
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

	s.Router = r
}

func (s *Server) Start() {
	log.Printf("Listening on port: %s", s.Port)
	http.ListenAndServe(":"+s.Port, s.Router)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
