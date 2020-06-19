import { expect as expectCDK, haveResource, SynthUtils } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import * as AWSElasticSearch from '../lib/index';

test('SQS Queue Created', () => {
    const app = new cdk.App();
    const stack = new cdk.Stack(app, "TestStack");
    // WHEN
    new AWSElasticSearch.AWSElasticSearch(stack, 'MyTestConstruct', {
        domainName: "test",
        masterUserName: "admin",
        masterUserPassword: "admin"
    });
    // THEN
    expectCDK(stack).to(haveResource("AWS::Elasticsearch::Domain"));
});

