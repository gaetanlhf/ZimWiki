package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/JojiiOfficial/ZimWiki/handlers"
	"github.com/JojiiOfficial/ZimWiki/utils"
	"github.com/JojiiOfficial/ZimWiki/zim"
	"github.com/briandowns/spinner"
	"github.com/gin-gonic/gin"
	ginlogrus "github.com/toorop/gin-logrus"
)

var (
	//go:embed html/*
	WebFS embed.FS

	//go:embed locale.zip
	LocaleByte []byte

	files []string

	srv  *http.Server
	ctx  context.Context
	stop context.CancelFunc
)

func main() {
	utils.SetupLogger()

	handlers.WebFS = WebFS

	handlers.LocaleByte = LocaleByte

	utils.Srv = srv

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

	startServer()
}
func startServer() {
	utils.Log.Info("Starting web server...")
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// Set release mode
	gin.SetMode(gin.ReleaseMode)
	// Create a gin engine
	webServer := gin.New()
	// Logrus logger handler for gin
	webServer.Use(ginlogrus.Logger(utils.Log), gin.Recovery())
	webServer.GET("/", func(c *gin.Context) {
		time.Sleep(20 * time.Second)
		c.String(http.StatusOK, "Welcome Gin Server")
	})
	srv = &http.Server{
		Addr:    utils.Config.Address + ":" + utils.Config.Port,
		Handler: webServer,
	}
	// Initializing the server in a goroutine so that it will not block the graceful shutdown handling
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			utils.Log.Fatalf("listen: %s\n", err)
		}
	}()

	utils.Log.Infof("HTTP server started on %s:%s", utils.Config.Address, utils.Config.Port)
	// Listen for the interrupt signal.
	<-ctx.Done()
	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	utils.Log.Info("Shutting down gracefully, press Ctrl+C again to force")
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(utils.Config.WaitingTimeWhenRestart)*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		s.Stop()
		utils.Log.Fatal("ZimWiki forced to shutdown: ", err)
	}
	s.Stop()
	utils.Log.Println("Server exiting")
}

// func startServer(zimService *zim.Handler, utils.Config configStruct) {
// 	router := NewRouter(zimService)
// 	server := createServer(router, utils.Config)

// 	// Start server
// 	go func() {
// 		err := server.ListenAndServe()
// 		if err != http.ErrServerClosed {
// 			utils.Log.Fatal(err)
// 		}
// 	}()

// 	utils.Log.Info("Server started")
// 	awaitExit(&server)
// }

// // Build a new Http server
// func createServer(router http.Handler, utils.Config configStruct) http.Server {
// 	return http.Server{
// 		Addr:    utils.Config.Address + ":" + utils.Config.Port,
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

// 	utils.Log.Info("Shutting down server")

// 	if httpServer != nil {
// 		err := httpServer.Shutdown(ctx)
// 		if err != nil {
// 			utils.Log.Warn(err)
// 		}

// 		utils.Log.Info("HTTP server shutdown complete")
// 	}

// 	utils.Log.Info("Shutting down complete")
// 	os.Exit(0)
// }
