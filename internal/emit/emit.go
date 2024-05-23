package emit

import (
	"context"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
)

type Callback func(ctx context.Context, persister operator.Persister) error
