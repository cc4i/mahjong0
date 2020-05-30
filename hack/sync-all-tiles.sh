#!/bin/bash

for t in `ls ../tiles-repo`
do
    if [ $t != "super" ] && [ $t != "cdk-tile" ] && [ $t != "app-tile" ]
    then
        for v in `ls ../tiles-repo/$t`
        do
            echo "$t - $v"
            ./sync-tile.sh $t $v
        done
    else 
        ./sync-super.sh
        ./sync-tile-base.sh
    fi
done