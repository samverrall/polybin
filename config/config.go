package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

const (
	ConfigFileName = "polybin.json"
)

type Config []ConfigEntry

type ConfigEntry struct {
	ProjectName string    `json:"project"`
	Services    []Service `json:"services"`
}

type Service struct {
	Type string `json:"type"`
	Dir  string `json:"dir"`
	// Binary is only required for services with the type `watch`.
	Binary *string  `json:"binary"`
	Args   []string `json:"args"`
}

func (c Config) FindProjectByName(projectName string) *ConfigEntry {
	for _, ce := range c {
		if ce.ProjectName == projectName {
			return &ce
		}
	}
	return nil
}

// Init creates a new config file at the supplied filepath directory.
// A basic example demo example JSON is used as the file contents.
func Init(filepath string) error {
	// os.Create()
	return nil
}

func CheckConfigFile(filepath string) error {
	file, err := os.ReadFile(filepath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("no config file exists in your home .config with filename: %s", ConfigFileName)

	case err != nil:
		return fmt.Errorf("failed to read config file: %s", err.Error())

	case len(file) == 0:
		return fmt.Errorf("supplied config file is empty")

	case !json.Valid(file):
		return errors.New("json in config is not valid")

	default:
		return nil
	}
}

func Parse(filepath string) (*Config, error) {
	file, err := os.ReadFile(filepath)
	log.Println(err)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
