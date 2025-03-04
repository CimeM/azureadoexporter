PROJECT_NAME		:= $(shell basename $(CURDIR))
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
GIT_TAG				:= $(shell git describe --tags --always)
DOCKERHUB_REPO		:= "cimartindev"

#######################################
# builds
#######################################

.PHONY: image
image: lint
	docker build -t $(PROJECT_NAME):$(GIT_TAG) .

#######################################
# run the image
#######################################


.PHONY: run
run: run
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

.PHONY: lint
lint:
	golangci-lint run
#######################################
# release assets
#######################################

.PHONY: release
release: release
	docker build -t $(DOCKERHUB_REPO)/$(PROJECT_NAME):$(GIT_TAG) .
	docker push $(DOCKERHUB_REPO)/$(PROJECT_NAME):$(GIT_TAG)