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
