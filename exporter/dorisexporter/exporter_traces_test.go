// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dorisexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dorisexporter"

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
	"go.uber.org/zap"
)

func TestPushTraceData(t *testing.T) {
	port, err := findRandomPort()
	require.NoError(t, err)

	config := createDefaultConfig().(*Config)
	config.Endpoint = fmt.Sprintf("http://127.0.0.1:%d", port)
	config.CreateSchema = false

	err = config.Validate()
	require.NoError(t, err)

	exporter := newTracesExporter(zap.NewNop(), config, componenttest.NewNopTelemetrySettings())

	ctx := context.Background()

	client, err := createDorisHTTPClient(ctx, config, nil, componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)
	require.NotNil(t, client)

	exporter.client = client

	defer func() {
		_ = exporter.shutdown(ctx)
	}()

	server := &http.Server{
		ReadTimeout: 3 * time.Second,
		Addr:        fmt.Sprintf(":%d", port),
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/otel/otel_traces/_stream_load", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"Status":"Success"}`))
		})
		server.Handler = mux
		err = server.ListenAndServe()
		assert.Equal(t, http.ErrServerClosed, err)
	}()

	err0 := fmt.Errorf("Not Started")
	for i := 0; err0 != nil && i < 10; i++ { // until server started
		err0 = exporter.pushTraceData(ctx, simpleTraces(10))
		time.Sleep(100 * time.Millisecond)
	}
	require.NoError(t, err0)

	_ = server.Shutdown(ctx)
}

func TestPushTraceDataRetry(t *testing.T) {
	port, err := findRandomPort()
	require.NoError(t, err)

	config := createDefaultConfig().(*Config)
	config.Endpoint = fmt.Sprintf("http://127.0.0.1:%d", port)
	config.CreateSchema = false

	err = config.Validate()
	require.NoError(t, err)

	exporter := newTracesExporter(zap.NewNop(), config, componenttest.NewNopTelemetrySettings())

	ctx := context.Background()

	client, err := createDorisHTTPClient(ctx, config, nil, componenttest.NewNopTelemetrySettings())
	require.NoError(t, err)
	require.NotNil(t, client)

	exporter.client = client

	defer func() {
		_ = exporter.shutdown(ctx)
	}()

	server := &http.Server{
		ReadTimeout: 3 * time.Second,
		Addr:        fmt.Sprintf(":%d", port),
	}

	times := 0
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/otel/otel_traces/_stream_load", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			times++
			if times < 3 {
				_, _ = w.Write([]byte(`{"Status":"Fail"}`))
				return
			}
			_, _ = w.Write([]byte(`{"Status":"Success"}`))
		})
		server.Handler = mux
		err = server.ListenAndServe()
		assert.Equal(t, http.ErrServerClosed, err)
	}()

	isRetryError := func(err error) bool {
		if err == nil {
			return false
		}
		return strings.HasPrefix(err.Error(), "failed to push trace data, response:")
	}

	traces := simpleTraces(10)
	dataAddress := dataAddress(traces)
	_, ok := retryMaps[labelTrace].Get(dataAddress)
	require.False(t, ok)

	err0 := fmt.Errorf("Not Started")
	for i := 0; !isRetryError(err0) && i < 10; i++ { // until server started
		err0 = exporter.pushTraceData(ctx, traces)
		time.Sleep(100 * time.Millisecond)
	}
	require.True(t, isRetryError(err0))

	label, ok := retryMaps[labelTrace].Get(dataAddress)
	require.True(t, ok)
	require.NotEqual(t, "", label)

	// first retry: fail
	err0 = exporter.pushTraceData(ctx, traces)
	require.True(t, isRetryError(err0))

	labelRetry, ok := retryMaps[labelTrace].Get(dataAddress)
	require.True(t, ok)
	require.Equal(t, label, labelRetry)

	// second retry: success
	err0 = exporter.pushTraceData(ctx, traces)
	require.NoError(t, err0)
	_, ok = retryMaps[labelTrace].Get(dataAddress)
	require.False(t, ok)

	_ = server.Shutdown(ctx)
}

func simpleTraces(count int) ptrace.Traces {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	rs.SetSchemaUrl("https://opentelemetry.io/schemas/1.4.0")
	rs.Resource().SetDroppedAttributesCount(10)
	rs.Resource().Attributes().PutStr("service.name", "test-service")
	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName("io.opentelemetry.contrib.doris")
	ss.Scope().SetVersion("1.0.0")
	ss.SetSchemaUrl("https://opentelemetry.io/schemas/1.7.0")
	ss.Scope().SetDroppedAttributesCount(20)
	ss.Scope().Attributes().PutStr("lib", "doris")
	timestamp := time.Now()
	for i := 0; i < count; i++ {
		s := ss.Spans().AppendEmpty()
		s.SetTraceID([16]byte{1, 2, 3, byte(i)})
		s.SetSpanID([8]byte{1, 2, 3, byte(i)})
		s.TraceState().FromRaw("trace state")
		s.SetParentSpanID([8]byte{1, 2, 4, byte(i)})
		s.SetName("call db")
		s.SetKind(ptrace.SpanKindInternal)
		s.SetStartTimestamp(pcommon.NewTimestampFromTime(timestamp))
		s.SetEndTimestamp(pcommon.NewTimestampFromTime(timestamp.Add(time.Minute)))
		s.Attributes().PutStr(semconv.AttributeServiceName, "v")
		s.Status().SetMessage("error")
		s.Status().SetCode(ptrace.StatusCodeError)
		event := s.Events().AppendEmpty()
		event.SetName("event1")
		event.SetTimestamp(pcommon.NewTimestampFromTime(timestamp))
		event.Attributes().PutStr("level", "info")
		link := s.Links().AppendEmpty()
		link.SetTraceID([16]byte{1, 2, 5, byte(i)})
		link.SetSpanID([8]byte{1, 2, 5, byte(i)})
		link.TraceState().FromRaw("error")
		link.Attributes().PutStr("k", "v")
	}
	return traces
}
