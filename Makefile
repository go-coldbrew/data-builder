.PHONY: build test doc lint bench test-v
build:
	go build ./...

test:
	go test -race ./...

test-v:
	go test -v -race ./...

lint:
	go tool golangci-lint run

bench:
	go test -run=^$ -bench=. -benchmem ./...

doc:
	go tool gomarkdoc --output '{{.Dir}}/README.md' ./...
