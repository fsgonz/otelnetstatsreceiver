package scraper

import (
	"io"
)

// NetworkStats represents the network statistics containing the number of received and transmitted bytes.
type NetworkStats struct {
	// ReceivedBytes holds the number of bytes received over the network.
	ReceivedBytes uint64

	// TransmittedBytes holds the number of bytes transmitted over the network.
	TransmittedBytes uint64
}

// NetworkStatsScraper defines an interface for scraping network stats data from an io.Reader.
type NetworkStatsScraper interface {
	// Scrape reads data from the provided io.Reader and scrapes it,
	// returning the count of received bytes, the transmitted bytes,
	// and an error if any.
	//
	// Parameters:
	//   data: The input data to be scraped, provided as an io.Reader.
	//
	// Returns:
	//   networkStats: The scraped network stats.
	//   error: An error, if any occurred during scraping.
	Scrape(data io.Reader) (networkStats NetworkStats, error error)
}
