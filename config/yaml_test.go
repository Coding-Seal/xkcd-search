package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	_, err := NewConfig("../test/config/right.yaml")
	assert.NoError(t, err)
	_, err = NewConfig("../test/config/no_file.yaml")
	assert.Error(t, err)
}

func TestNewConfig_FileNotFound(t *testing.T) {
	cfg, err := NewConfig("/nonexistent/path/config.yaml")
	// Error loading file, but returns config with defaults
	assert.Error(t, err)
	assert.NotNil(t, cfg)
}

func TestNewConfig_DefaultPort(t *testing.T) {
	cfg, _ := NewConfig("/nonexistent.yaml")
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestNewConfig_DefaultFetcherURL(t *testing.T) {
	cfg, _ := NewConfig("/nonexistent.yaml")
	assert.NotEmpty(t, cfg.Fetcher.SourceURL)
}
