package main

import (
	"log"
	"time"

	"github.com/dvergnes/log-collector/http"
	"github.com/dvergnes/log-collector/internal/version"

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

	httpServer := http.NewServer(http.Config{
		Port:            8888,
		ShutdownTimeout: time.Second,
	}, logger)

	errCh := make(chan error, 1)
	go func() {
		err := httpServer.Start()
		if err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	<-errCh
}
