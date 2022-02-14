# Build info
BUILD_INFO_IMPORT_PATH=github.com/dvergnes/log-collector/internal/version
VERSION=$(shell git describe --always --match "v[0-9]*" HEAD)
BUILD_INFO=-ldflags "-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)"
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

PKGS=$(shell go list ./...)

RUN_CONFIG?=config/config.yml

.PHONY: gotidy
gotidy:
	rm -fr go.sum
	go mod tidy -go=1.17

.PHONY: goinstall
goinstall:
	go install github.com/vektra/mockery/v2@latest

.PHONY: gomoddownload
gomoddownload:
	go mod download

.PHONY: gotest
gotest: mocks
	@set -e; go test $(GOTAGS) -timeout 4m ${PKGS}

mocks: goinstall
	mockery --dir processor --name TailReader --case underscore
	mockery --dir processor --name EventProcessor --case underscore

.PHONY: run
run:
	GO111MODULE=on go run --race ./cmd/... --config=${RUN_CONFIG}