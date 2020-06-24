#!/bin/bash

while true
do
    endpoint=`aws es describe-elasticsearch-domain \
        --domain-name $1 |jq -r '.DomainStatus.Endpoint'`
    if [ $endpoint != "null" ]
    then 
        
        echo "Domain : $1 is ready : $endpoint."
        break
    fi
    
    echo "Domain : $1 is not ready yet."

    sleep 10
done

