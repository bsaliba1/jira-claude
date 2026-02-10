.PHONY: build clean test run

BINARY_NAME=jira-claude
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

install:
	go install .

tidy:
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run
