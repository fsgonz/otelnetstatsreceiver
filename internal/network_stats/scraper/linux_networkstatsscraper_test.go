package scraper

import (
	"os"
	"testing"
)

func TestLinuxNetworkStatsScraper(t *testing.T) {
	t.Run("Network stats parsed from the file with default interface", func(t *testing.T) {
		const wantedReceivedBytes = 3862937603
		const wantedTransmitBytes = 281882792
		assertExpectedNetUsageBytes("testdata/eth0_test.data", t, wantedReceivedBytes, wantedTransmitBytes, "")
	})

	t.Run("Network stats parsed from the file with second interface in the file", func(t *testing.T) {
		const wantedReceivedBytes = 3862937603
		const wantedTransmitBytes = 281882792
		assertExpectedNetUsageBytes("testdata/eth0_test.data", t, wantedReceivedBytes, wantedTransmitBytes, "eth0")
	})

	t.Run("Network stats parsed from the file with third interface in the file", func(t *testing.T) {
		const wantedReceivedBytes = 2247549264
		const wantedTransmitBytes = 255567044
		assertExpectedNetUsageBytes("testdata/eth0_test.data", t, wantedReceivedBytes, wantedTransmitBytes, "eth1")
	})

	t.Run("Network stats parsed from the file with lo interface in the file", func(t *testing.T) {
		const wantedReceivedBytes = 1982736
		const wantedTransmitBytes = 1982736
		assertExpectedNetUsageBytes("testdata/eth0_test.data", t, wantedReceivedBytes, wantedTransmitBytes, "lo")
	})
}

func assertExpectedNetUsageBytes(testFile string, t *testing.T, wantedReceivedBytes uint64, wantedTransmitBytes uint64, interfaceName string) {
	f, err := os.Open(testFile)

	if err != nil {
		t.Errorf("The following error occurred on retrieving the test file: %s", err.Error())
	}

	networkStats, err := NewLinuxNetworkDevicesFileScraperWithInterface(interfaceName).Scrape(f)

	if err != nil {
		t.Errorf("Error on scraping the net stats: %s", err.Error())
	}

	if networkStats.ReceivedBytes != wantedReceivedBytes {
		t.Errorf("Error on received bytes. Expected: %d, Got: %d", wantedReceivedBytes, networkStats.ReceivedBytes)
	}

	if networkStats.TransmittedBytes != wantedTransmitBytes {
		t.Errorf("Error on transmitted bytes. Expected: %d, Got: %d", wantedTransmitBytes, networkStats.TransmittedBytes)
	}
}
