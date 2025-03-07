
tidy:
	go mod tidy

install:
	go mod download

test:
	go test -race -v ./...

lint:
	go vet ./...

build:
	go build ./...
