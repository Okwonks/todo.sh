package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Okwonks/go-todo/internal/api"
	"github.com/Okwonks/go-todo/internal/tui"
)

func main() {
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
