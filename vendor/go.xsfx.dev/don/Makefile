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
lint:
	golangci-lint run --enable-all --disable=exhaustivestruct

.PHONY: test
test:
	go test -v -race -cover -tags=integration ./... -timeout=120m

.PHONY: tidy
tidy:
	go mod tidy
	go mod vendor
