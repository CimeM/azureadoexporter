PROJECT_NAME		:= $(shell basename $(CURDIR))
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
GIT_TAG				:= $(shell git describe --dirty --tags --always)
DOCKERHUB_REPO		:= "cimartindev"
#######################################
# builds
#######################################

.PHONY: image
image: image
	docker build -t $(PROJECT_NAME):$(GIT_TAG) .

#######################################
# quality checks
#######################################

.PHONY: test
test:
	time go test ./...

#######################################
# release assets
#######################################

.PHONY: release
release: release
	docker build -t $(DOCKERHUB_REPO)/$(PROJECT_NAME):$(GIT_TAG) .
	docker push $(DOCKERHUB_REPO)/$(PROJECT_NAME):$(GIT_TAG)