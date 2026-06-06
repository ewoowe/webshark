package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"webshark/internal/handler"
	"webshark/internal/server"
)

func main() {
	fmt.Println("Starting WebShark...")

	router := handler.SetupRouter()
	srv := server.NewServer(":38081", router)

	go func() {
		fmt.Println("Server is running on http://localhost:38081")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")
}
