PROJECT_ID = avid-life-273910
CLUSTER_NAME = cluster-1

APP_DIRECTORY = ../app
APP_FILES = $(wildcard ../app/*)

EXECUTOR_DIRECTORY = ../executor
EXECUTOR_FILES = $(wildcard ../executor/*)

deploy: gke-connect install-rabbitmq app-docker executor-docker # install-keda
	kubectl create -f ../clickhouse-service.yaml
	kubectl create -f ../app-service.yaml
	kubectl create -f ../autoscaler.yaml

tear-down:
	kubectl delete -f ../clickhouse-service.yaml
	kubectl delete -f ../app-service.yaml
	kubectl delete TriggerAuthentication rabbitmq-consumer-trigger
	kubectl delete ScaledObject rabbitmq-consumer
	# kubectl delete deploy rabbitmq-consumer
	kubectl delete secret rabbitmq-consumer-secret

install-keda: gke-connect
	helm repo add kedacore https://kedacore.github.io/charts
	helm repo update
	helm install keda kedacore/keda 
	touch install-keda

install-rabbitmq: gke-connect
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	helm install my-release bitnami/rabbitmq
	touch install-rabbitmq

app-docker: gke-connect ${APP_DIRECTORY} ${APP_FILES}
	docker build -t gcr.io/${PROJECT_ID}/app ../app
	docker push gcr.io/${PROJECT_ID}/app
	touch app-docker

executor-docker: gke-connect ${EXECUTOR_DIRECTORY} ${EXECUTOR_FILES}
	docker build -t gcr.io/${PROJECT_ID}/executor ../executor
	docker push gcr.io/${PROJECT_ID}/executor
	touch executor-docker

gke-connect: 
	gcloud container clusters get-credentials ${CLUSTER_NAME}
	gcloud auth configure-docker
	touch gke-connect