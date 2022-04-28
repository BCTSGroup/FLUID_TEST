#!/bin/sh

echo ["$(date)"] daemon start >>daemon.log
echo pwd: `pwd` >>daemon.log

# 每次检测程序运行状态的间隔时间,10s
INTERVAL=10

echo http port: $PORT >>daemon.log

while :
do
  sleep $INTERVAL
  str=`lsof -i:$PORT`
  if [ -z "$str" ]; then
    echo ["$(date)"] 检测到程序崩溃 >>daemon.log
    prefix=$(date +%Y%m%d.%H.%M)
    mv test.log $prefix.test.log
    mv err.log $prefix.err.log
    ./main_runnable -memprof=on -daemon -daemonPid=$$ >test.log 2>err.log &
  fi
done
