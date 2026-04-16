BIN := flagpole

build:
	go build -o $(BIN) ./src

run:
	go run ./src

test:
	go test ./src/...

clean:
	rm -f $(BIN)

.PHONY: build run test clean
