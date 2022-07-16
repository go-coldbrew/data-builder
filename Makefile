.PHONY: build test doc
build:
	go build ./...

test:
	go test ./... -race

test-v:
	go test -v ./... -race

doc:
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc
	gomarkdoc --output '{{.Dir}}/README.md' ./...
