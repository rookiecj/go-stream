help:  ## show this help
	@# '*' is a GNU extension
	@grep $(MAKEFILE_LIST) -e '^[a-zA-Z0-9_-]*:.*##' | \
	sed 's/^\(.*\):.*##\(.*\)/\1\t\2/'

build:	## build
	go build -v ./...
test:	## test with coverage
	go test ./stream/... -coverprofile=coverage.txt -covermode=atomic -v -count=1 -timeout=30s -parallel=4 -failfast 

test-race:	## test race
	go test ./stream/... --race

lint:	## static analysis
	@go fmt $(shell go list ./stream/...)
	@go vet $(shell go list ./stream/...)
