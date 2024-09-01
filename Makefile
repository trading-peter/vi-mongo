BUILD_DIR := .build
SVC_NAME := vi-mongo

.PHONY: build run

all: build run

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(SVC_NAME) .

run:
	env $$(cat .env) $(BUILD_DIR)/$(SVC_NAME)

test:
	go test -v ./...

debug:
	if [ -f /proc/sys/kernel/yama/ptrace_scope ]; then \
		sudo sysctl kernel.yama.ptrace_scope=0; \
	fi
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(SVC_NAME) .
	$(BUILD_DIR)/$(SVC_NAME)

lint:
	golangci-lint run

release: check-version
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

check-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is not set"; \
		exit 1; \
	fi
