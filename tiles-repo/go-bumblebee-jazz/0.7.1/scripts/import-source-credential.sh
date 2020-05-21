#!/bin/bash

aws codebuild import-source-credentials \
    --server-type GITHUB \
    --auth-type PERSONAL_ACCESS_TOKEN \
    --token <cc4i_token> \
    --username <cc4i>