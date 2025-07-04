VERSION ?= $(shell ./hack/get_version_from_git.sh)
INSTALLATIONSOURCE ?= "Community"
LDFLAGS = -X github.com/trento-project/agent/version.Version="$(VERSION)"
LDFLAGS := $(LDFLAGS) -X github.com/trento-project/agent/version.InstallationSource="$(INSTALLATIONSOURCE)"
CURRENT_ARCH := $(shell go env GOARCH)
ARCHS ?= amd64 ppc64le s390x
DEBUG ?= 0
BUILD_DIR := ./build
BUILD_OUTPUT ?= $(BUILD_DIR)/$(CURRENT_ARCH)/trento-agent
LOCAL_RABBITMQ_SSL_URL="amqps://guest:guest@localhost:5677?certfile=$$(pwd)/container_fixtures/rabbitmq/certs/client_agent.trento.local_certificate.pem&keyfile=$$(pwd)/container_fixtures/rabbitmq/certs/client_agent.trento.local_key.pem&verify=verify_peer&cacertfile=$$(pwd)/container_fixtures/rabbitmq/certs/ca_certificate.pem"
TEST_MODULES := $(shell go list ./... | grep -v /mocks)

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
	$(GO_BUILD) -o $(BUILD_OUTPUT)

.PHONY: build-plugin-examples
build-plugin-examples:
	$(GO_BUILD) -o $(BUILD_DIR)/$(CURRENT_ARCH)/plugin_examples/dummy ./plugin_examples/dummy/dummy.go
	$(GO_BUILD) -o $(BUILD_DIR)/$(CURRENT_ARCH)/plugin_examples/sleep ./plugin_examples/sleep/sleep.go

.PHONY: cross-compiled $(ARCHS)
cross-compiled: $(ARCHS)
$(ARCHS):
	@mkdir -p build/$@
	GOOS=linux GOARCH=$@ $(GO_BUILD) -o $(BUILD_DIR)/$@/trento-agent

.PHONY: bundle
bundle:
	set -x
	find $(BUILD_DIR) -maxdepth 1 -mindepth 1 -type d -exec sh -c 'tar -zcf build/trento-agent-$$(basename {}).tgz -C {} trento-agent -C $$(pwd)/packaging/systemd trento-agent.service' \;
	
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
	mockery

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: vet-check
vet-check:
	go vet ./...

.PHONY: test
test:
	go test -v -p 1 -race ./...

.PHONY: rabbitmq-local-ssl-test
rabbitmq-local-ssl-test:
	RABBITMQ_URL=$(LOCAL_RABBITMQ_SSL_URL) go test -v -p 1 -race ./internal/factsengine/factsengine_integration_test.go

.PHONY: test-short
test-short:
	go test -short -v -p 1 -race ./...

.PHONY: test-coverage
test-coverage: 
	go test -v -p 1 -race -covermode atomic -coverprofile=covprofile $(TEST_MODULES)

.PHONY: test-build
test-build:
	bats -r ./test
