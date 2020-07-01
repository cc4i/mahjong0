import * as cdk from '@aws-cdk/core';
import * as ec from '@aws-cdk/aws-elasticache'
import ec2 = require('@aws-cdk/aws-ec2');

export interface AWSElastiCacheRedisProps {
  vpcId: string;
  subnetIds: string[];
  redisClusterName?: string;
  replicasPerNodeGroup?: number;
  numNodeGroups?: number;
  engineVersion?: string;
  autoMinorVersionUpgrade?: boolean;
}

export class AWSElastiCacheRedis extends cdk.Construct {

  public readonly redisClusterName:string;
  public readonly redisEndpoint: string;

  constructor(scope: cdk.Construct, id: string, props: AWSElastiCacheRedisProps) {
    super(scope, id);
    let uuid = Math.random().toString(36).substr(2,5);

    const subnetGroup = new ec.CfnSubnetGroup(scope, "SubnetGroup", {
      description: "subnet group for redis-cluster",
      subnetIds: props.subnetIds,
      cacheSubnetGroupName: "redis-subnetgroup-"+uuid,
    });

    let vpc = ec2.Vpc.fromLookup(scope, "vpc", {
      vpcId: props.vpcId
    })
    const redisClusterSG = new ec2.SecurityGroup(scope, "SecurityGroup", {
      securityGroupName: "redis-cluster-sg-"+uuid,
      vpc: vpc,
    });

    redisClusterSG.addIngressRule(ec2.Peer.ipv4(vpc.vpcCidrBlock), ec2.Port.tcp(6379));

    const redis = new ec.CfnReplicationGroup(scope, "ReplicationGroup", {
      replicationGroupId: props.redisClusterName+"-"+uuid,
      replicationGroupDescription: props.redisClusterName+"-"+uuid,
      replicasPerNodeGroup: props.replicasPerNodeGroup || 2,
      numNodeGroups: props.numNodeGroups || 2,
      engine: "redis",
      cacheNodeType: "cache.t3.medium",
      engineVersion: props.engineVersion || "5.0.6",
      autoMinorVersionUpgrade: props.autoMinorVersionUpgrade || true,
      automaticFailoverEnabled: true,
      securityGroupIds: [redisClusterSG.securityGroupId],
      cacheSubnetGroupName: subnetGroup.cacheSubnetGroupName,
      tags: [
        {
          key: "member",
          value: "redis-cluster-"+uuid
        }
      ]
            
    });
    redis.addDependsOn(subnetGroup);

    this.redisClusterName = redis.replicationGroupId!
    this.redisEndpoint = redis.attrConfigurationEndPointAddress+":"+redis.attrConfigurationEndPointPort
    
    /** Output from CF */
    new cdk.CfnOutput(scope,"redisClusterName", {value: this.redisClusterName})
    new cdk.CfnOutput(scope,"redisEndpoint", {value: this.redisEndpoint})

  }
  
}
