#!/bin/bash

ROLE_NAME=ElasticsearchRoleforApp


aws iam get-role --role-name ${ROLE_NAME} || error=true
if [ ${error} ]
then
    echo "Nothing has been done and all clear."
else

    # aws iam delete-policy --policy-arn ${policyArn} || true
    aws iam delete-role \
        --role-name ${ROLE_NAME}
    echo "Deleted ${ROLE_NAME}"

fi
if [ $? -ne 0 ] 
then
    echo "Failed to delete role : exit 0"
    exit 0
fi

roleArn=`aws iam create-role \
    --role-name ${ROLE_NAME} \
    --assume-role-policy-document file://assume-role.json | jq -r '.Role.Arn'`
echo "Created ${ROLE_NAME} = ${roleArn}"


echo ${roleArn}>role.arn
