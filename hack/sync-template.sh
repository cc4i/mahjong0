#!/bin/bash


aws_profile=sin
s3_bucket=cc-mahjong-0
local_hu_repo=../repo/hu

for t in `ls ${local_hu_repo}`
do 
    file_size_kb=`du -k ${local_hu_repo}/${t} | cut -f1`
    if [ $file_size_kb -eq 0 ]
    then 
    
        echo "${t} is ZERO & ignored."
    else
    
        echo "Syncing ${t} to s3::${s3_bucket}"
        aws s3 cp ${local_hu_repo}/${t} \
            s3://${s3_bucket}/templates/${t} \
            --profile ${aws_profile} \
            --acl public-read
    fi

done