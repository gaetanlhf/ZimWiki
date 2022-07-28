package handlers

import (
	"embed"
	"path/filepath"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
)

var (
	TemplateFS embed.FS
)

func embeddedFileHandler(config goview.Config, tmpl string) (string, error) {
	path := filepath.Join(config.Root, tmpl)
	bytes, err := TemplateFS.ReadFile(path + config.Extension)
	return string(bytes), err
}

func AddTemplate(WebServer *gin.Engine) *ginview.ViewEngine {
	gv := ginview.New(goview.Config{
		Root:      "templates",
		Extension: ".html",
		Master:    "layouts/master",
	})

	gv.ViewEngine.SetFileHandler(embeddedFileHandler)

	return gv
}
