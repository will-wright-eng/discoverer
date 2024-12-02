# Build configuration
BINARY_NAME=discoverer
INSTALL_PATH=/usr/local/bin
LOG_DIR=/usr/local/var/log
LAUNCH_AGENTS_DIR=$(HOME)/Library/LaunchAgents
PLIST_NAME=com.username.$(BINARY_NAME).plist

# Go settings
GOBASE=$(shell pwd)
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

#* Setup
.PHONY: $(shell sed -n -e '/^$$/ { n ; /^[^ .\#][^ ]*:/ { s/:.*$$// ; p ; } ; }' $(MAKEFILE_LIST))
.DEFAULT_GOAL := help

help: ## list make commands
	@echo ${MAKEFILE_LIST}
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

#* Go Commands
all: build ## build binary

build: ## build binary
	$(GOBUILD) -o $(GOBASE)/dist/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

clean: ## remove binary and logs
	rm -f $(GOBASE)/dist/$(BINARY_NAME)
	rm -f $(LOG_DIR)/$(BINARY_NAME).log
	rm $(HOME)/.config/$(BINARY_NAME)/$(BINARY_NAME).db

deps: ## download dependencies
	$(GOMOD) download
	$(GOGET) github.com/mattn/go-sqlite3

test: ## run tests
	$(GOTEST) -v ./...

run: ## run monitor using main.go
	$(GORUN) $(GOBASE)/cmd/$(BINARY_NAME)/main.go

fmt: ## format code
	find . -name "*.go" -exec gofmt -w {} +

#* Install Commands
install: build ## install binary and LaunchAgent
	# Create necessary directories
	sudo mkdir -p $(INSTALL_PATH)
	sudo mkdir -p $(LOG_DIR)

	# Install binary
	sudo cp $(GOBASE)/dist/$(BINARY_NAME) $(INSTALL_PATH)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)

	# Install and load LaunchAgent
	cp $(PLIST_NAME) $(LAUNCH_AGENTS_DIR)/
	launchctl unload $(LAUNCH_AGENTS_DIR)/$(PLIST_NAME) 2>/dev/null || true
	launchctl load $(LAUNCH_AGENTS_DIR)/$(PLIST_NAME)

uninstall: ## uninstall binary and LaunchAgent
	# Unload and remove LaunchAgent
	launchctl unload $(LAUNCH_AGENTS_DIR)/$(PLIST_NAME) 2>/dev/null || true
	rm -f $(LAUNCH_AGENTS_DIR)/$(PLIST_NAME)

	# Remove binary and logs
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	sudo rm -f $(LOG_DIR)/$(BINARY_NAME).log

#* LaunchAgent Commands
start: ## start LaunchAgent
	launchctl start $(PLIST_NAME)

stop: ## stop LaunchAgent
	launchctl stop $(PLIST_NAME)

status: ## check LaunchAgent status
	launchctl list | grep $(PLIST_NAME)

logs: ## tail logs
	tail -f $(LOG_DIR)/$(BINARY_NAME).log

#* Docker Commands
docker-kill: ## kill all docker containers
	for id in $$(docker ps --format "{{.ID}}"); do docker kill $$id; done

up: ## start app with `docker-compose up --build`
	docker-compose up --build --remove-orphans
