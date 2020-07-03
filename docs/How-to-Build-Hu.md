# How to build a Hu

Hu is a full solution and combination of Tiles, key is to build a turn key solution leverage all we have. 

## Prerequisite
- Install [Docker Desktop](https://docs.docker.com/desktop/#download-and-install)
- Install [AWS CDK](https://docs.aws.amazon.com/cdk/latest/guide/getting_started.html#getting_started_install) 
- Specify [AWS Configuration and Credential](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) setting

## Execution Order

Writing order in the deployment specification is critical path and depedent Tiles is a small branch at same stage. Dependent Tile will be execute first.

Using dry run to check out execution plan. Call following API to retrieve more detail of your Hu. 

```bash

# List all deployments
curl http://127.0.0.1:9090/v1alpha1/ts

# Get execution plan
curl http://127.0.0.1:9090/v1alpha1/ts/[session id]/plan

# Get the order of execution plan
curl http://127.0.0.1:9090/v1alpha1/ts/[session id]/plan/order

# Get the parallel order of execution plan
curl http://127.0.0.1:9090/v1alpha1/ts/[session id]/plan/order/parallel

```


## Useful Tips

1. Each family group doesn't allow repeated kind of Tile, and using different family group to deploy same kind of Tile.

## Hu example: GitOps & CD on EKS
