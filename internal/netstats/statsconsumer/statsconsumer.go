package statsconsumer

import (
	"context"
	"github.com/fsgonz/otelnetstatsreceiver/internal/emit"
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

	pollInterval time.Duration
	emit         emit.Callback
}

func (m *Manager) Start(persister operator.Persister) error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	// Start polling goroutine
	m.startPoller(ctx, persister)

	return nil
}

// startPoller kicks off a goroutine that will poll for net stats periodically.
func (m *Manager) startPoller(ctx context.Context, persister operator.Persister) {
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

			m.poll(ctx, persister)
		}
	}()
}

// poll checks all the watched paths for new entries
func (m *Manager) poll(ctx context.Context, persister operator.Persister) {
	err := m.emit(ctx, persister)
	if err != nil {
		m.set.Logger.Debug("Error on consuming stats", zap.Error(err))
		return
	}

}

func (m *Manager) Stop() error {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.wg.Wait()
	return nil
}

type Config struct {
	PollInterval time.Duration `mapstructure:"poll_interval,omitempty"`
}

func (c Config) Build(set component.TelemetrySettings, emit emit.Callback) (*Manager, error) {
	return &Manager{
		set:          set,
		pollInterval: c.PollInterval,
		emit:         emit,
	}, nil
}
