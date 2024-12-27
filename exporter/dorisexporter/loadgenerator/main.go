// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
)

type Config struct {
	Port      int
	Duration  int
	Speed     int
	BatchSize int
	Parallel  int
	Type      string
}

func parse(cfg *Config) {
	flag.IntVar(&cfg.Port, "port", 4317, "The port to send data to")
	flag.IntVar(&cfg.Duration, "duration", 10, "The duration of the test in seconds")
	flag.IntVar(&cfg.Speed, "speed", 100, "The speed of the test in items per second")
	flag.IntVar(&cfg.BatchSize, "batchsize", 100, "The size of each batch")
	flag.IntVar(&cfg.Parallel, "parallel", 1, "The number of parallel clients")
	flag.StringVar(&cfg.Type, "type", "log", "The type of data to send (log, trace, metric)")
	flag.Parse()
}

func main() {
	cfg := &Config{}
	parse(cfg)

	options := testbed.LoadOptions{
		DataItemsPerSecond: cfg.Speed,
		ItemsPerBatch:      cfg.BatchSize,
		Parallel:           cfg.Parallel,
		MaxDelay:           60 * time.Second,
	}
	dataProvider := testbed.NewPerfTestDataProvider(options)

	var sender testbed.DataSender
	if cfg.Type == "log" {
		sender = testbed.NewOTLPLogsDataSender("0.0.0.0", cfg.Port)
	} else if cfg.Type == "trace" {
		sender = testbed.NewOTLPTraceDataSender("0.0.0.0", cfg.Port)
	} else if cfg.Type == "metric" {
		sender = testbed.NewOTLPMetricDataSender("0.0.0.0", cfg.Port)
	} else {
		panic("Invalid type")
	}

	loadGenerator, err := testbed.NewLoadGenerator(dataProvider, sender)
	if err != nil {
		panic(err)
	}

	loadGenerator.Start(options)

	for loadGenerator.DataItemsSent() <= 0 {
		fmt.Println("Waiting for load generator to be ready...")
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Load generator is ready")

	for i := 0; i < cfg.Duration-2; i++ {
		fmt.Println("Stats:", loadGenerator.GetStats())
		time.Sleep(1 * time.Second)
	}

	loadGenerator.Stop()
}
