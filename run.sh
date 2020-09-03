#!/bin/bash

if ! minikube status; then
    minikube start --vm-driver=virtualbox
fi

kubectl delete -f clickhouse-service.yaml
kubectl delete -f app-service.yaml

eval $(minikube docker-env)

cd app
docker build -t app . 
cd ..

cd clickhouse-executor
docker build -t clickhouse-executor .
cd ..

kubectl create -f clickhouse-service.yaml
kubectl create -f app-service.yaml
