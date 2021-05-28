.DEFAULT_GOAL := default

.PHONY: default
default: build

.PHONY: test
test:
	go test -v ./...

.PHONY: watch
watch:
	ginkgo watch ./...

.PHONY: build
build:
	mkdir -p builds && go build -o ./builds/dragond ./cmd/dragondaemon/

.PHONY: build-with-mat-profile
build-with-mat-profile:
	mkdir -p builds && go build -tags matprofile -o ./builds/dragond ./cmd/dragondaemon/

.PHONY: build-install-start
build-install-start: build
	sudo ./builds/dragond stop && sudo ./builds/dragond remove && sudo ./builds/dragond install && sudo ./builds/dragond start

.PHONY: run-build
run-build: build
	./builds/dragond

.PHONY: run-build-mat-profile
run-build-mat-profile: build-with-mat-profile
	./builds/dragond
