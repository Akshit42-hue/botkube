.DEFAULT_GOAL := build
.PHONY: container-image test test-integration-slack test-integration-discord build pre-build publish lint lint-fix go-import-fmt system-check save-images load-and-push-images

# Show this help.
help:
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t

lint-fix: go-import-fmt
	@go mod tidy
	@go mod verify
	@golangci-lint run --fix "./..."

go-import-fmt:
	@./hack/fmt-imports.sh

# test
test: system-check
	@go test -v  -race ./...

test-integration-slack: system-check
	@go test -v -tags=integration -race -count=1 ./test/... -run "TestSlack"

test-integration-discord: system-check
	@go test -v -tags=integration -race -count=1 ./test/... -run "TestDiscord"

# Build the binary
build: pre-build
	@cd cmd/botkube;GOOS_VAL=$(shell go env GOOS) CGO_ENABLED=0 GOARCH_VAL=$(shell go env GOARCH) go build -o $(shell go env GOPATH)/bin/botkube
	@echo "Build completed successfully"

# Build the image
container-image: pre-build
	@echo "Building docker image"
	@./hack/goreleaser.sh build
	@echo "Docker image build successful"

# Build the image
container-image-single: pre-build
	@echo "Building single target docker image"
	@./hack/goreleaser.sh build_single
	@echo "Single target docker image build successful"

# Build the e2e test image
container-image-single-e2e: pre-build
	@echo "Building single target docker image for e2e test"
	@./hack/goreleaser.sh build_single_e2e
	@echo "Single target docker image build for e2e tests successful"

# Build project and push dev images with v9.99.9-dev tag
release-snapshot:
	@./hack/goreleaser.sh release_snapshot

# Build project and save images with IMAGE_TAG tag
save-images:
	@./hack/goreleaser.sh save_images

# Load project and push images with IMAGE_TAG tag
load-and-push-images:
	@./hack/goreleaser.sh load_and_push_images

# system checks
system-check:
	@echo "Checking system information"
	@if [ -z "$(shell go env GOOS)" ] || [ -z "$(shell go env GOARCH)" ] ; \
	then \
	echo 'ERROR: Could not determine the system architecture.' && exit 1 ; \
	else \
	echo 'GOOS: $(shell go env GOOS)' ; \
	echo 'GOARCH: $(shell go env GOARCH)' ; \
	echo 'System information checks passed.'; \
	fi ;

# Pre-build checks
pre-build: system-check

# Run chart lint & helm-docs
process-chart:
	@./hack/process-chart.sh
