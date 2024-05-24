package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/file"
	"github.com/fsgonz/otelnetstatsreceiver/internal/lumberjack"
	"github.com/fsgonz/otelnetstatsreceiver/internal/stats/sampler"
	"github.com/fsgonz/otelnetstatsreceiver/internal/stats/scraper"
	"github.com/google/uuid"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	LAST_COUNT_KEY          = "LAST_COUNT"
	FORMAT                  = "v1"
	SCHEMA_ID               = "schema_id"
	NETWORK_SCHEMA_ID       = "network_schema_id"
	FILE_LOGGER_OUTPUT      = "file_logger"
	PIPELINE_EMITTER_OUTPUT = "pipeline_emitter"
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

type SamplerEmitter interface {
	Emit(context.Context)
}

type FileLoggerSamplerEmitter struct {
	URI           string
	metricsLogger log.Logger
	persister     operator.Persister
	sampler       sampler.Sampler
}

func (e FileLoggerSamplerEmitter) Emit(ctx context.Context) {
	jsonEntry := logEntry(ctx, e.persister, e.sampler)
	e.metricsLogger.Println(string(jsonEntry))

}

type PipelineConsumerSamplerEmitter struct {
	Emitter   helper.LogEmitter
	persister operator.Persister
	sampler   sampler.Sampler
	input     file.Input
}

func (e PipelineConsumerSamplerEmitter) Emit(ctx context.Context) {
	jsonEntry := logEntry(ctx, e.persister, e.sampler)
	e.input.Emit(ctx, jsonEntry, map[string]any{})
}

func SamplerEmitterFactory(output string, uri string, persister operator.Persister, emitter *helper.LogEmitter, input file.Input) (SamplerEmitter, error) {
	fileBasedSampler := sampler.NewFileBasedSampler("/proc/net/dev", scraper.NewLinuxNetworkDevicesFileScraper())

	switch output {
	case FILE_LOGGER_OUTPUT:
		metricsLogger := log.New(&lumberjack.Logger{
			Filename:   uri,
			MaxSize:    100, // kilobytes
			MaxBackups: 20,
		}, "", 0)

		return FileLoggerSamplerEmitter{
			output,
			*metricsLogger,
			persister,
			fileBasedSampler,
		}, nil
	case PIPELINE_EMITTER_OUTPUT:
		return PipelineConsumerSamplerEmitter{
			*emitter,
			persister,
			fileBasedSampler,
			input,
		}, nil
	default:
		return nil, fmt.Errorf("unknown output type: %s", output)
	}
}

func logEntry(ctx context.Context, persister operator.Persister, sampler sampler.Sampler) []byte {
	byteSlice, _ := persister.Get(ctx, LAST_COUNT_KEY)

	var last_count uint64 = 0

	if byteSlice != nil {
		// Parse the string to an integer
		counter, _ := strconv.ParseUint(string(byteSlice), 10, 64)
		last_count = counter
	}

	samp, _ := sampler.Sample()

	persister.Set(ctx, LAST_COUNT_KEY, []byte(strconv.FormatUint(samp, 10)))

	orgID := os.Getenv("ORG_ID")
	envID := os.Getenv("ENV_ID")
	deploymentID := os.Getenv("DEPLOYMENT_ID")
	rootOrgID := os.Getenv("ROOT_ORG_ID")
	billingEnabled := os.Getenv("MULE_BILLING_ENABLED") == "true"
	workerID := "worker-" + strings.ReplaceAll(os.Getenv("POD_NAME"), os.Getenv("APP_NAME")+"-", "")
	ts := time.Now().Unix() * 1000

	u, _ := uuid.NewRandom()

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

	logEntry := networkIOLogEntry{
		Format: FORMAT,
		Time:   ts,
		Events: []networkIOLogEntryEvent{evt},
		Metadata: map[string]string{
			SCHEMA_ID: NETWORK_SCHEMA_ID,
		},
	}

	jsonEntry, _ := json.Marshal(logEntry)
	return jsonEntry
}
