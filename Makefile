.PHONY: build build-all run test test-coverage lint clean release-dry snapshot install tag

BINARY=migrate-tool
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Local build
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/migrate-tool

# Build for all platforms (local test)
build-all:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/migrate-tool
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 ./cmd/migrate-tool
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/migrate-tool
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/migrate-tool

run:
	go run ./cmd/migrate-tool $(ARGS)

test:
	go test -v ./...

test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

lint:
	golangci-lint run

clean:
	rm -rf bin/ dist/ coverage.txt coverage.html

# GoReleaser dry-run (no publish)
release-dry:
	goreleaser release --snapshot --clean --skip=publish

# GoReleaser snapshot
snapshot:
	goreleaser build --snapshot --clean

# Install locally
install: build
	cp bin/$(BINARY) $(GOPATH)/bin/

# Create release tag
tag:
	@read -p "Version (e.g., v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	echo "Created tag $$version. Push with: git push origin $$version"
