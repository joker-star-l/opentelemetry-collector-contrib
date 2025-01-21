// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dorisexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dorisexporter"

import "go.opentelemetry.io/collector/pdata/pmetric"

type metricModel interface {
	metricType() pmetric.MetricType
	tableSuffix() string
	add(pm pmetric.Metric, dm *dMetric, e *metricsExporter) error
	raw() any
	size() int
	bytes() ([]byte, error)
	dataType() dataType
	label() string
}

func generateMetricLabel(m metricModel, cfg *Config, dataAddress string) string {
	label := popRetryData(m.dataType(), dataAddress)
	if label == "" {
		label = generateLabel(cfg, cfg.Table.Metrics+m.tableSuffix())
	}
	return label
}

// dMetric Basic Metric
type dMetric struct {
	ServiceName        string         `json:"service_name"`
	ServiceInstanceID  string         `json:"service_instance_id"`
	MetricName         string         `json:"metric_name"`
	MetricDescription  string         `json:"metric_description"`
	MetricUnit         string         `json:"metric_unit"`
	ResourceAttributes map[string]any `json:"resource_attributes"`
	ScopeName          string         `json:"scope_name"`
	ScopeVersion       string         `json:"scope_version"`
}

// dExemplar Exemplar to Doris
type dExemplar struct {
	FilteredAttributes map[string]any `json:"filtered_attributes"`
	Timestamp          string         `json:"timestamp"`
	Value              float64        `json:"value"`
	SpanID             string         `json:"span_id"`
	TraceID            string         `json:"trace_id"`
}
