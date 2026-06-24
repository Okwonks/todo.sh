package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Okwonks/go-todo/internal/api"
	"github.com/Okwonks/go-todo/internal/tui"
)

func main() {
	// Determine the standard user cache directory (e.g., ~/Library/Caches on macOS, ~/.cache on Linux)
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = "." // Fallback to current working directory
	}

	logDir := filepath.Join(cacheDir, "todosh")
	// Ensure the directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logFilePath := filepath.Join(logDir, "todo.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Redirect standard logger output to the log file
	log.SetOutput(logFile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiChan := make(chan error, 1)
	go func() {
		apiChan <- api.Start(ctx)
	}()

	url := "http://localhost:8080"
	if err := waitForApi(url, 30*time.Second); err != nil {
		log.Fatalf("API service unreachable: %v", err)
	}

	if err := tui.Start(); err != nil {
		log.Printf("Couldn't start TUI: %v", err)
	}

	cancel() // sends signal to API server

	if err := <-apiChan; err != nil {
		log.Printf("API shutdown error: %v", err)
	}
}

func waitForApi(url string, timeout time.Duration) error {
	c := &http.Client{Timeout: 1 * time.Second}
	t := time.Now().Add(timeout)

	for time.Now().Before(t) {
		resp, err := c.Get(url + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("API did not become ready within %v", timeout)
}
