import { expect as expectCDK, matchTemplate, MatchStyle } from '@aws-cdk/assert';
import * as cdk from '@aws-cdk/core';
import Super = require('../lib/super-stack');

test('Empty Stack', () => {
    const app = new cdk.App();
    // WHEN
    const stack = new Super.SuperStack(app, 'MyTestStack');
    // THEN
    expectCDK(stack).to(matchTemplate({
      "Resources": {}
    }, MatchStyle.EXACT))
});
