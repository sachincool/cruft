.PHONY: fmt test vet check build snapshot

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

vet:
	go vet ./...

check: fmt test vet
	git diff --exit-code

build:
	go build -o bin/cruft ./cmd/cruft

snapshot:
	goreleaser release --snapshot --clean
