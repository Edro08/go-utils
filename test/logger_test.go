package test

import (
	"testing"

	"github.com/edro08/go-utils/logger"
)

func TestLogger_TEXT(t *testing.T) {
	opts := logger.Options{
		MinLevel: logger.DEBUG,
		Format:   logger.FormatText,
	}

	newLogger, _ := logger.New(opts)

	newLogger.Debug("TEST DEBUG", "Format", "TEXT")
}

func TestLogger_JSON(t *testing.T) {
	opts := logger.Options{
		MinLevel: logger.DEBUG,
		Format:   logger.FormatJSON,
	}
	newLogger, _ := logger.New(opts)

	newLogger.Debug("TEST DEBUG", "Format", "JSON")
}

func TestLogger_HiddenDebug(t *testing.T) {
	opts := logger.Options{
		MinLevel: logger.INFO,
		Format:   logger.FormatText,
	}
	newLogger, _ := logger.New(opts)

	newLogger.Debug("TEST DEBUG", "Format", "HIDDEN DEBUG")
	newLogger.Info("TEST INFO", "Format", "HIDDEN DEBUG")
}
