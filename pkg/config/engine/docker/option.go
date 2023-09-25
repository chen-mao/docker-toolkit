package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

type builder struct {
	path string
}

// Option defines a function that can be used to configure the config builder
type Option func(*builder)

// WithPath sets the path for the config builder
func WithPath(path string) Option {
	return func(b *builder) {
		b.path = path
	}
}

func (b *builder) build() (*Config, error) {
	if b.path == "" {
		empty := make(Config)
		return &empty, nil
	}

	return loadConfig(b.path)
}

// loadConfig loads the docker config from disk
func loadConfig(configFilePath string) (*Config, error) {
	log.Infof("Loading docker config from %v", configFilePath)

	info, err := os.Stat(configFilePath)
	if os.IsExist(err) && info.IsDir() {
		return nil, fmt.Errorf("config file is a directory")
	}

	cfg := make(Config)

	if os.IsNotExist(err) {
		log.Infof("Config file does not exist, creating new one")
		return &cfg, nil
	}

	readBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read config: %v", err)
	}

	reader := bytes.NewReader(readBytes)
	if err := json.NewDecoder(reader).Decode(&cfg); err != nil {
		return nil, err
	}

	log.Infof("Successfully loaded config")
	return &cfg, nil
}
