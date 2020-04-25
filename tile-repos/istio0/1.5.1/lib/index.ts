import * as sns from '@aws-cdk/aws-sns';
import * as subs from '@aws-cdk/aws-sns-subscriptions';
import * as sqs from '@aws-cdk/aws-sqs';
import * as cdk from '@aws-cdk/core';

export interface Istio0Props {
  /**
   * The visibility timeout to be configured on the SQS Queue, in seconds.
   *
   * @default Duration.seconds(300)
   */
  visibilityTimeout?: cdk.Duration;
}

/**
 * 
 
helm template install/kubernetes/operator/operator-chart/ \
  --set hub=docker.io/istio \
  --set tag=1.5.1 \
  --set operatorNamespace=istio-operator \
  --set istioNamespace=istio-system | kubectl apply -f -


 */

export class Istio0 extends cdk.Construct {
  /** @returns the ARN of the SQS queue */
  public readonly queueArn: string;

  constructor(scope: cdk.Construct, id: string, props: Istio0Props = {}) {
    super(scope, id);

    const queue = new sqs.Queue(this, 'Istio0Queue', {
      visibilityTimeout: props.visibilityTimeout || cdk.Duration.seconds(300)
    });

    const topic = new sns.Topic(this, 'Istio0Topic');

    topic.addSubscription(new subs.SqsSubscription(queue));

    this.queueArn = queue.queueArn;
  }
}
