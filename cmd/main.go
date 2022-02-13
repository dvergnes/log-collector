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

package main

import (
	"flag"
	"github.com/dvergnes/log-collector/http"
	"github.com/dvergnes/log-collector/internal/version"
	"github.com/spf13/afero"
	"io/ioutil"
	"log"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger %+v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	sugar := logger.Sugar()
	sugar.Infow("starting log-collector",
		"version", version.Version,
	)

	confFile := flag.String("config", "config.yml", "configuration file for the log collector")
	flag.Parse()
	data, err := ioutil.ReadFile(*confFile)
	if err != nil {
		log.Fatalf("failed to read configuration file %+v", err)
	}
	fs := afero.NewOsFs()
	conf, err := http.LoadConfig(data, fs)
	if err != nil {
		log.Fatalf("failed to parse configuration %+v", err)
	}

	httpServer := http.NewServer(conf, fs, logger)

	errCh := make(chan error, 1)
	go func() {
		err := httpServer.Start()
		if err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	// TODO: handle signals (SIGTERM)
	err = <-errCh
	if err != nil {
		log.Fatalf("failed to start http server %+v", err)
	}
}
