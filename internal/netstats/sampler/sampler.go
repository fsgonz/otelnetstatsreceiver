package sampler

import (
	"github.com/fsgonz/otelnetstatsreceiver/internal/netstats/scraper"
	"log"
	"os"
)

// Sampler is an interface that defines a sampler for a uint64 measurement.
// It is used to sample a measurement for a metric and return the sampled value
// along with any error that occurred during the sampling process.
//
// Methods:
// - Sample: Samples a measurement for a metric.
//
// Example usage:
//
//	type MySampler struct {}
//
//	func (s *MySampler) Sample() (uint64, error) {
//	    // Implement the sampling logic here
//	    var sampleValue uint64 = 42 // Example sample value
//	    return sampleValue, nil     // Return the sample value and no error
//	}
//
//	func main() {
//	    var sampler Sampler = &MySampler{}
//	    value, err := sampler.Sample()
//	    if err != nil {
//	        fmt.Println("Error sampling:", err)
//	        return
//	    }
//	    fmt.Println("Sampled value:", value)
//	}
type Sampler interface {
	// Sample samples a measurement for a metric.
	// Returns:
	// - sampleValue: The sampled uint64 value.
	// - error: An error that occurred during sampling, or nil if no error occurred.
	Sample() (sampleValue uint64, err error)
}

// Storage for measurements.
type Storage interface {
	// Save the sample
	Save(lastSample uint64) error
	// Load the sample
	Load() (uint64, error)
}

// FileBasedDeltaSampler is a struct that manages the sampling of network statistics
// from a file, calculates deltas between samples, and stores the last measurements in a storage backend.
//
// Fields:
//   - FileBasedSampler: An instance of FileBasedSampler which handles the basic sampling
//     operations from a file.
//   - Storage: An instance of the Storage interface which is responsible for storing
//     the last measurements.
//
// Example usage:
//
//	// Assuming implementations of NetworkStatsScraper and Storage interfaces
//	var scraper scraper.NetworkStatsScraper
//	var storage Storage
//
//	// URI of the file to be sampled
//	uri := "/path/to/network/stats/file"
//
//	// Create a new FileBasedDeltaSampler
//	deltaSampler := NewFileBasedDeltaSampler(uri, scraper, storage)
//
//	// Use deltaSampler to start sampling deltas
type FileBasedDeltaSampler struct {
	FileBasedSampler FileBasedSampler
	Storage          Storage
}

// NewFileBasedDeltaSampler creates a new instance of FileBasedDeltaSampler.
// This sampler is responsible for sampling network statistics, calculating deltas,
// and storing the results in a storage backend.
//
// Parameters:
//   - uri: The URI of the file where the network statistics will be sampled from.
//   - scraper: An implementation of the NetworkStatsScraper interface used to scrape
//     network statistics.
//   - storage: An implementation of the Storage interface where the sampled deltas
//     will be stored.
//
// Returns:
// - A pointer to an instance of FileBasedDeltaSampler.
func NewFileBasedDeltaSampler(uri string, scraper scraper.NetworkStatsScraper, storage Storage) *FileBasedDeltaSampler {
	return &FileBasedDeltaSampler{
		FileBasedSampler: *NewFileBasedSampler(uri, scraper),
		Storage:          storage,
	}
}

func (s *FileBasedDeltaSampler) Sample() (uint64, error) {
	lastSample, err := s.Storage.Load()

	if err != nil {
		return 0, err
	}

	sample, err := s.FileBasedSampler.Sample()

	if err != nil {
		return 0, err
	}

	delta := sample - lastSample

	err = s.Storage.Save(sample)

	if err != nil {
		return 0, err
	}

	return delta, nil
}

// FileBasedSampler is a struct that handles the sampling of network statistics
// from a file specified by a URI using a given scraper.
//
// Fields:
//   - uri: The URI of the file from which network statistics will be sampled.
//   - scraper: An implementation of the NetworkStatsScraper interface that retrieves
//     the value from the specified file.
//
// Example usage:
//
//	// Assuming an implementation of the NetworkStatsScraper interface
//	var scraper scraper.NetworkStatsScraper
//
//	// URI of the file to be sampled
//	uri := "/path/to/network/stats/file"
//
//	// Create a new FileBasedSampler
//	sampler := FileBasedSampler{
//	    uri:     uri,
//	    scraper: scraper,
//	}
//
//	// Use sampler to retrieve network statistics
//	value, err := sampler.scraper.Scrape(uri)
//	if err != nil {
//	    fmt.Println("Error scraping network statistics:", err)
//	    return
//	}
//	fmt.Println("Scraped value:", value)
type FileBasedSampler struct {
	// uri is the URI for the file from which network statistics will be sampled.
	uri string
	// scraper is an implementation of the NetworkStatsScraper interface used to
	// retrieve the value from the specified file.
	scraper scraper.NetworkStatsScraper
}

// NewFileBasedSampler creates a new instance of FileBasedSampler.
// This function initializes a FileBasedSampler with the provided URI and scraper.
//
// Parameters:
//   - uri: The URI of the file from which network statistics will be sampled.
//   - statsScraper: An implementation of the NetworkStatsScraper interface that will be used
//     to retrieve the network statistics from the specified file.
//
// Returns:
// - A pointer to an instance of FileBasedSampler initialized with the given URI and scraper.
//
// Example usage:
//
//	// Assuming an implementation of the NetworkStatsScraper interface
//	var scraper scraper.NetworkStatsScraper
//
//	// URI of the file to be sampled
//	uri := "/path/to/network/stats/file"
//
//	// Create a new FileBasedSampler
//	sampler := NewFileBasedSampler(uri, scraper)
//
//	// Use the sampler to retrieve network statistics
//	value, err := sampler.scraper.Scrape(sampler.uri)
//	if err != nil {
//	    fmt.Println("Error scraping network statistics:", err)
//	    return
//	}
//	fmt.Println("Scraped value:", value)
func NewFileBasedSampler(uri string, statsScraper scraper.NetworkStatsScraper) *FileBasedSampler {
	return &FileBasedSampler{
		uri:     uri,
		scraper: statsScraper,
	}
}

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
