
#!/bin/bash
set -x

# List all instance & lifecycle
aws ec2 describe-instances \
    --filters Name=instance-state-name,Values=running Name=tag:Name,Values=nodes-asg-mahjong-spot-cluster0  \
    --output json \
    | jq '.Reservations[].Instances[] | .InstanceId, .InstanceLifecycle, .PrivateDnsName ' | sed -e 's/"//g' > spot.out

exec 5< spot.out

# Tag instnace & label nodes
while read instanceId <&5 ; do
    read instanceLifecycle <&5
    read privateDnsName <&5
    echo "$instanceId Lifecycle = $instanceLifecycle"
    if [ $instanceLifecycle = "spot" ]
    then
        echo "tag $instanceId"
        aws ec2 create-tags --resources $instanceId --tags Key=lifecycle,Value=spot || true
        kubectl label nodes/$privateDnsName lifecycle=spot
    fi
done