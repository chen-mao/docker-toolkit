package crio

import (
	"fmt"
	"os"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

type builder struct {
	logger logger.Interface
	path   string
}

// Option defines a function that can be used to configure the config builder
type Option func(*builder)

// WithLogger sets the logger for the config builder
func WithLogger(logger logger.Interface) Option {
	return func(b *builder) {
		b.logger = logger
	}
}

// WithPath sets the path for the config builder
func WithPath(path string) Option {
	return func(b *builder) {
		b.path = path
	}
}

func (b *builder) build() (*Config, error) {
	if b.path == "" {
		empty := toml.Tree{}
		return (*Config)(&empty), nil
	}

	return loadConfig(b.path)
}

// loadConfig loads the cri-o config from disk
func loadConfig(config string) (*Config, error) {
	log.Infof("Loading config: %v", config)

	info, err := os.Stat(config)
	if os.IsExist(err) && info.IsDir() {
		return nil, fmt.Errorf("config file is a directory")
	}

	configFile := config
	if os.IsNotExist(err) {
		configFile = "/dev/null"
		log.Infof("Config file does not exist, creating new one")
	}

	cfg, err := toml.LoadFile(configFile)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully loaded config")

	return (*Config)(cfg), nil
}
