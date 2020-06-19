MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.SUFFIXES:

EXAMPLES := $(wildcard examples/*)

.PHONY: help
help:
	@echo targets: $(EXAMPLES)

.PHONY: $(EXAMPLES)
$(EXAMPLES):
	go run . $@

.PHONY: test
test:
	go test ./pkg