import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');

export interface Network0Props {
  cidr?: string
  
}

export class Network0 extends cdk.Construct {
  /** @returns the ARN of the SQS queue */
  public readonly vpcId: string;
  public readonly publicSubnetId1: string;
  public readonly publicSubnetId2: string;
  public readonly privateSubnetId1: string;
  public readonly privateSubnetId2: string;
  public readonly availabilityZones: string[];

  public readonly baseVpc: ec2.Vpc;

  constructor(scope: cdk.Construct, id: string, props: Network0Props) {
    super(scope, id);

    // Create the base VPC
    this.baseVpc = new ec2.Vpc(this, "BaseVPC", {
      cidr: props.cidr || "10.0.0.0/16",
      maxAzs: 2
    });

    this.publicSubnetId1 = this.baseVpc.publicSubnets[0].subnetId;
    this.publicSubnetId2 = this.baseVpc.publicSubnets[1].subnetId;
    this.privateSubnetId1 = this.baseVpc.privateSubnets[0].subnetId
    this.privateSubnetId2 = this.baseVpc.privateSubnets[1].subnetId;
    this.availabilityZones = this.baseVpc.availabilityZones;

    // Tags for EKS 
    this.baseVpc.publicSubnets[0].node.applyAspect(new cdk.Tag("kubernetes.io/role/elb","1"));
    this.baseVpc.publicSubnets[1].node.applyAspect(new cdk.Tag("kubernetes.io/role/elb","1"));
    this.baseVpc.privateSubnets[0].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));
    this.baseVpc.privateSubnets[1].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));


    
  }
}
