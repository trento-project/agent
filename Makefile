VERSION ?= $(shell ./hack/get_version_from_git.sh)
INSTALLATIONSOURCE ?= "Community"
LDFLAGS = -X github.com/trento-project/agent/version.Version="$(VERSION)"
LDFLAGS := $(LDFLAGS) -X github.com/trento-project/agent/version.InstallationSource="$(INSTALLATIONSOURCE)"
ARCHS ?= amd64 arm64 ppc64le s390x
DEBUG ?= 0

ifeq ($(DEBUG), 0)
	LDFLAGS += -s -w
	GO_BUILD = CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -trimpath
else
	GO_BUILD = CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)"
endif

.PHONY: default
default: clean mod-tidy fmt vet-check test build

.PHONY: build
build: agent
agent:
	$(GO_BUILD) -o trento-agent

.PHONY: cross-compiled $(ARCHS)
cross-compiled: $(ARCHS)
$(ARCHS):
	@mkdir -p build/$@
	GOOS=linux GOARCH=$@ $(GO_BUILD) -o build/$@/trento-agent

.PHONY: clean
clean: clean-binary 

.PHONY: clean-binary
clean-binary:
	go clean
	rm -rf build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: fmt-check
fmt-check:
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

.PHONY: lint
lint:
	golangci-lint -v run

.PHONY: generate
generate:
ifeq (, $(shell command -v mockery 2> /dev/null))
	$(error "'mockery' command not found. You can install it locally with 'go install github.com/vektra/mockery/v2'.")
endif
ifeq (, $(shell command -v swag 2> /dev/null))
	$(error "'swag' command not found. You can install it locally with 'go install github.com/swaggo/swag/cmd/swag'.")
endif
	go generate ./...

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: vet-check
vet-check:
	go vet ./...

.PHONY: test
test:
	GIN_MODE=test go test -v -p 1 ./...

.PHONY: test-coverage
test-coverage: build/coverage.out
build/coverage.out:
	@mkdir -p build
	GIN_MODE=test go test -cover -coverprofile=build/coverage.out ./...
	go tool cover -html=build/coverage.out
