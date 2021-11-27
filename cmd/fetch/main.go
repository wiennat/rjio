package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wiennat/rjio/pkg/rjio2"
)

var trackingPrefixVar string
var templateVar string
var outputPathVar string

func main() {
	var sourceVar = flag.String("s", "feed.yaml", "path to source yaml")
	var trackingPrefixVar = flag.String("p", "", "(Optional) Tracking prefix")
	var templateVar = flag.String("t", "template.xml", "path to template xml")
	var outputPathVar = flag.String("o", "output", "output path")
	debugVar := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debugVar {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	logger := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	logger.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s\t", i)
	}
	log.Logger = log.Output(logger)

	rjio2.Execute(&rjio2.FetchOption{
		SourcePath:     *sourceVar,
		TemplatePath:   *templateVar,
		OutputPath:     *outputPathVar,
		TrackingPrefix: *trackingPrefixVar,
	})
}
