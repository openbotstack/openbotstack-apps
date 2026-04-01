.PHONY: build test test-v lint clean demo

build:
	go build ./...

test:
	go test ./...

test-v:
	go test -v ./...

lint:
	go vet ./...

clean:
	go clean ./...

demo:
	go run ./apps/demo
