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
	buf lint -v

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: test-integration
test-integration: release
	go test -v -tags=integration -timeout=120m

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
	sudo ./scripts/build.sh

.PHONY: install-tools
install-tools:
	go install -v \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/bufbuild/buf/cmd/buf \
		github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking \
		github.com/bufbuild/buf/cmd/protoc-gen-buf-lint

.PHONY: grpc-gen
grpc-gen:
	buf beta mod update
	buf generate -v
