build:
	go build -o ./bin/shortener ./cmd/shortener

test:
	go test -cover ./...
	go test ./... -v

gen:
	go generate ./...
	mockgen -destination=internal/storage/storage_mocks/mock_repository.go -package=storage_mocks github.com/cmrd-a/shortener/internal/storage Repository
	mockgen -destination=internal/service/service_mocks/mock_generator.go -package=service_mocks github.com/cmrd-a/shortener/internal/service Generator
	make format

format:
	goimports -l -w .

lint:
	go vet ./...
	staticcheck -checks=all,-ST1000, ./...

tidy:
	go mod tidy

check: build tidy format lint test cover

cover:
	go test -cover ./...
	go test ./... -coverpkg='./internal/...', -coverprofile coverage.out.tmp
	cat coverage.out.tmp | grep -v "_easyjson.go\|mocks" > coverage.out
	rm coverage.out.tmp


cover-html: cover
	go tool cover -html=coverage.out

cover-cli: cover
	go tool cover -func=coverage.out

install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

