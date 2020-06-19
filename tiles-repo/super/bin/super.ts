#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');

//
// BO: import-libs 
//
{{range .TsLibs}}
{{if and (ne .TileCategory "ContainerApplication") (ne .TileCategory "Application")}}
import { {{.TileConstructName}} } from '../lib/{{.TileFolder}}/lib/index'
{{end}}
{{end}}
//
// EO: 
//

const app = new cdk.App();


//
// BO: import-stacks
//
{{range .TsStacks}}
{{if and (ne .TileCategory "ContainerApplication") (ne .TileCategory "Application")}}
export class {{.TileStackName}} extends cdk.Stack {
    public readonly {{.TileVariable}}: {{.TileConstructName}};

    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
      super(scope, id, props);
  
      this.{{.TileVariable}} = new {{.TileConstructName}}(this, '{{.TileConstructName}}', {
          {{ $map := .InputParameters }}
          {{ range $key, $value := $map }}
          {{ if (ne $value.IsOverrideField "yes") }}
          {{ $key }}: {{ $value.InputValueForTemplate }},
          {{ end }}
          {{ end }}
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
{{end}}
//
// EO: 
//
