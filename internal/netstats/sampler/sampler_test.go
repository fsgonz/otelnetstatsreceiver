package sampler

import (
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/scraper"

	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func TestFileBasedSampler(t *testing.T) {
	t.Run("retrieves the sum of the received and transmit from a file.", func(t *testing.T) {
		sampler := NewFileBasedSampler("testdata/test1.data", &BreakLineScraper{})

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
		sampler := NewFileBasedSampler("nonExistingFile.data", &BreakLineScraper{})

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
		sampler := NewFileBasedSampler("testdata/test1.data", alwaysFailScraper)

		_, err := sampler.Sample()

		if err == nil {
			t.Errorf("An error was expected but err was nil")
		}

		errorMessage := err.Error()

		if errorMessage != alwaysFailScraper.Error {
			t.Errorf("Expected error was '%s' but it was '%s'", alwaysFailScraper.Error, errorMessage)
		}
	})

	t.Run("when the scraper fails, the error is propagated to the sampler", func(t *testing.T) {
		alwaysFailScraper := &AlwaysFailScraper{Error: "Test error"}
		sampler := NewFileBasedSampler("testdata/test1.data", alwaysFailScraper)

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

func TestFileBasedDeltaSampler(t *testing.T) {
	t.Run("retrieves the delta of the sum of the received and transmit from a file.", func(t *testing.T) {
		tempFile, err := ioutil.TempFile("", "TestFileBasedDeltaSampler-*.txt")

		if err != nil {
			t.Errorf("Error on creating temporary file %s", err.Error())
			return
		}
		tempFilePath := tempFile.Name()
		defer os.Remove(tempFilePath)

		err = addValuesToTempFile(tempFile, 200, 400)

		if err != nil {
			t.Errorf("Error on adding values to temp file %s", err.Error())
			return
		}

		sampler := NewFileBasedDeltaSampler(tempFilePath, &BreakLineScraper{}, &TestStorage{})

		want := uint64(600)

		got, err := sampler.Sample()

		if err != nil {
			t.Errorf("Error on sampling %s", err.Error())
			return
		}
		if got != want {
			t.Errorf("got %d want %d", got, want)
		}

		tempFile, err = os.OpenFile(tempFile.Name(), os.O_WRONLY|os.O_TRUNC, 0666)
		err = addValuesToTempFile(tempFile, 400, 600)

		want = uint64(400)

		got, err = sampler.Sample()

		if err != nil {
			t.Errorf("Error on sampling %s", err.Error())
			return
		}
		if got != want {
			t.Errorf("got %d want %d", got, want)
			return
		}
	})
}

func addValuesToTempFile(tempFile *os.File, readBytes uint64, transmitBytes uint64) error {
	// Write the numbers to the file, each on a new line
	content := fmt.Sprintf("%d\n%d", readBytes, transmitBytes)
	if _, err := tempFile.Write([]byte(content)); err != nil {
		return err
	}

	// Close the file
	if err := tempFile.Close(); err != nil {
		return err
	}
	return nil
}

// Scrapers for testing

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

// Storage for testing

type TestStorage struct {
	LastCount uint64
}

func (s *TestStorage) Load() (uint64, error) {
	return s.LastCount, nil
}

func (s *TestStorage) Save(value uint64) error {
	s.LastCount = value
	return nil
}
