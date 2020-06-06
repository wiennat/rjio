package main

import (
	"flag"
	"log"
	"os"

	"github.com/wiennat/rjio/feed"
	"github.com/wiennat/rjio/web"
	yaml "gopkg.in/yaml.v2"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/core"
	"xorm.io/xorm"
)

// VERSION represents rjio version
const VERSION string = "0.1.0"

func main() {
	log.Printf("Starting rjio %s", VERSION)

	// parse arguments
	helpPtr := flag.Bool("h", false, "Display help")
	configPtr := flag.String("c", "config.yml", "Config file path.")
	portPtr := flag.String("p", "3000", "Port")

	flag.Parse()
	if *helpPtr {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *configPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	// read configuration
	f, err := os.Open(*configPtr)
	if err != nil {
		log.Fatalf("Cannot read config file, error=%v", err)
		os.Exit(2)
	}

	var cfg feed.Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalf("Cannot marshal config file, error=%v", err)
		os.Exit(2)
	}
	// start program

	log.Printf("driver = %s, filename = %s", cfg.Database.Driver, cfg.Database.Filename)
	engine, err := xorm.NewEngine(cfg.Database.Driver, cfg.Database.Filename)
	if err != nil {
		log.Fatalf("Cannot initialize database connection, error=%v", err)
		os.Exit(2)
	}
	engine.SetMapper(core.GonicMapper{})

	log.Printf("port = %s", *portPtr)

	fs, err := feed.NewService(engine)

	fetcher := feed.NewFetcher(&cfg, fs)
	fetcher.Start()

	server := web.NewServer()
	server.Engine = engine
	server.Config = &cfg
	server.Port = *portPtr
	server.Fetcher = fetcher
	server.Routes()
	server.SetupSession("rjio-session")
	server.Start()

}
