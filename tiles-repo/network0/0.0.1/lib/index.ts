import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');
import { Tag } from '@aws-cdk/core';

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
  public readonly isolatedSubnetId1: string;
  public readonly isolatedSubnetId2: string;
  public readonly availabilityZones: string[];
  public readonly publicSubnet1: ec2.ISubnet;
  public readonly publicSubnet2: ec2.ISubnet;
  public readonly privateSubnet1: ec2.ISubnet;
  public readonly privateSubnet2: ec2.ISubnet;
  public readonly isolatedSubnet1: ec2.ISubnet;
  public readonly isolatedSubnet2: ec2.ISubnet;

  constructor(scope: cdk.Construct, id: string, props: Network0Props) {
    super(scope, id);

    /**  Create the base VPC */
    this.baseVpc = new ec2.Vpc(this, "BaseVPC", {
      cidr: props.cidr || "10.0.0.0/16",
      maxAzs: 2,
      subnetConfiguration: [
        {
          cidrMask: 19,
          name: 'ingress',
          subnetType: ec2.SubnetType.PUBLIC,
        },
        {
          cidrMask: 19,
          name: 'application',
          subnetType: ec2.SubnetType.PRIVATE,
        },
        {
          cidrMask: 19,
          name: 'rds',
          subnetType: ec2.SubnetType.ISOLATED,
        }
      ]
    });


    /** Tags for EKS */
    this.baseVpc.publicSubnets[0].node.applyAspect(new cdk.Tag("kubernetes.io/role/elb","1"));
    this.baseVpc.publicSubnets[1].node.applyAspect(new cdk.Tag("kubernetes.io/role/elb","1"));
    this.baseVpc.privateSubnets[0].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));
    this.baseVpc.privateSubnets[1].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));
    this.baseVpc.isolatedSubnets[0].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));
    this.baseVpc.isolatedSubnets[1].node.applyAspect(new cdk.Tag("kubernetes.io/role/internal-elb","1"));

    this.vpcId = this.baseVpc.vpcId;
    this.publicSubnetId1 = this.baseVpc.publicSubnets[0].subnetId;
    this.publicSubnetId2 = this.baseVpc.publicSubnets[1].subnetId;
    this.privateSubnetId1 = this.baseVpc.privateSubnets[0].subnetId;
    this.privateSubnetId2 = this.baseVpc.privateSubnets[1].subnetId;
    this.isolatedSubnetId1 = this.baseVpc.isolatedSubnets[0].subnetId;
    this.isolatedSubnetId2 = this.baseVpc.isolatedSubnets[1].subnetId;
    this.availabilityZones = this.baseVpc.availabilityZones;
    this.publicSubnet1 = this.baseVpc.publicSubnets[0]
    this.publicSubnet2 = this.baseVpc.publicSubnets[1];
    this.privateSubnet1 = this.baseVpc.privateSubnets[0];
    this.privateSubnet2 = this.baseVpc.privateSubnets[1];
    this.isolatedSubnet1 = this.baseVpc.isolatedSubnets[0];
    this.isolatedSubnet2 = this.baseVpc.isolatedSubnets[1];

    /** Added CF Output */
    new cdk.CfnOutput(this,"vpcId", {value: this.vpcId})
    new cdk.CfnOutput(this,"publicSubnetId1", {value: this.publicSubnetId1})
    new cdk.CfnOutput(this,"publicSubnetId2", {value: this.publicSubnetId2})
    new cdk.CfnOutput(this,"privateSubnetId1", {value: this.privateSubnetId1})
    new cdk.CfnOutput(this,"privateSubnetId2", {value: this.privateSubnetId2})
    new cdk.CfnOutput(this,"isolatedSubnetId1", {value: this.isolatedSubnetId1})
    new cdk.CfnOutput(this,"isolatedSubnetId2", {value: this.isolatedSubnetId2})
    let i=1
    this.availabilityZones.forEach( z => {
      new cdk.CfnOutput(this,"availabilityZones_"+z, {value: z});
      i++;
    });
    
    
  }
}
