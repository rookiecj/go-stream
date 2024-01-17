help:  ## show this help
	@# '*' is a GNU extension
	@grep $(MAKEFILE_LIST) -e '^[a-zA-Z0-9_-]*:.*##' | \
	sed 's/^\(.*\):.*##\(.*\)/\1\t\2/'

tidy:  ## update deps
	go mod tidy

build: ## build
	go build $(shell go list ./... | grep -v /exampl | grep -v /cmd)


lint:	## static analysis
	@go fmt $(shell go list ./stream/...)
	@go vet $(shell go list ./stream/...)


clean: 	## clean
	-rm store.test
	go clean -cache

test:	## test with coverage
	go test -v -timeout=10s $(shell go list ./... | grep -v /example | grep -v /cmd)

coverage:	## test with coverage
	#go test --converage ./store/...
	go test -coverprofile=coverage.txt -covermode=atomic -v -count=1 -timeout=30s -parallel=4 -failfast $(shell go list ./... | grep -v /example  | grep -v /cmd)

test-race:	## test race
	go test ./stream/... --race

