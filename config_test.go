package main

import (
	"os"
	"reflect"
	"testing"
)

func TestReadConfigFromFile(t *testing.T) {
	expected := []ServiceConfigModel{
		{
			Listen: "localhost:8080",
			UpstreamEndpoints: []string{
				"http://localhost:8081",
				"http://localhost:8082",
			},
			SourceAddresses: []string{
				"192.168.1.1",
				"192.168.1.2",
			},
		},
		{
			Listen: "localhost:8080",
			UpstreamEndpoints: []string{
				"http://localhost:8081",
				"http://localhost:8082",
			},
			SourceAddresses: []string{
				"192.168.1.1",
				"192.168.1.2",
			},
		},
	}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	// Write test data to the temporary file
	testData := []byte(`
- listen: "localhost:8080"
  upstream_endpoints:
    - "http://localhost:8081"
    - "http://localhost:8082"
  source_addresses:
    - "192.168.1.1"
    - "192.168.1.2"
- listen: "localhost:8080"
  upstream_endpoints:
    - "http://localhost:8081"
    - "http://localhost:8082"
  source_addresses:
    - "192.168.1.1"
    - "192.168.1.2"`)
	if _, err := tmpfile.Write(testData); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Call ReadConfigFromFile with the temporary file path
	servicesConfig := ReadConfigFromFile(tmpfile.Name())

	// Check if the returned servicesConfig is equal to the expected slice
	if !reflect.DeepEqual(servicesConfig, expected) {
		t.Errorf("ReadConfigFromFile() = %v, want %v", servicesConfig, expected)
	}
}
