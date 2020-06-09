import * as cdk from '@aws-cdk/core';
import secretsManager = require('@aws-cdk/aws-secretsmanager');
import rds = require('@aws-cdk/aws-rds');
import ssm = require('@aws-cdk/aws-ssm');
import ec2 = require('@aws-cdk/aws-ec2');

export interface AWSAuroraMysqlProps {
  vpc: ec2.Vpc;
  masterUser: string;
  clusterIdentifier: string;
  defaultDatabaseName?: string;

}

export class AWSAuroraMysql extends cdk.Construct {

  public readonly clusterIdentifier: string;
  public readonly clusterEndpoint: string;
  public readonly clusterReadEndpoint: string;
  public readonly defaultDatabaseName: string;

  constructor(scope: cdk.Construct, id: string, props: AWSAuroraMysqlProps) {
    super(scope, id);
    let uuid = Math.random().toString(36).substr(2,5);
    
    const databaseCredentialsSecret = new secretsManager.Secret(this, 'DBCredentialsSecret', {
      secretName: props.clusterIdentifier+`-aurora-mysql-credentials`,
      generateSecretString: {
        secretStringTemplate: JSON.stringify({
          username: props.masterUser,
        }),
        excludePunctuation: true,
        includeSpace: false,
        generateStringKey: 'password'
      }
    });

    new ssm.StringParameter(this, 'DBCredentialsArn', {
      parameterName: props.clusterIdentifier+`-aurora-mysql-credentials-arn`,
      stringValue: databaseCredentialsSecret.secretArn,
    });

    const p = new rds.ClusterParameterGroup(scope, "ClusterParameterGroup", {
      family: "aurora-mysql5.7",
      parameters: {
        "max_connections": "200"
      }
    })

    const mysql = new rds.DatabaseCluster(scope, "AuroraMySQL", {
      clusterIdentifier: props.clusterIdentifier,
      engine: rds.DatabaseClusterEngine.AURORA_MYSQL,
      engineVersion: '5.7.mysql_aurora.2.08.0',
      parameterGroup: p,
      instanceProps: {
        instanceType: ec2.InstanceType.of(ec2.InstanceClass.R5, ec2.InstanceSize.LARGE),
        vpcSubnets: {
          subnetType: ec2.SubnetType.PRIVATE
        },
        vpc: props.vpc,
      },
      masterUser: {
        username: databaseCredentialsSecret.secretValueFromJson('username').toString(),
        password: databaseCredentialsSecret.secretValueFromJson('password'),

      },
      defaultDatabaseName: props.defaultDatabaseName || uuid
    })
    mysql.node.addDependency(p)
    
    
    new cdk.CfnOutput(this,"clusterIdentifier", {value: mysql.clusterIdentifier})
    new cdk.CfnOutput(this,"clusterEndpoint", {value: mysql.clusterEndpoint.hostname +":"+ mysql.clusterEndpoint.port})
    new cdk.CfnOutput(this,"clusterReadEndpoint", {value: mysql.clusterReadEndpoint.hostname +":"+ mysql.clusterReadEndpoint.port})
    new cdk.CfnOutput(this,"defaultDatabaseName", {value: props.defaultDatabaseName || uuid})

    this.clusterIdentifier = mysql.clusterIdentifier;
    this.clusterEndpoint = mysql.clusterEndpoint.hostname +":"+ mysql.clusterEndpoint.port;
    this.clusterReadEndpoint = mysql.clusterReadEndpoint.hostname +":"+ mysql.clusterReadEndpoint.port;
    this.defaultDatabaseName = props.defaultDatabaseName || uuid
    
    

  }
}
