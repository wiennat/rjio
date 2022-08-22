package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "net/http/pprof"

	_ "github.com/joho/godotenv/autoload"

	"github.com/wiennat/rjio/feed"
	yaml "gopkg.in/yaml.v2"
)

// VERSION represents rjio version
const VERSION string = "0.1.0"

func main() {
	fmt.Println("Starting rjio " + VERSION)

	// parse arguments
	helpPtr := flag.Bool("h", false, "Display help")
	configPtr := flag.String("c", "config.yml", "Config file path.")
	portPtr := flag.String("p", "3000", "Port")
	runFetcherPtr := flag.Bool("f", true, "Enable fetcher.")
	runServerPtr := flag.Bool("s", true, "Enable server.")

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
	if cfg.Database.Driver == "firebase" {
		s, err := feed.SetupFirebaseStorage()
		feed.SetupStorage(s)
		if err != nil {
			log.Fatalf("cannot initialize firebase storage. error=%v", err)
		}
	} else if cfg.Database.Driver == "sqlite" {
		s := feed.SetupSqlStorage(&cfg)
		feed.SetupStorage(s)
		if err != nil {
			log.Fatalf("cannot initialize sql storage. error=%v", err)
		}
	}

	if *runFetcherPtr {
		fetcher := feed.SetupFetcher(&cfg)
		fetcher.Start()
	}

	if *runServerPtr {
		mux := feed.SetupHandler(&cfg)
		fmt.Println("Serving content at port :" + *portPtr)
		http.ListenAndServe(":"+*portPtr, mux)
	}
}
