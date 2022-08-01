// Server принимает OS метрики от agent и сохраняет их в репозиториях.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/server/repositories"
	"github.com/ustkit/cmas/internal/server/router"
	"github.com/ustkit/cmas/internal/server/tools"
	"github.com/ustkit/cmas/internal/types"
)

// @Title CMAS API
// @Version 1.0
// @Contact.name Mark Vaisman
// @License.name MIT
// @License.url https://github.com/ustkit/cmas/blob/main/LICENSE
// @BasePath /
func main() {
	serverConfig := &config.Config{}
	flag.StringVar(&serverConfig.Address, "a", "localhost:8080", "server address")
	flag.BoolVar(&serverConfig.Restore, "r", true, "restore data")
	flag.StringVar(&serverConfig.StoreInterval, "i", "300s", "store interval")
	flag.StringVar(&serverConfig.StoreFile, "f", "/tmp/cmas-metrics-db.json", "store file")
	flag.StringVar(&serverConfig.Key, "k", "", "key")
	flag.StringVar(&serverConfig.DataBaseDSN, "d", "", "database dsn")
	flag.Parse()

	err := env.Parse(serverConfig)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var repository types.MetricRepo

	if serverConfig.DataBaseDSN == "" {
		repository = repositories.NewRepositoryInMemory(serverConfig)
	} else {
		tools.Migration("file://internal/server/migrations/", serverConfig.DataBaseDSN)
		repository, err = repositories.NewRepositoryPostgreSQL(serverConfig)
		if err != nil {
			log.Printf("database: %s", err)
		}
	}

	defer repository.Close()

	err = repository.Restore()
	if err != nil {
		log.Printf("restore data: %s", err)
	}

	if serverConfig.StoreInterval != "0" && serverConfig.StoreFile != "" {
		metricSaver := func(ctx context.Context, storeInterval time.Duration) {
			ticker := time.NewTicker(storeInterval)

			for {
				select {
				case <-ticker.C:
					err = repository.SaveToFile()
					if err != nil {
						log.Println(err)
						stop()
					}
				case <-ctx.Done():
					ticker.Stop()

					return
				}
			}
		}

		storeInterval, err := time.ParseDuration(serverConfig.StoreInterval)
		if err == nil {
			go metricSaver(ctx, storeInterval)
		} else {
			log.Printf("invalid store interval parameter: %s", err)
		}
	}

	go func() {
		err := http.ListenAndServe(serverConfig.Address, router.NewRouter(serverConfig, repository))
		if err != nil {
			log.Println(err)
			stop()
		}
	}()

	<-ctx.Done()
}
