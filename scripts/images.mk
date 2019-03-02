include scripts/tools.mk

DOCKER_REPO := arhatdev

.build-image:
	$(eval TARGET := $(MAKECMDGOALS:-image=))
	$(DOCKERBUILD) --build-arg TARGET=$(TARGET) \
		-t $(DOCKER_REPO)/$(TARGET):latest \
		-f cicd/docker/app.dockerfile .

pty-device-plugin-image: .build-image
pty-client-image: .build-image
kubectl-pty-image: .build-image

.check-image:
	$(eval TARGET := $(MAKECMDGOALS:-image-check=))
	$(DIVE) $(DOCKER_REPO)/$(TARGET):latest

pty-device-plugin-image-check: .check-image
pty-client-image-check: .check-image
kubectl-pty-image-check: .check-image

.push-image:
	$(eval TARGET := $(MAKECMDGOALS:-image-push=))
	$(DOCKERPUSH) $(DOCKER_REPO)/$(TARGET):latest

pty-device-plugin-image-push: .push-image
pty-client-image-push: .push-image
kubectl-pty-image-push: .push-image
