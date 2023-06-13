#!/bin/bash

# Check for Go
if ! command -v go &> /dev/null
then
    echo "Go could not be found. Please install it first."
    exit 1
fi

go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/segmentio/golines@latest
golangci-lint run
