.PHONY: build

build:
	make clean
	go build -ldflags "-s -w" -o build/auto-deb ./src

build-linux:
	make clean
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o build/auto-deb-linux-amd64 ./src

deps:
	go mod tidy

clean:
	rm -rf build
	go clean
