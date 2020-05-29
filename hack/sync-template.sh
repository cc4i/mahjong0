#!/bin/bash


aws_profile=sin
s3_bucket=cc-mahjong-0

for t in `ls ../templates`
do 
    echo "Syncing ${t} to s3::${s3_bucket}"
    aws s3 cp ../templates/${t} \
        s3://${s3_bucket}/templates/${t} \
        --profile ${aws_profile} \
        --acl public-read

done