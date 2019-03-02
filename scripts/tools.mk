GO := go
DOCKER := docker
GIT := git
DIVE := dive
DOCKERBUILD := docker build
DOCKERPUSH := docker push

COMMIT = $(shell $(GIT) rev-parse HEAD 2>/dev/null)
BRANCH = $(shell $(GIT) rev-parse --abbrev-ref HEAD 2>/dev/null)
VERSION = $(shell $(GIT) describe --tags 2>/dev/null)

ifeq ($(COMMIT),HEAD)
	COMMIT = none
else ifeq ($(COMMIT),)
	COMMIT = none
endif

ifeq ($(BRANCH),HEAD)
	BRANCH = none
else ifeq ($(BRANCH),)
	BRANCH = none
endif

ifeq ($(VERSION),)
	VERSION = none
endif

LDFLAGS := \
			-X arhat.dev/kube-host-pty/pkg/version.branch=$(BRANCH) \
			-X arhat.dev/kube-host-pty/pkg/version.commit=$(COMMIT) \
			-X arhat.dev/kube-host-pty/pkg/version.version=$(VERSION)

BUILD_ARGS := -ldflags='$(LDFLAGS)' -mod=vendor
RELEASE_LDFLAGS := -w -s $(LDFLAGS)

TEST_ARGS := \
			-v -race -failfast \
			-covermode=atomic \
			-coverprofile=coverage.txt

GOBUILD := $(GO) build $(BUILD_ARGS)
GOTEST := $(GO) test $(BUILD_ARGS)
