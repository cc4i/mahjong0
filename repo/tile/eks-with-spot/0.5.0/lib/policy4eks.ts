

import cdk = require('@aws-cdk/core');
import iam = require('@aws-cdk/aws-iam');


export interface PolicyProps {}

export class NodePolicies extends cdk.Construct {
    
    public eksInlinePolicy : { [name: string]:iam.PolicyDocument}

    constructor(scope: cdk.Construct, id: string, props: PolicyProps) {
        super(scope, id);

        this.eksInlinePolicy = {
            "Autoscaler4eks": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "autoscaling:DescribeAutoScalingGroups",
                            "autoscaling:DescribeAutoScalingInstances",
                            "autoscaling:DescribeLaunchConfigurations",
                            "autoscaling:DescribeTags",
                            "autoscaling:SetDesiredCapacity",
                            "autoscaling:TerminateInstanceInAutoScalingGroup",
                            "ec2:DescribeLaunchTemplateVersions"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }), 
            "ALBIngress": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "acm:DescribeCertificate",
                            "acm:ListCertificates",
                            "acm:GetCertificate",
                            "ec2:AuthorizeSecurityGroupIngress",
                            "ec2:CreateSecurityGroup",
                            "ec2:CreateTags",
                            "ec2:DeleteTags",
                            "ec2:DeleteSecurityGroup",
                            "ec2:DescribeAccountAttributes",
                            "ec2:DescribeAddresses",
                            "ec2:DescribeInstances",
                            "ec2:DescribeInstanceStatus",
                            "ec2:DescribeInternetGateways",
                            "ec2:DescribeNetworkInterfaces",
                            "ec2:DescribeSecurityGroups",
                            "ec2:DescribeSubnets",
                            "ec2:DescribeTags",
                            "ec2:DescribeVpcs",
                            "ec2:ModifyInstanceAttribute",
                            "ec2:ModifyNetworkInterfaceAttribute",
                            "ec2:RevokeSecurityGroupIngress",
                            "elasticloadbalancing:AddListenerCertificates",
                            "elasticloadbalancing:AddTags",
                            "elasticloadbalancing:CreateListener",
                            "elasticloadbalancing:CreateLoadBalancer",
                            "elasticloadbalancing:CreateRule",
                            "elasticloadbalancing:CreateTargetGroup",
                            "elasticloadbalancing:DeleteListener",
                            "elasticloadbalancing:DeleteLoadBalancer",
                            "elasticloadbalancing:DeleteRule",
                            "elasticloadbalancing:DeleteTargetGroup",
                            "elasticloadbalancing:DeregisterTargets",
                            "elasticloadbalancing:DescribeListenerCertificates",
                            "elasticloadbalancing:DescribeListeners",
                            "elasticloadbalancing:DescribeLoadBalancers",
                            "elasticloadbalancing:DescribeLoadBalancerAttributes",
                            "elasticloadbalancing:DescribeRules",
                            "elasticloadbalancing:DescribeSSLPolicies",
                            "elasticloadbalancing:DescribeTags",
                            "elasticloadbalancing:DescribeTargetGroups",
                            "elasticloadbalancing:DescribeTargetGroupAttributes",
                            "elasticloadbalancing:DescribeTargetHealth",
                            "elasticloadbalancing:ModifyListener",
                            "elasticloadbalancing:ModifyLoadBalancerAttributes",
                            "elasticloadbalancing:ModifyRule",
                            "elasticloadbalancing:ModifyTargetGroup",
                            "elasticloadbalancing:ModifyTargetGroupAttributes",
                            "elasticloadbalancing:RegisterTargets",
                            "elasticloadbalancing:RemoveListenerCertificates",
                            "elasticloadbalancing:RemoveTags",
                            "elasticloadbalancing:SetIpAddressType",
                            "elasticloadbalancing:SetSecurityGroups",
                            "elasticloadbalancing:SetSubnets",
                            "elasticloadbalancing:SetWebACL",
                            "iam:CreateServiceLinkedRole",
                            "iam:GetServerCertificate",
                            "iam:ListServerCertificates",
                            "waf-regional:GetWebACLForResource",
                            "waf-regional:GetWebACL",
                            "waf-regional:AssociateWebACL",
                            "waf-regional:DisassociateWebACL",
                            "tag:GetResources",
                            "tag:TagResources",
                            "waf:GetWebACL",
                            "wafv2:GetWebACL",
                            "wafv2:GetWebACLForResource",
                            "wafv2:AssociateWebACL",
                            "wafv2:DisassociateWebACL",
                            "shield:DescribeProtection",
                            "shield:GetSubscriptionState",
                            "shield:DeleteProtection",
                            "shield:CreateProtection",
                            "shield:DescribeSubscription",
                            "shield:ListProtections"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "AppMesh": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "appmesh:*",
                            "servicediscovery:CreateService",
                            "servicediscovery:GetService",
                            "servicediscovery:RegisterInstance",
                            "servicediscovery:DeregisterInstance",
                            "servicediscovery:ListInstances",
                            "servicediscovery:ListNamespaces",
                            "servicediscovery:ListServices",
                            "route53:GetHealthCheck",
                            "route53:CreateHealthCheck",
                            "route53:UpdateHealthCheck",
                            "route53:ChangeResourceRecordSets",
                            "route53:DeleteHealthCheck"
            
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "CertManagerChangeSet": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "route53:ChangeResourceRecordSets"
                        ],
                        resources: ["arn:aws:route53:::hostedzone/*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "CertManagerGetChange": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "route53:GetChange"
                        ],
                        resources: ["arn:aws:route53:::change/*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "CertManagerHostedZone": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "route53:ListHostedZones",
                            "route53:ListResourceRecordSets",
                            "route53:ListHostedZonesByName",
                            "route53:ListTagsForResource"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "EBS": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "ec2:AttachVolume",
                            "ec2:CreateSnapshot",
                            "ec2:CreateTags",
                            "ec2:CreateVolume",
                            "ec2:DeleteSnapshot",
                            "ec2:DeleteTags",
                            "ec2:DeleteVolume",
                            "ec2:DescribeAvailabilityZones",
                            "ec2:DescribeInstances",
                            "ec2:DescribeSnapshots",
                            "ec2:DescribeTags",
                            "ec2:DescribeVolumes",
                            "ec2:DescribeVolumesModifications",
                            "ec2:DetachVolume",
                            "ec2:ModifyVolume"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "EFS": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "elasticfilesystem:*"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "EFSEC2": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "ec2:DescribeSubnets",
                            "ec2:CreateNetworkInterface",
                            "ec2:DescribeNetworkInterfaces",
                            "ec2:DeleteNetworkInterface",
                            "ec2:ModifyNetworkInterfaceAttribute",
                            "ec2:DescribeNetworkInterfaceAttribute"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "FSX": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "fsx:*"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "ServiceLinkRole": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "iam:CreateServiceLinkedRole",
                            "iam:AttachRolePolicy",
                            "iam:PutRolePolicy"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            }),
            "XRay": new iam.PolicyDocument({
                statements: [
                    new iam.PolicyStatement({
                        actions: [
                            "xray:PutTraceSegments",
                            "xray:PutTelemetryRecords",
                            "xray:GetSamplingRules",
                            "xray:GetSamplingTargets",
                            "xray:GetSamplingStatisticSummaries"
                        ],
                        resources: ["*"],
                        effect: iam.Effect.ALLOW
                    }),
                ]
            })
        }

    }
}