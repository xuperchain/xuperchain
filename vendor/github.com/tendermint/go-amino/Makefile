GOTOOLS = \
	github.com/golang/dep/cmd/dep  \
	gopkg.in/alecthomas/gometalinter.v2
GOTOOLS_CHECK = dep gometalinter.v2

all: check_tools get_vendor_deps test metalinter

########################################
###  Build

build:
	# Nothing to build!

install:
	# Nothing to install!


########################################
### Tools & dependencies

check_tools:
	@# https://stackoverflow.com/a/25668869
	@echo "Found tools: $(foreach tool,$(GOTOOLS_CHECK),\
        $(if $(shell which $(tool)),$(tool),$(error "No $(tool) in PATH")))"

get_tools:
	@echo "--> Installing tools"
	go get -u -v $(GOTOOLS)
	@gometalinter.v2 --install

update_tools:
	@echo "--> Updating tools"
	@go get -u $(GOTOOLS)

get_vendor_deps:
	@rm -rf vendor/
	@echo "--> Running dep ensure"
	@dep ensure


########################################
### Testing

test:
	go test $(shell go list ./... | grep -v vendor)

gofuzz_binary:
	rm -rf tests/fuzz/binary/corpus/
	rm -rf tests/fuzz/binary/crashers/
	rm -rf tests/fuzz/binary/suppressions/
	go run tests/fuzz/binary/init-corpus/main.go --corpus-parent=tests/fuzz/binary
	go-fuzz-build github.com/tendermint/go-amino/tests/fuzz/binary
	go-fuzz -bin=./fuzz_binary-fuzz.zip -workdir=tests/fuzz/binary

gofuzz_json:
	rm -rf tests/fuzz/json/corpus/
	rm -rf tests/fuzz/json/crashers/
	rm -rf tests/fuzz/json/suppressions/
	go-fuzz-build github.com/tendermint/go-amino/tests/fuzz/json
	go-fuzz -bin=./fuzz_json-fuzz.zip -workdir=tests/fuzz/json


########################################
### Formatting, linting, and vetting

fmt:
	@go fmt ./...

metalinter:
	@echo "==> Running linter"
	gometalinter.v2 --vendor --deadline=600s --disable-all  \
		--enable=deadcode \
		--enable=goconst \
		--enable=goimports \
		--enable=gosimple \
		--enable=ineffassign \
		--enable=megacheck \
		--enable=misspell \
		--enable=staticcheck \
		--enable=safesql \
		--enable=structcheck \
		--enable=unconvert \
		--enable=unused \
		--enable=varcheck \
		--enable=vetshadow \
		./...

		#--enable=maligned \
		#--enable=gas \
		#--enable=aligncheck \
		#--enable=dupl \
		#--enable=errcheck \
		#--enable=gocyclo \
		#--enable=golint \ <== comments on anything exported
		#--enable=gotype \
		#--enable=interfacer \
		#--enable=unparam \
		#--enable=vet \

metalinter_all:
	protoc $(INCLUDE) --lint_out=. types/*.proto
	gometalinter.v2 --vendor --deadline=600s --enable-all --disable=lll ./...


test_golang1.10rc:
	docker run -it -v "$(CURDIR):/go/src/github.com/tendermint/go-amino" -w "/go/src/github.com/tendermint/go-amino" golang:1.10-rc /bin/bash -ci "make get_tools all"

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: build install check_tools get_tools update_tools get_vendor_deps test fmt metalinter metalinter_all
