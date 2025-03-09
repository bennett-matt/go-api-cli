package templates

import (
	"embed"
)

//go:embed structure/*
var EmbeddedTemplates embed.FS
