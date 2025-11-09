BINARY=robot-secret

export SECRET ?= "Hidden beneath the old oak tree, golden coins patiently await discovery."
export NBR_OF_ROBOTS ?= 6
export BUFFER_SIZE ?= 100
export OUTPUT_FILE ?= "secret.txt"
export PERCENTAGE_OF_LOST ?= 0
export PERCENTAGE_OF_DUPLICATED ?= 0
export DUPLICATED_NUMBER ?= 0

GO=go

.PHONY: all build run clean test

all: build

build:
	$(GO) build -o $(BINARY) .

run: build
	SECRET="$(SECRET)" \
	NBR_OF_ROBOTS="$(NBR_OF_ROBOTS)" \
	BUFFER_SIZE="$(BUFFER_SIZE)" \
	OUTPUT_FILE="$(OUTPUT_FILE)" \
	PERCENTAGE_OF_LOST="$(PERCENTAGE_OF_LOST)" \
	PERCENTAGE_OF_DUPLICATED="$(PERCENTAGE_OF_DUPLICATED)" \
	DUPLICATED_NUMBER="$(DUPLICATED_NUMBER)" \
	./$(BINARY)

test:
	$(GO) test -v ./...

clean:
	@rm -f $(BINARY) $(OUTPUT_FILE)
