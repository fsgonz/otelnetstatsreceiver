// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package adapter // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"

import (
	"github.com/fsgonz/otelnetstatsreceiver/internal/consumerretry"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.opentelemetry.io/collector/component"
)

// BaseConfig is the common configuration of a stanza-based receiver
type BaseConfig struct {
	Operators      []operator.Config    `mapstructure:"operators"`
	StorageID      *component.ID        `mapstructure:"storage"`
	RetryOnFailure consumerretry.Config `mapstructure:"retry_on_failure"`

	// currently not configurable by users, but available for benchmarking
	numWorkers    int
	maxBatchSize  uint
	flushInterval time.Duration
}
