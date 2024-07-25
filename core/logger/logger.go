package logger

import (
	"os"

	"http_test/core/config"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	env := config.GetConfig().App.Environment

	if env == "production" {
		// Log as JSON instead of the default ASCII formatter.
		logger.SetFormatter(&logrus.JSONFormatter{})
		// Only log the info severity or above.
		logger.SetLevel(logrus.InfoLevel)
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		logger.SetFormatter(&logrus.TextFormatter{
			ForceQuote:    true,
			FullTimestamp: true,
		})
		logger.SetLevel(logrus.InfoLevel)
	}
	// Output to stdout instead of the default stderr
	logger.SetOutput(os.Stdout)
}

func GetLogger() *logrus.Logger {
	return logger
}
