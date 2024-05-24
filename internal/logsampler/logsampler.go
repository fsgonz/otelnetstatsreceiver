package logsampler

import "time"

type Config struct {
	LogSamplers []LogSampler `mapstructure:"log_samplers"`
}

type LogSampler struct {
	Metric       string        `mapstructure:"metric"`
	Output       string        `mapstructure:"output"`
	URI          string        `mapstructure:"uri"`
	PollInterval time.Duration `mapstructure:"poll_interval,omitempty"`
}

func (cfg *Config) Validate() error {
	return nil
}
