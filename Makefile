.PHONY: all build clean fmt

all: fmt build

fmt:
	@echo 'Reformat Go code ...'
	find . -type f -name '*.go' -exec go fmt {} \;

build:
	@echo 'Build seiso binary ...'
	go build

clean:
	@echo 'Clean up build artifacts ...'
	rm -f seiso
