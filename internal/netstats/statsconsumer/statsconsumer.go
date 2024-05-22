package statsconsumer

import (
	"time"
)

type Config struct {
	PollInterval time.Duration `mapstructure:"poll_interval,omitempty"`
	StartAt      string        `mapstructure:"start_at,omitempty"`
}
