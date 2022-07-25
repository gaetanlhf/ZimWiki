package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/JojiiOfficial/ZimWiki/config"
	"github.com/JojiiOfficial/ZimWiki/handlers"
	"github.com/JojiiOfficial/ZimWiki/log"
	"github.com/JojiiOfficial/ZimWiki/server"
	"github.com/JojiiOfficial/ZimWiki/zim"
)

var (
	//go:embed templates/*
	TemplateFS embed.FS

	//go:embed locale.zip
	LocaleByte []byte

	files []string
)

func main() {
	log.SetupLogger()

	server.TemplateFS = TemplateFS

	handlers.LocaleByte = LocaleByte

	config.LoadConfig()

	// Verify library path
	for _, ele := range config.Config.LibPath {
		_, err := os.Stat(ele)
		if err != nil {
			log.Errorf("'%s' is invalid: %s", ele, err)
			return
		}
		path, _ := filepath.Abs(ele)
		files = append(files, path)
	}
	service := zim.New(files)
	err := service.Start(config.Config.IndexPath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	server.StartServer(service)
}
