# Kubernetes from scratch
I have started this repository as a practice on Docker and Kubernetes.

My goal is to be able to deploy an HTTP API, a RabbitMQ server and a worker as containers in a kubernetes cluster.
- deployment must be doable in a single command such as `mage deploy:all`
- the HTTP API will accept requests, validate a simple payload and send a message to the worker by publishing it in a RabbitMQ exchange
- the worker will subscribe to a queue bound to the exchange

## Images / Docker
To build the api (`-t` tags/names it "api"):
```bash
docker build -t api -f images/api/Dockerfile .
docker tag api $DOCKER_USER/api:latest
```
This creates an image viewable using `docker images` and that can be run into a container using `docker run api`.

### Container Registry

#### Local : does not work?
Each pod running in the cluster will need to pull its image from a container registry. We will use a local registry i.e. a container running on our local machine.
```bash
docker run -d -p 5000:5000 --name registry registry:latest
```
Note that port 5000 may be used on MacOS by AirPlay. This can be deactivated in the System Preferences > Sharing > AirPlay.

#### Docker Login
```bash
docker login -u $DOCKER_USER -p "$DOCKER_PASS"
docker push $DOCKER_USER/api:latest
kubectl create secret docker-registry docker-cred --docker-server=https://index.docker.io/v1/ --docker-username="$DOCKER_USER" --docker-password="$DOCKER_PASS"
```

## Kubernetes
```bash
kubectl apply -f deployments/api.yaml # pulls from my remote personal docker image registry
kubectl get pods
```

## Services
A service connects a container with other containers in the same cluster or the outside world.
A service of type LoadBalancer connects to the outside, but will need running [`minikube tunnel`](https://minikube.sigs.k8s.io/docs/commands/tunnel/) to allow connecting from the local machine (e.g. `curl routedIP:nodePort` where routedIP is on the right of the route line in minikube tunnel output).
=> `curl 192.168.64.2:30000`

## Database

### postgres - container
See yaml, connect with `psql $DATABASE_URL`

To inject the postgresql password in our kube deployments (e.g. worker), we use a kube secret:
```bash
kubectl create secret generic postgres-password --from-literal=password="$PG_PASSWORD"
```

### dbmate - migrations
See [repo](https://github.com/amacneil/dbmate)

Use:
- `dbmate new`
- `dbmate up`

Input: db/migrations

Output: db/schema.sql

### sqlc - SQL -> Go code generation
See [docs](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html)

Use:
- `sqlc -f db/sqlc.yaml generate`

Config file: db/sqlc.yaml

Input: db/schema.sql, db/queries

Output: db/sqlcgen