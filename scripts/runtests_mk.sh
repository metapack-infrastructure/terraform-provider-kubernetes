#!/bin/bash

TEST_FILTER=$1

export TF_ACC=1
export KUBE_CTX=minikube
export KUBE_CTX_CLUSTER=minikube
export KUBE_CTX_AUTH_INFO=minikube

/usr/local/bin/go test -v -parallel 1 -timeout 3600s github.com/metapack-infrastructure/terraform-provider-kubernetes/kubernetes -run "^TestAccKubernetes${TEST_FILTER}.*"
