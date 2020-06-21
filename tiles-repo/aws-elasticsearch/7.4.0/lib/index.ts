import * as cdk from '@aws-cdk/core';
import * as es from '@aws-cdk/aws-elasticsearch';
import { Stack } from '@aws-cdk/core';

export interface AWSElasticSearchProps {
  elasticsearchVersion?: string;
  domainName: string;
  dataInstanceCount?: number;
  dataInstanceType?: string;
  instanceVolumeSize?: number;
  instanceVolumeType?: string;
  masterInstanceCount?: number;
  masterInstanceType?: string;
  masterUserName: string;
  masterUserPassword: string;
  kmsKeyId?: string;

}

export class AWSElasticSearch extends cdk.Construct {

  public readonly domainName: string;
  public readonly domainEndpoint: string;

  constructor(scope: cdk.Construct, id: string, props: AWSElasticSearchProps) {
    super(scope, id);

    let partition = cdk.Aws.PARTITION;
    let region = cdk.Aws.REGION;
    let accountId = cdk.Aws.ACCOUNT_ID;

    const search = new es.CfnDomain(scope, "domain", {
      domainName: props.domainName,
      elasticsearchClusterConfig: {
        dedicatedMasterEnabled: true,
        instanceCount: props.dataInstanceCount || 2,
        zoneAwarenessEnabled: true,
        instanceType: props.dataInstanceType || "r5.large.elasticsearch",
        dedicatedMasterType: props.masterInstanceType || "r5.large.elasticsearch",
        dedicatedMasterCount: props.masterInstanceCount || 3,
      },
      elasticsearchVersion: props.elasticsearchVersion || "7.4",
      ebsOptions: {
        ebsEnabled: true,
        volumeSize: props.instanceVolumeSize || 40,
        volumeType: props.instanceVolumeType || "gp2"
      },
      encryptionAtRestOptions: {
        enabled: true
      },
      nodeToNodeEncryptionOptions: {
        enabled: true
      },
      accessPolicies: {
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Principal": {
              "AWS": accountId
            },
            "Action": "es:*",
            "Resource": "arn:"+partition+":es:"+region+":"+accountId+":domain/"+props.domainName+"/*",
            "Condition": {
              "IpAddress": {
                "aws:SourceIp": [
                  "127.0.0.1"
                ]
              }
            }
          }
        ]
      }
    });

    new cdk.CfnOutput(scope,"domainName", { value: search.domainName! });
    new cdk.CfnOutput(scope,"domainEndpoint", {value: search.attrDomainEndpoint });
    
    this.domainName = search.domainName!;
    this.domainEndpoint = search.attrDomainEndpoint;
    
  }
}
