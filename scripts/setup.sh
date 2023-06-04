#!/bin/bash

go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/segmentio/golines@latest

cp scripts/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

