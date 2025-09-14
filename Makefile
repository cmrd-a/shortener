build:
	go build -o ./bin/shortener -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$$(date +'%Y/%m/%d %H:%M:%S')' -X main.buildCommit=$$(git rev-parse HEAD)" ./cmd/shortener

build-staticlint:
	go build -o ./bin/staticlint ./cmd/staticlint

staticlint: build-staticlint
	./bin/staticlint ./...

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

check: staticlint build tidy format lint test cover

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
	go install github.com/golang/mock/mockgen@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/cmd/goimports@latest