#!/bin/bash

conf="/root/config/trconfig_replay.json"
hw="e4:ce:8f:01:4c:54"
trace="/out/clean_dump.pcap"

while getopts ":c:w:t:r:" opt; do
  case $opt in
    c) conf="$OPTARG"
    ;;
    w) hw="$OPTARG"
    ;;
    t) trace="$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done

/usr/bin/tr -conf $conf -hw $hw &

pid=$!

#Do tcp replay
echo "Starting replay for trace $trace"
tcpreplay -i lo $trace

#kill netm
echo "Killing process $pid"
kill -2 $pid

folder=/tmp/ta.*.out
for f in $folder; do
    echo "Copying file $f to /out/"
    mv $f /out/
done
chmod 666 /out/*.out