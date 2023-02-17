package utils

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type WebServerConfig struct{}

var WebServer = WebServerConfig{}

func (w WebServerConfig) StartServer(port string) (func(), error) {

	server := &http.Server{Addr: ":" + port}
	_, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("[WEBSERVER] HTTP server failed: %s\n", err)
		}
	}()

	shutdown := func() {
		cancel()

		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Printf("[WEBSERVER] HTTP server shutdown failed: %s\n", err)
		}
		wg.Wait()
	}

	time.Sleep(100 * time.Millisecond)

	return shutdown, nil
}
