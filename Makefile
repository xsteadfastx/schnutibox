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

.PHONY: install-tools
install-tools:
	go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc

.PHONY: protoc-gen
protoc-gen:
	protoc \
		--proto_path=api/proto/v1 \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_out=pkg/api/v1 \
		--go_out=pkg/api/v1 \
		--grpc-gateway_opt=logtostderr=true \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=generate_unbound_methods=true \
		--grpc-gateway_out=pkg/api/v1 \
		api/proto/v1/schnutibox.proto
