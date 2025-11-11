BINARY=robot-secret

export SECRET ?= "Hidden beneath the old oak tree, golden coins patiently await discovery."
export NBR_OF_ROBOTS ?= 6
export BUFFER_SIZE ?= 10
export END_OF_SECRET ?= "."
export OUTPUT_FILE ?= "secret.txt"
export PERCENTAGE_OF_LOST ?= 0
export PERCENTAGE_OF_DUPLICATED ?= 0
export DUPLICATED_NUMBER ?= 0
export MAX_ATTEMPTS ?= 50
export TIMEOUT ?= 10s
export QUIET_PERIOD ?= 5s

GO=go

.PHONY: all build run clean test

all: build

build:
	$(GO) build -o $(BINARY) .

run: build
	SECRET="$(SECRET)" \
	NBR_OF_ROBOTS="$(NBR_OF_ROBOTS)" \
	BUFFER_SIZE="$(BUFFER_SIZE)" \
	END_OF_SECRET="$(END_OF_SECRET)" \
	OUTPUT_FILE="$(OUTPUT_FILE)" \
	PERCENTAGE_OF_LOST="$(PERCENTAGE_OF_LOST)" \
	PERCENTAGE_OF_DUPLICATED="$(PERCENTAGE_OF_DUPLICATED)" \
	DUPLICATED_NUMBER="$(DUPLICATED_NUMBER)" \
	MAX_ATTEMPTS="$(MAX_ATTEMPTS)" \
	TIMEOUT="$(TIMEOUT)" \
	QUIET_PERIOD="$(QUIET_PERIOD)" \
	./$(BINARY)

test:
	$(GO) test -v ./...

clean:
	@rm -f $(BINARY) $(OUTPUT_FILE)
