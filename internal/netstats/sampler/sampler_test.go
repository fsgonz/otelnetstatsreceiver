package sampler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/scraper"
	"io"
	"strconv"
	"testing"
)

func TestFileBasedSampler(t *testing.T) {
	t.Run("retrieves the sum of the received and transmit from a file.", func(t *testing.T) {
		sampler := FileBasedSampler{
			uri:     "testdata/test1.data",
			scraper: &BreakLineScraper{},
		}
		want := uint64(1030)

		got, err := sampler.Sample()

		if err != nil {
			t.Errorf("Error on sampling %s", err.Error())
		}
		if got != want {
			t.Errorf("got %d want %d", got, want)
		}
	})

	t.Run("when a file does not exists an error is raised", func(t *testing.T) {
		const wanted = "open nonExistingFile.data: no such file or directory"

		sampler := FileBasedSampler{
			uri:     "nonExistingFile.data",
			scraper: &BreakLineScraper{},
		}

		_, err := sampler.Sample()

		if err == nil {
			t.Errorf("An error was expected but err was nil")
		}

		errorMessage := err.Error()

		if errorMessage != wanted {
			t.Errorf("Expected error was '%s' but it was '%s'", wanted, errorMessage)
		}
	})

	t.Run("when the scraper fails, the error is propagated to the sampler", func(t *testing.T) {
		alwaysFailScraper := &AlwaysFailScraper{Error: "Test error"}

		sampler := FileBasedSampler{
			uri:     "testdata/test1.data",
			scraper: alwaysFailScraper,
		}

		_, err := sampler.Sample()

		if err == nil {
			t.Errorf("An error was expected but err was nil")
		}

		errorMessage := err.Error()

		if errorMessage != alwaysFailScraper.Error {
			t.Errorf("Expected error was '%s' but it was '%s'", alwaysFailScraper.Error, errorMessage)
		}

	})
}

type AlwaysFailScraper struct {
	Error string
}

func (s *AlwaysFailScraper) Scrape(io.Reader) (scraper.NetworkStats, error) {
	return scraper.NetworkStats{}, errors.New(s.Error)
}

type BreakLineScraper struct{}

func (s *BreakLineScraper) Scrape(r io.Reader) (scraper.NetworkStats, error) {
	// Read all content from the reader
	content, err := io.ReadAll(r)
	if err != nil {
		return scraper.NetworkStats{}, fmt.Errorf("failed to read content: %w", err)
	}

	// Split the content by newline
	parts := bytes.SplitN(content, []byte("\n"), 2)
	if len(parts) < 2 {
		return scraper.NetworkStats{}, errors.New("expected two parts separated by a newline")
	}

	// Convert the first part to uint64
	firstPart, err := strconv.ParseUint(string(parts[0]), 10, 64)
	if err != nil {
		return scraper.NetworkStats{}, fmt.Errorf("failed to parse first part as uint64: %w", err)
	}

	// Convert the second part to uint64
	secondPart, err := strconv.ParseUint(string(parts[1]), 10, 64)
	if err != nil {
		return scraper.NetworkStats{}, fmt.Errorf("failed to parse second part as uint64: %w", err)
	}

	return scraper.NetworkStats{ReceivedBytes: firstPart, TransmittedBytes: secondPart}, nil
}
