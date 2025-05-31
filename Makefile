build:
	go build -o ./bin/shortener ./cmd/shortener

test:
	go test ./... -v

generate:
	go generate ./...
	make format

format:
	goimports -l -w .

lint:
	go vet ./...
	staticcheck -checks=all,-ST1000, ./...

tidy:
	go mod tidy

check: build tidy format lint test

cover:
	go test ./... -coverpkg='./internal/...', -coverprofile coverage.out.tmp
	cat coverage.out.tmp | grep -v "_easyjson.go\|mocks" > coverage.out
	rm coverage.out.tmp

cover-html: cover
	go tool cover -html=coverage.out

cover-cli: cover
	go tool cover -func=coverage.out

mock:
	mockgen -destination=internal/storage/mocks/mock_repository.go -package=mocks github.com/cmrd-a/shortener/internal/storage Repository