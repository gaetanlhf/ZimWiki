package log

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

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func Fatalln(args ...interface{}) {
	Log.Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}
