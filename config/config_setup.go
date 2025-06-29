package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	data map[string]interface{}
}

// NewConfig
// Read and deserialize the YAML file into a variable of type map[string]interface{}
func NewConfig(file string) *Config {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var data map[string]interface{}

	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		panic(err)
	}

	return &Config{data: data}
}
