# Eks0
The Tile repreents a basic EKS cluster, which uses EKS 1.16 as default and depends on network Tile. The worker nodes will be managed nodes.

## Dependent Tile

- Network0 [ v0.0.1 ]

## Inputs

- name: cidr
- name: vpc
- name: vpcSubnets
- name: clusterName
- name: clusterVersion
- name: capacityInstance
- name: capacity



## Outputs 
- name: clusterName
- name: clusterVersion
- name: clusterArn
- name: clusterEndpoint
- name: masterRoleARN
- name: capacityInstance
- name: capacity

## Notice