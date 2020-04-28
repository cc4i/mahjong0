#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');

//
//{{BO: import libs area}}
//

// generated
import Eks0 from '../lib/eks0/lib/index'
//
// generated
import Network0 from '../lib/network0/lib/index'
//
//{{EO:}}

const app = new cdk.App();


//
//{{BO: import stacks area}}
//
// generated
export class Network0Stack extends cdk.Stack {
    public readonly network0: Network0;

    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
      super(scope, id, props);
  
      this.network0 = new Network0(this, "eks0", {
          cidr: '10.0.0.0/16'
      })
    }
}

const network0statck = new Network0Stack(app, 'Network0Stack', {
    env: {
        account: process.env.CDK_DEFAULT_ACCOUNT,
        region: process.env.CDK_DEFAULT_REGION
    }
});
//

// generated
export class Eks0Stack extends cdk.Stack {
    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
      super(scope, id, props);
  
      const eks0 = new Eks0(this, "eks0", {
          vpc: network0statck.network0.baseVpc,
          clusterName: 'hello-eks'
      })
    }
  }
new Eks0Stack(app, 'Eks0Stack', {
    env: {
        account: process.env.CDK_DEFAULT_ACCOUNT,
        region: process.env.CDK_DEFAULT_REGION
    }
});  
//


//{{EO:}}


// generated


// generated
