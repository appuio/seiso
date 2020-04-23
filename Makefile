.PHONY: all build clean dist fmt tests

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

dist:
	@echo 'Build all distributions ...'
	goreleaser release --snapshot --rm-dist --skip-sign

clean:
	@echo 'Clean up build artifacts ...'
	rm -rf seiso dist/
