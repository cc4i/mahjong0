import { expect as expectCDK, haveResource, SynthUtils } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');
import * as AwsAuroraMysql from '../lib/index';


test('Aurora MySQL Created', () => {
  const app = new cdk.App();
  const stack = new cdk.Stack(app, "TestStack");
  const vpc = new ec2.Vpc(stack,"vpc",{})
  // WHEN
  new AwsAuroraMysql.AWSAuroraMysql(stack, 'MyTestConstruct',{
    vpc: vpc,
    masterUser: "dbadmin",
    clusterIdentifier: "DbMySQL"
  });
  // THEN
  expectCDK(stack).to(haveResource("AWS::RDS::DBCluster"));
});
