# Build info
BUILD_INFO_IMPORT_PATH=github.com/dvergnes/log-collector/internal/version
VERSION=$(shell git describe --always --match "v[0-9]*" HEAD)
BUILD_INFO=-ldflags "-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)"
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

PKGS=$(shell go list ./...)

.PHONY: gotidy
gotidy:
	rm -fr go.sum
	go mod tidy -go=1.17

.PHONY: gomoddownload
gomoddownload:
	go mod download

.PHONY: gotest
gotest:
	@set -e; go test $(GOTAGS) -timeout 4m ${PKGS}

mock:
	mockery --dir processor --name TailReader --case underscore