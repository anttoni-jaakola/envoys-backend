#!/bin/bash

case "$1" in
start)
   nohup ./bin > /dev/null 2>&1&
   echo $!>/var/run/envoys.pid
   ;;
stop)
   kill `cat /var/run/envoys.pid`
   rm /var/run/envoys.pid
   ;;
restart)
   $0 stop
   $0 start
   ;;
status)
   if [ -e /var/run/envoys.pid ]; then
      echo run.sh is running, pid=`cat /var/run/envoys.pid`
   else
      echo run.sh is NOT running
      exit 1
   fi
   ;;
*)
   echo "Usage: $0 {start|stop|status|restart}"
esac

exit 0