#!/bin/bash

echo "Creating symmetric key ..."
aws kms create-key \
    --tags TagKey=Purpose,TagValue=Tiles TagKey=Builder,TagValue=Mahjong \
    --description "Master symmetric key for Tiles" > kms.out
SYMMETRIC_KEY_ID=`cat kms.out | jq -r '.KeyMetadata.KeyId'`
SYMMETRIC_KEY_ARN=`cat kms.out | jq -r '.KeyMetadata.Arn'`
echo "Done"
if [ $SYMMETRIC_KEY_ALIAS = "" ]
then
    echo "Attach alias to symmetric key ..."
    aws kms create-alias \
        --alias-name $SYMMETRIC_KEY_ALIAS \
        --target-key-id $SYMMETRIC_KEY_ID
    echo "Done"
fi
echo "Creating asymmetric key ..."
aws kms create-key \
    --tags TagKey=Purpose,TagValue=Tiles TagKey=Builder,TagValue=Mahjong \
    --key-usage ENCRYPT_DECRYPT \
    --customer-master-key-spec RSA_4096 \
    --description "Master asymmetric key for Tiles" > akms.out

ASYMMETRIC_KEY_ID=`cat akms.out | jq -r '.KeyMetadata.KeyId'`
ASYMMETRIC_KEY_ARN=`cat akms.out | jq -r '.KeyMetadata.Arn'`
echo "Done"
if [ $ASYMMETRIC_KEY_ALIAS = "" ]
then
    echo "Attach alias to asymmetric key ..."
    aws kms create-alias \
        --alias-name $ASYMMETRIC_KEY_ALIAS \
        --target-key-id $ASYMMETRIC_KEY_ID
    echo "Done"
fi

echo symmetricKeyID=$SYMMETRIC_KEY_ID
echo symmetricKeyArn=$SYMMETRIC_KEY_ARN
echo symmetricKeyAlias=$SYMMETRIC_KEY_ALIAS
echo asymmetricKeyID=$ASYMMETRIC_KEY_ID
echo asymmetricKeyArn=$ASYMMETRIC_KEY_ARN
echo asymmetricKeyAlias=$ASYMMETRIC_KEY_ALIAS