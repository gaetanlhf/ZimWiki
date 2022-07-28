package handlers

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	StaticFS embed.FS
)

func AddStatic(WebServer *gin.Engine) {
	WebServer.StaticFS("/public", http.FS(StaticFS))
}
