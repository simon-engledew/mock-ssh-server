MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.SUFFIXES:

.PHONY: example
example:
	go run . example.star