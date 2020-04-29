import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');

/** Input parameters */
export interface Network0Props {
  cidr?: string
  
}

export class Network0 extends cdk.Construct {

  /** Directly exposed to other stack  */
  public readonly baseVpc: ec2.Vpc;
  public readonly vpcId: string;
  public readonly publicSubnetId1: string;
  public readonly publicSubnetId2: string;
  public readonly privateSubnetId1: string;
  public readonly privateSubnetId2: string;
  public readonly availabilityZones: string[];
  public readonly publicSubnet1: ec2.ISubnet;
  public readonly publicSubnet2: ec2.ISubnet;
  public readonly privateSubnet1: ec2.ISubnet;
  public readonly privateSubnet2: ec2.ISubnet;

  constructor(scope: cdk.Construct, id: string, props: Network0Props) {
    super(scope, id);

    /**  Create the base VPC */
    this.baseVpc = new ec2.Vpc(this, "BaseVPC", {
      cidr: props.cidr || "10.0.0.0/16",
      maxAzs: 2
    });


    /** Tags for EKS */
    this.baseVpc.publicSubnets[0].node.applyAspect(new cdk.Tag("kubernetes.io/role/elb","1"));
    this.baseVpc.publicSubnets[1].node.applyAspect(new cdk.Tag("kubernetes.io/role/elb","1"));
    this.baseVpc.privateSubnets[0].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));
    this.baseVpc.privateSubnets[1].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));

   
    this.publicSubnetId1 = this.baseVpc.publicSubnets[0].subnetId;
    this.publicSubnetId2 = this.baseVpc.publicSubnets[1].subnetId;
    this.privateSubnetId1 = this.baseVpc.privateSubnets[0].subnetId;
    this.privateSubnetId2 = this.baseVpc.privateSubnets[1].subnetId;
    this.availabilityZones = this.baseVpc.availabilityZones;
    this.publicSubnet1 = this.baseVpc.publicSubnets[0]
    this.publicSubnet2 = this.baseVpc.publicSubnets[1];
    this.privateSubnet1 = this.baseVpc.privateSubnets[0];
    this.privateSubnet2 = this.baseVpc.privateSubnets[1];


    
  }
}
