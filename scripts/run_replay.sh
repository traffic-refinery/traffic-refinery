#!/bin/bash

conf="/root/config/trconfig_replay.json"
hw="a0:ce:c8:0d:2b:a7"
trace="/out/clean_dump.pcap"
dlevel=""

while getopts ":c:w:t:r:" opt; do
  case $opt in
    c) conf="$OPTARG"
    ;;
    w) hw="$OPTARG"
    ;;
    t) trace="$OPTARG"
    ;;
    d) dlevel="-$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done

/usr/bin/tr -conf $conf -hw $hw $dlevel &

pid=$!

#Do tcp replay
echo "Starting replay for trace $trace"
tcpreplay -i lo $trace
echo "Concluded replay"

#kill netm
echo "Killing process $pid"
kill -2 $pid

folder=/tmp/tr.*.out
ls /tmp/
for f in $folder; do
    echo "Copying file $f to /out/"
    mv $f /out/
done
chmod 666 /out/*.out