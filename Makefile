.PHONY: build test doc lint bench
build:
	go build ./...

test:
	go test ./... -race

test-v:
	go test -v ./... -race

install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/princjef/gomarkdoc/cmd/gomarkdoc

lint:
	golangci-lint run

bench:
	go test -bench=. -benchmem ./...

doc: install
	gomarkdoc --output '{{.Dir}}/README.md' ./...
