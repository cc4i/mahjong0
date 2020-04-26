#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');

//
//{{ BO: import-libs }}
//
{{ import-libs }}
//
//{{ EO: }}
//

const app = new cdk.App();


//
//{{ BO: import-stacks }}
//
{{ import-stacks }}
//
//{{ EO: }}
//
