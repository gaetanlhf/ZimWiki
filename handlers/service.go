package handlers

import (
	"github.com/JojiiOfficial/ZimWiki/zim"
	"github.com/gin-gonic/gin"
)

func AddService(WebServer *gin.Engine, ZimService *zim.Handler) {
	WebServer.Use(func(c *gin.Context) {
		c.Set("ZimService", ZimService)
	})
}
