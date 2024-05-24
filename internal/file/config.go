package file

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.opentelemetry.io/collector/component"
)

const (
	operatorType = "file_input"
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

// Build will build a stats input operator from the supplied configuration
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
