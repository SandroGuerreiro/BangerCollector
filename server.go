package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	srvAddr = "http://localhost:8080"
)

var shutdown chan bool

func startServer() {
	var wg sync.WaitGroup

	// Initialize the server
	srv := &http.Server{Addr: ":8080"}

	// Register the handlers
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/import", importHandler)
	http.HandleFunc("/delete", deleteHandler)

	// Increment the WaitGroup counter.
	wg.Add(1)

	// Serve in a goroutine
	go func() {
		log.Println("Server is starting...")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server is shutting down...")

	// Shutdown the server
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes.

		// Timeout context for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server Shutdown Failed:%+v", err)
		}
		log.Println("Server exited properly")
	}()

	// Continue with the rest of your program
	wg.Wait() // Wait for the HTTP server to shut down.

	// Now that the server is down, continue with the rest of the program
}
