.PHONY: run build test test-race bench vet fmt clean

run:
	go run ./cmd/gocachedb

build:
	go build -o gocachedb ./cmd/gocachedb

test:
	go test ./...

test-race:
	go test -race ./...

bench:
	go test ./internal/store/... -bench=. -benchmem -run=^$$

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -f gocachedb dump.rdb
