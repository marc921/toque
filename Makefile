include .env
export

.PHONY: deploy
deploy:
	docker build -t $(COMPONENT) -f images/$(COMPONENT)/Dockerfile .
	docker tag $(COMPONENT) $(DOCKER_USER)/$(COMPONENT):latest
	docker login -u $(DOCKER_USER) -p "$(DOCKER_PASS)"
	docker push $(DOCKER_USER)/$(COMPONENT):latest
	-kubectl delete deployment $(COMPONENT)-deployment
	kubectl apply -f deployments/$(COMPONENT).yaml

.PHONY: db
db:
	kubectl apply -f deployments/postgres.yaml

.PHONY: msg
msg:
	kubectl apply -f deployments/rabbitmq.yaml

.PHONY: api
api:
	make deploy COMPONENT=api

.PHONY: worker
worker:
	make deploy COMPONENT=worker

.PHONY: scanner
scanner:
	make deploy COMPONENT=scanner

.PHONY: k8s-up
k8s-up:
	minikube status 2>/dev/null | grep -q Running || minikube start

.PHONY: k8s-down
k8s-down:
	kubectl delete all --all
	minikube stop

.PHONY: docker-up
docker-up:
	open -a Docker

.PHONY: all
all:
	make k8s-up
	make db
	make msg
	make deploy COMPONENT=api
	make deploy COMPONENT=worker
	make deploy COMPONENT=scanner