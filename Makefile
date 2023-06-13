setup:
	@echo "Setting up the environment"
	@./scripts/setup.sh

cibuild:
	./scripts/cibuild.sh

#####################################

BINARY=goinit
SRC=./main.go
BIN_DIR=./bin
.DEFAULT_GOAL := build
BUILD_CMD=CGO_ENABLED=0 go build -mod=readonly -ldflags="-s -w" -gcflags=all=-l -trimpath=true

build:
	@$(BUILD_CMD) -o $(BIN_DIR)/$(BINARY) $(SRC)

run: build
	$(BIN_DIR)/$(BINARY)

test:
	go test ./... -v

clean:
	go clean
	rm -rf $(BIN_DIR)

