# Eks0
The Tile repreents a basic EKS cluster, which uses EKS 1.15 as default and depends on Tile - Network0. It will use default nodegroup as capacity.

## Dependent Tile

- Network0 [ v0.0.1 ]

## Inputs

- name: vpc
- name: vpcSubnets
- name: clusterName
- name: clusterVersion
- name: capacityInstance
- name: capacity



## Outputs 
- name: clusterName
- name: clusterArn
- name: clusterEndpoint
- name: masterRoleARN

## Notice