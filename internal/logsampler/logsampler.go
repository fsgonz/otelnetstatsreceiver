package logsampler

import (
	"time"
)

type Config struct {
	LogSamplers []LogSampler `mapstructure:"log_samplers"`
}

type LogSamplerError struct {
	Msg string
}

func (e *LogSamplerError) Error() string {
	return e.Msg
}

type LogSampler struct {
	Metric       string        `mapstructure:"metric"`
	Output       string        `mapstructure:"output"`
	URI          string        `mapstructure:"uri"`
	PollInterval time.Duration `mapstructure:"poll_interval,omitempty"`
}

func (cfg *Config) Validate() error {
	if len(cfg.LogSamplers) > 1 {
		return &LogSamplerError{"No more than one sampler supported in this version"}
	}

	if len(cfg.LogSamplers) == 1 {
		logSampler := cfg.LogSamplers[0]

		if logSampler.Metric != "netstats" {
			return &LogSamplerError{"Incorrect metric in sampler. Possible Values: [netstats]"}
		}
		switch logSampler.Output {
		case "file_logger", "pipeline_emitter":
			break
		default:
			return &LogSamplerError{"Incorrect output in sampler. Possible Values: [file_logger, pipeline_emitter]"}
		}
	}
	return nil
}
