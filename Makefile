# init project PATH
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output
COMPILECACHEDIR := $(HOMEDIR)/.compile_cache
XVMDIR  := $(COMPILECACHEDIR)/xvm
TESTNETDIR := $(HOMEDIR)/testnet
LICENSEEYE   := license-eye
GOINSTALL    := go install
PIP          := pip3
PIPINSTALL   := $(PIP) install


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
	go env -w GOPROXY=https://goproxy.cn,direct
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

# make deps
deps:
	$(call PIP_INSTALL_PKG, pre-commit)
	$(call INSTALL_PKG, license-eye, github.com/apache/skywalking-eyes/cmd/license-eye@latest)

# go install package
# $(1) package name
# $(2) package address
define INSTALL_PKG
	@echo installing $(1)
	$(GOINSTALL) $(2)
	@echo $(1) installed
endef

define PIP_INSTALL_PKG
	@echo installing $(1)
	$(PIPINSTALL) $(1)
	@echo $(1) installed
endef

# make license-check, check code file's license declaration
license-check:
	$(LICENSEEYE) header check

# make license-fix, fix code file's license declaration
license-fix:
	$(LICENSEEYE) header fix

# avoid filename conflict and speed up build
.PHONY: all compile test clean
