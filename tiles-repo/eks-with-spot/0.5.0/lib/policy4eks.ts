

import cdk = require('@aws-cdk/core');
import ec2 = require('@aws-cdk/aws-ec2');
import eks = require('@aws-cdk/aws-eks');
import autoscaling = require('@aws-cdk/aws-autoscaling');
import iam = require('@aws-cdk/aws-iam');


export interface PolicyProps {
}

export class EksNodesSpot extends cdk.Construct {
    
    public eksInlinePolicy : { [name: string]:iam.PolicyDocument}

    constructor(scope: cdk.Construct, id: string, props: PolicyProps={}) {
        super(scope, id);

        const p = iam.PolicyDocument

    }
}