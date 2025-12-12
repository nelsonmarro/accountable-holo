// package logging provides logging utilities for the application.
package logging

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func Init() (*log.Logger, *log.Logger, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, nil, err
	}

	dirs := filepath.Dir(executablePath)
	logsDirPath := filepath.Join(dirs, "logs")

	os.MkdirAll(logsDirPath, 0o755)

	logFileName := time.Now().Format("2006-01-02")
	newLogFile := filepath.Join(logsDirPath, logFileName)
	err := os.OpenFile(newLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
}
