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
	"net"
	"net/http"

	"go.uber.org/zap"
)

// Server is a HTTP server that implements the REST API to read events located in log files
type Server struct {
	config Config
	server *http.Server

	logger *zap.Logger
}

// NewServer creates a Server with the given Config
func NewServer(config Config, parentLogger *zap.Logger) *Server {
	router := routes()
	return &Server{
		config: config,
		logger: parentLogger.Named("http-server"),
		server: &http.Server{
			Handler: router,
		},
	}
}

// Start starts the HTTP server on the port defined in Config
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start HTTP server %w", err)
	}
	s.logger.Sugar().Infow("starting http server",
		"port", s.config.Port)
	if err := s.server.Serve(ln); err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server %w", err)
	}
	return nil
}

// Stop stops the HTTP server. It waits up to the timeout defined by Config.ShutdownTimeout. After this delay, the HTTP
// server is stopped forcibly.
func (s *Server) Stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	s.logger.Sugar().Info("stopping http server")

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Warn("failed to stop http server", zap.Error(err))
	} else {
		s.logger.Info("http server is stopped")
	}
}
