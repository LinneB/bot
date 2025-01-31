default: build

build: test
	go build cmd/bot.go

test:
	go test -v ./...

run: test
	go run cmd/bot.go

fmt:
	gofmt -s -w -l .
