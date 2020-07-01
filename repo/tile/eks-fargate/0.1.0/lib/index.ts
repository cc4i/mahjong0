import * as cdk from '@aws-cdk/core';
import eks = require('@aws-cdk/aws-eks');
import ec2 = require('@aws-cdk/aws-ec2');
import iam = require('@aws-cdk/aws-iam');
import { ManagedPolicy } from '@aws-cdk/aws-iam';

/** Input parameters */
export interface EksFargateProps {
  vpc: ec2.Vpc,
  clusterName: string,
  clusterVersion?: string,
}

export class EksFargate extends cdk.Construct {
  
  /** Directly exposed to other stack */
  public readonly clusterName: string;
  public readonly clusterEndpoint: string;
  public readonly masterRoleARN: string;
  public readonly clusterArn: string;

  constructor(scope: cdk.Construct, id: string, props: EksFargateProps) {
    super(scope, id);

    const eksRole = new iam.Role(this, 'EksClusterMasterRole', {
      assumedBy: new iam.AccountRootPrincipal(),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEKSServicePolicy"),
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEKSClusterPolicy"),
      ]
    });


    // Innitial EKS cluster
    const cluster = new eks.Cluster (this, "EKSFargateCluster", {
      vpc: props.vpc,
      vpcSubnets: [{subnets: props.vpc.privateSubnets}, {subnets: props.vpc.publicSubnets}],
      clusterName: props.clusterName,
      version: props.clusterVersion || '1.16',
      // Master role as initial permission to run Kubectl
      mastersRole: eksRole, 
      defaultCapacity: 0
    });
    
    const fargateProfileRole = new iam.Role(this, 'EksFargateProfileRole', {
      assumedBy: new iam.ServicePrincipal('eks-fargate-pods.amazonaws.com'),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEKSFargatePodExecutionRolePolicy")
      ]
    });
    const fargetProfile = cluster.addFargateProfile("FargateProfile", {
      fargateProfileName: "DefaultProfile",
      selectors: [
        { namespace: 'default' },
        { namespace: 'kube-system' }
      ],
      podExecutionRole: fargateProfileRole,
      subnetSelection: {subnets: props.vpc.privateSubnets}
    })


    /** Added CF Output */
    new cdk.CfnOutput(this,"clusterName", {value: cluster.clusterName})
    new cdk.CfnOutput(this,"masterRoleARN", {value: eksRole.roleArn})
    new cdk.CfnOutput(this,"clusterEndpoint", {value: cluster.clusterEndpoint})
    new cdk.CfnOutput(this,"clusterArn", {value: cluster.clusterArn})
    
    this.clusterName = cluster.clusterName;
    this.masterRoleARN = eksRole.roleArn;
    this.clusterEndpoint = cluster.clusterEndpoint;
    this.clusterArn = cluster.clusterArn;

  }


}
