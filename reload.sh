#!/bin/bash

# Load environment variables
source .env

deploy() {
	docker build -t $1 -f images/$1/Dockerfile .
	docker tag $1 $DOCKER_USER/$1:latest
	docker login -u $DOCKER_USER -p "$DOCKER_PASS"
	docker push $DOCKER_USER/$1:latest
	kubectl delete deployment $1-deployment
	kubectl apply -f deployments/$1.yaml
}

# If no argument is passed, deploy all
if [ -z "$1" ]; then
	# Start database
	kubectl apply -f deployments/postgres.yaml

	# Start message broker
	kubectl apply -f deployments/rabbitmq.yaml

	# Rebuild, push and deploy the api and worker
	deploy api
	deploy worker
	exit 0
fi

deploy $1