package handlers

import (
	"net/http"

	"github.com/JojiiOfficial/ZimWiki/zim"
	"github.com/gin-gonic/gin"
)

var (
	version   string
	buildTime string
)

type HomeCards struct {
	Image string
	Title string
	Text  string
	Link  string
}

// Index handle index route
func Index(ctx *gin.Context) {
	ZimService := ctx.MustGet("ZimService").(*zim.Handler)
	var cards []HomeCards

	// Create cards
	for i := range ZimService.GetFiles() {
		file := &ZimService.GetFiles()[i]

		info := file.GetInfos()

		// Get Faviconlink
		var favurl string
		favIcon, err := file.Favicon()
		if err == nil {
			favurl = zim.GetRawWikiURL(file, favIcon)
		}

		// Create homeCard
		cards = append(cards, HomeCards{
			Text:  info.GetDescription(),
			Title: info.Title,
			Image: favurl,
			Link:  zim.GetMainpageURL(file),
		})
	}

	ctx.HTML(http.StatusOK, "home", gin.H{"Cards": cards, "Version": version, "Buildtime": buildTime})

}
