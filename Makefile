.PHONY: clean test build

all: bin/blackbox_decode

clean:
	rm -rf bin/blackbox_decode

bin/blackbox_decode: src/exporter/blackbox_decode/main.go
	go build -i -o $@ src/exporter/blackbox_decode/*.go

run: clean bin/blackbox_decode
	./bin/blackbox_decode

test: all
	V=$(V) go test -cover ./...
