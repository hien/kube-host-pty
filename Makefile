include scripts/tools.mk

# set default build target if empty
MAKECMDGOALS ?= pty-device-plugin
.build:
	$(GOBUILD) -o build/$(MAKECMDGOALS) cmd/$(MAKECMDGOALS)/*.go

pty-device-plugin: .build
pty-client: .build
kubectl-pty: .build

include scripts/images.mk

.test:
	$(GOTEST) ./$(PKG)

test: $(eval PKG := ...) .test
test-pkg: $(eval PKG := pkg/...) .test
