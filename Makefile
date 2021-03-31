all: clean test build

build:
	go build -o output/tribo cmd/tribo/main.go

clean:
	rm -f output/*

test:
	go test ./...
