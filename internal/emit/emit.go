package emit

import (
	"context"
)

type Callback func(ctx context.Context) error
