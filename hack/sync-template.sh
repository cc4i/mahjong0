#!/bin/bash


aws_profile=sin
s3_bucket=cc-mahjong-0

for t in `ls ../templates`
do 
    file_size_kb=`du -k ../templates/${t} | cut -f1`
    if [ $file_size_kb -eq 0 ]
    then 
    
        echo "${t} is ZERO & ignored."
    else
    
        echo "Syncing ${t} to s3::${s3_bucket}"
        aws s3 cp ../templates/${t} \
            s3://${s3_bucket}/templates/${t} \
            --profile ${aws_profile} \
            --acl public-read
    fi

done