#!/bin/bash

# Create secrets into Secret Manager
# Name for the secret
SECRET_NAME=$1
# GitHub user
GITHUB_USER=$2
# GitHub access token
ACCESS_TOKEN=$3
# Repo to storage manifest in GitHub
APP_CONF_REPO=$4
# Master key ID for those secrets
KMS_KEY_ID=$5


aws secretsmanager get-secret-value --secret-id $SECRET_NAME
if [ $? -eq '0' ]
then
    aws secretsmanager update-secret --secret-id $SECRET_NAME \
        --secret-string '{"githubUser":"'$GITHUB_USER'", "accessToken":"'$ACCESS_TOKEN'", "appConfRepo":"'$APP_CONF_REPO'"}'

else 
    aws secretsmanager create-secret --name $SECRET_NAME \
        --secret-string '{"githubUser":"'$GITHUB_USER'", "accessToken":"'$ACCESS_TOKEN'", "appConfRepo":"'$APP_CONF_REPO'"}'
    # aws secretsmanager create-secret --name $SECRET_NAME \
    #     --kms-key-id $KMS_KEY_ID \
    #     --secret-string '{"githubUser":"'$GITHUB_USER'", "accessToken":"'$ACCESS_TOKEN'", "appConfRepo":"'$APP_CONF_REPO'"}'

fi 
