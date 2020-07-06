# Hu & Tiles

Check out following handy solutions from Hu, or build your own with Tiles.

## Hu

|        Hu    | Version | Description      |
|-----------------|---------|------------------|
| Simple EKS| [v0.1.0](./templates/eks-simple.yaml)| Quick launch with few lines of yaml.|
| EKS with Spot instance| [v0.1.0](./templates/eks-spot-simple.yaml)| Quick launch EKS cluster with mixed spot and on-demand instances, as well as handling spot termination, cluster auto scaler and HPA. |
| Simple ArgoCD | [v0.1.0](./templates/argocd-simple.yaml) | Setup ArgoCD on EKS with simple configuration.|
| Basic CD with ArgoCD | [v0.1.0](./templates/argocd-with-app.yaml) | Building a modern CD with example applicaiton on GitHub, all you need is a GitHub token.|
| Perfect Microservice on EKS | [v0.1.0]() |  Implement a handy containerized Microsercices architecture on EKS with all major componnets and demo applications. |


## Tiles

|        Tiles    | Version | Description      |
|-----------------|---------|------------------|
| Basic Network | [v0.0.1](./tiles-repo/network0/0.0.1)  | The classic network pattern cross multiple availibilty zone with public and private subnets, NAT, etc. |
| Simple EKS| [v0.0.1](./tiles-repo/eks0/0.0.1)| The basic EKS cluster, which uses EKS 1.15 as default version and depends on Network0. |
| | [v0.0.5](./tiles-repo/eks0/0.0.5)| Update EKS default version to 1.16 and expose more options. |
| EKS on Spot | [v0.5.0](./tiles-repo/eks-with-spot/0.5.0)| Provison EKS 1.16 as default and using auto scaling group with mixed spot and normal (4:1) instances. Also has Cluster Autoscaler, Horizontal Pod Autoscaler and Spot Instance Handler setup. |
|EFS | [v0.1.0](./tiles-repo/efs/0.1.0)|The basic EFS conpoment and based on Network0. EFS is a perfect choice as storage option for Kubernetes. |
|ArgoCD | [v1.5.2](./tiles-repo/argocd0/1.5.2)|The Argocd0 is basic component to help build up GitOps based CI/CD capability, which depends on Tile - Eks0 & Network0.|
|Go-Bumblebee-ONLY| [v0.0.1](./tiles-repo/go-bumblebee-only/0.0.1) | This is demo application, which can be deploy to Kubernetes cluster to demostrate rich capabilities.|
|Istio | [v1.5.4](./tiles-repo/istio0/1.5.4) | Setup Istio 1.6 on EKS with all necessary features. Managed by Istio operator and Egress Gateway was off by default. |
|AWS KMS | [v0.1.0](./tiles-repo/aws-kms-keygenerator/0.1.0) | Generate both symmetric key and asymmetric key for down stream applications or services |
|AWS ElastiCache Redis | [v5.0.6](./tiles-repo/aws-elasticache-redis/5.0.6) | Setup a redis cluster with replcation group with flexiable options. |
|AWS Aurora Mtsql | [v2.07.2](./tiles-repo/aws-aurora-mysql/2.07.2) | Provision a Aurora MySQL cluster and integrated with Secret Manager to automate secret ratation. |
| Go-BumbleBee-Jazz | [v0.7.1](./tiles-repo/go-bumblebee-jazz/0.7.1) | Modern cloud native application with tipycal features to try out how great your Kubernetes cluster are.|




