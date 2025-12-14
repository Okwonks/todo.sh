package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Okwonks/go-todo/internal"
)

func main() {
	srv := internal.Server()

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down...")

		ctx, release := context.WithTimeout(context.Background(), 5*time.Second)
		defer release()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server Shutdown: %v", err)
		}

		close(idleConnsClosed)
		log.Println("Shutdown completed.")
	}()

	fmt.Printf("TODO api v1 listening on port %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-idleConnsClosed
}
