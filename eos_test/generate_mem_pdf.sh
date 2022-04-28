#!/usr/bin/env bash
cd ./test/node/

dir=$(ls -l ./ |awk '/^d/ {print $NF}')
echo ${dir}
for i in ${dir}
do
	cd ${i}
	prof=$(ls ./ | grep mem)
	echo ${prof}
	for q in ${prof}
	do
		go tool pprof -pdf ./main_runnable ${q} >  ${q}.pdf
	done
	cd ..
done
