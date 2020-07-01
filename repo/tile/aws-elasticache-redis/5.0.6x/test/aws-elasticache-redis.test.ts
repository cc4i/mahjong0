import { expect as expectCDK, haveResource, SynthUtils } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');
import * as AWSElastiCacheRedis from '../lib/index';

test('Redis Cluster Created', () => {
    const app = new cdk.App();
    const stack = new cdk.Stack(app, "TestStack");
    // WHEN
    const vpc = new ec2.Vpc(stack, "VPC",{});
    new AWSElastiCacheRedis.AWSElastiCacheRedis(stack, 'MyTestConstruct', {
      vpcId: vpc.vpcId,
      subnetIds: [vpc.privateSubnets[0].subnetId]
    });
    // THEN
    expectCDK(stack).to(haveResource("AWS::ElastiCache::ReplicationGroup"));
});

