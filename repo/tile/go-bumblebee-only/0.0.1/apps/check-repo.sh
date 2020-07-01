#!/bin/bash

retcode=`curl -H "Authorization: token $GIT_ACCESS_TOKEN" -s -o /dev/null -w "%{http_code}" https://api.github.com/repos/$GITHUB_USER/$APP_CONF_REPO`
if [ $retcode -ne 200 ]
then
    echo 'failed to create repo" >&2
    exit 1
fi
echo '$APP_CONF_REPO is ready to go!'