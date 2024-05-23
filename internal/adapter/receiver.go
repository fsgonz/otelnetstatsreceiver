// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package adapter // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/sampler"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/scraper"
	"github.com/google/uuid"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"log"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/extension/experimental/storage"
	rcvr "go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/pipeline"
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

type receiver struct {
	set                 component.TelemetrySettings
	metricsPollInternal time.Duration
	MetricsOutputFile   string
	id                  component.ID
	wg                  sync.WaitGroup
	cancel              context.CancelFunc

	pipe      pipeline.Pipeline
	emitter   *helper.LogEmitter
	consumer  consumer.Logs
	converter *Converter
	obsrecv   *receiverhelper.ObsReport

	storageID     *component.ID
	storageClient storage.Client
}

// Ensure this receiver adheres to required interface
var _ rcvr.Logs = (*receiver)(nil)

// Start tells the receiver to start
func (r *receiver) Start(ctx context.Context, host component.Host) error {
	rctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.set.Logger.Info("Starting stanza receiver")

	if err := r.setStorageClient(ctx, host); err != nil {
		return fmt.Errorf("storage client: %w", err)
	}

	if err := r.pipe.Start(r.storageClient); err != nil {
		return fmt.Errorf("start stanza: %w", err)
	}

	r.converter.Start()

	// Below we're starting 2 loops:
	// * one which reads all the logs produced by the emitter and then forwards
	//   them to converter
	// ...
	r.wg.Add(1)
	go r.emitterLoop(rctx)

	// ...
	// * second one which reads all the logs produced by the converter
	//   (aggregated by Resource) and then calls consumer to consumer them.
	r.wg.Add(1)
	go r.consumerLoop(rctx)

	// Those 2 loops are started in separate goroutines because batching in
	// the emitter loop can cause a flush, caused by either reaching the max
	// flush size or by the configurable ticker which would in turn cause
	// a set of log entries to be available for reading in converter's out
	// channel. In order to prevent backpressure, reading from the converter
	// channel and batching are done in those 2 goroutines.

	go r.startMetricsGeneration(rctx, r.storageClient)

	return nil
}

// emitterLoop reads the log entries produced by the emitter and batches them
// in converter.
func (r *receiver) emitterLoop(ctx context.Context) {
	defer r.wg.Done()

	// Don't create done channel on every iteration.
	doneChan := ctx.Done()
	for {
		select {
		case <-doneChan:
			r.set.Logger.Debug("Receive loop stopped")
			return

		case e, ok := <-r.emitter.OutChannel():
			if !ok {
				continue
			}

			if err := r.converter.Batch(e); err != nil {
				r.set.Logger.Error("Could not add entry to batch", zap.Error(err))
			}
		}
	}
}

// consumerLoop reads converter log entries and calls the consumer to consumer them.
func (r *receiver) consumerLoop(ctx context.Context) {
	defer r.wg.Done()

	// Don't create done channel on every iteration.
	doneChan := ctx.Done()
	pLogsChan := r.converter.OutChannel()
	for {
		select {
		case <-doneChan:
			r.set.Logger.Debug("Consumer loop stopped")
			return

		case pLogs, ok := <-pLogsChan:
			if !ok {
				r.set.Logger.Debug("Converter channel got closed")
				continue
			}
			obsrecvCtx := r.obsrecv.StartLogsOp(ctx)
			logRecordCount := pLogs.LogRecordCount()
			cErr := r.consumer.ConsumeLogs(ctx, pLogs)
			if cErr != nil {
				r.set.Logger.Error("ConsumeLogs() failed", zap.Error(cErr))
			}
			r.obsrecv.EndLogsOp(obsrecvCtx, "stanza", logRecordCount, cErr)
		}
	}
}

// Shutdown is invoked during service shutdown
func (r *receiver) Shutdown(ctx context.Context) error {
	if r.cancel == nil {
		return nil
	}

	r.set.Logger.Info("Stopping stanza receiver")
	pipelineErr := r.pipe.Stop()
	r.converter.Stop()
	r.cancel()
	r.wg.Wait()

	if r.storageClient != nil {
		clientErr := r.storageClient.Close(ctx)
		return multierr.Combine(pipelineErr, clientErr)
	}
	return pipelineErr
}

func (r *receiver) startMetricsGeneration(ctx context.Context, persister operator.Persister) {
	metricsLogger := log.New(&Logger{
		Filename:   r.MetricsOutputFile,
		MaxSize:    100, // kilobytes
		MaxBackups: 20,
	}, "", 0)

	ticker := time.NewTicker(r.metricsPollInternal)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			persistMetrics(metricsLogger, ctx, persister)
		case <-ctx.Done():
			return
		}
	}
}

func persistMetrics(logger *log.Logger, ctx context.Context, persister operator.Persister) {
	byteSlice, _ := persister.Get(ctx, "last_count")

	var last_count uint64 = 0

	if byteSlice != nil {
		// Parse the string to an integer
		counter, err := strconv.ParseUint(string(byteSlice), 10, 64)
		last_count = counter
		if err != nil {

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

	if err != nil {

	}

	logger.Println(string(b))

}
