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
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dvergnes/log-collector/api"
	"github.com/dvergnes/log-collector/processor"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func validateFileParameter(file string) error {
	if len(file) == 0 {
		return httpError{
			code: invalidParameter,
			details: "file name must not be empty",
			httpStatus: http.StatusBadRequest,
		}
	}
	basename := filepath.Base(file)
	if basename != file {
		return httpError{
			code: invalidParameter,
			details: "file name must not contain any relative or absolute path reference",
			httpStatus: http.StatusBadRequest,
		}
	}
	return nil
}

func parseLimit(max uint, limit string) (uint, error) {
	if len(limit) == 0 {
		return 0, nil
	}
	l, err := strconv.Atoi(limit)
	if err != nil {
		return 0, httpError{
			code: invalidParameter,
			details: "limit is not a valid integer",
			httpStatus: http.StatusBadRequest,
		}
	}
	if l <= 0 {
		return 0, httpError{
			code: invalidParameter,
			details: "limit must be strictly positive",
			httpStatus: http.StatusBadRequest,
		}
	}
	if uint(l) > max {
		return 0, httpError{
			code: invalidParameter,
			details: fmt.Sprintf("limit must be equal or less than %d", max),
			httpStatus: http.StatusBadRequest,
		}
	}
	return uint(l), nil
}

type httpError struct {
	code       string
	details    string
	httpStatus int
}

func (err httpError) Error() string {
	return err.details
}

var internalErr = httpError{
	code:       internalError,
	httpStatus: http.StatusInternalServerError,
	details:    internalErrorDetails,
}

func checkFile(fs afero.Fs, path string) error {
	exist, err := afero.Exists(fs, path)
	if err != nil {
		return internalErr
	}
	if !exist {
		return httpError{
			code:       invalidParameter,
			httpStatus: http.StatusNotFound,
			details:    fmt.Sprintf("file %s was not found", path),
		}
	}
	isDir, err := afero.IsDir(fs, path)
	if err != nil {
		return internalErr
	}
	if isDir {
		return httpError{
			code:       invalidParameter,
			httpStatus: http.StatusBadRequest,
			details:    fmt.Sprintf("file %s is a directory", path),
		}
	}
	return nil
}

func logHandler(fs afero.Fs, config *Config, parentLogger *zap.Logger) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	logger := parentLogger.Named("log-handler")
	return func(w http.ResponseWriter, request *http.Request, params httprouter.Params) {
		query := request.URL.Query()
		name := query.Get("file")
		err := validateFileParameter(name)
		if err != nil {
			handleError(w, err, logger)
			return
		}

		limit, err := parseLimit(config.MaxEvents, query.Get("limit"))
		if err != nil {
			handleError(w, err, logger)
			return
		}
		if limit == 0 {
			limit = config.MaxEvents
		}

		path := filepath.Join(config.LogFolder, name)
		if err := checkFile(fs, path); err != nil {
			logger.Error("failed to verify that file can be processed", zap.Error(err))
			handleError(w, err, logger)
			return
		}

		reader, err := processor.NewTailReader(fs, path)
		if err != nil {
			logger.Error("failed to open reader", zap.Error(err))
			handleError(w, err, logger)
			return
		}
		defer reader.Close()

		filter := query.Get("filter")
		logger.Sugar().Infow("processing file",
			"file", path,
			"filter", filter,
			"limit", limit)
		p := createProcessor(reader, config, filter, limit)

		events, err := processFile(request.Context(), p)
		if err != nil {
			logger.Error("failed to process file", zap.Error(err))
			handleError(w, err, logger)
			return
		}
		writeResponse(w, api.LogResponse{
			File:   path,
			Events: events,
		}, logger)

	}
}

func handleError(w http.ResponseWriter, err error, logger *zap.Logger) {
	if httpErr, ok := err.(httpError); ok {
		writeErrorResponse(w, httpErr.httpStatus, api.ErrorResponse{
			Code:    httpErr.code,
			Details: httpErr.details,
		}, logger)
	} else {
		writeErrorResponse(w, http.StatusInternalServerError, api.ErrorResponse{
			Code:    internalError,
			Details: internalErrorDetails,
		}, logger)
	}
}

func processFile(ctx context.Context, p processor.EventProcessor) ([]string, error) {
	var acc []string
	for {
		select {
		case <-ctx.Done():
			return nil, httpError{
				code:       requestCanceled,
				details:    "client canceled request",
				httpStatus: http.StatusBadRequest,
			}
		default:
		}

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

func createProcessor(reader processor.TailReader, config *Config, filter string, limit uint) processor.EventProcessor {
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
