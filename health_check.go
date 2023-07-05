package utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/vault/api"
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

type HealthCheckRedisConfig struct {
	Urls []string
}

type HealthCheckVaultConfig struct {
	Url  string
	Cert string
}

type HealthCheckConfig struct {
	HealthCheckDBConfig       []HealthCheckDBConfig
	HealthCheckRabbitMQConfig *HealthCheckRabbitMQConfig
	HealthCheckRedisConfig    *HealthCheckRedisConfig
	HealthCheckVaultConfig    *HealthCheckVaultConfig
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
					fmt.Printf("[UTILS][WEBSERVER] RabbitMQ server is down: %v\n", rabbitErr)
					return
				}
			}

			if len(HealthCheck.HealthCheckDBConfig) > 0 {
				for _, dbConfig := range HealthCheck.HealthCheckDBConfig {
					dbErr := dbHealthCheck(dbConfig.Type, dbConfig.Url)
					if dbErr != nil {
						w.WriteHeader(http.StatusServiceUnavailable)
						w.Write([]byte("FAIL"))
						fmt.Printf("[UTILS][WEBSERVER] DB server is down: %s %v\n", dbConfig.Type, dbErr)
						return
					}
				}
			}

			if HealthCheck.HealthCheckRedisConfig != nil {
				if err := redisHealthCheck(*HealthCheck.HealthCheckRedisConfig); err != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("FAIL"))
					fmt.Printf("[UTILS][WEBSERVER] Redis cluster is down %v\n", err)
					return
				}
			}

			if HealthCheck.HealthCheckVaultConfig != nil {
				vaultErr := vaultHealthCheck(HealthCheck.HealthCheckVaultConfig)
				if vaultErr != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("FAIL"))
					fmt.Printf("[UTILS][WEBSERVER] Vault server is down %v\n", vaultErr)
					return
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

func redisHealthCheck(cfg HealthCheckRedisConfig) error {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: cfg.Urls,
	})
	defer client.Close()

	if err := client.Ping().Err(); err != nil {
		return err
	}

	return nil
}

func vaultHealthCheck(config *HealthCheckVaultConfig) error {

	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(config.Cert)
	if err != nil {
		return fmt.Errorf("unable to read Vault certificate: %v", err)
	}
	certs.AppendCertsFromPEM(pemData)

	vaultConfig := &api.Config{
		Address: config.Url,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: certs,
				},
			},
		},
	}

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return err
	}

	// Make a request to the health endpoint of Vault
	resp, err := client.Sys().Health()
	if err != nil {
		return err
	}

	if !resp.Initialized || resp.Sealed {
		return fmt.Errorf("vault is not healthy: %v", resp)
	}

	return nil
}
