#!/bin/bash

argocd app create cc-go-bumblebee \
--repo https://github.com/cc4i/go-bumblebee-tile.git \
--revision master \
--path kustomize \
--dest-server https://kubernetes.default.svc \
--dest-namespace __Namespace__

argocd app sync cc-go-bumblebee

