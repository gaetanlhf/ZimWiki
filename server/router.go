package server

import (
	"embed"
	"net/http"
	"path/filepath"
	"time"

	"github.com/JojiiOfficial/ZimWiki/config"
	"github.com/JojiiOfficial/ZimWiki/handlers"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	TemplateFS embed.FS
	StaticFS   embed.FS
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

	memorystore := persistence.NewInMemoryStore(time.Duration(config.Config.SearchCacheDuration) * time.Minute)

	WebServer.HTMLRender = gv

	WebServer.Use(func(c *gin.Context) {
		c.Set("hd", hd)
	})

	WebServer.Use(func(c *gin.Context) {
		c.Writer.Header().Add("Cache-Control", "max-age=31536000")
	})

	WebServer.Use(gzip.Gzip(gzip.DefaultCompression))
	WebServer.StaticFS("/public", http.FS(StaticFS))
	WebServer.GET("/", cache.CachePage(memorystore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, handlers.ShowIndexPage))
	WebServer.GET("/wiki/raw/*raw", cache.CachePage(memorystore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, handlers.WikiRaw))
	WebServer.GET("/wiki/view/*view", cache.CachePage(memorystore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, handlers.WikiView))
	WebServer.GET("/wiki/search", cache.CachePage(memorystore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, handlers.Search))

}
