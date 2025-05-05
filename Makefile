build:
	go build -o ./bin/shortener ./cmd/shortener

test:
	go test ./... -v -cover -vet=all

gen:
	go generate ./...

format:
	goimports -l -w .

lint:
	go vet ./...
	staticcheck -checks=all,-ST1000, ./...

tidy:
	go mod tidy

check: build tidy format lint
	go test ./...