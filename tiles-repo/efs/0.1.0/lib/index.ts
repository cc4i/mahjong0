import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');
import efs = require('@aws-cdk/aws-efs');

export interface EfsProps {
  vpc: ec2.Vpc,
}

export class Efs extends cdk.Construct {
  /** @returns file system Id of EFS */
  public readonly fileSystemId: string;

  constructor(scope: cdk.Construct, id: string, props: EfsProps) {
    super(scope, id);

    const fileSystem = new efs.FileSystem(this, "EfsFileSystem", {
      vpc: props.vpc,
      encrypted: true,
      lifecyclePolicy: efs.LifecyclePolicy.AFTER_14_DAYS,
      performanceMode: efs.PerformanceMode.GENERAL_PURPOSE,
      throughputMode: efs.ThroughputMode.BURSTING
    })

    this.fileSystemId =  fileSystem.fileSystemId

    new cdk.CfnOutput(scope, "fileSystemId", {value: this.fileSystemId})
  }
}
