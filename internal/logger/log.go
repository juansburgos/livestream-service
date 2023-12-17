package logger

import (
	"log"
	"os"
	"sync"
)

var logger *log.Logger
var once sync.Once

// Get singleton instance of logger
func Get() *log.Logger {
	once.Do(func() {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	})
	return logger
}
