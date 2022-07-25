package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/JojiiOfficial/ZimWiki/zim"
	"github.com/gin-gonic/gin"
	gzim "github.com/tim-st/go-zim"
)

func parseWikiRequest(ctx *gin.Context, hd HandlerData, isRedirect bool) (*zim.File, *gzim.Namespace, *gzim.DirectoryEntry, bool) {
	sPath := strings.Split(parseURLPath(ctx.Request.URL), "/")

	var reqWikiID string
	var z *zim.File

	// We can use zim.File for getting
	// The desired wiki and mainpage
	if len(sPath) > 2 {
		// WikiID represents the zim UUID
		reqWikiID = sPath[2]

		hd.ZimService.Mx.Lock()
		// Find requested wiki file by given ID
		z = hd.ZimService.FindWikiFile(reqWikiID)
		hd.ZimService.Mx.Unlock()
		if z == nil {
			return nil, nil, nil, false
		}
	}

	// Something in the request is missing
	if len(sPath) < 5 {
		newLoc := "/"

		// Try to use main page if
		// the page is the only
		// thing missing
		if len(sPath) >= 3 {
			if mainpage := zim.GetMainpageURL(z); len(mainpage) > 0 {
				newLoc = mainpage
			}
		}

		// Something is missing in the given URL
		ctx.Redirect(http.StatusMovedPermanently, newLoc)
		return nil, nil, nil, false
	}

	// Throw error for invalid namespaces
	reqNamespace := sPath[3]
	if !strings.ContainsAny(reqNamespace, "ABIJMUVWX-") || len(reqNamespace) > 1 {
		ctx.AbortWithStatus(http.StatusNotFound)
		return nil, nil, nil, false
	}

	// Parse namespace
	namespace := gzim.Namespace(reqNamespace[0])

	switch namespace {
	case gzim.NamespaceLayout, gzim.NamespaceArticles, gzim.NamespaceImagesFiles, gzim.NamespaceImagesText:
	default:
		ctx.AbortWithStatus(http.StatusNotFound)
		return nil, nil, nil, false
	}

	// reqFileURL is the url of the
	// requested file inside a wiki
	reqFileURL := strings.Join(sPath[4:], "/")

	z.Mx.Lock()
	entry, _, found := z.EntryWithURL(namespace, []byte(reqFileURL))
	z.Mx.Unlock()
	if !found {

		// Only try to find match
		// if not already redirected
		if !isRedirect {
			var useAltURL bool

			// Try to add/remove plural suffix
			if strings.HasSuffix(reqFileURL, "s") {
				useAltURL = true
				ctx.Request.URL.Path = ctx.Request.URL.Path[:len(ctx.Request.URL.Path)-1]
			} else {
				useAltURL = true
				ctx.Request.URL.Path = ctx.Request.URL.Path + "s"
			}

			if useAltURL {
				return parseWikiRequest(ctx, hd, true)
			}
		}

		ctx.AbortWithStatus(http.StatusNotFound)
		return nil, nil, nil, false
	}

	// Follow redirect
	if entry.IsRedirect() {
		z.Mx.Lock()
		entry, _ = z.FollowRedirect(&entry)
		z.Mx.Unlock()

		// Use preview url if
		// entry is article
		var destURL string
		if entry.IsArticle() {
			destURL = zim.GetWikiURL(z, entry)
		} else {
			destURL = zim.GetRawWikiURL(z, entry)
		}

		ctx.Redirect(http.StatusMovedPermanently, destURL)
		return nil, nil, nil, false
	}

	return z,
		&namespace,
		&entry,
		true
}

// WikiRaw handle direct wiki requests, without embedding into the webUI
func WikiRaw(ctx *gin.Context) {
	hd := ctx.MustGet("hd").(HandlerData)
	// Find file and dirEntry
	z, _, entry, success := parseWikiRequest(ctx, hd, false)
	if !success {
		// We already handled
		// http errors & redirects
		ctx.AbortWithStatus(http.StatusNotFound)
		return

	}

	// Create reader from requested file
	z.Mx.Lock()
	defer z.Mx.Unlock()
	blobReader, _, err := z.BlobReader(entry)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Set Mimetype accordingly
	if mimetypeList := z.MimetypeList(); int(entry.Mimetype()) < len(mimetypeList) {
		ctx.Header("Content-Type", mimetypeList[entry.Mimetype()])
	}

	// // Cache files
	// w.Header().Set("Cache-Control", "max-age=31536000, public")

	// Send raw file
	buff := make([]byte, 1024*1024)
	_, err = io.CopyBuffer(ctx.Writer, blobReader, buff)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)

		return
	}
}

// WikiView sends a human friendly preview page for a WIKI site
func WikiView(ctx *gin.Context) {
	hd := ctx.MustGet("hd").(HandlerData)
	// Find file and dirEntry
	z, namespace, entry, success := parseWikiRequest(ctx, hd, false)
	if !success {
		// We already handled
		// http errors & redirects
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	var favurl, favType string
	z.Mx.Lock()
	favIcon, err := z.Favicon()
	z.Mx.Unlock()
	if err == nil {
		if mimetypeList := z.MimetypeList(); int(favIcon.Mimetype()) < len(mimetypeList) {
			favType = mimetypeList[favIcon.Mimetype()]
		}
		favurl = zim.GetRawWikiURL(z, favIcon)
	}

	ctx.HTML(http.StatusOK, "viewPage", gin.H{
		"FavIcon":   favurl,
		"Favtype":   favType,
		"Wiki":      z.GetID(),
		"Namespace": string(*namespace),
		"Source":    zim.GetRawWikiURL(z, *entry),
	})

}

func parseURLPath(u *url.URL) string {
	path := u.Path
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	return path
}
