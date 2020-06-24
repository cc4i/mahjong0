#!/bin/bash


endpoint=`aws es describe-elasticsearch-domain \
    --domain-name $1 |jq -r '.DomainStatus.Endpoint'`
if [ $endpoint = "null" ]
then 
    echo "Domain : $1 is not ready yet."
    exit 1
fi
echo "Domain : $1 is ready : $endpoint."
