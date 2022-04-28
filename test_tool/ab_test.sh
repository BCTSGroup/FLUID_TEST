#!/usr/bin/env bash

while true

do

ab -n 200 -c 1 -p ab_body.txt -T 'application/json' 'http://127.0.0.1:7999/RequestData'

sleep 1

done

#ab -n 10000 -c 1 -p ab_body.txt -T 'application/json' 'http://127.0.0.1:7999/TestTps'

# DAG AB 测试命令
#ab -n 1000 -c 1 -p ab_body.txt -T 'application/json' 'http://127.0.0.1:7999/RequestData'
#ab -n 100 -c 1 -p ab_body.txt -T 'application/json' 'http://127.0.0.1:8002/EpochTag'
# SingleChain AB 测试命令
#ab -n 1000 -c 1  -p ab_singlechain.txt -T 'application/json' 'http://127.0.0.1:7999/RequestData'
#for ((i=1; i<=1000; i++))
#do
#ab -n 10 -c 1 -p ab_body.txt -T 'application/json' 'http://127.0.0.1:8002/EpochTag'
#sleep 0.001
#done