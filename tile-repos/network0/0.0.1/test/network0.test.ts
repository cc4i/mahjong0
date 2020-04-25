import { expect as expectCDK, haveResource, SynthUtils } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import Network0 = require('../lib/index');

test('VPC Created', () => {
    const app = new cdk.App();
    const stack = new cdk.Stack(app, "TestStack");
    // WHEN
    new Network0.Network0(stack, 'MyTestConstruct', {});
    // THEN
    expectCDK(stack).to(haveResource("AWS::EC2::VPC"));
});

