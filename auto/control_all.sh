#!/bin/bash
# Send control message to all nodes

work_root=$(pwd)
usage="sh ./control_all.sh {stop|start|restart|forcestop}"

case "$1" in
    start | stop | forcestop | restart)
        for node_name in $(ls "${work_root}" | grep node);
        do
          cd "${work_root}/${node_name}"
          sh control.sh "$1"
        done
        ;;
    *)
        echo "$usage"
        ;;
esac
