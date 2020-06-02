#!/bin/bash

# create secrets into Secret Manager
SECRET_NAME=$1
GITHUB_USER=$2
ACCESS_TOKEN=$3
APP_CONF_REPO=$4


aws secretsmanager get-secret-value --secret-id $SECRET_NAME
if [ $? -eq '0' ]
then

    aws secretsmanager update-secret --secret-id $SECRET_NAME \
        --secret-string '[{"githubUser":"'$GITHUB_USER'"},{"accessToken":"'$ACCESS_TOKEN'"}, {"appConfRepo":"'$APP_CONF_REPO'"}]'

else 
    aws secretsmanager create-secret --name $SECRET_NAME \
        --secret-string '[{"githubUser":"'$GITHUB_USER'"},{"accessToken":"'$ACCESS_TOKEN'"}, {"appConfRepo":"'$APP_CONF_REPO'"}]'

fi 
