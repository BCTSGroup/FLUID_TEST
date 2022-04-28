#!/usr/bin/env bash
go build -o main_runnable main.go


cd test/node/
dir=$(ls -l ./ |awk '/^d/ {print $NF}')
for i in ${dir}
do
    cd ${i}
    #现在处于test/node/test*
    pwd
    cp ../../../main_runnable .
    cd ..
done
#现在处于test/node

cd ../../
rm main_runnable
