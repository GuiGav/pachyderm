# Set as you wish
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := example-spout101
CONTAINER_VERSION := 2.0.0-beta.5
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)

docker-image:
	@docker build -t $(CONTAINER_TAG) .
	@docker push $(CONTAINER_TAG)

deploy:
	pachctl create pipeline -f ./pipelines/spout.json
	pachctl create pipeline -f ./pipelines/processor.json
	pachctl create pipeline -f ./pipelines/reducer.json
