package utils

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	ginlogrus "github.com/toorop/gin-logrus"
)

var (
	srv  *http.Server
	ctx  context.Context
	stop context.CancelFunc
)

func StartServer() {
	Log.Info("Starting web server...")
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// Set release mode
	gin.SetMode(gin.ReleaseMode)
	// Create a gin engine
	webServer := gin.New()
	// Logrus logger handler for gin
	webServer.Use(ginlogrus.Logger(Log), gin.Recovery())
	webServer.GET("/", func(c *gin.Context) {
		time.Sleep(20 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})
	srv = &http.Server{
		Addr:    Config.Address + ":" + Config.Port,
		Handler: webServer,
	}
	// Initializing the server in a goroutine so that it will not block the graceful shutdown handling
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			Log.Fatalf("listen: %s\n", err)
		}
	}()

	Log.Infof("HTTP server started on %s:%s", Config.Address, Config.Port)
	// Listen for the interrupt signal.
	<-ctx.Done()
	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	Log.Info("Shutting down gracefully, press Ctrl+C again to force")
	s.Start()
	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(Config.WaitingTimeWhenRestart)*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		s.Stop()
		Log.Fatal("ZimWiki forced to shutdown: ", err)
	}
	s.Stop()
	Log.Println("Server exiting")
}
