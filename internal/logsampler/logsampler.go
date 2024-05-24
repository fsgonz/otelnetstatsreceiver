package logsampler

type Config struct {
	LogSamplers []LogSampler `mapstructure:"log_samplers"`
}

type LogSampler struct {
	Metric string `mapstructure:"metric"`
	Output string `mapstructure:"output"`
	URI    string `mapstructure:"uri"`
}

func (cfg *Config) Validate() error {
	return nil
}
