import cdk = require('@aws-cdk/core');
import ec2 = require('@aws-cdk/aws-ec2');
import eks = require('@aws-cdk/aws-eks');
import autoscaling = require('@aws-cdk/aws-autoscaling');
import iam = require('@aws-cdk/aws-iam');
import { NodePolicies } from './policy4eks'


export interface EksNodesProps {
    clusterName: string;
    clusterVersion: string;
    clusterEndpoint: string;
    clusterCertificateAuthorityData: string
    vpc: ec2.Vpc;
    publicSubnetId1: string;
    publicSubnetId2: string;
    privateSubnetId1: string;
    privateSubnetId2: string;
    capacityInstance?: string[];
    capacity?: number;
    keyPair4EC2: string;
    maxSizeASG: string;
    minSizeASG: string;
    desiredCapacityASG: string;
    cooldownASG: string;
    onDemandPercentage: number;
    controlPlaneSG: ec2.SecurityGroup;
}

export class EksNodesSpot extends cdk.Construct {

    nodesLaunchTemplate: ec2.CfnLaunchTemplate;
    autoScalingGroup: autoscaling.CfnAutoScalingGroup;
    nodesRole: iam.Role;
    
    nodesSG: ec2.SecurityGroup;


    constructor(scope: cdk.Construct, id: string, props: EksNodesProps) {
        super(scope, id);
        let uuid = Math.random().toString(36).substr(2,5);

       

        /** work nodes security group */ 
        this.nodesSG = new ec2.SecurityGroup(this, "NodesSecurityGroup",{
            securityGroupName: "nodes-for-eks-sg-"+uuid,
            vpc: props.vpc,
        });
        /**ssh access to nodes */
        this.nodesSG.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(22));
        /** control plance access to nodes */ 
        this.nodesSG.addIngressRule(props.controlPlaneSG, ec2.Port.allTraffic())
        this.nodesSG.addIngressRule(this.nodesSG, ec2.Port.allTraffic())
        //access to control panel
        props.controlPlaneSG.addIngressRule(this.nodesSG, ec2.Port.allTraffic())
        props.controlPlaneSG.addIngressRule(props.controlPlaneSG, ec2.Port.allTraffic())


        /** Tag security group  */
        this.nodesSG.node.applyAspect(new cdk.Tag("kubernetes.io/cluster/"+props.clusterName,"owned"));

        
        /** organize managed polices */
        let region = process.env.CDK_DEFAULT_REGION
        let managedPolicies = []
        if (region == "cn-north-1" || region == "cn-northwest-1" ) {
            managedPolicies = [
                {managedPolicyArn: "arn:aws-cn:iam::aws:policy/AmazonEKSWorkerNodePolicy"},
                {managedPolicyArn: "arn:aws-cn:iam::aws:policy/AmazonEKS_CNI_Policy"},
                {managedPolicyArn: "arn:aws-cn:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"},
                {managedPolicyArn: "arn:aws-cn:iam::aws:policy/CloudWatchAgentServerPolicy"}
          ]
        } else {
            managedPolicies = [
                {managedPolicyArn: "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"},
                {managedPolicyArn: "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"},
                {managedPolicyArn: "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"},
                {managedPolicyArn: "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"}
          ]
          
        }
        /** create role and attach policies */
        this.nodesRole = new iam.Role(this, "NodesRole", {
            roleName: "nodes-for-eks-role-"+uuid,
            assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
            managedPolicies: managedPolicies,
            inlinePolicies: new NodePolicies(this, "policy", {}).eksInlinePolicy
        });
        /** attach policy to instance profile */
        let ec2Profile = new iam.CfnInstanceProfile(this,"ec2Profile",{
            roles: [this.nodesRole.roleName]
        });

        /** define launch template */
        let imageId = new eks.EksOptimizedImage({
            nodeType: eks.NodeType.STANDARD,
            kubernetesVersion: props.clusterVersion
        }).getImage(scope).imageId
        this.nodesLaunchTemplate = new ec2.CfnLaunchTemplate(this, "NodesLaunchTemplate", {
            launchTemplateName: "NodesLaunchTemplate-"+uuid,
            launchTemplateData: {
                instanceType: "c5.large",
                imageId: imageId,
                keyName: props.keyPair4EC2,
                iamInstanceProfile: {arn: ec2Profile.attrArn},
                securityGroupIds: [this.nodesSG.securityGroupId],
                blockDeviceMappings: [{
                    deviceName: "/dev/xvda",
                    ebs: {
                        volumeSize: 40,
                        deleteOnTermination: true
                    }
                }]
                
            },
        });
        
        
        /** covert to property */
        let newOverrides: autoscaling.CfnAutoScalingGroup.LaunchTemplateOverridesProperty[] = []
        props.capacityInstance?.forEach(ci => {
            newOverrides.push({instanceType: ci})
        });
        /** default capacity instances */
        if (newOverrides.length == 0) {
            newOverrides = [{instanceType: "c5.large"},{instanceType: "r5.large"},{instanceType: "m5.large"}]
        }
        
        /** define auto scaling group */
        this.autoScalingGroup = new autoscaling.CfnAutoScalingGroup(this, "NodesAutoScalingGroup", {

            autoScalingGroupName: props.clusterName+"-nodegroup-"+uuid,
            vpcZoneIdentifier: [
                props.publicSubnetId1,
                props.publicSubnetId2,
                props.privateSubnetId1,
                props.privateSubnetId2
            ],
            desiredCapacity: props.desiredCapacityASG,
            cooldown: props.cooldownASG,
            healthCheckType: "EC2",
            maxSize: props.maxSizeASG,
            minSize: props.minSizeASG,
            mixedInstancesPolicy: {
                instancesDistribution: {
                    onDemandBaseCapacity: 0,
                    onDemandPercentageAboveBaseCapacity: props.onDemandPercentage,
                    /* @see https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-autoscaling-autoscalinggroup-instancesdistribution.html#cfn-autoscaling-autoscalinggroup-instancesdistribution-spotallocationstrategy */
                    /* Valid values: lowest-price | capacity-optimized */
                    spotAllocationStrategy: "capacity-optimized",
                    //spotInstancePools: 2
                },
                launchTemplate: {
                    launchTemplateSpecification: {
                        launchTemplateName: this.nodesLaunchTemplate.launchTemplateName,
                        version: this.nodesLaunchTemplate.attrLatestVersionNumber
                    },
                    overrides: newOverrides
                }
            },
            // Pretty important to add tag - "kubernetes.io/cluster/", otherwaise nodes 
            // would be failed to register into EKS cluster
            tags: [
                {
                    key: "Name",
                    value: props.clusterName+"-nodegroup-"+uuid,
                    propagateAtLaunch: true
                },
                {
                    key: "member",
                    value: props.clusterName+"-nodegroup-"+uuid,
                    propagateAtLaunch: true
                },
                {
                    key: "kubernetes.io/cluster/"+props.clusterName,
                    value: "owned",
                    propagateAtLaunch: true
                },
                {
                    key: "k8s.io/cluster-autoscaler/enabled",
                    value: "true",
                    propagateAtLaunch: true
                },
                {
                    key: " k8s.io/cluster-autoscaler/"+props.clusterName,
                    value: "owned",
                    propagateAtLaunch: true
                }
  
            ]
        });
        
        this.autoScalingGroup.node.addDependency(this.nodesLaunchTemplate);

        this.autoScalingGroup.addOverride("UpdatePolicy",{
            "AutoScalingScheduledAction": {
                "IgnoreUnmodifiedGroupSizeProperties": true
            },
            "AutoScalingRollingUpdate": {
                "MinInstancesInService": "1",
                "MaxBatchSize": "1",
                "WaitOnResourceSignals": true,
                "MinSuccessfulInstancesPercent": "100",
            }
        });
        this.autoScalingGroup.addOverride("CreationPolicy",{
            "ResourceSignal": {
                "Count": props.desiredCapacityASG,
                "Timeout": "PT15M"
            }
        });


        /** moved here due to lookup 'logicalId' */ 
        this.nodesLaunchTemplate.addPropertyOverride("LaunchTemplateData.UserData",cdk.Fn.base64([
            `#!/bin/bash`,
            `set -ex`,
            `B64_CLUSTER_CA=${props.clusterCertificateAuthorityData}`,
            `API_SERVER_URL=${props.clusterEndpoint}`,
            `/etc/eks/bootstrap.sh ${props.clusterName} --kubelet-extra-args '--node-labels=eks.amazonaws.com/nodegroup-image=${imageId},eks.amazonaws.com/nodegroup=${this.autoScalingGroup.autoScalingGroupName}' --b64-cluster-ca $B64_CLUSTER_CA --apiserver-endpoint $API_SERVER_URL`,
            /* `/opt/aws/bin/cfn-init -v --stack ${cdk.Aws.STACK_NAME} --resource ${this.nodesLaunchTemplate.logicalId} --region ${cdk.Aws.REGION}`, */
            `/opt/aws/bin/cfn-signal -e $? --stack ${cdk.Aws.STACK_NAME} --resource ${this.autoScalingGroup.logicalId} --region ${cdk.Aws.REGION}`,
        ].join('\n')));

        
    }

}