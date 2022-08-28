package templates

import "embed"

//go:embed *.xml *.html
var TemplateBox embed.FS
