VERSION = $(shell git rev-parse HEAD)
GIT_TAG = $(shell git rev-list --tags --max-count=1)
VERSION_TAG = $(if $(GIT_TAG),$(shell git describe --tags $(GIT_TAG)),v0)
CMD_DIR = "$(shell pwd)"/cmd/webrtc-signaling
BIN_DIR = "$(shell pwd)"/bin
RELEASE_DIR = "$(shell pwd)"/release
TARGET_NAME = webrtc-signaling-go
TARGET_OS = $(shell go env GOOS)
TARGET_ARCH = $(shell go env GOARCH)

build:
	@echo "Building for OS=$(TARGET_OS) Arch=$(TARGET_ARCH)"
	@echo "Version $(VERSION_TAG)-$(VERSION)"
	@mkdir -p $(BIN_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -ldflags="-X 'github.com/ownerofglory/webrtc-signaling-go/internal/handler.AppVersion=$(VERSION_TAG)'" -o $(BIN_DIR)/$(TARGET_NAME) $(CMD_DIR)/main.go

test:
	@echo "Testing..."
	go test -coverprofile=coverage.out ./...

clean:
	@rm -rf ./bin
	@rm -rf ./release