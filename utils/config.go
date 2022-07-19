package utils

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/teamwork/reload"
)

var (
	Config configStruct
	Srv    *http.Server

	s = spinner.New(spinner.CharSets[26], 100*time.Millisecond)
)

type configStruct struct {
	LibPath                []string
	Address                string
	Port                   string
	IndexPath              string
	EnableSearchCache      bool
	SearchCacheDuration    int
	EnableAutoRestart      bool
	WaitingTimeWhenRestart int
}

func LoadConfig() {
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
			Log.Fatal("Config file not found")
		} else {
			// Config file was found but another error was produced
			Log.Fatal("Error when reading the config file")
		}
		// Failed to read configuration file
		Log.Fatal(err)
	}

	// Load the configuration from the configuration file
	LibPath := viper.GetStringSlice("librarypath")
	Address := viper.GetString("Address")
	Port := viper.GetString("Port")
	IndexPath := viper.GetString("indexpath")
	EnableSearchCache := viper.GetBool("enableasearchcache")
	SearchCacheDuration := viper.GetInt("searchcacheduration")
	EnableAutoRestart := viper.GetBool("enableautorestart")
	WaitingTimeWhenRestart := viper.GetInt("waitingtimebeforerestart")

	Config = configStruct{LibPath: LibPath, Address: Address, Port: Port, IndexPath: IndexPath, EnableSearchCache: EnableSearchCache, SearchCacheDuration: SearchCacheDuration, EnableAutoRestart: EnableAutoRestart, WaitingTimeWhenRestart: WaitingTimeWhenRestart}

	Log.Info("Config loaded successfully")

	if EnableAutoRestart {
		// Viper should check the configuration file for changes
		viper.WatchConfig()
		// When the configuration file is updated
		viper.OnConfigChange(func(e fsnotify.Event) {
			Log.Warn("The configuration file has been updated: ", e.Name)
			if Srv == nil {
				Log.Warn("ZimWiki will be restarted in " + strconv.Itoa(Config.WaitingTimeWhenRestart) + " second(s)...")
			} else {
				Log.Warn("ZimWiki will restart within a maximum of " + strconv.Itoa(Config.WaitingTimeWhenRestart) + " seconds if the current http requests do not end")
			}
			// Show a spinner
			s.Start()

			// Wait some time, the file can be updated successively
			if Srv != nil {
				// Create a deadline for the await
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				if Srv != nil {
					err := Srv.Shutdown(ctx)
					if err != nil {
						s.Stop()
						Log.Warn("The HTTP server has been killed after " + strconv.Itoa(Config.WaitingTimeWhenRestart) + " seconds of waiting")
					} else {
						Log.Info("The HTTP server has been successfully stopped")
					}
				}
			} else {
				time.Sleep(time.Duration(Config.WaitingTimeWhenRestart) * time.Second)
			}
			// Stop the spinner
			s.Stop()
			// Restart the program
			Log.Info("Restart in progress...")
			reload.Exec()
		})
	}
}
