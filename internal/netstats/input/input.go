package netstats

import (
	"context"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/sampler"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/scraper"
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

	ent, err := i.NewEntry(samp - last_count)
	if err != nil {
		return fmt.Errorf("create entry: %w", err)
	}
	i.Write(ctx, ent)
	last_count++
	persister.Set(ctx, "counter", []byte(strconv.FormatUint(last_count, 10)))
	return nil
}
