# init project PATH
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output
COMPILECACHEDIR := $(HOMEDIR)/.compile_cache
XVMDIR  := $(COMPILECACHEDIR)/xvm
TESTNETDIR := $(HOMEDIR)/testnet

VERSION:=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)
COMMIT_ID:=$(shell git rev-parse --short HEAD 2>/dev/null ||echo unknown)

# init command params
export GO111MODULE=on
X_ROOT_PATH := $(HOMEDIR)
export X_ROOT_PATH
export PATH := $(OUTDIR)/bin:$(XVMDIR):$(PATH)

# make, make all
all: clean compile

# make compile, go build
compile: xvm xchain
xchain:
	VERSION=$(VERSION) COMMIT_ID=$(COMMIT_ID) bash $(HOMEDIR)/auto/build.sh
prepare:
	go mod download
# make xvm
xvm:
	bash $(HOMEDIR)/auto/build_xvm.sh

# make test, test your code
test: xvm unit
unit:
	go test -coverprofile=coverage.txt -covermode=atomic ./...

# make clean
cleanall: clean cleantest cleancache
clean:
	rm -rf $(OUTDIR)
cleantest:
	rm -rf $(TESTNETDIR)
cleancache:
	rm -rf $(COMPILECACHEDIR)

# deploy test network
testnet:
	bash $(HOMEDIR)/auto/deploy_testnet.sh

# Docker related tasks
build-image:
	docker build -t xchain:dev .
# avoid filename conflict and speed up build
.PHONY: all compile test clean
