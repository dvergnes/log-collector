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
	"encoding/json"
	"net/http"

	"github.com/dvergnes/log-collector/api"

	"go.uber.org/zap"
)

const (
	invalidParameter = "invalid.parameter"
	internalError    = "internal.error"
	requestCanceled  = "request.canceled"
)

const internalErrorDetails = "Oops, try again later. If the problem persist, please contact your administrator"

func writeErrorResponse(w http.ResponseWriter, httpCode int, resp api.ErrorResponse, logger *zap.Logger) {
	payload, err := json.Marshal(resp)
	if err != nil {
		logger.Error("failed to serialize error response", zap.Error(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	if _, err := w.Write(payload); err != nil {
		logger.Error("failed to write error response", zap.Error(err))
	}
}



func writeResponse(w http.ResponseWriter, resp api.LogResponse, logger *zap.Logger) {
	payload, err := json.Marshal(resp)
	if err != nil {
		logger.Error("failed to serialize response", zap.Error(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(payload); err != nil {
		logger.Error("failed to write response", zap.Error(err))
	}
}
