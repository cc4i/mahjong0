#!/bin/bash

for t in `ls ../tiles-repo`
do
    if [ $t != "super" ]
    then
        for v in `ls ../tiles-repo/$t`
        do
            echo "$t - $v"
            ./sync-tile.sh $t $v
        done
    else 
        ./sync-super.sh
    fi
done