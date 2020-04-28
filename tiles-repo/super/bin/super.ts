#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import ec2 = require('@aws-cdk/aws-ec2');

//
//{{ BO: import-libs }}
//
{{range .ImportLibs}}
import {{.TileName}} from ('../lib/{{.TileName}}/lib/index') 
{{end}}
//
//{{ EO: }}
//

const app = new cdk.App();


//
//{{ BO: import-stacks }}
//
export class {{.TileStack}} extends cdk.Stack {
    public readonly {{.TileVariable}}: {{.Tile}};

    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
      super(scope, id, props);
  
      this.{{.TileVariable}} = new {{.Tile}}(this, '{{Tile}}', {
          {{range .InputParameters}}
          {{.inputName}}: {{.inputValue}}
          {{end}}
      })
    }
}

const {{.TileStackVariable}} = new {{.TileStack}}(app, '{{.TileStack}}', {
    env: {
        account: process.env.CDK_DEFAULT_ACCOUNT,
        region: process.env.CDK_DEFAULT_REGION
    }
});
//
//{{ EO: }}
//
