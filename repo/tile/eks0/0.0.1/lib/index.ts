import * as cdk from '@aws-cdk/core';
import eks = require('@aws-cdk/aws-eks');
import ec2 = require('@aws-cdk/aws-ec2');
import iam = require('@aws-cdk/aws-iam');
import region = require('@aws-cdk/region-info');
import { CfnOutput } from '@aws-cdk/core';

/** Input parameters */
export interface Eks0Props {
  vpc: ec2.Vpc,
  vpcSubnets?: ec2.ISubnet[],
  clusterName: string,
  capacity?: number,
  capacityInstance?: string,
  version?: string,
}

export class Eks0 extends cdk.Construct {
  
  /** Directly exposed to other stack */
  public readonly clusterName: string;
  public readonly clusterEndpoint: string;
  public readonly masterRoleARN: string;
  public readonly clusterArn: string;

  constructor(scope: cdk.Construct, id: string, props: Eks0Props) {
    super(scope, id);

    let region = process.env.CDK_DEFAULT_REGION
    let policies = []
    if (region == "cn-north-1" || region == "cn-northwest-1" ) {
      policies = [
        {managedPolicyArn:  "arn:"+":iam::aws:policy/AmazonEKSServicePolicy"},
        {managedPolicyArn: "arn:aws-cn:iam::aws:policy/AmazonEKSClusterPolicy"}
      ]
    } else {
      policies = [
        {managedPolicyArn:  "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"},
        {managedPolicyArn: "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"}
      ]
      
    }

    const eksRole = new iam.Role(this, 'EksClusterMasterRole', {
      assumedBy: new iam.AccountRootPrincipal(),
      managedPolicies: policies
    });

    eksRole.addToPolicy(
        new iam.PolicyStatement({
            actions: ["elasticloadbalancing:*","ec2:CreateSecurityGroup","ec2:Describe*"],
            resources: ["*"]
        })
      );

    // Instance type for node group
    let capacityInstance: ec2.InstanceType;
    if (props.capacityInstance == undefined) {
      capacityInstance=ec2.InstanceType.of(ec2.InstanceClass.C5, ec2.InstanceSize.LARGE);
    } else {
      capacityInstance=new ec2.InstanceType(props.capacityInstance);
    }

    // Prepared subnet for node group
    let vpcSubnets: ec2.SubnetSelection[];
    if (props.vpcSubnets == undefined){
      vpcSubnets = [{subnets: props.vpc.publicSubnets}, {subnets: props.vpc.privateSubnets}]
    } else {
      vpcSubnets = [{subnets: props.vpcSubnets}]
    }
    // Innitial EKS cluster
    const cluster = new eks.Cluster (this, "BasicEKSCluster", {
      vpc: props.vpc,
      vpcSubnets: vpcSubnets,
      clusterName: props.clusterName,
      defaultCapacity: props.capacity,
      defaultCapacityInstance: capacityInstance,
      version: props.version || '1.15',
      // Master role as initial permission to run Kubectl
      mastersRole: eksRole,
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
