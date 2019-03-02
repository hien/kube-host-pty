include scripts/tools.mk

.PHONY: build-image
.build-image:
	$(eval TARGET := $(MAKECMDGOALS:-image=))
	$(DOCKERBUILD) --build-arg TARGET=$(TARGET) \
		-t arhatdev/$(TARGET):latest \
		-f cicd/docker/app.dockerfile .

pty-device-plugin-image: .build-image
pty-client-image: .build-image
kubectl-pty-image: .build-image

.PHONY: .check-image
.check-image:
	$(eval TARGET := $(MAKECMDGOALS:-image-check=))
	$(DIVE) arhatdev/$(TARGET):latest

pty-device-plugin-image-check: .check-image
pty-client-image-check: .check-image
kubectl-pty-image-check: .check-image
