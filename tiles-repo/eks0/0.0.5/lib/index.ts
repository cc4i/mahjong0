import * as cdk from '@aws-cdk/core';
import eks = require('@aws-cdk/aws-eks');
import ec2 = require('@aws-cdk/aws-ec2');
import iam = require('@aws-cdk/aws-iam');
import {NodePolicies} from './policy4eks'
import { ManagedPolicy, ServicePrincipal, PolicyDocument, PolicyStatement } from '@aws-cdk/aws-iam';

/** Input parameters */
export interface Eks0Props {
  vpc: ec2.Vpc,
  vpcSubnets?: ec2.ISubnet[],
  clusterName: string,
  capacity?: number,
  capacityInstance?: string,
  clusterVersion?: string,
}

export class Eks0 extends cdk.Construct {
  
  /** Directly exposed to other stack */
  public readonly clusterName: string;
  public readonly clusterEndpoint: string;
  public readonly masterRoleARN: string;
  public readonly clusterArn: string;
  public readonly capacity: number;
  public readonly capacityInstance: string;

  constructor(scope: cdk.Construct, id: string, props: Eks0Props) {
    super(scope, id);

    const eksRole = new iam.Role(this, 'EksClusterMasterRole', {
      assumedBy: new iam.AccountRootPrincipal(),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEKSServicePolicy"),
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEKSClusterPolicy"),
      ]
    });


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
      vpcSubnets = [{subnets: props.vpc.publicSubnets}, {subnets: props.vpc.privateSubnets}];
    } else {
      vpcSubnets = [{subnets: props.vpcSubnets}];
    }
    // Innitial EKS cluster
    const cluster = new eks.Cluster (this, "BasicEKSCluster", {
      vpc: props.vpc,
      vpcSubnets: vpcSubnets,
      clusterName: props.clusterName,
      defaultCapacity: 0,
      version: props.clusterVersion || '1.16',
      // Master role as initial permission to run Kubectl
      mastersRole: eksRole,
    });


    /** managed nodegroup */
    const nodegroupRole = new iam.Role(scope, 'NodegroupRole', {
      assumedBy: new iam.ServicePrincipal("ec2.amazonaws.com"),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEKSWorkerNodePolicy"),
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEKS_CNI_Policy"),
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEC2ContainerRegistryReadOnly"),
      ],
      inlinePolicies: new NodePolicies(scope, "inlinePolicies", {}).eksInlinePolicy
    });
    const managed = cluster.addNodegroup("managed-node", {
      instanceType: capacityInstance,
      minSize: Math.round(props.capacity!/2),
      maxSize: props.capacity,
      nodeRole: nodegroupRole
    });
    

    /** Added CF Output */
    new cdk.CfnOutput(this,"clusterName", {value: cluster.clusterName})
    new cdk.CfnOutput(this,"masterRoleARN", {value: eksRole.roleArn})
    new cdk.CfnOutput(this,"clusterEndpoint", {value: cluster.clusterEndpoint})
    new cdk.CfnOutput(this,"clusterArn", {value: cluster.clusterArn})
    new cdk.CfnOutput(this,"capacity", {value: String(props.capacity) || "0"})
    new cdk.CfnOutput(this,"capacityInstance", {value: capacityInstance.toString() })
    
    this.clusterName = cluster.clusterName;
    this.masterRoleARN = eksRole.roleArn;
    this.clusterEndpoint = cluster.clusterEndpoint;
    this.clusterArn = cluster.clusterArn;
    this.capacity =  props.capacity || 0
    this.capacityInstance = capacityInstance.toString()

  }


}
