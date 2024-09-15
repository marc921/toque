include .env
export

.PHONY: build
build:
	docker build -t $(COMPONENT) -f images/$(COMPONENT)/Dockerfile .
	docker tag $(COMPONENT):latest $(CONTAINER_REGISTRY)/$(COMPONENT):latest
	@docker login $(CONTAINER_REGISTRY) -u nologin --password-stdin <<< "$(SCW_SECRET_KEY)"
	docker push $(CONTAINER_REGISTRY)/$(COMPONENT):latest

.PHONY: k8s-up
k8s-up:
	minikube status 2>/dev/null | grep -q Running || minikube start

.PHONY: down
down:
	kubectl delete all --all
	# minikube stop

.PHONY: docker-up
docker-up:
	@if ! docker info > /dev/null 2>&1; then \
		echo "Docker daemon is not running. Starting Docker..."; \
		open -a Docker; \
		echo "Waiting for Docker daemon to be ready..."; \
		while ! docker info > /dev/null 2>&1; do \
			sleep 1; \
		done; \
		echo "Docker daemon is now ready."; \
	else \
		echo "Docker daemon is already running."; \
	fi

.PHONY: helm-apply
helm-apply:
	helm --kubeconfig $(KUBECONFIG) upgrade --install $(RELEASE) ./toque-helmchart

.PHONY: all
all:
	# make k8s-up
	go mod tidy
	make docker-up
	make build COMPONENT=api
	make build COMPONENT=worker
	make build COMPONENT=scanner
	make helm-apply RELEASE=toque-release