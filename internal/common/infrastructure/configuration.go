package infrastructure

import (
	"fmt"
	"gopkg.in/yaml.v3"
	_ "gopkg.in/yaml.v3"
	"os"
)

type Configuration struct {
	MainDatabase struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"main_database"`
	TimeSeriesDatabase struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"time_series_database"`
}

func (configuration *Configuration) LoadFromFile() error {
	file, err := os.Open("configs/config.yml")
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(configuration)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	return nil
}
