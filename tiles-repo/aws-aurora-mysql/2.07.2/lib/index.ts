import * as cdk from '@aws-cdk/core';
import secretsManager = require('@aws-cdk/aws-secretsmanager');
import rds = require('@aws-cdk/aws-rds');
import ssm = require('@aws-cdk/aws-ssm');
import ec2 = require('@aws-cdk/aws-ec2');

export interface AwsAuroraMysqlProps {
  vpc: ec2.IVpc;
  username: string;
  dbname: string;

}

export class AwsAuroraMysql extends cdk.Construct {


  constructor(scope: cdk.Construct, id: string, props: AwsAuroraMysqlProps) {
    super(scope, id);

    const databaseCredentialsSecret = new secretsManager.Secret(this, 'DBCredentialsSecret', {
      secretName: props.dbname+`-aurora-mysql-credentials`,
      generateSecretString: {
        secretStringTemplate: JSON.stringify({
          username: props.username,
        }),
        excludePunctuation: true,
        includeSpace: false,
        generateStringKey: 'password'
      }
    });

    new ssm.StringParameter(this, 'DBCredentialsArn', {
      parameterName: props.dbname+`-aurora-mysql-credentials-arn`,
      stringValue: databaseCredentialsSecret.secretArn,
    });

    const mysql = new rds.DatabaseCluster(scope, "AuroraMySQL", {
      engine: rds.DatabaseClusterEngine.AURORA_MYSQL,
      engineVersion: '2.07.2',
      instanceProps: {
        instanceType: ec2.InstanceType.of(ec2.InstanceClass.R5, ec2.InstanceSize.LARGE),
        vpc: props.vpc,
      },
      masterUser: {
        username: databaseCredentialsSecret.secretValueFromJson('username').toString(),
        password: databaseCredentialsSecret.secretValueFromJson('password'),

      }
      
    })

  }
}
