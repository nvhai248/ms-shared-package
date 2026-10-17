# ms-shared-package

Shared core library and Protobuf / gRPC schemas for Go microservices.

## Structure

```
├── Makefile                # Protoc code generation and test helpers
├── common
│   ├── apperror            # Domain error models, i18n support, and HTTP <-> gRPC code mapping
│   ├── ctx                 # Context propagation (X-Request-ID, user metadata, locale)
│   └── response            # Uniform generic HTTP JSON response schema
└── grpc
    ├── grpcclient          # HTTP/2 gRPC client with keepalive, pooling, and tracing interceptors
    ├── pb                  # Generated Go code from Protobuf (.pb.go, _grpc.pb.go)
    │   ├── auth/v1
    │   ├── order/v1
    │   └── product/v1
    └── proto               # Protobuf definition files (.proto)
        ├── auth/v1
        ├── order/v1
        └── product/v1
```

## How to Compile Protobuf

Ensure `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` are installed:
```bash
make install-proto-tools
```

Generate protobuf files:
```bash
make gen-proto
```

## Local Development with `replace` in `go.mod`

In services consuming this shared package before git tagging:
```go
replace github.com/nvhai248/ms-shared-package => /Users/hainguyenvan/Desktop/PW/ms-shared-package
```
