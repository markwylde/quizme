BINARY := interrogate
GO ?= go

# The plugin payload is the authoritative skill; the copy under .claude/ exists
# only so this repository's own agents see it without installing the plugin.
SKILL_SRC := skills/interrogate/SKILL.md
SKILL_COPY := .claude/skills/interrogate/SKILL.md

.PHONY: build test run fmt vet clean install skill

build:
	$(GO) build -o $(BINARY) .

test:
	$(GO) test ./...

run:
	@test -n "$(FILE)" || (echo "usage: make run FILE=path/to/questionnaire.yaml" >&2; exit 2)
	$(GO) run . $(FILE)

# skill regenerates the local copy from the authoritative one.
skill:
	@mkdir -p $(dir $(SKILL_COPY))
	@cp $(SKILL_SRC) $(SKILL_COPY)
	@echo "regenerated $(SKILL_COPY) from $(SKILL_SRC)"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

install:
	$(GO) install .

clean:
	rm -f $(BINARY)
