// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package component // import "go.opentelemetry.io/collector/component"

import (
	"context"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
)

// Exporter exports telemetry data from the collector to a destination.
type Exporter interface {
	Component
}

// TracesExporter is an Exporter that can consume traces.
type TracesExporter interface {
	Exporter
	consumer.Traces
}

// MetricsExporter is an Exporter that can consume metrics.
type MetricsExporter interface {
	Exporter
	consumer.Metrics
}

// LogsExporter is an Exporter that can consume logs.
type LogsExporter interface {
	Exporter
	consumer.Logs
}

// ExporterCreateSettings configures Exporter creators.
type ExporterCreateSettings struct {
	TelemetrySettings

	// BuildInfo can be used by components for informational purposes
	BuildInfo BuildInfo
}

// ExporterFactory is factory interface for exporters.
//
// This interface cannot be directly implemented. Implementations must
// use the NewExporterFactory to implement it.
type ExporterFactory interface {
	Factory

	// CreateDefaultConfig creates the default configuration for the Exporter.
	// This method can be called multiple times depending on the pipeline
	// configuration and should not cause side-effects that prevent the creation
	// of multiple instances of the Exporter.
	// The object returned by this method needs to pass the checks implemented by
	// 'configtest.CheckConfigStruct'. It is recommended to have these checks in the
	// tests of any implementation of the Factory interface.
	CreateDefaultConfig() config.Exporter

	// CreateTracesExporter creates a TracesExporter based on this config.
	// If the exporter type does not support tracing or if the config is not valid,
	// an error will be returned instead.
	CreateTracesExporter(ctx context.Context, set ExporterCreateSettings, cfg config.Exporter) (TracesExporter, error)

	// TracesExporterStability gets the stability level of the TracesExporter.
	TracesExporterStability() StabilityLevel

	// CreateMetricsExporter creates a MetricsExporter based on this config.
	// If the exporter type does not support metrics or if the config is not valid,
	// an error will be returned instead.
	CreateMetricsExporter(ctx context.Context, set ExporterCreateSettings, cfg config.Exporter) (MetricsExporter, error)

	// MetricsExporterStability gets the stability level of the MetricsExporter.
	MetricsExporterStability() StabilityLevel

	// CreateLogsExporter creates a LogsExporter based on the config.
	// If the exporter type does not support logs or if the config is not valid,
	// an error will be returned instead.
	CreateLogsExporter(ctx context.Context, set ExporterCreateSettings, cfg config.Exporter) (LogsExporter, error)

	// LogsExporterStability gets the stability level of the LogsExporter.
	LogsExporterStability() StabilityLevel
}

// ExporterFactoryOption apply changes to ExporterOptions.
type ExporterFactoryOption interface {
	// applyExporterFactoryOption applies the option.
	applyExporterFactoryOption(o *exporterFactory)
}

var _ ExporterFactoryOption = (*exporterFactoryOptionFunc)(nil)

// exporterFactoryOptionFunc is an ExporterFactoryOption created through a function.
type exporterFactoryOptionFunc func(*exporterFactory)

func (f exporterFactoryOptionFunc) applyExporterFactoryOption(o *exporterFactory) {
	f(o)
}

// ExporterCreateDefaultConfigFunc is the equivalent of ExporterFactory.CreateDefaultConfig().
type ExporterCreateDefaultConfigFunc func() config.Exporter

// CreateDefaultConfig implements ExporterFactory.CreateDefaultConfig().
func (f ExporterCreateDefaultConfigFunc) CreateDefaultConfig() config.Exporter {
	return f()
}

// CreateTracesExporterFunc is the equivalent of ExporterFactory.CreateTracesExporter().
type CreateTracesExporterFunc func(context.Context, ExporterCreateSettings, config.Exporter) (TracesExporter, error)

// CreateTracesExporter implements ExporterFactory.CreateTracesExporter().
func (f CreateTracesExporterFunc) CreateTracesExporter(ctx context.Context, set ExporterCreateSettings, cfg config.Exporter) (TracesExporter, error) {
	if f == nil {
		return nil, ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg)
}

// CreateMetricsExporterFunc is the equivalent of ExporterFactory.CreateMetricsExporter().
type CreateMetricsExporterFunc func(context.Context, ExporterCreateSettings, config.Exporter) (MetricsExporter, error)

// CreateMetricsExporter implements ExporterFactory.CreateMetricsExporter().
func (f CreateMetricsExporterFunc) CreateMetricsExporter(ctx context.Context, set ExporterCreateSettings, cfg config.Exporter) (MetricsExporter, error) {
	if f == nil {
		return nil, ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg)
}

// CreateLogsExporterFunc is the equivalent of ExporterFactory.CreateLogsExporter().
type CreateLogsExporterFunc func(context.Context, ExporterCreateSettings, config.Exporter) (LogsExporter, error)

// CreateLogsExporter implements ExporterFactory.CreateLogsExporter().
func (f CreateLogsExporterFunc) CreateLogsExporter(ctx context.Context, set ExporterCreateSettings, cfg config.Exporter) (LogsExporter, error) {
	if f == nil {
		return nil, ErrDataTypeIsNotSupported
	}
	return f(ctx, set, cfg)
}

type exporterFactory struct {
	baseFactory
	ExporterCreateDefaultConfigFunc
	CreateTracesExporterFunc
	CreateMetricsExporterFunc
	CreateLogsExporterFunc
}

func (e exporterFactory) TracesExporterStability() StabilityLevel {
	return e.getStabilityLevel(config.TracesDataType)
}

func (e exporterFactory) MetricsExporterStability() StabilityLevel {
	return e.getStabilityLevel(config.MetricsDataType)
}

func (e exporterFactory) LogsExporterStability() StabilityLevel {
	return e.getStabilityLevel(config.LogsDataType)
}

// WithTracesExporter overrides the default "error not supported" implementation for CreateTracesExporter and the default "undefined" stability level.
func WithTracesExporter(createTracesExporter CreateTracesExporterFunc, sl StabilityLevel) ExporterFactoryOption {
	return exporterFactoryOptionFunc(func(o *exporterFactory) {
		o.stability[config.TracesDataType] = sl
		o.CreateTracesExporterFunc = createTracesExporter
	})
}

// WithMetricsExporter overrides the default "error not supported" implementation for CreateMetricsExporter and the default "undefined" stability level.
func WithMetricsExporter(createMetricsExporter CreateMetricsExporterFunc, sl StabilityLevel) ExporterFactoryOption {
	return exporterFactoryOptionFunc(func(o *exporterFactory) {
		o.stability[config.MetricsDataType] = sl
		o.CreateMetricsExporterFunc = createMetricsExporter
	})
}

// WithLogsExporter overrides the default "error not supported" implementation for CreateLogsExporter and the default "undefined" stability level.
func WithLogsExporter(createLogsExporter CreateLogsExporterFunc, sl StabilityLevel) ExporterFactoryOption {
	return exporterFactoryOptionFunc(func(o *exporterFactory) {
		o.stability[config.LogsDataType] = sl
		o.CreateLogsExporterFunc = createLogsExporter
	})
}

// NewExporterFactory returns a ExporterFactory.
func NewExporterFactory(cfgType config.Type, createDefaultConfig ExporterCreateDefaultConfigFunc, options ...ExporterFactoryOption) ExporterFactory {
	f := &exporterFactory{
		baseFactory:                     baseFactory{cfgType: cfgType, stability: make(map[config.DataType]StabilityLevel)},
		ExporterCreateDefaultConfigFunc: createDefaultConfig,
	}
	for _, opt := range options {
		opt.applyExporterFactoryOption(f)
	}
	return f
}
