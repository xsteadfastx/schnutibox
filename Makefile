.PHONY: build
build:
	goreleaser build --rm-dist --snapshot

.PHONY: release
release:
	goreleaser release --rm-dist --snapshot --skip-publish

.PHONY: generate
generate:
	go generate

.PHONY: lint
lint: lint-go lint-buf lint-js

.PHONY: lint-go
lint-go:
	golangci-lint run --timeout 10m --enable-all --disable=exhaustivestruct,godox

.PHONY: lint-buf
lint-buf:
	buf lint -v

.PHONY: lint-js
lint-js:
	eslint assets/web/files/**.js

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: test-integration
test-integration: build
	go test -v -tags=integration -timeout=240m

.PHONY: test-all
test-all: test test-integration

.PHONY: readme
readme:
	goreadme -title revive > README.md

.PHONY: tidy
tidy:
	go mod tidy
	go mod vendor

.PHONY: build-image
build-image:
	./scripts/build.sh

.PHONY: install-tools
install-tools:
	  @cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: grpc-gen
grpc-gen:
	buf generate -v

.PHONY: air
air:
	air

.PHONY: drone-local
drone-local:
	drone exec --trusted --timeout 5h

.PHONY: clean
clean:
	rm -rf node_modules
	rm -rf package* || exit 0
