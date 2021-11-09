.DEFAULT_GOAL := default

.PHONY: default
default: build

.PHONY: ci-run
ci-run: test lint build


.PHONY: test
test:
	gotestsum ./...

.PHONY: test-verbose
test-verbose:
	gotestsum --format standard-verbose ./...

.PHONY: benchmark
benchmark:
	go test -run="none" -bench=. ./...

.PHONY: coverage
coverage:
	go test -coverpkg=./... -coverprofile=profile.cov ./... && go tool cover -func profile.cov && rm profile.cov

.PHONY: install-gotestsum
install-gotestsum:
	go get github.com/gotestyourself/gotestsum

.PHONY: install-linter
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.41.1

.PHONY: lint
lint:
	golangci-lint run

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
