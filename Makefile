.PHONY: clean
clean: 
	if [ -d "bin" ]; then rm -r bin; fi

build: clean
	go build -o bin/exchange

start: build
	./bin/exchange

test:
	@go test -v ./...
