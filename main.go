package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/samverrall/polybin/cmd"
	"github.com/samverrall/polybin/config"
)

const (
	configDir = ".config"
)

type flags struct {
	Service string
}

// parseFlags parses the provided flags, and validates the input. If a non-nil
// error is returned, this is considered a flag parse failure.
func parseFlags() (flags, error) {
	var userFlags flags
	flag.StringVar(&userFlags.Service, "service", "", "The name of the group of processes to run")
	flag.Parse()

	if err := validateFlags(userFlags); err != nil {
		return flags{}, err
	}

	return userFlags, nil
}

// validateFlags validates the user inputted flags, and returns an error
// if a flag is wrongly provided.
func validateFlags(flags flags) error {
	if strings.TrimSpace(flags.Service) == "" {
		return errors.New("service flag must be provided")
	}
	return nil
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		fmt.Printf("-> Failed to parse flags: %s", err.Error())
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("polybin: %v", err)
		os.Exit(1)
	}
	configFilePath := filepath.Join(homeDir, configDir, config.ConfigFileName)

	fmt.Printf("-> Looking for config file in %s\n", configFilePath)

	if err := config.CheckConfigFile(configFilePath); err != nil {
		fmt.Printf("-> Failed to read config file: %v", err)
		os.Exit(1)
	}

	config, err := config.Parse(configFilePath)
	if err != nil {
		fmt.Printf("-> Failed to parse config: %s", err.Error())
		os.Exit(1)
	}

	fmt.Printf("-> Starting polybin\n")

	if err := cmd.Polybin(config, flags.Service); err != nil {
		fmt.Printf("-> Polybin failed: %v", err)
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	<-stopChan
}
