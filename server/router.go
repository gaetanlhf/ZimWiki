package server

import (
	"embed"
	"path/filepath"

	"github.com/JojiiOfficial/ZimWiki/handlers"
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

func GetRoutes() {
	hd := handlers.HandlerData{
		ZimService: ZimService,
	}

	gv := ginview.New(goview.Config{
		Root:      "templates",
		Extension: ".html",
		Master:    "layouts/master",
	})

	gv.ViewEngine.SetFileHandler(embeddedFileHandler)

	WebServer.HTMLRender = gv

	WebServer.Use(func(c *gin.Context) {
		c.Set("hd", hd)
	})

	WebServer.GET("/", handlers.ShowIndexPage)

}
