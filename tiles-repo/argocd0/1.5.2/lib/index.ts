import * as cdk from '@aws-cdk/core';

export interface Argocd0Props {
  clusterName: string,
  masterRoleARN: string,
  namespace?: string,
  manifest?: string[],

}

export class Argocd0 extends cdk.Construct {
  
  public readonly adminUser: string;
  public readonly adminPassword: string;
  public readonly endpoint: string;


  constructor(scope: cdk.Construct, id: string, props: Argocd0Props) {
    super(scope, id);

  }
}
