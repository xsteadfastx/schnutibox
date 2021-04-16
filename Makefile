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
	golangci-lint run --timeout 10m --enable-all --disable=exhaustivestruct

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: test-integration
test-integration: release
	go test -v -tags=integration -timeout=120m

.PHONY: readme
readme:
	goreadme -title revive > README.md

.PHONY: tidy
tidy:
	go mod tidy
	go mod vendor

.PHONY: build-image
build-image:
	sudo ./scripts/build.sh
