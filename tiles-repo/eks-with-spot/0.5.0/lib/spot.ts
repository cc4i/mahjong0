import cdk = require('@aws-cdk/core');
import ec2 = require('@aws-cdk/aws-ec2');
import eks = require('@aws-cdk/aws-eks');
import autoscaling = require('@aws-cdk/aws-autoscaling');
import iam = require('@aws-cdk/aws-iam');


export interface EksNodesProps {
    clusterName: string,
    clusterVersion: string,
    vpc: ec2.Vpc,
    publicSubnetId1: string,
    publicSubnetId2: string,
    privateSubnetId1: string,
    privateSubnetId2: string,
    capacityInstance?: string[],
    capacity?: number,
    keyPair4EC2: string,
    maxSizeASG: string,
    minSizeASG: string,
    desiredCapacityASG: string,
    cooldownASG: string,
    onDemandPercentage: number,
}

export class EksNodesSpot extends cdk.Construct {

    nodesLaunchTemplate: ec2.CfnLaunchTemplate
    autoScalingGroup: autoscaling.CfnAutoScalingGroup
    nodesRole: iam.Role
    
    controlPlaneSG: ec2.SecurityGroup
    nodesSG: ec2.SecurityGroup
    nodesSharedSG: ec2.SecurityGroup


    constructor(scope: cdk.Construct, id: string, props: EksNodesProps) {
        super(scope, id)

        /** control panel security group  */ 
        this.controlPlaneSG = new ec2.SecurityGroup(this, `EksControlPlaneSG`, {
            vpc: props.vpc
          });

        /** work nodes security group */ 
        this.nodesSG = new ec2.SecurityGroup(this, "NodesSecurityGroup",{
            securityGroupName: "nodes-for-eks-sg",
            vpc: props.vpc
        });
        /**s sh access to nodes */
        this.nodesSG.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(22));
        /** control plance access to nodes */ 
        this.nodesSG.addIngressRule(this.controlPlaneSG, ec2.Port.tcpRange(1025,65535))
        this.nodesSG.addIngressRule(this.controlPlaneSG, ec2.Port.tcp(443))
        //access to control panel
        this.controlPlaneSG.addIngressRule(this.nodesSG, ec2.Port.tcp(443))

        /** work nodes shared scurity group */
        this.nodesSharedSG = new ec2.SecurityGroup(this, "NodesSharedSecurityGroup",{
            securityGroupName: "nodes-shared-for-eks-sg",
            vpc: props.vpc
        });
        this.nodesSharedSG.addIngressRule(this.nodesSharedSG, ec2.Port.allTcp())
         
        this.nodesRole = new iam.Role(this, "NodesRole", {
            roleName: "nodes-for-eks-role",
            assumedBy: new iam.ServicePrincipal('ec2.amazonaws.com'),
            managedPolicies: [
                {managedPolicyArn: "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"},
                {managedPolicyArn: "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"},
                {managedPolicyArn: "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"}
            ]
        });
        let ec2Profile = new iam.CfnInstanceProfile(this,"ec2Profile",{
            roles: [this.nodesRole.roleName]
        });

        this.nodesLaunchTemplate = new ec2.CfnLaunchTemplate(this, "NodesLaunchTemplate", {
            launchTemplateName: "NodesLaunchTemplate",
            launchTemplateData: {
                instanceType: "c5.large",
                imageId: new eks.EksOptimizedImage({
                    nodeType: eks.NodeType.STANDARD,
                    kubernetesVersion: props.clusterVersion
                }).getImage(scope).imageId,
                keyName: props.keyPair4EC2,
                iamInstanceProfile: {arn: ec2Profile.attrArn},
                securityGroupIds: [this.nodesSG.securityGroupId, this.nodesSharedSG.securityGroupId],
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
        

        this.autoScalingGroup = new autoscaling.CfnAutoScalingGroup(this, "NodesAutoScalingGroup", {
            
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
                    value: "nodes-asg-"+props.clusterName,
                    propagateAtLaunch: true
                },
                {
                    key: "kubernetes.io/cluster/"+props.clusterName,
                    value: "owned",
                    propagateAtLaunch: true
                },

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
            `set -e`,
            `sudo yum update -y`,
            `sudo yum install -y aws-cfn-bootstrap aws-cli jq wget`,
            `/etc/eks/bootstrap.sh ${props.clusterName}`,
            /* `/opt/aws/bin/cfn-init -v --stack ${cdk.Aws.STACK_NAME} --resource ${this.nodesLaunchTemplate.logicalId} --region ${cdk.Aws.REGION}`, */
            `/opt/aws/bin/cfn-signal -e $? --stack ${cdk.Aws.STACK_NAME} --resource ${this.autoScalingGroup.logicalId} --region ${cdk.Aws.REGION}`,
        ].join('\n')));

        
    }

}