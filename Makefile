GO ?= $(shell command -v go 2>/dev/null)

ifeq ($(GO),)
GO := mise exec -- go
endif

.PHONY: test

test:
	$(GO) test ./...
