.PHONY: all build clean fmt tests

all: clean fmt tests build

fmt:
	@echo 'Reformat Go code ...'
	find . -type f -name '*.go' -exec go fmt {} \;

tests:
	@echo 'Run all tests ...'
	go test ./...

build:
	@echo 'Build seiso binary ...'
	go build

clean:
	@echo 'Clean up build artifacts ...'
	rm -f seiso
