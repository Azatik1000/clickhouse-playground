PROJECT_ID = avid-life-273910
CLUSTER_NAME = cluster-2

APP_DIRECTORY = ../app
APP_FILES = $(wildcard ../app/*)

EXECUTOR_DIRECTORY = ../executor
EXECUTOR_FILES = $(wildcard ../executor/*)

TEMPLATER_DIRECTORY = ../templater
TEMPLATER_FILES = $(wildcard ../templater/*)

CONFIG_DIRECTORY = ../config
CONFIG_FILES = $(wildcard ../config/*)

DEPLOY_TEMPLATE_FILES = $(wildcard ../*-template.yaml)

deploy: gke-connect install-rabbitmq install-keda app-docker executor-docker configure-deployment 
	kubectl create -f ../executor-rbac.yaml
	kubectl create -f ../executor-deployments.yaml
	kubectl create -f ../app-service.yaml
	kubectl create -f ../rabbitmq-auth.yaml
	kubectl create -f ../autoscalers.yaml

tear-down:
	kubectl delete -f ../autoscalers.yaml
	kubectl delete -f ../rabbitmq-auth.yaml
	kubectl delete -f ../app-service.yaml
	kubectl delete -f ../executor-deployments.yaml
	kubectl delete -f ../executor-rbac.yaml

configure-deployment: ${DEPLOY_TEMPLATE_FILES} ${TEMPLATER_DIRECTORY} ${TEMPLATER_FILES}
	go run ../templater/main.go
	touch configure-deployment

install-keda: gke-connect
	helm repo add kedacore https://kedacore.github.io/charts
	helm repo update
	helm install keda kedacore/keda 
	touch install-keda

install-rabbitmq: gke-connect
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	helm install my-release bitnami/rabbitmq
	kubectl exec my-release-rabbitmq-0 -- sh -c "rabbitmq-plugins enable rabbitmq_management"
	kubectl exec my-release-rabbitmq-0 -- sh -c "rabbitmqctl add_user test test"
	kubectl exec my-release-rabbitmq-0 -- sh -c "rabbitmqctl set_user_tags test administrator"
	kubectl exec my-release-rabbitmq-0 -- sh -c 'rabbitmqctl set_permissions -p / test ".*" ".*" ".*"'
	touch install-rabbitmq

app-docker: gke-connect move-configurations ${APP_DIRECTORY} ${APP_FILES}
	docker build -t gcr.io/${PROJECT_ID}/app ../app
	docker push gcr.io/${PROJECT_ID}/app
	touch app-docker

executor-docker: gke-connect ${EXECUTOR_DIRECTORY} ${EXECUTOR_FILES}
	docker build -t gcr.io/${PROJECT_ID}/executor ../executor
	docker push gcr.io/${PROJECT_ID}/executor
	touch executor-docker

move-configurations: ${CONFIG_DIRECTORY} ${CONFIG_FILES} 
	cp -r ${CONFIG_DIRECTORY} "${APP_DIRECTORY}/config"
	touch move-configurations

gke-connect: 
	gcloud container clusters get-credentials ${CLUSTER_NAME}
	gcloud auth configure-docker
	touch gke-connect
