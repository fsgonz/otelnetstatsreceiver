package scraper

import (
	"io"
)

// NetworkStats represents received and transmitted byte counts for a network interface.
type NetworkStats struct {
	ReceivedBytes    uint64
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
