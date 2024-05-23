package netstats

import (
	"context"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/statsconsumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"strconv"
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
	byteSlice, err := persister.Get(ctx, "counter")

	var counter2 = 0

	if byteSlice != nil {
		// Parse the string to an integer
		counter, err := strconv.Atoi(string(byteSlice))
		counter2 = counter
		if err != nil {
			// Handle the error if the conversion fails
			fmt.Println("Error:", err)
		} else {
			// Successfully converted to an integer
			fmt.Println("The integer value is:", counter)
		}
	}

	ent, err := i.NewEntry(counter2)
	if err != nil {
		return fmt.Errorf("create entry: %w", err)
	}
	i.Write(ctx, ent)
	persister.Set(ctx, "counter", []byte(strconv.Itoa(counter2)))
	return nil
}
