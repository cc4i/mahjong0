#!/bin/bash

ROLE_NAME=go-bumblebee-jazz-codebuild-role
ROLE_POLICY=go-bumblebee-jazz-codebuild-policy


aws iam get-role --role-name ${ROLE_NAME} || error=true
if [ ${error} ]
then
    echo "Nothing has been done and all clear."
else
    ROLE_POLICY_ARN=`aws iam list-attached-role-policies --role-name ${ROLE_NAME} | jq -r '.AttachedPolicies[].PolicyArn'`
    aws iam detach-role-policy \
        --role-name ${ROLE_NAME} \
        --policy-arn ${ROLE_POLICY_ARN}
    echo "Detached ${ROLE_NAME} and ${ROLE_POLICY_ARN}"
    aws iam delete-policy \
        --policy-arn ${ROLE_POLICY_ARN}
    echo "Deleted ${ROLE_POLICY} = ${ROLE_POLICY_ARN}"
    aws iam delete-role \
        --role-name ${ROLE_NAME}
    echo "Deleted ${ROLE_NAME}"

fi


aws iam create-role \
    --role-name ${ROLE_NAME} \
    --assume-role-policy-document file://assume-role.json
echo "Created ${ROLE_NAME}"

ROLE_POLICY_ARN=`aws iam create-policy \
    --policy-name ${ROLE_POLICY} \
    --policy-document file://iam-role-policy.json | jq -r '.Policy.Arn'`
echo "Created ${ROLE_POLICY} = ${ROLE_POLICY_ARN}"

aws iam attach-role-policy \
    --role-name ${ROLE_NAME} \
    --policy-arn ${ROLE_POLICY_ARN}
echo "Atteched ${ROLE_POLICY_ARN} to ${ROLE_NAME} "
