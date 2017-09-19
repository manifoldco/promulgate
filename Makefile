CMD=promulgate
VERSION?=$(shell git describe --tags --dirty | sed 's/^v//')
ZIP=$(shell go list | awk -F / '{ print $$NF }')
GO_BUILD=CGO_ENABLED=0 go build -i --ldflags="-w -X $(shell go list)/version=$(VERSION)"

rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) \
    $(filter $(subst *,%,$2),$d))

LINTERS=\
	gofmt \
	golint \
	vet \
	misspell \
	ineffassign \
	deadcode

all: ci

ci: $(LINTERS) build

.PHONY: all ci

# ################################################
# Bootstrapping for base golang package deps
# ################################################

CMD_PKGS=\
	github.com/golang/lint/golint \
	github.com/dominikh/go-tools/simple \
	github.com/client9/misspell/cmd/misspell \
	github.com/gordonklaus/ineffassign \
	github.com/tsenart/deadcode \
	github.com/alecthomas/gometalinter

define VENDOR_BIN_TMPL
vendor/bin/$(notdir $(1)): vendor
	go build -o $$@ ./vendor/$(1)
VENDOR_BINS += vendor/bin/$(notdir $(1))
endef

$(foreach cmd_pkg,$(CMD_PKGS),$(eval $(call VENDOR_BIN_TMPL,$(cmd_pkg))))
$(patsubst %,%-bin,$(filter-out gofmt vet,$(LINTERS))): %-bin: vendor/bin/%
gofmt-bin vet-bin:

bootstrap:
	which dep || go get -u github.com/golang/dep/cmd/dep

vendor: Gopkg.lock
	dep ensure

.PHONY: bootstrap $(CMD_PKGS)

# ################################################
# Test and linting
# ###############################################

$(LINTERS): %: vendor/bin/gometalinter %-bin vendor
	PATH=`pwd`/vendor/bin:$$PATH gometalinter --tests --disable-all --vendor \
	    --deadline=5m -s data ./... --enable $@

.PHONY: $(LINTERS)

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

OS_ARCH= \
	darwin_amd64 \
	linux_amd64 \
	windows_amd64

os=$(word 1,$(subst _, ,$1))
arch=$(word 2,$(subst _, ,$1))
sfx=$(patsubst windows_%,.exe,$(filter windows_%,$1))

$(OS_ARCH:%=os-build/%/bin/$(CMD)): os-build/%/bin/$(CMD):
	PREFIX=build/$*/ GOOS=$(call os,$*) GOARCH=$(call arch,$*) make build/$*/bin/$(CMD)$(call sfx,$*)

$(OS_ARCH:%=build/$(ZIP)_$(VERSION)_%.zip): build/$(ZIP)_$(VERSION)_%.zip: os-build/%/bin/$(CMD)$(call sfx,$*)
	cd build/$*/bin; zip -r ../../$(ZIP)_$(VERSION)_$*.zip $(CMD)$(call sfx,$*)

zips: $(OS_ARCH:%=build/$(ZIP)_$(VERSION)_%.zip)

.PHONY: zips $(OS_ARCH:%=os-build/%/bin/manifold)

# ################################################
# Cleaning
# ################################################

clean:
	rm -rf bin
	rm -rf build
