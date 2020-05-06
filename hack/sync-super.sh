#!/bin/bash

aws_profile=sin
s3_bucket=cc-mahjong-0


super_tgz=/tmp/super.tgz

echo "Syncing < ${super_tgz} > to S3::${s3_bucket}"

cd ../tiles-repo/super
tar --exclude='./node_modules' -zcvf ${super_tgz} ./*
aws s3 cp ${super_tgz} s3://${s3_bucket}/tiles-repo/super/super.tgz --profile ${aws_profile}

echo "Synced < ${super_tgz} > to S3::${s3_bucket}"