package otelnetstatsreceiver

import (
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	require.NotNil(t, cfg, "failed to create default config")
	require.NoError(t, componenttest.CheckConfigStruct(cfg))
}
