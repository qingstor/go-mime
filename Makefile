SHELL := /bin/bash

.PHONY: all check format vet lint build test generate tidy integration_test

# golint: go get -u golang.org/x/lint/golint
# go-bindata: go get -u github.com/kevinburke/go-bindata/...
tools := golint go-bindata

$(tools):
	@command -v $@ >/dev/null 2>&1 || echo "$@ is not found, plese install it."

check: vet lint

format:
	@echo "go fmt"
	@go fmt ./...
	@echo "ok"

vet:
	@echo "go vet"
	@go vet ./...
	@echo "ok"

lint: golint
	@echo "golint"
	@golint ./...
	@echo "ok"

generate_iana_media_types: go-bindata
	@pushd internal/cmd \
		&& go generate ./... \
		&& go build -o ../bin/iana ./iana \
		&& popd
	@./internal/bin/iana
	@echo "Done"

build: tidy check
	@echo "build storage"
	@go build ./...
	@echo "ok"

tidy:
	@pushd internal/cmd && go mod tidy && go mod verify && popd
	@go mod tidy && go mod verify
