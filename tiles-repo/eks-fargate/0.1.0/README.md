# Eks0
The Tile repreents a basic EKS cluster, which uses EKS 1.16 as default and depends on netwrok Tile. The worker nodes will make up with unmanaged nodes, managed nodes and fargate.

## Dependent Tile

- Network0 [ v0.0.1 ]

## Inputs

- name: vpc
- name: clusterName
- name: clusterVersion



## Outputs 
- name: clusterName
- name: clusterArn
- name: clusterEndpoint
- name: masterRoleARN

## Notice