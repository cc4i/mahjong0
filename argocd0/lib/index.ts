import * as cdk from '@aws-cdk/core';

export interface Argocd0Props {

}

export class Argocd0 extends cdk.Construct {
  
  public readonly argocdServer: string;
  public readonly argocdServerCmd: string;
  public readonly argocdInstall: string;

  constructor(scope: cdk.Construct, id: string, props: Argocd0Props = {}) {
    super(scope, id);

  }
}
