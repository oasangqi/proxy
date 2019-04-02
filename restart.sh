#!/bin/sh

cat proxy.pid  |xargs -n 1 kill -sigterm 1>/dev/null 2>&1
LOG_DFT="logs"
logdir=`grep "logdir" proxy.toml |cut -d '"' -f 2`
if test -z $logdir;then
	logdir=$LOG_DFT
	echo "logdir is empty, use $logdir"
	mkdir $logdir 1>/dev/null 2>&1
fi
nohup ./proxy --log_dir $logdir &
