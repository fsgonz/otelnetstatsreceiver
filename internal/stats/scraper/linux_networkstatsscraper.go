package scraper

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// LinuxNetworkDevicesFileScraper is a struct that represents a scraper for
// network devices files on Linux systems.
//
// Fields:
// - InterfaceName: The name of the network interface whose statistics will be scraped.
//
// This scraper is specifically designed to work with Linux network device files,
// typically found in the /proc/net/dev directory or similar locations.
//
// Example usage:
//
//	// Create a new LinuxNetworkDevicesFileScraper for a specific network interface
//	scraper := LinuxNetworkDevicesFileScraper{
//	    InterfaceName: "eth0",
//	}
//
//	// Implement the scraping logic using the InterfaceName field
//	stats, err := scraper.Scrape()
//	if err != nil {
//	    fmt.Println("Error scraping network device statistics:", err)
//	    return
//	}
//	fmt.Println("Scraped network statistics:", stats)
type LinuxNetworkDevicesFileScraper struct {
	InterfaceName string
}

// NewLinuxNetworkDevicesFileScraperWithInterface creates a new instance of LinuxNetworkDevicesFileScraper
// for the specified network interface.
//
// Parameters:
//   - interfaceName: The name of the network interface whose statistics will be scraped. If an empty
//     string is provided, the default interface "eth0" will be used.
//
// Returns:
//   - A pointer to an instance of LinuxNetworkDevicesFileScraper initialized with the specified or default
//     network interface name.
//
// Example usage:
//
//	// Create a scraper for a specific network interface
//	scraper := NewLinuxNetworkDevicesFileScraperWithInterface("wlan0")
//
//	// Use the scraper to retrieve network statistics
//	stats, err := scraper.Scrape()
//	if err != nil {
//	    fmt.Println("Error scraping network device statistics:", err)
//	    return
//	}
//	fmt.Println("Scraped network statistics:", stats)
//
//	// Create a scraper with the default network interface
//	defaultScraper := NewLinuxNetworkDevicesFileScraperWithInterface("")
//	defaultStats, err := defaultScraper.Scrape()
//	if err != nil {
//	    fmt.Println("Error scraping network device statistics:", err)
//	    return
//	}
//	fmt.Println("Scraped default network statistics:", defaultStats)
func NewLinuxNetworkDevicesFileScraperWithInterface(interfaceName string) *LinuxNetworkDevicesFileScraper {
	if interfaceName == "" {
		return NewLinuxNetworkDevicesFileScraperWithInterface("eth0")
	}
	return &LinuxNetworkDevicesFileScraper{
		InterfaceName: interfaceName,
	}
}

// NewLinuxNetworkDevicesFileScraper creates a new instance of LinuxNetworkDevicesFileScraper
// with the default network interface name set to "eth0".
//
// Returns:
//   - A pointer to an instance of LinuxNetworkDevicesFileScraper initialized with the default
//     network interface name "eth0".
//
// Example usage:
//
//	// Create a new LinuxNetworkDevicesFileScraper with the default network interface "eth0"
//	scraper := NewLinuxNetworkDevicesFileScraper()
//
//	// Use the scraper to retrieve network statistics
//	stats, err := scraper.Scrape()
//	if err != nil {
//	    fmt.Println("Error scraping network device statistics:", err)
//	    return
//	}
//	fmt.Println("Scraped network statistics:", stats)
func NewLinuxNetworkDevicesFileScraper() *LinuxNetworkDevicesFileScraper {
	return NewLinuxNetworkDevicesFileScraperWithInterface("eth0")
}

// Scrape reads network statistics from the provided data reader, which is expected to contain
// information in the format of /proc/net/dev, and extracts statistics for the specified network interface.
//
// Parameters:
// - data: An io.Reader that provides the content of the network devices file (e.g., /proc/net/dev).
//
// Returns:
// - networkStats: A struct containing the received and transmitted bytes for the specified network interface.
// - error: An error if the specified network interface is not found or if there are issues parsing the data.
func (s *LinuxNetworkDevicesFileScraper) Scrape(data io.Reader) (networkStats NetworkStats, err error) {
	// Create a new scanner to read the data line by line
	scanner := bufio.NewScanner(data)

	// Iterate over each line in the data
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the current line contains the specified network interface name followed by a colon
		if strings.Contains(line, s.InterfaceName+":") {
			// Split the line into fields using whitespace as the delimiter
			fields := strings.Fields(line)

			// Parse the received bytes (second field)
			receivedBytes, _ := strconv.ParseUint(fields[1], 10, 64)

			// Parse the transmitted bytes (tenth field)
			transmittedBytes, _ := strconv.ParseUint(fields[9], 10, 64)

			// Return the parsed network statistics
			return NetworkStats{
				ReceivedBytes:    receivedBytes,
				TransmittedBytes: transmittedBytes,
			}, nil
		}
	}

	// If the loop completes without finding the interface, return an error
	return NetworkStats{}, fmt.Errorf("interface '%s' not found in file info", s.InterfaceName)
}
