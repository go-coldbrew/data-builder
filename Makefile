.PHONY: build test doc lint bench install test-v
build:
	go build ./...

test:
	go test -race ./...

test-v:
	go test -v -race ./...

install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/princjef/gomarkdoc/cmd/gomarkdoc

lint: install
	golangci-lint run

bench:
	go test -run=^$ -bench=. -benchmem ./...

doc: install
	gomarkdoc --output '{{.Dir}}/README.md' ./...
