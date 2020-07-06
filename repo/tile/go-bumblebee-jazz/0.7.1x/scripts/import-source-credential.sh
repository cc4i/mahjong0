#!/bin/bash

aws codebuild import-source-credentials \
    --server-type GITHUB \
    --auth-type PERSONAL_ACCESS_TOKEN \
    --token $GIT_ACCESS_TOKEN \
    --username $GITHUB_USER