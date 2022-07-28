package handlers

import (
	"time"

	"github.com/JojiiOfficial/ZimWiki/config"
	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
)

func AddRoutes(WebServer *gin.Engine) {
	WebServer.GET("/", cache.CachePage(MemoryStore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, Index))
	WebServer.GET("/wiki/raw/*raw", cache.CachePage(MemoryStore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, WikiRaw))
	WebServer.GET("/wiki/view/*view", cache.CachePage(MemoryStore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, WikiView))
	WebServer.GET("/wiki/search", cache.CachePage(MemoryStore, time.Duration(config.Config.SearchCacheDuration)*time.Minute, Search))
}
