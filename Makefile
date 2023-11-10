help:  ## show this help
	@# '*' is a GNU extension
	@grep $(MAKEFILE_LIST) -e '^[a-zA-Z0-9_-]*:.*##' | \
	sed 's/^\(.*\):.*##\(.*\)/\1\t\2/'

test:	## test
	go test ./stream/...


lint:	## static analysis
	@go fmt $(shell go list ./stream/...)
	@go vet $(shell go list ./stream/...)
