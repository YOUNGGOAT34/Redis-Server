BIN = cachedb

build:
	go build -o $(BIN) ./app

run:
	./$(BIN) --port 6379

replica:
	./$(BIN) --port 6380 --replicaof "localhost 6379"
test:
	go test ./app/tester -count=1

test-v:
	go test ./app/tester -count=1 -v

test-race:
	go test -race ./app/tester -count=1

test-all:
	go build ./...
	go test ./... -count=1

clean:
	rm -f $(BIN)

.PHONY: build run replica test test-race test-all clean
