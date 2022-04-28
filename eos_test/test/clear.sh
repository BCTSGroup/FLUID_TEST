#!/usr/bin/env bash

cd node/
dir=$(ls -l ./ |awk '/^d/ {print $NF}')
echo ${dir}
for i in ${dir}
do
    cd ${i}
    rm -rf db
    rm -rf shareRule
    rm -rf shareTable
    rm -rf tempTable
    rm err.log
    rm test.*
    pwd

    cd ..

done
