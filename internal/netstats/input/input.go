package netstats

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

type networkIOLogEntry struct {
	// Format is the schema version
	Format string `json:"format"`
	// Time is the time this entry was created in unix epoch milliseconds
	Time     int64                    `json:"time"`
	Events   []networkIOLogEntryEvent `json:"events"`
	Metadata map[string]string        `json:"metadata"`
}

type networkIOLogEntryEvent struct {
	ID string `json:"id"`
	// Timestamp is the time this entry was created in unix epoch milliseconds
	Timestamp  int64  `json:"timestamp"`
	RootOrgID  string `json:"root_org_id"`
	OrgID      string `json:"org_id"`
	EnvID      string `json:"env_id"`
	AssetID    string `json:"asset_id"`
	WorkerID   string `json:"worker_id"`
	UsageBytes uint64 `json:"usage_bytes"`
	Billable   bool   `json:"billable"`
}

type Input struct {
	helper.InputOperator
	consumer *fileconsumer.Manager
}

func (i *Input) Start(persister operator.Persister) error {
	return i.consumer.Start(persister)
}

// Stop will stop the file monitoring process
func (i *Input) Stop() error {
	return i.consumer.Stop()
}

func (i *Input) emit(ctx context.Context, token []byte, attrs map[string]any) error {
	if len(token) == 0 {
		return nil
	}

	ent, err := i.NewEntry(string(token))
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
