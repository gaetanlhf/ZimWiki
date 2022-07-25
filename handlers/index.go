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

// Index handle index route
func ShowIndexPage(ctx *gin.Context) {
	hd := ctx.MustGet("hd").(HandlerData)
	var cards []HomeCards

	// Create cards
	for i := range hd.ZimService.GetFiles() {
		file := &hd.ZimService.GetFiles()[i]

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
