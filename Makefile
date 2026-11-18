# Piclang Makefile v2
GO      ?= go
CC      ?= gcc
PICC    := ./picc
RUNNER  := ./run_shellcode
GOFLAGS ?=

.PHONY: all build test runner examples smoke clean help

all: build runner

## build: compila el compilador picc
build:
	$(GO) build $(GOFLAGS) -o $(PICC) ./cmd/picc

## runner: compila el runner C
runner:
	$(CC) -O0 -o $(RUNNER) run_shellcode.c

## test: ejecuta la suite de tests del compilador (compila y ejecuta shellcode)
test:
	$(GO) test ./... -count=1

## examples: compila todos los examples/*/main.pic a examples/*/out.bin
examples: build
	@set -e; \
	for f in examples/*/main.pic; do \
		[ -e "$$f" ] || continue; \
		dir=$$(dirname "$$f"); \
		out="$$dir/out.bin"; \
		echo "  [picc] $$f -> $$out"; \
		$(PICC) -rand-sections=false -o "$$out" "$$f"; \
	done

## smoke: compila y ejecuta cada example, reporta PASS/FAIL
smoke: build runner
	@./run_all_examples.sh

## clean: borra binarios generados
clean:
	rm -f $(PICC) $(RUNNER)
	rm -f examples/*/out.bin
	rm -f out.bin hello.bin bootstrap.bin

## help: muestra esta ayuda
help:
	@sed -n 's/^## //p' $(MAKEFILE_LIST)
