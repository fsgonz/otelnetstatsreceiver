package statsconsumer

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Manager struct {
	// Deprecated [v0.101.0]
	*zap.SugaredLogger

	set    component.TelemetrySettings
	wg     sync.WaitGroup
	cancel context.CancelFunc

	pollInterval  time.Duration
	fromBeginning bool
}

func (m *Manager) Start(persister operator.Persister) error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	// Start polling goroutine
	m.startPoller(ctx)

	return nil
}

// startPoller kicks off a goroutine that will poll for net stats periodically.
func (m *Manager) startPoller(ctx context.Context) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		globTicker := time.NewTicker(m.pollInterval)
		defer globTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-globTicker.C:
			}

			m.poll(ctx)
		}
	}()
}

// poll checks all the watched paths for new entries
func (m *Manager) poll(ctx context.Context) {
	m.set.Logger.Debug("Consuming stats")
}

type Config struct {
	PollInterval time.Duration `mapstructure:"poll_interval,omitempty"`
	StartAt      string        `mapstructure:"start_at,omitempty"`
}

func (c Config) Build(set component.TelemetrySettings, emit interface{}) (*Manager, error) {
	var startAtBeginning bool

	switch c.StartAt {
	case "beginning":
		startAtBeginning = true
	case "end":
		startAtBeginning = false
	default:
		return nil, fmt.Errorf("invalid start_at location '%s'", c.StartAt)
	}

	return &Manager{
		set:           set,
		pollInterval:  c.PollInterval,
		fromBeginning: startAtBeginning,
	}, nil
}
