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
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dvergnes/log-collector/processor"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func validateFileParameter(file string) error {
	if len(file) == 0 {
		return errors.New("file name must not be empty")
	}
	basename := filepath.Base(file)
	if basename != file {
		return errors.New("file name must not contain any relative or absolute path reference")
	}
	return nil
}

func parseLimit(max uint, limit string) (uint, error) {
	if len(limit) == 0 {
		return 0, nil
	}
	l, err := strconv.Atoi(limit)
	if err != nil {
		return 0, errors.New("limit is not an invalid integer")
	}
	if l <= 0 {
		return 0, errors.New("limit must be strictly positive")
	}
	if uint(l) > max {
		return 0, fmt.Errorf("limit must be less or equal than %d", max)
	}
	return uint(l), nil
}

func logHandler(fs afero.Fs, config Config, parentLogger *zap.Logger) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	logger := parentLogger.Named("log-handler")
	return func(w http.ResponseWriter, request *http.Request, params httprouter.Params) {
		query := request.URL.Query()
		name := query.Get("file")
		err := validateFileParameter(name)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, errorResponse{
				Code:    invalidParameter,
				Details: err.Error(),
			}, logger)
			return
		}

		limit, err := parseLimit(config.MaxEvents, query.Get("limit"))
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, errorResponse{
				Code:    invalidParameter,
				Details: err.Error(),
			}, logger)
			return
		}
		if limit == 0 {
			limit = config.MaxEvents
		}

		path := config.LogFolder + name
		if !isFileValid(w, fs, path, name, logger) {
			return
		}
		reader, err := processor.NewTailReader(fs, path)
		defer reader.Close()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, errorResponse{
				Code:    internalError,
				Details: err.Error(),
			}, logger)
			return
		}

		filter := query.Get("filter")
		p := createProcessor(reader, config, filter, limit)

		events, err := processFile(p)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, errorResponse{
				Code:    internalError,
				Details: err.Error(),
			}, logger)
			return
		}
		writeResponse(w, logResponse{
			File:   path,
			Events: events,
		}, logger)

	}
}

func processFile(p processor.EventProcessor) ([]string, error) {
	var acc []string
	for {
		s, err := p.Next()
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		acc = append(acc, s)
	}
	return acc, nil
}

func createProcessor(reader processor.TailReader, config Config, filter string, limit uint) processor.EventProcessor {
	p := processor.EventProcessor(processor.NewEventBreaker(reader, processor.ReverseScanLines, config.BufferSize))
	if len(filter) != 0 {
		predicate := func(s string) bool {
			return strings.Contains(s, filter)
		}
		p = processor.WithFilter(p, predicate)
	}
	p = processor.WithLimit(p, limit)
	return p
}

func isFileValid(w http.ResponseWriter, fs afero.Fs, path string, name string, logger *zap.Logger) bool {
	exist, err := afero.Exists(fs, path)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, errorResponse{
			Code:    internalError,
			Details: internalErrorDetails,
		}, logger)
		return false
	}
	if !exist {
		writeErrorResponse(w, http.StatusNotFound, errorResponse{
			Code:    invalidParameter,
			Details: fmt.Sprintf("File %s was not found", name),
		}, logger)
		return false
	}
	isDir, err := afero.IsDir(fs, path)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, errorResponse{
			Code:    internalError,
			Details: internalErrorDetails,
		}, logger)
		return false
	}
	if isDir {
		writeErrorResponse(w, http.StatusBadRequest, errorResponse{
			Code:    invalidParameter,
			Details: fmt.Sprintf("File %s is a directory", name),
		}, logger)
		return false
	}
	return true
}
