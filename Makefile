BIN = graylog-project
BIN_LINUX = $(BIN).linux
BIN_DARWIN_AMD64 = $(BIN).darwin-amd64
BIN_DARWIN_ARM64 = $(BIN).darwin-arm64
BIN_WINDOWS_AMD64 = $(BIN).windows-amd64.exe

GIT_REV=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%S%z)
GIT_TAG=$(shell git describe --tags --abbrev=0 --dirty 2>/dev/null || echo "0.0.0+dev")
BUILD_OPTS = -mod=vendor -ldflags "-s -X github.com/Graylog2/graylog-project-cli/cmd.gitRevision=$(GIT_REV) -X github.com/Graylog2/graylog-project-cli/cmd.buildDate=$(BUILD_DATE) -X github.com/Graylog2/graylog-project-cli/cmd.gitTag=$(GIT_TAG)"

all: test build

build: build-linux build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_OPTS) -o $(BIN_LINUX) main.go

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_OPTS) -o $(BIN_DARWIN_AMD64) main.go

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_OPTS) -o $(BIN_DARWIN_ARM64) main.go

build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_OPTS) -o $(BIN_WINDOWS_AMD64) main.go

sign-binaries: sign-binary-windows-amd64
	# This needs to run in a Docker container with the graylog/internal-codesigntool image
	codesigntool sign $(BIN_WINDOWS_AMD64)

sign-binary-windows-amd64:

install: install-linux install-darwin-amd64

install-darwin-amd64: build-darwin-amd64
	install -m 555 $(BIN_DARWIN_AMD64) $(HOME)/bin/$(BIN)

install-darwin-arm64: build-darwin-arm64
	install -m 555 $(BIN_DARWIN_ARM64) $(HOME)/bin/$(BIN)

install-linux: build-linux
	install -m 555 $(BIN_LINUX) $(HOME)/bin/$(BIN)

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	CGO_ENABLED=0 go test -mod=vendor ./...

clean:
	rm -f $(BIN_LINUX) $(BIN_DARWIN_AMD64) $(BIN_DARWIN_ARM64) $(BIN_WINDOWS_AMD64)
