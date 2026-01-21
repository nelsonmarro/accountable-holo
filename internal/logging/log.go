// Package logging provides logging utilities for the application.
package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Init() (*log.Logger, *log.Logger, error) {
	// Usar AppData (Windows) o .config (Linux) para escribir logs sin requerir Admin
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, nil, err
	}

	// Logs en: %AppData%/Verith/logs
	logsDirPath := filepath.Join(configDir, "Verith", "logs")

	err = os.MkdirAll(logsDirPath, 0o755)
	if err != nil {
		return nil, nil, err
	}

	logFileName := time.Now().Format("2006-01-02")
	newLogFile := filepath.Join(logsDirPath, logFileName)

	file, err := os.OpenFile(newLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, nil, err
	}

	infoMultiWriter := io.MultiWriter(file, os.Stdout)
	errorMultiWriter := io.MultiWriter(file, os.Stderr)

	infoLogger := log.New(infoMultiWriter, "INFO\t ", log.Ldate|log.Ltime)
	errorLogger := log.New(errorMultiWriter, "ERROR\t ", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLogger, errorLogger, nil
}
