NAME := dcos-services
REVISION := $(shell git describe --tags --always --dirty="-dev")

export GO15VENDOREXPERIMENT=1

LDFLAGS := -X github.com/dcos/dcos-oauth/version.REVISION=$(REVISION)

install:
	go install -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' ./...

save:
	godep save ./...

docker:
	docker build -t $(NAME) .

test: docker
	go test ./...

.PHONY: install save docker test

DIRS=$(subst $(space),$(newline),$(shell go list ./... | grep -v /vendor/))
TEST=$(subst $(space),$(newline),$(shell go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.Dir}}{{end}}' ./...))
NOTEST=$(filter-out $(TEST),$(DIRS))

test-compile: $(addsuffix .test-compile, $(TEST))

%.test-compile:
	cd $* && go test -p 1 -v -c .
