package netstats

import (
	"context"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/statsconsumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
)

type Input struct {
	helper.InputOperator
	consumer *statsconsumer.Manager
}

func (i *Input) Start(persister operator.Persister) error {
	return i.consumer.Start(persister)
}

// Stop will stop the file monitoring process
func (i *Input) Stop() error {
	return i.consumer.Stop()
}

func (i *Input) emit(ctx context.Context, persister operator.Persister) error {
	ent, err := i.NewEntry("hola")
	if err != nil {
		return fmt.Errorf("create entry: %w", err)
	}
	i.Write(ctx, ent)
	return nil
}
