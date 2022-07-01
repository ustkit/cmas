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
	"github.com/ustkit/cmas/internal/agent"
	"github.com/ustkit/cmas/internal/agent/config"
)

func main() {
	agentConfig := &config.Config{}
	flag.StringVar(&agentConfig.Sever, "a", "localhost:8080", "server address")
	flag.StringVar(&agentConfig.PollInterval, "p", "2s", "poll interval")
	flag.StringVar(&agentConfig.ReportInterval, "r", "10s", "report interval")
	flag.StringVar(&agentConfig.DataType, "t", "jsonbatch", "data type")
	flag.StringVar(&agentConfig.Key, "k", "", "data signing key")
	flag.Parse()

	err := env.Parse(agentConfig)
	if err != nil {
		panic(err)
	}

	pollInterval, err := time.ParseDuration(agentConfig.PollInterval)
	if err != nil {
		log.Fatal(err)
	}

	reportInterval, err := time.ParseDuration(agentConfig.ReportInterval)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	metrics := agent.NewMetrics()

	metricRuntimeUpdater := func(ctx context.Context) {
		ticker := time.NewTicker(pollInterval)

		for {
			select {
			case <-ticker.C:
				err := metrics.RuntimeUpdate()
				if err != nil {
					log.Print(err)
				}
			case <-ctx.Done():
				ticker.Stop()

				return
			}
		}
	}

	go metricRuntimeUpdater(ctx)

	metricGopsutilUpdater := func(ctx context.Context) {
		ticker := time.NewTicker(pollInterval)

		for {
			select {
			case <-ticker.C:
				err := metrics.GopsutilUpdate()
				if err != nil {
					log.Print(err)
				}
			case <-ctx.Done():
				ticker.Stop()

				return
			}
		}
	}

	go metricGopsutilUpdater(ctx)

	metricSender := func(ctx context.Context) {
		client := &http.Client{}
		client.Timeout = reportInterval

		ticker := time.NewTicker(reportInterval)

		for {
			select {
			case <-ticker.C:
				if agentConfig.DataType == "jsonbatch" {
					metrics.SendBatch(ctx, client, agentConfig)
				} else {
					metrics.Send(ctx, client, agentConfig)
				}
			case <-ctx.Done():
				ticker.Stop()

				return
			}
		}
	}

	go metricSender(ctx)

	<-ctx.Done()
}
