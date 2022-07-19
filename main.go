package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/JojiiOfficial/ZimWiki/handlers"
	"github.com/JojiiOfficial/ZimWiki/utils"
	"github.com/JojiiOfficial/ZimWiki/zim"
)

var (
	//go:embed html/*
	WebFS embed.FS

	//go:embed locale.zip
	LocaleByte []byte

	files []string
)

func main() {
	utils.SetupLogger()

	handlers.WebFS = WebFS

	handlers.LocaleByte = LocaleByte

	utils.LoadConfig()

	// Verify library path
	for _, ele := range utils.Config.LibPath {
		_, err := os.Stat(ele)
		if err != nil {
			utils.Log.Errorf("'%s' is invalid: %s", ele, err)
			return
		}
		path, _ := filepath.Abs(ele)
		files = append(files, path)
	}
	service := zim.New(files)
	err := service.Start(utils.Config.IndexPath)
	if err != nil {
		utils.Log.Fatalln(err)
		return
	}

	utils.StartServer()
}
