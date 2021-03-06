PROJECT_ID = avid-life-273910
CLUSTER_NAME = cluster-2

APP_DIRECTORY = ../app
APP_FILES = $(wildcard ../app/*)

EXECUTOR_DIRECTORY = ../executor
EXECUTOR_FILES = $(wildcard ../executor/*)

EXECUTOR_DIRECTORY = ../executor
EXECUTOR_FILES = $(wildcard ../executor/*)

FRONTEND_DIRECTORY = ../frontend
FRONTEND_FILES = $(wildcard ../frontend/*)

TEMPLATER_DIRECTORY = ../templater
TEMPLATER_FILES = $(wildcard ../templater/*)

CONFIG_DIRECTORY = ../config
CONFIG_FILES = $(wildcard ../config/*)

DEPLOY_TEMPLATE_FILES = $(wildcard ../*-template.yaml)

deploy: gke-connect install-rabbitmq install-keda app-docker executor-docker configure-deployment
	kubectl create -f ../executor-rbac.yaml
	kubectl create -f ../executor-deployments.yaml
	kubectl create -f ../rabbitmq-auth.yaml
	kubectl create -f ../autoscalers.yaml
	kubectl create -f ../app-service.yaml
	kubectl create -f ../frontend-deployment.yaml
	kubectl create -f ../frontend-service.yaml
	kubectl create -f ../ingress.yaml

tear-down:
	kubectl delete -f ../ingress.yaml
	kubectl delete -f ../frontend-service.yaml
	kubectl delete -f ../frontend-deployment.yaml
	kubectl delete -f ../app-service.yaml
	kubectl delete -f ../autoscalers.yaml
	kubectl delete -f ../rabbitmq-auth.yaml
	kubectl delete -f ../executor-deployments.yaml
	kubectl delete pod -l component=worker
	kubectl delete -f ../executor-rbac.yaml

configure-deployment: ${DEPLOY_TEMPLATE_FILES} ${TEMPLATER_DIRECTORY} ${TEMPLATER_FILES}
	go run ../templater/main.go
	touch configure-deployment

install-postgresql: gke-connect
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm install --set postgresqlUsername=postgres --set postgresqlPassword=postgres my-postgres bitnami/postgresql
	touch install-postgresql

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

frontend-docker: gke-connect ${FRONTEND_DIRECTORY} ${FRONTEND_FILES}
	docker build -t gcr.io/${PROJECT_ID}/frontend:latest ../frontend
	docker push gcr.io/${PROJECT_ID}/frontend:latest
	touch frontend-docker

update-frontend: frontend-docker
	kubectl set image deployments/frontend frontend=gcr.io/${PROJECT_ID}/frontend
	kubectl rollout restart deployment/frontend

app-docker: gke-connect move-configurations ${APP_DIRECTORY} ${APP_FILES}
	docker build -t gcr.io/${PROJECT_ID}/app ../app
	docker push gcr.io/${PROJECT_ID}/app
	touch app-docker

update-app: app-docker
	kubectl set image deployments/app-deployment app=gcr.io/${PROJECT_ID}/app
	kubectl rollout restart deployment/app-deployment

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
