# AWS Solutions Assembler

[![CircleCI](https://circleci.com/gh/cc4i/mahjong0.svg?style=svg)](https://circleci.com/gh/cc4i/mahjong0)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/cc4i/mahjong0)
[![Go Report Card](https://goreportcard.com/badge/github.com/cc4i/mahjong0)](https://goreportcard.com/report/github.com/cc4i/mahjong0)

## Description
AWS Solutions Assembler is also known as [Mahjong](./docs/All-Concept.md), which has built-in mechanism to leverage pattern based abstracts to build up any solution.

Builders can use [Mahjong](./docs/All-Concept.md) to share solutions with the best industry practice. Customers can quickly experience those solutions or build their own.


## Prerequisite

- Install [Docker](https://docs.docker.com/desktop/#download-and-install)
- Install [CDK](https://github.com/aws/aws-cdk)
- [Setup AWS configuration and credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)
- Download [mctl](https://github.com/awslabs/aws-solutions-assembler/releases)

## Quick Start

```bash

# Run dice as coantainer
docker run -d -v ~/.aws:/root/.aws -p 9090:9090 mahjongs/dice

# Kick start browser for first trial (On Darwin)
open http://127.0.0.1:9090/toy

# Paste the solution and send to provision 

```

## Hu

- Containerized microservices on EKS
> Modernized microservices on EKS with built-in automated release pipeline, service mesh, log, metrics, tracing, secret management, and more, which's a one-stop solution for containerized microservices.


## Develope your own

If you want to share your expertise or build your favorite things from scratch, following guide would be helpful.

- [How to build the Hu](./docs/How-to-Build-Hu.md) 

- [How to build the Tile](./docs/How-to-Build-Tile.md)

- [All available Hu and Tile](./repo/README.md)

## What's coming

- [X] Data pipeline on EKS
- [X] Serverless on EKS
- [X] AI on EKS


## Referenes

- [Node.js](https://nodejs.org/en/download/) ( ≥ 10.12.0 ) 
- [AWS CLI 2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) 
- [CDK](https://github.com/aws/aws-cdk)
- [aws-iam-authenticator](https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html)
- [Kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html)
- [Kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md)
- [Helm](https://helm.sh/docs/intro/install/)

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This project is licensed under the Apache-2.0 License.