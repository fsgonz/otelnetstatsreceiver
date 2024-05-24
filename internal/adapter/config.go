// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package adapter // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"

import (
	"github.com/fsgonz/otelnetstatsreceiver/internal/consumerretry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.opentelemetry.io/collector/component"
)

const (
	operatorType = "net_stats_input"
)

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewFileInputConfig() })
}

// NewConfig creates a new input config with default values
func NewFileInputConfig() *FileInputConfig {
	return NewFileInputConfigWithID(operatorType)
}

// NewConfigWithID creates a new input config with default values
func NewFileInputConfigWithID(operatorID string) *FileInputConfig {
	return &FileInputConfig{
		InputConfig: helper.NewInputConfig(operatorID, operatorType),
		Config:      *fileconsumer.NewConfig(),
	}
}

// Build will build a netstats input operator from the supplied configuration
func (c FileInputConfig) Build(set component.TelemetrySettings) (operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(set)

	if err != nil {
		return nil, err
	}

	input := &Input{
		InputOperator: inputOperator,
	}

	input.consumer, err = c.Config.Build(set, input.emit)
	if err != nil {
		return nil, err
	}

	return input, nil
}

type FileInputConfig struct {
	helper.InputConfig  `mapstructure:",squash"`
	fileconsumer.Config `mapstructure:",squash"`
}

// BaseConfig is the common configuration of a stanza-based receiver
type BaseConfig struct {
	Operators                     []operator.Config    `mapstructure:"operators"`
	StorageID                     *component.ID        `mapstructure:"storage"`
	RetryOnFailure                consumerretry.Config `mapstructure:"retry_on_failure"`
	MetricsGenerationPollInterval time.Duration        `mapstructure:"metrics_generation_poll_interval,omitempty"`
	MetricsOutputFile             string               `mapstructure:"metrics_output_file,omitempty"`

	// currently not configurable by users, but available for benchmarking
	numWorkers    int
	maxBatchSize  uint
	flushInterval time.Duration
}

type LogSamplerConfig struct {
	LogSamplers []LogSampler `mapstructure:"log_samplers"`
}

type LogSampler struct {
	Metric string `mapstructure:"metric"`
	Output string `mapstructure:"output"`
	URI    string `mapstructure:"uri"`
}

func (cfg *LogSamplerConfig) Validate() error {
	return nil
}
