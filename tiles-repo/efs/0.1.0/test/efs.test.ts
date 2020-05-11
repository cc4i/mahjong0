import { expect as expectCDK, haveResource, SynthUtils } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as Efs from '../lib/index';
import * as ec2 from '@aws-cdk/aws-ec2';

test('EFS Created', () => {
    const app = new cdk.App();
    const stack = new cdk.Stack(app, "TestStack");
    const vpc = new ec2.Vpc(stack, "vpc", {})
    // WHEN
    new Efs.Efs(stack, 'MyTestConstruct', {
      vpc: vpc,
    });
    // THEN
    expectCDK(stack).to(haveResource("AWS::EFS::FileSystem"));
});


