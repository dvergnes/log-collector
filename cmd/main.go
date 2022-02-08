package main

import (
	"log"

	"github.com/dvergnes/log-collector/internal/version"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err!=nil {
		log.Fatalf("failed to initialize logger %+v", err)
	}
	defer logger.Sync() // flushes buffer, if any

	sugar := logger.Sugar()
	sugar.Infow("starting log-collector",
		"version", version.Version,
	)

}
