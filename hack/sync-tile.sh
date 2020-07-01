#!/bin/bash

aws_profile=sin
s3_bucket=cc-mahjong-0

if [ $# -ne 2 ]; then
    echo "Usage: sync-tile.sh [tile name] [tile version]"
    exit 0
fi

tile_name=$1
tile_version=$2
tile_dir=`echo ${tile_name} | tr '[:upper:]' '[:lower:]'`
tile_name_lowercase=`echo ${tile_name} | tr '[:upper:]' '[:lower:]'`
tile_tgz=/tmp/${tile_name_lowercase}.tgz
local_tile_repo=../repo/tile

echo "Syncing < ${tile_name} - ${tile_version} > to S3::${s3_bucket}"

cd ${local_tile_repo}/${tile_dir}/${tile_version}
tar --exclude='./node_modules' --exclude='.DS_Store' --exclude='role.arn' -zcvf ${tile_tgz} ./*
aws s3 cp ${tile_tgz} \
    s3://${s3_bucket}/tiles-repo/${tile_name_lowercase}/${tile_version}/${tile_name_lowercase}.tgz \
    --profile ${aws_profile} \
    --acl public-read

aws s3 cp tile-spec.yaml \
    s3://${s3_bucket}/tiles-repo/${tile_name_lowercase}/${tile_version}/tile-spec.yaml \
    --profile ${aws_profile} \
    --acl public-read


echo "Synced < ${tile_name} - ${tile_version} > to S3::${s3_bucket}"
