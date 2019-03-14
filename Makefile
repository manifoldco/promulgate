CMD=promulgate
VERSION?=$(shell git describe --tags --dirty | sed 's/^v//')
ZIP=$(shell go list | awk -F / '{ print $$NF }')
GO_BUILD=CGO_ENABLED=0 go build -i --ldflags="-w -X $(shell go list)/version=$(VERSION)"

rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) \
    $(filter $(subst *,%,$2),$d))

all: ci

ci: $(LINTERS) build cover

.PHONY: all ci

# ################################################
# Bootstrapping for base golang package deps
# ################################################
BOOTSTRAP=\
	github.com/golang/dep/cmd/dep \
	github.com/alecthomas/gometalinter \
	github.com/jteeuwen/go-bindata

$(BOOTSTRAP):
	go get -u $@

bootstrap: $(BOOTSTRAP)
	gometalinter --install

vendor: Gopkg.lock
	dep ensure -v -vendor-only

.PHONY: bootstrap

# ################################################
# Test and linting
# ###############################################
LINTERS=\
	gofmt \
	golint \
	gosimple \
	vet \
	misspell \
	ineffassign \
	deadcode

METALINT=gometalinter --tests --disable-all --vendor --deadline=5m -e "zz_.*\.go" \
	 ./... --enable

lint: $(LINTERS)

$(LINTERS): vendor
	$(METALINT) $@

.PHONY: $(LINTERS) lint

COVER_TEST_PKGS:=$(shell find . -type f -name '*_test.go' | grep -v vendor | grep -v generated | rev | cut -d "/" -f 2- | rev | sort -u)
$(COVER_TEST_PKGS:=-cover): %-cover: all-cover.txt
	@CGO_ENABLED=0 go test -coverprofile=$@.out -covermode=atomic ./$*
	@if [ -f $@.out ]; then \
	    grep -v "mode: atomic" < $@.out >> all-cover.txt; \
	    rm $@.out; \
	fi

all-cover.txt:
	echo "mode: atomic" > all-cover.txt

cover: vendor all-cover.txt $(COVER_TEST_PKGS:=-cover)

.PHONY: cover $(LINTERS) $(COVER_TEST_PKGS:=-cover)

# ################################################
# Building
# ###############################################$

PREFIX?=
SUFFIX=
ifeq ($(GOOS),windows)
    SUFFIX=.exe
endif

build: $(PREFIX)bin/$(CMD)$(SUFFIX)

$(PREFIX)bin/$(CMD)$(SUFFIX): vendor
	$(GO_BUILD) -o $(PREFIX)bin/$(CMD)$(SUFFIX) .

.PHONY: build

#################################################
# Releasing
#################################################

NO_WINDOWS= \
	darwin_amd64 \
	linux_amd64
OS_ARCH= \
	$(NO_WINDOWS) \
	windows_amd64

os=$(word 1,$(subst _, ,$1))
arch=$(word 2,$(subst _, ,$1))
sfx=$(patsubst windows_%,.exe,$(filter windows_%,$1))

$(OS_ARCH:%=os-build/%/bin/$(CMD)): os-build/%/bin/$(CMD):
	PREFIX=build/$*/ GOOS=$(call os,$*) GOARCH=$(call arch,$*) make build/$*/bin/$(CMD)$(call sfx,$*)

$(OS_ARCH:%=build/$(ZIP)_$(VERSION)_%.zip): build/$(ZIP)_$(VERSION)_%.zip: os-build/%/bin/$(CMD)$(call sfx,$*)
	cd build/$*/bin; zip -r ../../$(ZIP)_$(VERSION)_$*.zip $(CMD)$(call sfx,$*)
$(NO_WINDOWS:%=build/$(ZIP)_$(VERSION)_%.tar.gz): build/$(ZIP)_$(VERSION)_%.tar.gz: os-build/%/bin/$(CMD)$(call sfx,$*)
	cd build/$*/bin; tar -czf ../../$(ZIP)_$(VERSION)_$*.tar.gz $(CMD)$(call sfx,$*)

zips: $(NO_WINDOWS:%=build/$(ZIP)_$(VERSION)_%.tar.gz) build/$(ZIP)_$(VERSION)_windows_amd64.zip

.PHONY: zips $(OS_ARCH:%=os-build/%/bin/manifold)

# ################################################
# Cleaning
# ################################################

clean:
	rm -rf bin
	rm -rf build
