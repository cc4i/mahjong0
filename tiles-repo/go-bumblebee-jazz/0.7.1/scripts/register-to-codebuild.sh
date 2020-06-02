#!/bin/bash


roleArn=`cat role.arn`
projects[0]="air"
projects[1]="was"
projects[2]="gql"
projects[3]="front"
projects[4]="locust"

for project in ${projects[@]}
do 
    echo "Deleting ${project} from CodeBuild ..."
    aws codebuild delete-project --name ${project}-service-build || true
    echo "Done."
    sleep 5

    echo "Re-creating ${project} into CodeBuild ..."
    aws codebuild create-project \
        --cli-input-json file://${project}-codebuild.json \
        --service-role ${roleArn}   
    echo "Done." 
    sleep 5

    echo "Activing webhook on Github with all events ..."
    aws codebuild create-webhook \
        --project-name ${project}-service-build \
        --filter-groups '[[{"type": "EVENT", "pattern": "PUSH", "excludeMatchedPattern": false},{"type":"FILE_PATH","pattern": "src/'${project}'", "excludeMatchedPattern": false}]]'
    echo "Done." 
done




