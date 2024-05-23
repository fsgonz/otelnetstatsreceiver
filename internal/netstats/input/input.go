package netstats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/sampler"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/scraper"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/statsconsumer"
	"github.com/google/uuid"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"strconv"
	"time"
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
	byteSlice, err := persister.Get(ctx, "last_count")

	var last_count uint64 = 0

	if byteSlice != nil {
		// Parse the string to an integer
		counter, err := strconv.ParseUint(string(byteSlice), 10, 64)
		last_count = counter
		if err != nil {
			i.Logger().Error("Error")
		}
	}

	basedSampler := sampler.NewFileBasedSampler("/Users/fabian.gonzalez/logs", scraper.NewLinuxNetworkDevicesFileScraper())

	samp, _ := basedSampler.Sample()

	orgID := "org_id"               // os.Getenv("ORG_ID")
	envID := "env_id"               // os.Getenv("ENV_ID")
	deploymentID := "deployment_id" // os.Getenv("DEPLOYMENT_ID")
	rootOrgID := "root_org_id"      // os.Getenv("ROOT_ORG_ID")
	billingEnabled := true          // os.Getenv("MULE_BILLING_ENABLED") == "true"
	workerID := "worker-"           // + strings.ReplaceAll(os.Getenv("POD_NAME"), os.Getenv("APP_NAME")+"-", "")
	ts := time.Now().Unix() * 1000

	u, err := uuid.NewRandom()

	evt := networkIOLogEntryEvent{
		ID:         u.String(),
		Timestamp:  ts,
		RootOrgID:  rootOrgID,
		OrgID:      orgID,
		EnvID:      envID,
		AssetID:    deploymentID,
		WorkerID:   workerID,
		UsageBytes: samp - last_count,
		Billable:   billingEnabled,
	}

	e := networkIOLogEntry{
		Format: "v1",
		Time:   ts,
		Events: []networkIOLogEntryEvent{evt},
		Metadata: map[string]string{
			"schema_id": "network_schema_id",
		},
	}

	b, err := json.Marshal(e)

	ent, err := i.NewEntry(string(b))
	if err != nil {
		return fmt.Errorf("create entry: %w", err)
	}
	i.Write(ctx, ent)
	last_count++
	persister.Set(ctx, "last_count", []byte(strconv.FormatUint(samp, 10)))
	return nil
}
