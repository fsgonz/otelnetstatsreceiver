package netstats

import (
	"context"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/statsconsumer"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

type Input struct {
	helper.InputOperator
	consumer *statsconsumer.Manager
}

func (i *Input) emit(ctx context.Context, token []byte, attrs map[string]any) error {
	if len(token) == 0 {
		return nil
	}

	ent, err := i.NewEntry("hola")
	if err != nil {
		return fmt.Errorf("create entry: %w", err)
	}

	for k, v := range attrs {
		if err := ent.Set(entry.NewAttributeField(k), v); err != nil {
			i.Logger().Error("set attribute", zap.Error(err))
		}
	}
	i.Write(ctx, ent)
	return nil
}
