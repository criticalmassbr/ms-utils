package utils

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

type HealthCheckDBConfig struct {
	Type string
	Url  string
}

type HealthCheckRabbitMQConfig struct {
	Url string
}

type HealthCheckConfig struct {
	HealthCheckDBConfig       []HealthCheckDBConfig
	HealthCheckRabbitMQConfig *HealthCheckRabbitMQConfig
}

var HealthCheck = HealthCheckConfig{}

func (h HealthCheckConfig) StartServer(port string) (func(), error) {

	server := &http.Server{Addr: ":" + port}
	_, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {

			if HealthCheck.HealthCheckRabbitMQConfig != nil {
				rabbitErr := rabbitMQHealthCheck(HealthCheck.HealthCheckRabbitMQConfig.Url)
				if rabbitErr != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("FAIL"))
					fmt.Printf("[UTILS][WEBSERVER] RabbitMQ server is down")
					return
				}
			}

			if len(HealthCheck.HealthCheckDBConfig) > 0 {
				for _, dbConfig := range HealthCheck.HealthCheckDBConfig {
					dbErr := dbHealthCheck(dbConfig.Type, dbConfig.Url)
					if dbErr != nil {
						w.WriteHeader(http.StatusServiceUnavailable)
						w.Write([]byte("FAIL"))
						fmt.Printf("[UTILS][WEBSERVER] DB server is down: %s\n", dbConfig.Type)
						return
					}
				}
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("[UTILS][WEBSERVER] HTTP server failed: %s\n", err)
		}
	}()

	shutdown := func() {
		cancel()

		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Printf("[UTILS][WEBSERVER] HTTP server shutdown failed: %s\n", err)
		}
		wg.Wait()
	}

	time.Sleep(100 * time.Millisecond)

	return shutdown, nil
}

func rabbitMQHealthCheck(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Channel()
	if err != nil {
		return err
	}

	return nil
}

func dbHealthCheck(dbType string, url string) error {
	db, err := sql.Open(dbType, url)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}
