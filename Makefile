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
	golangci-lint run

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: readme
readme:
	goreadme -title revive > README.md
