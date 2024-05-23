package netstats

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.opentelemetry.io/collector/component"
)

const operatorType = "net_stats_input"

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewConfig() })
}

// NewConfig creates a new input config with default values
func NewConfig() *Config {
	return NewConfigWithID(operatorType)
}

// NewConfigWithID creates a new input config with default values
func NewConfigWithID(operatorID string) *Config {
	return &Config{
		InputConfig: helper.NewInputConfig(operatorID, operatorType),
		Config:      *fileconsumer.NewConfig(),
	}
}

type Config struct {
	helper.InputConfig  `mapstructure:",squash"`
	fileconsumer.Config `mapstructure:",squash"`
}

// Build will build a netstats input operator from the supplied configuration
func (c Config) Build(set component.TelemetrySettings) (operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(set)

	if err != nil {
		return nil, err
	}

	input := &Input{
		InputOperator: inputOperator,
	}

	c.Config.Include = append(c.Config.Include, "/Users/fabian.gonzalez/utyman.log")
	input.consumer, err = c.Config.Build(set, input.emit)
	if err != nil {
		return nil, err
	}

	return input, nil
}
