# Go parameters
GO         = go
GO_BUILD   = $(GO) build
GO_CLEAN   = $(GO) clean
GO_TEST    = $(GO) test
GO_GET     = $(GO) get
GO_INSTALL = $(GO) install

GO_BIN_DIR    := $(shell go env GOPATH)/bin
GOLANGCI_LINT := $(GO_BIN_DIR)/golangci-lint

BINARY_NAME    := goStatic
BINARY_VERSION := $(shell git describe --tags --always)
BINARY_BUILD   := $(shell git rev-parse HEAD)
BINARY_DATE    := $(shell date +%FT%T%z)

LDFLAGS := -ldflags "-X main.VERSION=$(BINARY_VERSION) -X main.BUILD=$(BINARY_BUILD) -X main.BUILDDATE=$(BINARY_DATE)"
PLATFORMS := linux-amd64 linux-386 linux-arm linux-arm64 darwin-amd64 windows-amd64 windows-386

BINARIES = $(foreach PLATFORM, $(PLATFORMS), $(BINARY_NAME)-$(PLATFORM:windows-%=windows-%.exe))

CURRENT_PLATFORM = $(patsubst %.exe,%,$(patsubst $(BINARY_NAME)-%,%,$(@)))
OS               = $(word 1, $(subst -, ,$(CURRENT_PLATFORM)))
ARCH             = $(word 2, $(subst -, ,$(CURRENT_PLATFORM)))

.PHONY: all build build-all test test-all clean deps install linter release

build: deps test $(BINARY_NAME)

all: deps test build build-all

build-all: $(BINARIES)

test: linter
	$(GO_TEST) -v ./...

test-all: linter
	$(GO_TEST) -v all

linter: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run -v

clean:
	$(GO_CLEAN)
	$(GO) mod tidy
	rm -f $(BINARIES)

deps:
	$(GO_GET) -u -v ./...

install:
	$(GO_INSTALL)

release: deps test test-all build-all

$(BINARY_NAME):
	CGO_ENABLED=0 $(GO_BUILD) $(LDFLAGS) -o $@ -v ./...

$(PLATFORMS:%=$(BINARY_NAME)-%):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO_BUILD) $(LDFLAGS) -o $@ -v ./...

$(BINARY_NAME)-windows-%.exe:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO_BUILD) $(LDFLAGS) -o $@ -v ./...

# install golangci-lint if not exist
$(GOLANGCI_LINT):
	$(GO_GET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
