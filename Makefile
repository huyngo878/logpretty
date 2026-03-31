BINARY     := logpretty
BUILD_DIR  := dist
GOFLAGS    := CGO_ENABLED=0
LDFLAGS    := -ldflags="-s -w"
TRIMPATH   := -trimpath

.PHONY: build test lint vuln release-dry clean

build:
	$(GOFLAGS) go build $(TRIMPATH) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .

test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run

vuln:
	govulncheck ./...

release-dry:
	goreleaser release --snapshot --clean

clean:
	rm -rf $(BUILD_DIR) coverage.out
