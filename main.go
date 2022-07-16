package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/JojiiOfficial/ZimWiki/handlers"
	"github.com/JojiiOfficial/ZimWiki/zim"
	"github.com/briandowns/spinner"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/teamwork/reload"
)

var (
	//go:embed html/*
	WebFS embed.FS

	//go:embed locale.zip
	LocaleByte []byte

	config configStruct
	files  []string

	srv  *http.Server
	ctx  context.Context
	stop context.CancelFunc
)

type configStruct struct {
	libPath                []string
	address                string
	port                   string
	indexPath              string
	EnableSearchCache      bool
	SearchCacheDuration    int
	enableAutoRestart      bool
	waitingTimeWhenRestart int
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

	// Load the configuration from the configuration file
	libPath := viper.GetStringSlice("librarypath")
	address := viper.GetString("address")
	port := viper.GetString("port")
	indexPath := viper.GetString("indexpath")
	EnableSearchCache := viper.GetBool("enableasearchcache")
	SearchCacheDuration := viper.GetInt("searchcacheduration")
	enableAutoRestart := viper.GetBool("enableautorestart")
	waitingTimeWhenRestart := viper.GetInt("waitingtimebeforerestart")

	config = configStruct{libPath: libPath, address: address, port: port, indexPath: indexPath, EnableSearchCache: EnableSearchCache, SearchCacheDuration: SearchCacheDuration, enableAutoRestart: enableAutoRestart, waitingTimeWhenRestart: waitingTimeWhenRestart}

	log.Info("Config loaded successfully")

	if enableAutoRestart {
		// Viper should check the configuration file for changes
		viper.WatchConfig()
		// When the configuration file is updated
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Warn("The configuration file has been updated: ", e.Name)
			log.Warn("ZimWiki will be restarted in " + strconv.Itoa(config.waitingTimeWhenRestart) + " second(s)...")
			// Show a spinner
			s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
			s.Start()
			s.Color("white")
			// Wait some time, the file can be updated successively
			time.Sleep(time.Duration(config.waitingTimeWhenRestart) * time.Second)
			// Stop the spinner
			s.Stop()
			// Restart the program
			reload.Exec()
		})
	}

	handlers.EnableSearchCache = EnableSearchCache
	handlers.SearchCacheDuration = SearchCacheDuration

	// Verify library path
	for _, ele := range config.libPath {
		_, err := os.Stat(ele)
		if err != nil {
			log.Errorf("'%s' is invalid: %s", ele, err)
			return
		}
		path, _ := filepath.Abs(ele)
		files = append(files, path)
	}
	service := zim.New(files)
	err = service.Start(config.indexPath)
	if err != nil {
		log.Fatalln(err)
		return
	}

	startServer()
}

func startServer() {
	log.Info("Starting web server...")
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// Set release mode
	gin.SetMode(gin.ReleaseMode)
	// Create a gin engine
	webServer := gin.New()
	webServer.GET("/", func(c *gin.Context) {
		time.Sleep(20 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	srv = &http.Server{
		Addr:    config.address + ":" + config.port,
		Handler: webServer,
	}
	// Initializing the server in a goroutine so that it will not block the graceful shutdown handling
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Infof("HTTP server started on %s:%s", config.address, config.port)

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.waitingTimeWhenRestart)*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		s.Stop()
		log.Fatal("ZimWiki forced to shutdown: ", err)
	}
	s.Stop()
	log.Println("Server exiting")
}

// func startServer(zimService *zim.Handler, config configStruct) {
// 	router := NewRouter(zimService)
// 	server := createServer(router, config)

// 	// Start server
// 	go func() {
// 		err := server.ListenAndServe()
// 		if err != http.ErrServerClosed {
// 			log.Fatal(err)
// 		}
// 	}()

// 	log.Info("Server started")
// 	awaitExit(&server)
// }

// // Build a new Http server
// func createServer(router http.Handler, config configStruct) http.Server {
// 	return http.Server{
// 		Addr:    config.address + ":" + config.port,
// 		Handler: router,
// 	}
// }

// // Shutdown server gracefully
// func awaitExit(httpServer *http.Server) {
// 	signalChan := make(chan os.Signal, 1)
// 	signal.Notify(signalChan, syscall.SIGINT, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

// 	// await os signal
// 	<-signalChan

// 	// Create a deadline for the await
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
// 	defer cancel()

// 	// Remove that ugly '^C'
// 	fmt.Print("\r")

// 	log.Info("Shutting down server")

// 	if httpServer != nil {
// 		err := httpServer.Shutdown(ctx)
// 		if err != nil {
// 			log.Warn(err)
// 		}

// 		log.Info("HTTP server shutdown complete")
// 	}

// 	log.Info("Shutting down complete")
// 	os.Exit(0)
// }

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
