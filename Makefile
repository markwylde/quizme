BINARY := interrogate
GO ?= go

.PHONY: build test run fmt vet clean install

build:
	$(GO) build -o $(BINARY) .

test:
	$(GO) test ./...

run:
	@test -n "$(FILE)" || (echo "usage: make run FILE=path/to/questionnaire.yaml" >&2; exit 2)
	$(GO) run . $(FILE)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

install:
	$(GO) install .

clean:
	rm -f $(BINARY)
