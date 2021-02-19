all: clean build

build:
	go build -o output/tribo cmd/tribo/main.go

clean:
	rm output/*
