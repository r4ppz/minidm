package minidm

import (
	"log"
	"os"
)

var (
	debugMode bool
	logger    *log.Logger
)

func init() {
	debugMode = os.Getenv("MINIDM_DEBUG") == "1"
	logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
}

func Debugf(format string, args ...any) {
	if debugMode {
		logger.Printf("[DEBUG] "+format, args...)
	}
}

func Infof(format string, args ...any) {
	logger.Printf("[INFO] "+format, args...)
}

func Errorf(format string, args ...any) {
	logger.Printf("[ERROR] "+format, args...)
}
