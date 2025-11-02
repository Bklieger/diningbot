.PHONY: build test test-unit test-integration clean run

build:
	go build -o diningbot

test: test-unit test-integration

test-unit:
	go test ./... -cover

test-integration:
	go test -v -tags=integration -timeout 120s

clean:
	rm -f diningbot debug_*.html

run: build
	./diningbot "Branner Dining" "Lunch"

.DEFAULT_GOAL := build

