#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');

//
// BO: import-libs 
//
{{range .TsLibs}}
import { {{.TileName}} } from '../lib/{{.TileFolder}}/lib/index'
{{end}}
//
// EO: 
//

const app = new cdk.App();


//
// BO: import-stacks
//
{{range .TsStacks}}
export class {{.TileStackName}} extends cdk.Stack {
    public readonly {{.TileVariable}}: {{.TileName}};

    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
      super(scope, id, props);
  
      this.{{.TileVariable}} = new {{.TileName}}(this, '{{ .TileName }}', {
          {{ $map := .InputParameters }}
          {{ range $key, $value := $map }}
          {{ $key }}: {{ $value }},
          {{end}}
      })
    }
}

const {{.TileStackVariable}} = new {{.TileStackName}}(app, '{{.TileStackName}}', {
    env: {
        account: process.env.CDK_DEFAULT_ACCOUNT,
        region: process.env.CDK_DEFAULT_REGION
    }
});
{{end}}
//
// EO: 
//
