package sampler

import (
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/scraper"
	"log"
	"os"
)

// Sampler allows to sample network usage. It returns the sum of received and transmitted bytes
type Sampler interface {
	Sample() (uint64, error)
}

type SamplerStorage interface {
}

type FileBasedDeltaSampler struct {
	FileBasedSampler FileBasedSampler
	Storage          SamplerStorage
}

// An implementation of the Sampler based on a file
type FileBasedSampler struct {
	uri     string
	scraper scraper.NetworkStatsScraper
}

// Will sample from a file
func (s *FileBasedSampler) Sample() (uint64, error) {
	f, err := os.Open(s.uri)
	if err != nil {
		return 0, err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("Error on closing the stats file.")
		}
	}(f)

	networkUsageStats, err := s.scraper.Scrape(f)

	if err != nil {
		return 0, err
	}

	netIo := networkUsageStats.ReceivedBytes + networkUsageStats.TransmittedBytes

	return netIo, nil
}
