package statsconsumer

import (
	"context"
	"github.com/fsgonz/otelnetstatsreceiver/internal/emit"
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

func (m *Manager) Start() error {
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
	err := m.emit(ctx)
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
