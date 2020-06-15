import * as cdk from '@aws-cdk/core';
import eks = require('@aws-cdk/aws-eks');
import ec2 = require('@aws-cdk/aws-ec2');
import iam = require('@aws-cdk/aws-iam');
import {NodePolicies} from './policy4eks'

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

    let region = process.env.CDK_DEFAULT_REGION
    let policies = []
    if (region == "cn-north-1" || region == "cn-northwest-1" ) {
      policies = [
        {managedPolicyArn:  "arn:aws-cn:iam::aws:policy/AmazonEKSServicePolicy"},
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
      managedPolicies: policies,
      inlinePolicies: new NodePolicies(scope, "inlinePolicy", {}).eksInlinePolicy
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
    /** unmanaged nodegroup */
    cluster.addCapacity("unmanaged-node", {
      instanceType: capacityInstance,
      minCapacity:  Math.round(props.capacity!/2),
      maxCapacity: props.capacity
    })
    /** managed nodegroup */
    cluster.addNodegroup("managed-node", {
      instanceType: capacityInstance,
      minSize: (props.capacity! - Math.round(props.capacity!/3)),
      maxSize: props.capacity
    })

    

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
