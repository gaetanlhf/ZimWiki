package utils

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Log = logrus.New()
)

func SetupLogger() {
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: false,
		TimestampFormat:  time.Stamp,
		FullTimestamp:    true,
		ForceColors:      true,
		DisableColors:    false,
	})
}
