#!/bin/bash

# Load environment variables
source .env

# Start rabbitmq server
kubectl apply -f deployments/rabbitmq.yaml

# Rebuild, push and deploy the api and worker services
deploy() {
	docker build -t $1 -f images/$1/Dockerfile .
	docker tag $1 $DOCKER_USER/$1:latest
	docker login -u $DOCKER_USER -p "$DOCKER_PASS"
	docker push $DOCKER_USER/$1:latest
	kubectl delete deployment $1-deployment
	kubectl apply -f deployments/$1.yaml
}

deploy api
deploy worker