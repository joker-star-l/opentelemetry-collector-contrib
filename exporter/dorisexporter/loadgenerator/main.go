// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
)

type Config struct {
	Port       int
	Duration   int
	Speed      int
	BatchSize  int
	BatchDelay int
	Parallel   int
	Type       string
}

func main() {
	cfg := Config{
		Port:      4317,
		Duration:  3,
		Speed:     1,
		BatchSize: 1,
		Parallel:  1,
		Type:      "metric",
	}

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

	for !loadGenerator.IsReady() {
		fmt.Println("Waiting for load generator to be ready...")
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("Load generator is ready")

	for i := 0; i < cfg.Duration; i++ {
		fmt.Println("Stats:", loadGenerator.GetStats())
		time.Sleep(1 * time.Second)
	}

	loadGenerator.Stop()
}
