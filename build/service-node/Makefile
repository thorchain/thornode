# Image Hashes
STATE_CHAIN_IMAGE_HASH=$(shell  grep "Successfully built" ./log/state-chain-build.log | cut -d ' ' -f 3)

# Turn off cache
ifeq (${NOCACHE}, TRUE)
CACHE=--no-cache
endif

config:
	@echo ENV: ${ENV}
	@echo NET: ${NET}
	@echo POOL_ADDRESS: ${POOL_ADDRESS}
	@echo START_BLOCK_HEIGHT: ${START_BLOCK_HEIGHT}
	@echo CACHE: ${CACHE}
	@echo ECS_CLUSTER: ${ECS_CLUSTER}
	@echo INFLUX_HOST: ${INFLUX_HOST}
	@echo REGISTRY_ID: ${REGISTRY_ID}
	@echo REPOSITORY: ${REPOSITORY}
	@echo ECR_STATECHAIN: ${ECR_STATECHAIN}
	@echo ECR_CHAIN_SVC: ${ECR_CHAIN_SVC}
	@echo STATE_CHAIN_IMAGE_HASH: ${STATE_CHAIN_IMAGE_HASH}
	@echo CHAIN_SERVICE_IMAGE_HASH: ${CHAIN_SERVICE_IMAGE_HASH}

.PHONY: clean
clean:
	rm -rf ./log

.PHONY: login
login:
	aws configure set aws_access_key_id ${AWS_ACCESS_KEY_ID}
	aws configure set aws_secret_access_key ${AWS_SECRET_ACCESS_KEY}
	aws configure set region ${AWS_REGION}
	$(shell aws ecr get-login --no-include-email --region ap-southeast-2)

.PHONY: logger
logger: login
	mkdir -p ./log

# Step 1
build: clean logger
	cd docker/statechain/ && ./build.sh ${ECR_STATECHAIN} | tee ../../log/state-chain-build.log

# Step 2
tag: build
	docker tag ${REGISTRY_ID}.${REPOSITORY}/${ECR_STATECHAIN}:latest ${REGISTRY_ID}.${REPOSITORY}/${ECR_STATECHAIN}:${STATE_CHAIN_IMAGE_HASH}

# Step 3
push: tag
	docker push ${REGISTRY_ID}.${REPOSITORY}/${ECR_STATECHAIN}:${STATE_CHAIN_IMAGE_HASH}
	docker push ${REGISTRY_ID}.${REPOSITORY}/${ECR_STATECHAIN}:latest

# Step 4
deploy: config push
	aws ecs update-service --force-new-deployment --cluster ${ECS_CLUSTER} --service ${ECR_STATECHAIN}
