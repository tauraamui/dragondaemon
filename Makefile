.DEFAULT_GOAL := default

.PHONY: default
default: build

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	mkdir -p builds && go build -o ./builds/dragond ./cmd/dragondaemon/

.PHONY: build-install-start
build-install-start: build
	./builds/dragond stop && ./builds/dragond remove && ./builds/dragond install && ./builds/dragond start

.PHONY: run-build
run-build: build
	./builds/dragond
