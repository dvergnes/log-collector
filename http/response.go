package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

const (
	invalidParameter = "invalid.parameter"
	internalError    = "internal.error"
)

const internalErrorDetails = "Oops, try again later. If the problem persist, please contact your administrator"

type errorResponse struct {
	Code    string `json:"code"`
	Details string `json:"details"`
}

func writeErrorResponse(w http.ResponseWriter, httpCode int, resp errorResponse, logger *zap.Logger) {
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

type logResponse struct {
	File   string   `json:"file"`
	Events []string `json:"events"`
}

func writeResponse(w http.ResponseWriter, resp logResponse, logger *zap.Logger) {
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
