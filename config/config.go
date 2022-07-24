package config

import (
	"github.com/JojiiOfficial/ZimWiki/log"
	"github.com/spf13/viper"
)

var (
	Config configStruct
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
			log.Fatal("Config file not found")
		} else {
			// Config file was found but another error was produced
			log.Fatal("Error when reading the config file")
		}
		// Failed to read configuration file
		log.Fatal(err)
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

	log.Info("Config loaded successfully")

}
