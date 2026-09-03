BINARY := interrogate
GO ?= go

# The plugin payload is the authoritative skill; the copy under .claude/ exists
# only so this repository's own agents see it without installing the plugin.
SKILL_SRC := skills/interrogate/SKILL.md
SKILL_COPY := .claude/skills/interrogate/SKILL.md

.PHONY: build test run fmt vet clean install skill icon bundle

build:
	$(GO) build -o $(BINARY) .

test:
	$(GO) test ./...

run:
	@test -n "$(FILE)" || (echo "usage: make run FILE=path/to/questionnaire.yaml" >&2; exit 2)
	$(GO) run . $(FILE)

# icon draws the raster the platform bundlers want from the drawing that is the
# real source. The window icon needs no such thing: the binary carries the SVG.
icon: icon.png

icon.png: icon.svg internal/tools/icon/main.go
	$(GO) run ./internal/tools/icon -size 1024

# bundle builds Interrogate.app, which is how macOS is given an icon: it takes
# the Dock tile from an application bundle, never from a bare binary. Run it
# with a questionnaire like so:
#
#	open -a ./Interrogate.app --args path/to/questions.yaml
bundle: icon.png
	$(GO) run fyne.io/tools/cmd/fyne@latest package -os darwin -icon icon.png -name Interrogate

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
	rm -f $(BINARY) icon.png
	rm -rf Interrogate.app
