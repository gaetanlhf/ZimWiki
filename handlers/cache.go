package handlers

import (
	"time"

	"github.com/JojiiOfficial/ZimWiki/config"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	MemoryStore = persistence.NewInMemoryStore(time.Duration(config.Config.SearchCacheDuration) * time.Minute)
)

func AddHTMLCache(WebServer *gin.Engine) {
	WebServer.Use(func(c *gin.Context) {
		c.Writer.Header().Add("Cache-Control", "max-age=31536000")
	})
}

func AddGzip(WebServer *gin.Engine) {
	WebServer.Use(gzip.Gzip(gzip.DefaultCompression))
}
