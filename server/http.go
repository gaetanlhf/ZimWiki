package server

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/JojiiOfficial/ZimWiki/config"
	"github.com/JojiiOfficial/ZimWiki/handlers"
	"github.com/JojiiOfficial/ZimWiki/log"
	"github.com/JojiiOfficial/ZimWiki/zim"
	"github.com/briandowns/spinner"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/teamwork/reload"
	ginlogrus "github.com/toorop/gin-logrus"
)

var (
	WebServer  *gin.Engine
	ZimService *zim.Handler
)

func StartServer(service *zim.Handler) {
	ZimService = service
	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)

	log.Info("Starting web server...")

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Set release mode
	gin.SetMode(gin.ReleaseMode)

	// Create a gin engine
	WebServer = gin.New()

	// Logrus logger handler for gin
	WebServer.Use(ginlogrus.Logger(log.Log), gin.Recovery())

	// router.GetRoutes(WebServer, service)
	srv := &http.Server{
		Addr:    config.Config.Address + ":" + config.Config.Port,
		Handler: WebServer,
	}

	// Initializing the server in a goroutine so that it will not block the graceful shutdown handling
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	handlers.AddHTMLCache(WebServer)
	handlers.AddGzip(WebServer)
	handlers.AddStatic(WebServer)
	handlers.AddService(WebServer, ZimService)
	WebServer.HTMLRender = handlers.AddTemplate(WebServer)
	handlers.AddRoutes(WebServer)

	log.Infof("HTTP server started on %s:%s", config.Config.Address, config.Config.Port)

	if config.Config.EnableAutoRestart {
		// Viper should check the configuration file for changes
		viper.WatchConfig()
		// When the configuration file is updated
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Warn("The configuration file has been updated: ", e.Name)
			log.Warn("ZimWiki will restart in a maximum of " + strconv.Itoa(config.Config.WaitingTimeWhenRestart) + " seconds if the potentially ongoing HTTP requests do not end before")

			// Show a spinner
			s.Start()

			// Create a deadline for the await
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Config.WaitingTimeWhenRestart)*time.Second)
			defer cancel()

			if srv != nil {
				err := srv.Shutdown(ctx)
				if err != nil {
					s.Stop()
					log.Warn("The " + strconv.Itoa(config.Config.WaitingTimeWhenRestart) + "-second time limit has been exceeded")
				} else {
					log.Warn("HTTP requests were successfully terminated before the " + strconv.Itoa(config.Config.WaitingTimeWhenRestart) + "second time limit was reached")
				}
			}

			// Stop the spinner
			s.Stop()

			// Restart the program
			log.Info("Restart in progress...")
			reload.Exec()
		})
	}

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Remove that ugly '^C'
	fmt.Print("\r")

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()

	log.Info("Shutting down gracefully, press Ctrl+C again to force")
	log.Info("ZimWiki will shutdown in a maximum of " + strconv.Itoa(config.Config.WaitingTimeWhenRestart) + " seconds if the potentially ongoing HTTP requests do not end before")

	// Show a spinner
	s.Start()

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Config.WaitingTimeWhenRestart)*time.Second)
	defer cancel()

	// Remove that ugly '^C'
	fmt.Print("\r")

	err := srv.Shutdown(ctx)
	if err != nil {
		s.Stop()
		log.Fatal("ZimWiki forced to shutdown: ", err)
	}

	// Stop the spinner
	s.Stop()
	log.Info("Shutdown...")
}
