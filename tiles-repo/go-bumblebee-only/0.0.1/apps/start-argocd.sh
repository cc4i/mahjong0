#!/bin/bash

export KUBECONFIG=`pwd`/../../../kube.config
kubectl port-forward svc/argocd-server -n argocd 8080:443
