package template

import (
	"embed"
)

//go:embed *
var Files embed.FS

const (
	TemplateResources = "resources.gohtml"
	TemplateSuccess   = "success.gohtml"
)
