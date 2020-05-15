# Mahjong
<img src='./docs/mahjong-table.png' width='100'>

## Description
[Mahjong](./docs/All-Concept.md) has built-in mechanism to abstract best practice away from traditional solutions, so builders can quickly build new abstract block or full solution based on other building blocks, called Tile.

People can use Hu to quickly spin up full solutions or resources on AWS with industry best practice and non-industry experience required.

## Prerequisite

- Install [Docker](https://docs.docker.com/desktop/#download-and-install)
- Install [CDK](https://github.com/aws/aws-cdk)
- [Setup AWS configuration and credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)
- Download latest [mctl](https://github.com/cc4i/mahjong0/releases) [ Linux / Darwin / Windows ]

## Quick Start

```bash

# Run dice as coantainer
docker run -d -v ~/.aws:/root/.aws -p 9090:9090 herochinese/dice

# Kick start browser for first trial (On Darwin)
open http://127.0.0.1:9090/toy

```

## Develope a Tile

Check out following for an EKS quick start, and click [here](./docs/How-to-Build-Tile.md) for more detail to develope Tile.

```bash

# Run dice on DEV mode in order to loading your Tile
docker run -it -v ~/mywork/mylabs/csdc/mahjong-0/tiles-repo:/workspace/tiles-repo \
    -v ~/.aws:/root/.aws \
    -e M_MODE=dev \
    -p 9090:9090 \
    herochinese/dice

# Initial a Tile project with your favorite name
mctl init tile sample-tile

# Deploy Tiles with your very first try. 
cd sample-tile
mctl deploy -f ./eks-simple.yaml

# Make your own bespoke Tiles ...

```


## Develope a Hu

Click [here](./docs/How-to-Build-Hu.md) for more detail to develop Hu.


## What to Next


## Referenes

- [Node.js](https://nodejs.org/en/download/) ( ≥ 10.12.0 ) 
- [AWS CLI 2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) 
- [CDK](https://github.com/aws/aws-cdk)
- [aws-iam-authenticator](https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html)
- [Kubectl](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html)
- [Kustomize](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md)
- [Helm](https://helm.sh/docs/intro/install/)