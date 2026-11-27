.PHONY: run build test cover fmt vet check swagger

run:
	go run ./cmd/api

build:
	go build ./...

test:
	go test ./...

cover:
	go test ./internal/service/... -cover

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt vet test build

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g main.go -d cmd/api,internal/handler,internal/order -o docs --parseInternal --outputTypes json
