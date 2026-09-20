.PHONY: build test run start fmt version

build:
	go build -o bin/vun ./cmd/agent

version:
	go run ./cmd/agent version

start:
	go run ./cmd/agent start -config configs/agent.example.yaml

test:
	go test ./...

run:
	go run ./cmd/agent -config configs/agent.example.yaml

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')
