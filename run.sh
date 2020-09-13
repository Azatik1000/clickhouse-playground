#!/bin/bash

if ! minikube status; then
    minikube start --vm-driver=virtualbox
fi

kubectl delete -f clickhouse-service.yaml
kubectl delete -f app-service.yaml
kubectl delete ScaledObject rabbitmq-consumer
kubectl delete deploy rabbitmq-consumer
kubectl delete secret rabbitmq-consumer-secret
kubectl delete TriggerAuthentication rabbitmq-consumer-trigger

eval $(minikube docker-env)

cd app
docker build -t app . 
cd ..

cd clickhouse-executor
docker build -t clickhouse-executor .
cd ..

kubectl create -f clickhouse-service.yaml
kubectl create -f app-service.yaml
kubectl create -f autoscaler.yaml

