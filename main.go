package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JojiiOfficial/ZimWiki/handlers"
	"github.com/JojiiOfficial/ZimWiki/zim"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	//go:embed html/*
	WebFS embed.FS

	//go:embed locale.zip
	LocaleByte []byte

	config []configStruct
)

type configStruct struct {
	libPath             string
	address             string
	port                string
	EnableSearchCache   bool
	SearchCacheDuration int
}

func main() {
	setupLogger()

	handlers.WebFS = WebFS

	handlers.LocaleByte = LocaleByte

	// Configuration file path is the actual folder
	viper.AddConfigPath(".")
	// Configuration file type is yaml
	viper.SetConfigType("yaml")
	// Configuration file name is config
	viper.SetConfigName("config")
	// Read configuration file
	err := viper.ReadInConfig()

	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			// Config file not found
			log.Fatal("Config file not found")
		} else {
			// Config file was found but another error was produced
			log.Fatal("Error when reading the config file")
		}
		// Failed to read configuration file
		log.Fatal(err)
	}

	log.Info("Config loaded successfully")

	// Load the configuration from the configuration file
	libPath := viper.GetString("librarypath")
	address := viper.GetString("address")
	port := viper.GetString("port")
	EnableSearchCache := viper.GetBool("enableasearchcache")
	SearchCacheDuration := viper.GetInt("searchcacheduration")

	config := configStruct{libPath: libPath, address: address, port: port, EnableSearchCache: EnableSearchCache, SearchCacheDuration: SearchCacheDuration}

	handlers.EnableSearchCache = EnableSearchCache
	handlers.SearchCacheDuration = SearchCacheDuration

	// Verify library path
	s, err := os.Stat(config.libPath)
	if err != nil {
		log.Errorf("Can't use '%s' as library path. %s", config.libPath, err)
		return
	}
	if !s.IsDir() {
		log.Error("Library must be a path!")
		return
	}

	service := zim.New(config.libPath)
	err = service.Start(config.libPath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	startServer(service, config)
}

func startServer(zimService *zim.Handler, config configStruct) {
	router := NewRouter(zimService)
	server := createServer(router, config)

	// Start server
	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.Info("Server started")
	awaitExit(&server)
}

// Build a new Http server
func createServer(router http.Handler, config configStruct) http.Server {
	return http.Server{
		Addr:    config.address + ":" + config.port,
		Handler: router,
	}
}

// Shutdown server gracefully
func awaitExit(httpServer *http.Server) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	// await os signal
	<-signalChan

	// Create a deadline for the await
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Remove that ugly '^C'
	fmt.Print("\r")

	log.Info("Shutting down server")

	if httpServer != nil {
		err := httpServer.Shutdown(ctx)
		if err != nil {
			log.Warn(err)
		}

		log.Info("HTTP server shutdown complete")
	}

	log.Info("Shutting down complete")
	os.Exit(0)
}

func setupLogger() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  time.Stamp,
		FullTimestamp:    true,
		ForceColors:      true,
		DisableColors:    false,
	})
}
