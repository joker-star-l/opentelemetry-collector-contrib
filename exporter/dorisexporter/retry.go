// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dorisexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dorisexporter"

import (
	"fmt"
	"strings"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type dataType int

const (
	labelLog                        = 0
	labelTrace                      = 1
	labelMetricGauge                = 2
	labelMetricSum                  = 3
	labelMetricHistogram            = 4
	labelMetricExponentialHistogram = 5
	labelMetricSummary              = 6
)

// retryMaps is a array of maps for storing retry data addresses and labels
var retryMaps [7]cmap.ConcurrentMap[string, string]

func init() {
	for i := 0; i < 7; i++ {
		retryMaps[i] = cmap.New[string]()
	}
}

// dataAddress returns the address of the orig in pmetric.Metrics, plog.Logs, or ptrace.Traces
func dataAddress(data any) string {
	s := fmt.Sprintf("%v", data)
	return s[1:strings.Index(s, " ")]
}

func addRetryData(t dataType, address string, label string) {
	retryMaps[t].Set(address, label)
	fmt.Printf("add {key: %s, value: %s}\n", address, label)
}

func popRetryData(t dataType, address string) string {
	label, ok := retryMaps[t].Pop(address)
	fmt.Printf("pop {key: %s, value: %s}\n", address, label)
	if !ok {
		return ""
	}
	return label
}
