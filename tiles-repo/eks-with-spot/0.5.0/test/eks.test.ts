import { expect as expectCDK, haveResource, SynthUtils } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as ec2 from '@aws-cdk/aws-ec2';
import { EksWithSpot } from '../lib/index';

test('EKS Created', () => {
    const app = new cdk.App();
    const stack = new cdk.Stack(app, "TestStack");
    const vpc = new ec2.Vpc(stack, "vpc", {})
    // WHEN
    const cluster = new EksWithSpot(stack, "MyTestConstruct", {
        vpc: vpc,
        clusterName: "testCluster",
        keyPair4EC2: "keypair",
        maxSizeASG: "1",
        minSizeASG: "1",
        desiredCapacityASG: "1",
        cooldownASG: "1",
        onDemandPercentage: 25,
    });
    // THEN
    expectCDK(stack).to(haveResource("AWS::CloudFormation::Stack"));
});


