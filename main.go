package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "net/http/pprof"

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

	feed.SetupDb(&cfg)
	fetcher := feed.SetupFetcher(&cfg)
	fetcher.Start()
	mux := feed.SetupHandler(&cfg)

	fmt.Println("Serving content at port :" + *portPtr)
	http.ListenAndServe(":"+*portPtr, mux)
}
