BIN := flagpole

build:
	go build -o $(BIN) ./src

run:
	go run ./src

test:
	go test ./src/...

clean:
	rm -f $(BIN)

swag:
	$(shell go env GOPATH)/bin/swag init -g src/main.go --output src/docs

.PHONY: build run test clean swag
