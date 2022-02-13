// Copyright (c) 2022 Denis Vergnes
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

const (
	defaultLogFolder  = "/var/log/"
	defaultBufferSize = 4096
	defaultMaxEvents  = 10_000
)

// Config contains the configuration for the HTTP server
type Config struct {
	// Port defines the listening port of the HTTP server
	Port uint `yaml:"port"`
	// ShutdownTimeout defines the timeout to wait for the server to shut down gracefully
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	// LogFolder defines the folder that contains the log files to read
	LogFolder string `yaml:"log_folder"`
	// BufferSize defines the size in bytes of the buffer used to read the log file
	BufferSize int `yaml:"buffer_size"`
	// MaxEvents defines the maximum number of events returned. That means the limit applies after filter is applied.
	MaxEvents uint `yaml:"max_events"`
}

func (c *Config) setDefaults() {
	if c.LogFolder == "" {
		c.LogFolder = defaultLogFolder
	}
	if c.BufferSize == 0 {
		c.BufferSize = defaultBufferSize
	}
	if c.MaxEvents == 0 {
		c.MaxEvents = defaultMaxEvents
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 30 * time.Second
	}
}

func (c *Config) validate(fs afero.Fs) error {
	if c.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be strictly positive")
	}

	ok, err := afero.Exists(fs, c.LogFolder)
	if err != nil {
		return fmt.Errorf("failed to verify log folder presence %w", err)
	}
	if !ok {
		return errors.New("log folder declared in configuration does not exist")
	}
	ok, err = afero.IsDir(fs, c.LogFolder)
	if err != nil {
		return fmt.Errorf("failed to verify that log folder is a directory %w", err)
	}
	if !ok {
		return errors.New("log folder declared in configuration is not a directory")
	}
	return nil
}

// LoadConfig loads the Config from the given bytes array, it sets defaults and verify that the config is valid.
// It returns an error if the config cannot be read or if it is invalid
func LoadConfig(data []byte, fs afero.Fs) (*Config, error) {
	conf := &Config{}
	if err := yaml.Unmarshal(data, conf); err != nil {
		return nil, err
	}
	conf.setDefaults()
	return conf, conf.validate(fs)
}
