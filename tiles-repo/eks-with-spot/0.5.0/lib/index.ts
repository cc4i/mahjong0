import * as cdk from '@aws-cdk/core';
import eks = require('@aws-cdk/aws-eks');
import ec2 = require('@aws-cdk/aws-ec2');
import iam = require('@aws-cdk/aws-iam');
import region = require('@aws-cdk/region-info');
import { CfnOutput } from '@aws-cdk/core';
import autoscaling = require('@aws-cdk/aws-autoscaling');
import { EksNodesSpot } from './spot'

/** Input parameters */
export interface EksSpotProps {
  vpc: ec2.Vpc;
  vpcSubnets?: ec2.ISubnet[];
  clusterName: string;
  clusterVersion?: string;
  keyPair4EC2: string;
  capacityInstance?: string[];
  maxSizeASG?: string;
  minSizeASG?: string;
  desiredCapacityASG?: string;
  cooldownASG?: string;
  onDemandPercentage?: number;
  
}

export class EksWithSpot extends cdk.Construct {
  
  /** Directly exposed to other stack */
  public readonly regionOfCluster: string
  public readonly clusterName: string;
  public readonly clusterVersion: string;
  public readonly clusterEndpoint: string;
  public readonly masterRoleARN: string;
  public readonly clusterArn: string;
  public readonly autoScalingGroupName: string;
  public readonly autoScalingGroupMaxSize: string;
  public readonly autoScalingGroupMinSize: string;
  public readonly autoScalingGroupDesiredCapacity: string;
  public readonly nodesRoleARN: string;
  public capacityInstance: string;
  public controlPlaneSG: ec2.SecurityGroup;

  constructor(scope: cdk.Construct, id: string, props: EksSpotProps) {
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
      managedPolicies: policies
    });



    // Prepared subnet for node group
    let vpcSubnets: ec2.SubnetSelection[];
    if (props.vpcSubnets == undefined){
      vpcSubnets = [{subnets: props.vpc.publicSubnets}, {subnets: props.vpc.privateSubnets}]
    } else {
      vpcSubnets = [{subnets: props.vpcSubnets}]
    }
     /** control panel security group  */ 
    this.controlPlaneSG = new ec2.SecurityGroup(this, `EksControlPlaneSG`, {
      vpc: props.vpc
    });

    // Innitial EKS cluster
    const cluster = new eks.Cluster(scope, "EksSpotCluster", {
      vpc: props.vpc,
      vpcSubnets: vpcSubnets,
      clusterName: props.clusterName,
      defaultCapacity: 0,
      version: props.clusterVersion || '1.16',
      // Master role as initial permission to run Kubectl
      mastersRole: eksRole,
      securityGroup: this.controlPlaneSG,
    })

    const spotNodeGroup = new EksNodesSpot(scope, "SpotNodeGroup", {
      clusterName: props.clusterName,
      clusterEndpoint: cluster.clusterEndpoint,
      clusterCertificateAuthorityData: cluster.clusterCertificateAuthorityData,
      clusterVersion: props.clusterVersion || '1.16',
      vpc: props.vpc,
      publicSubnetId1: props.vpc.publicSubnets[0].subnetId,
      publicSubnetId2: props.vpc.publicSubnets[1].subnetId,
      privateSubnetId1: props.vpc.privateSubnets[0].subnetId,
      privateSubnetId2: props.vpc.privateSubnets[0].subnetId,
      capacityInstance: props.capacityInstance,
      keyPair4EC2: props.keyPair4EC2,
      maxSizeASG: props.maxSizeASG || "6",
      minSizeASG: props.minSizeASG || "3",
      desiredCapacityASG: props.desiredCapacityASG || "3",
      cooldownASG: props.cooldownASG || "180",
      onDemandPercentage: props.onDemandPercentage || 25,
      controlPlaneSG: this.controlPlaneSG,
    });

    this.capacityInstance=""
    props.capacityInstance?.forEach(c => {
      this.capacityInstance = c + "/" + this.capacityInstance
    })
    this.capacityInstance=this.capacityInstance.substring(0,this.capacityInstance.length-1)

    /** Added CF Output */
    new cdk.CfnOutput(scope,"regionOfCluster", {value: process.env.CDK_DEFAULT_REGION || ""})
    new cdk.CfnOutput(scope,"clusterName", {value: cluster.clusterName})
    new cdk.CfnOutput(scope,"clusterVersion", {value: props.clusterVersion || '1.16'})
    new cdk.CfnOutput(scope,"masterRoleARN", {value: eksRole.roleArn})
    new cdk.CfnOutput(scope,"clusterEndpoint", {value: cluster.clusterEndpoint})
    new cdk.CfnOutput(scope,"clusterArn", {value: cluster.clusterArn})
    new cdk.CfnOutput(scope,"autoScalingGroupName", {value: spotNodeGroup.autoScalingGroup.autoScalingGroupName || ""})
    new cdk.CfnOutput(scope,"autoScalingGroupMaxSize", {value: spotNodeGroup.autoScalingGroup.maxSize })
    new cdk.CfnOutput(scope,"autoScalingGroupMinSize", {value: spotNodeGroup.autoScalingGroup.minSize })
    new cdk.CfnOutput(scope,"autoScalingGroupDesiredCapacity", {value: spotNodeGroup.autoScalingGroup.desiredCapacity || "" })
    new cdk.CfnOutput(scope,"nodesRoleARN", {value: spotNodeGroup.nodesRole.roleArn})
    new cdk.CfnOutput(scope,"capacityInstance", {value: this.capacityInstance})
    
    this.regionOfCluster = process.env.CDK_DEFAULT_REGION || "";
    this.clusterName = cluster.clusterName;
    this.clusterVersion = props.clusterVersion || '1.16';
    this.masterRoleARN = eksRole.roleArn;
    this.clusterEndpoint = cluster.clusterEndpoint;
    this.clusterArn = cluster.clusterArn;
    this.autoScalingGroupName = spotNodeGroup.autoScalingGroup.autoScalingGroupName || "";
    this.autoScalingGroupMaxSize = spotNodeGroup.autoScalingGroup.maxSize;
    this.autoScalingGroupMinSize = spotNodeGroup.autoScalingGroup.minSize;
    this.autoScalingGroupDesiredCapacity = spotNodeGroup.autoScalingGroup.desiredCapacity || "";
    this.nodesRoleARN = spotNodeGroup.nodesRole.roleArn;
    
    cluster.clusterEndpoint
    cluster.clusterCertificateAuthorityData

  }


}
