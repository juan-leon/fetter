# The go toolchain version we want to use
GOVERSION = 1.16.5

# Where we will install a modern go if none is available (note that this
# variable can be overriden in the environment, if needed)
INSTALL_PATH ?= $(HOME)/goroot

# Go commands
GOCMD = go
GOFMT = gofmt -l
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOLINT = golangci-lint run

PROJECT = github.com/juan-leon/fetter
GOBIN = bin
EXEC = bin/fetter

export PATH := $(INSTALL_PATH)/go/bin:$(HOME)/bin:$(PATH)

# PATH is not inherited by shells spawned by "shell" function
go_version := $(shell PATH=$(PATH) go version 2>/dev/null)
linter_version := $(shell PATH=$(PATH) golangci-lint --version 2>/dev/null)
now := $(shell date +'%Y-%m-%dT%T')
src := $(shell find -name '*.go')
sha := $(shell git log -1 --pretty=%H 2>/dev/null || echo unknown)

# Version can be overwritten via env var.  If not present, we figure it out from
# git.  The "word 1" is an ultra paranoid protection against spaces in tag name:
# those are not liked by the linker unless escaped.
version ?= $(word 1, $(shell git describe --abbrev --tags 2>/dev/null || echo unknown))

define install_go
	@echo Installing Go $(GOVERSION)
	mkdir -p $(INSTALL_PATH)
	curl -s https://storage.googleapis.com/golang/go$(GOVERSION).linux-amd64.tar.gz | tar -C $(INSTALL_PATH) -xz
	@echo Done installing Go $(GOVERSION)
endef

define install_linter
	@echo Installing linter
	mkdir -p $(HOME)/bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(HOME)/bin v1.41.1
	@echo Done installing linter
endef

.PHONY: clean test toolchain linter lint full-dist quick-dist


build: $(EXEC)

$(EXEC): $(src)
	$(GOBUILD) \
		--ldflags "-X main.Commit=$(sha) -X main.BuildDate=$(now) -X main.Version=$(version)" \
		-o $(EXEC) \
		github.com/juan-leon/fetter

full-dist: $(src)
	goreleaser build --rm-dist

quick-dist: $(src)
	goreleaser build --skip-validate --rm-dist --single-target

clean:
	rm -f $(EXEC)

# Format source code files
fmt: toolchain
	$(GOFMT) -w .

# Prints the source code files poorly formatted
lint:
	@echo Linting code
	$(GOLINT)

# Run tests
test:
	$(GOTEST) -coverprofile=.coverage.out $(PROJECT)/...
	@echo Code coverage
	@go tool cover -func=.coverage.out | tail -n 1
	@echo "Use 'go tool cover -html=.coverage.out' to inspect results"

# Make sure we have go installed, or install it otherwise
toolchain:
ifeq (, $(findstring $(GOVERSION), $(go_version)))
	$(call install_go)
endif

# Make sure we have go 1.24 installed, or install it otherwise
linter: toolchain
ifeq (, $(findstring 1.41, $(linter_version)))
	$(call install_linter)
endif
