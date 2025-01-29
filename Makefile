PROJECT_NAME		:= $(shell basename $(CURDIR))
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
GIT_TAG				:= $(shell git describe --dirty --tags --always)
DOCKERHUB_REPO		:= "cimem"
#######################################
# builds
#######################################

.PHONY: image
image: image
	docker build -t $(PROJECT_NAME):$(GIT_TAG) .

#######################################
# run the image
#######################################

.PHONY: image
image: image
	docker run -p 8080:8080 \
		-e ADO_ORGANIZATION=fabrikam \
		-e ADO_PROJECT=fabrikam-fiber-tfvc \
		-e ADO_URL=localhost \
		-e ADO_PERSONAL_ACCESS_TOKEN=$(PAT)
		$(PROJECT_NAME):$(GIT_TAG)

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