EXAMPLES = advertise alive bye monitor search

default: test

test:
	go test -v ./...

test-full:
	go test -v -race .

lint:
	go vet ./...
	@echo ""
	golint ./...

cyclo:
	-gocyclo -top 10 -avg .

report:
	@echo "misspell"
	@find . -name "*.go" | xargs misspell
	@echo ""
	-gocyclo -over 14 -avg .
	@echo ""
	go vet ./...
	@echo ""
	golint ./...

deps:
	go get -v -u -d -t ./...

tags:
	gotags -f tags -R .
.PHONY: tags

clean: examples-clean

examples: examples-build

examples-build: $(EXAMPLES)

examples-clean:
	rm -f $(EXAMPLES)

advertise: examples/advertise/*.go *.go
	go build ./examples/advertise

alive: examples/alive/*.go *.go
	go build ./examples/alive

bye: examples/bye/*.go *.go
	go build ./examples/bye

monitor: examples/monitor/*.go *.go
	go build ./examples/monitor

search: examples/search/*.go *.go
	go build ./examples/search

.PHONY: test test-full lint cyclo report deps clean \
	examples examples-build examples-clean
