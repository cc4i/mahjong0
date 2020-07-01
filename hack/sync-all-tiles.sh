#!/bin/bash

local_tile_repo=../repo/tile

for t in `ls ${local_tile_repo}`
do
    if [ $t != "super" ] && [ $t != "cdk-tile" ] && [ $t != "app-tile" ]
    then
        for v in `ls ${local_tile_repo}/$t`
        do
            echo "$t - $v"
            ./sync-tile.sh $t $v
        done
    else 
        ./sync-super.sh
        ./sync-tile-base.sh
    fi
done