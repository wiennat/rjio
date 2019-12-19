package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	text "text/template"

	"github.com/GeertJohan/go.rice"
	"github.com/wiennat/rjio/feed"
)

type ChannelHandler struct {
	feedService *feed.Service
	channelCfg  *feed.ChannelConfig
}

func NewChannelHandler(fs *feed.Service, channelCfg *feed.ChannelConfig) *ChannelHandler {
	return &ChannelHandler{
		feedService: fs,
		channelCfg:  channelCfg,
	}
}

func (h *ChannelHandler) GetChannelConfig(w http.ResponseWriter, r *http.Request)    {}
func (h *ChannelHandler) UpdateChannelConfig(w http.ResponseWriter, r *http.Request) {}
func (h *ChannelHandler) GetChannelCustomFeed(w http.ResponseWriter, r *http.Request) {

	d, err := h.feedService.GetChannelItems(0, 9999)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	txt, err := h.renderText("rss_raw.xml", map[string]interface{}{
		"Entries": d,
		"Config":  h.channelCfg,
	})

	if err != nil {
		log.Printf("\nRender Error: %v\n", err)
		return
	}
	w.Header().Set("content-type", "application/xml; charset=utf-8")
	w.Header().Set("content-length", fmt.Sprintf("%d", len(txt)))
	w.Write([]byte(txt))
}

func (h *ChannelHandler) HeadChannelCustomFeed(w http.ResponseWriter, r *http.Request) {

	d, err := h.feedService.GetChannelItems(0, 9999)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	txt, err := h.renderText("rss_raw.xml", map[string]interface{}{
		"Entries": d,
		"Config":  h.channelCfg,
	})

	if err != nil {
		log.Printf("\nRender Error: %v\n", err)
		return
	}
	w.Header().Set("content-type", "application/xml; charset=utf-8")
	w.Header().Set("content-length", fmt.Sprintf("%d", len(txt)))
	w.Write([]byte(""))
}

func (h *ChannelHandler) renderText(tmpl string, param map[string]interface{}) (string, error) {
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
	tmplMessage, err := text.New(tmpl).Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	var tpl bytes.Buffer
	if err := tmplMessage.Execute(&tpl, param); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
