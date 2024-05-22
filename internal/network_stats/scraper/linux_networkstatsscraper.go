package scraper

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// LinuxNetworkDevicesFileScraper A Network Devices File Scraper for Linux
type LinuxNetworkDevicesFileScraper struct {
	InterfaceName string
}

func NewLinuxNetworkDevicesFileScraperWithInterface(interfaceName string) *LinuxNetworkDevicesFileScraper {
	if interfaceName == "" {
		return NewLinuxNetworkDevicesFileScraperWithInterface("eth0")
	}
	return &LinuxNetworkDevicesFileScraper{
		InterfaceName: interfaceName,
	}
}

func NewLinuxNetworkDevicesFileScraper() *LinuxNetworkDevicesFileScraper {
	return NewLinuxNetworkDevicesFileScraperWithInterface("eth0")
}

func (s *LinuxNetworkDevicesFileScraper) Scrape(data io.Reader) (networkStats NetworkStats, error error) {
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, s.InterfaceName+":") {
			fields := strings.Fields(line)
			receivedBytes, _ := strconv.ParseUint(fields[1], 10, 64)
			transmittedBytes, _ := strconv.ParseUint(fields[9], 10, 64)
			return NetworkStats{
				ReceivedBytes:    receivedBytes,
				TransmittedBytes: transmittedBytes,
			}, nil
		}
	}

	return NetworkStats{}, fmt.Errorf("interface '%s' not found in file info", s.InterfaceName)
}
