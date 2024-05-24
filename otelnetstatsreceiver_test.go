package otelnetstatsreceiver

import (
	"github.com/fsgonz/otelnetstatsreceiver/internal/adapter"
	"github.com/fsgonz/otelnetstatsreceiver/internal/consumerretry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/parser/regex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	require.NotNil(t, cfg, "failed to create default config")
	require.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestLoadConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	sub, err := cm.Sub(component.MustNewID("filelog").String())
	require.NoError(t, err)
	require.NoError(t, component.UnmarshalConfig(sub, cfg))

	assert.NoError(t, component.ValidateConfig(cfg))
	assert.Equal(t, testdataConfigYaml(), cfg)
}

func testdataConfigYaml() *OtelNetStatsReceiverConfig {
	return &OtelNetStatsReceiverConfig{
		BaseConfig: adapter.BaseConfig{
			Operators: []operator.Config{
				{
					Builder: func() *regex.Config {
						cfg := regex.NewConfig()
						cfg.Regex = "^(?P<time>\\d{4}-\\d{2}-\\d{2}) (?P<sev>[A-Z]*) (?P<msg>.*)$"
						sevField := entry.NewAttributeField("sev")
						sevCfg := helper.NewSeverityConfig()
						sevCfg.ParseFrom = &sevField
						cfg.SeverityConfig = &sevCfg
						timeField := entry.NewAttributeField("time")
						timeCfg := helper.NewTimeParser()
						timeCfg.Layout = "%Y-%m-%d"
						timeCfg.ParseFrom = &timeField
						cfg.TimeParser = &timeCfg
						return cfg
					}(),
				},
			},
			RetryOnFailure: consumerretry.Config{
				Enabled:         false,
				InitialInterval: 1 * time.Second,
				MaxInterval:     30 * time.Second,
				MaxElapsedTime:  5 * time.Minute,
			},
		},
		InputConfig: adapter.FileInputConfig(func() file.Config {
			c := file.NewConfig()
			c.Include = []string{"testdata/simple.log"}
			c.StartAt = "beginning"
			return *c
		}()),
	}
}
