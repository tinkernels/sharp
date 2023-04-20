package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

var (
	ExecConfig []ServiceConfigModel
)

type ServiceConfigModel struct {
	Listen            string   `yaml:"listen"`
	UpstreamEndpoints []string `yaml:"upstream_endpoints"`
	SourceAddresses   []string `yaml:"source_addresses"`
}

func ReadConfigFromFile(path string) (servicesConfig []ServiceConfigModel) {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Read file error:", err)
		panic(err)
	}
	err = yaml.Unmarshal(file, &servicesConfig)
	if err != nil {
		fmt.Println("Unmarshal config file error:", err)
		panic(err)
	}
	ExecConfig = servicesConfig
	return
}
