#!/usr/bin/env bash

pkill main_runnable
cd ..
./build_linux.sh
cd test

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
    ./main_runnable &
    sleep 0.5s
    cd ..

done
