// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/component"
)

var (
	Type      = component.MustNewType("signaltometrics")
	ScopeName = "github.com/open-telemetry/opentelemetry-collector-contrib/connector/signaltometricsconnector"
)

const (
	TracesToMetricsStability  = component.StabilityLevelDevelopment
	LogsToMetricsStability    = component.StabilityLevelDevelopment
	MetricsToMetricsStability = component.StabilityLevelDevelopment
)