.PHONY: gen-proto install-proto-tools test

PROTO_SRC=grpc/proto
PB_OUT=grpc/pb

gen-proto:
	@mkdir -p $(PB_OUT)/auth/v1 $(PB_OUT)/product/v1 $(PB_OUT)/order/v1
	export PATH=$$PATH:$$(go env GOPATH)/bin && \
	protoc \
		--proto_path=$(PROTO_SRC) \
		--go_out=$(PB_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PB_OUT) --go-grpc_opt=paths=source_relative \
		$(PROTO_SRC)/auth/v1/*.proto \
		$(PROTO_SRC)/product/v1/*.proto \
		$(PROTO_SRC)/order/v1/*.proto
	@echo "Generated Protobuf & gRPC Go files successfully in $(PB_OUT)"

install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.8
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

test:
	go test -v ./...
