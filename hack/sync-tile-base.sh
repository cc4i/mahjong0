#!/bin/bash

aws_profile=sin
s3_bucket=cc-mahjong-0


cdk_tile_tgz=/tmp/cdk-tile.tgz
app_tile_tgz=/tmp/cdk-tile.tgz

echo "Syncing < ${cdk_tile_tgz} > to S3::${s3_bucket}"

cd ../tiles-repo/cdk-tile
tar --exclude='./node_modules' --exclude='.DS_Store' -zcvf ${cdk_tile_tgz} ./*
aws s3 cp ${cdk_tile_tgz} \
    s3://${s3_bucket}/tiles-repo/tile/cdk-tile.tgz \
    --profile ${aws_profile} \
    --acl public-read

echo "Synced < ${cdk_tile_tgz} > to S3::${s3_bucket}"

echo "---+++---"

cd ../tiles-repo/app-tile
tar --exclude='./node_modules' --exclude='.DS_Store' -zcvf ${app_tile_tgz} ./*
aws s3 cp ${app_tile_tgz} \
    s3://${s3_bucket}/tiles-repo/tile/app-tile.tgz \
    --profile ${aws_profile} \
    --acl public-read

echo "Synced < ${app_tile_tgz} > to S3::${s3_bucket}"
