.PHONY: clean test build cover

all: bin/blackbox_decode

clean:
	rm -rf bin/blackbox_decode

bin/blackbox_decode: src/exporter/blackbox_decode/main.go
	go build -i -o $@ src/exporter/blackbox_decode/*.go

run: clean bin/blackbox_decode
	./bin/blackbox_decode

test: all
	V=$(V) go test -cover ./...

benchmark: all
	V=$(V) go test -count=5 -run=- -bench=. -test.benchmem ./...

cover:
	@rm -rf coverage.txt
	@for d in `go list ./...`; do \
		t=$$(date +%s); \
		go test -coverprofile=cover.out -covermode=atomic $$d || exit 1; \
		echo "Coverage test $$d took $$(($$(date +%s)-t)) seconds"; \
		if [ -f cover.out ]; then \
			cat cover.out >> coverage.txt; \
			rm cover.out; \
		fi; \
	done
	@echo "Uploading coverage results..."
	@curl -s https://codecov.io/bash | bash