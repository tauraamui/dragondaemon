.DEFAULT_GOAL := default

.PHONY: default
default: build

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	mkdir -p builds && go build -o ./builds/dragond ./cmd/dragondaemon/

.PHONY: run-build
run-build: build
	./builds/dragond
