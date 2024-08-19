default: build

build: test
	go build

test:
	go test -v ./...

run:
	go run main.go

fmt:
	gofmt -s -w -l .
