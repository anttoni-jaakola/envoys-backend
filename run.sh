#!/bin/bash

case "$1" in
start)
   nohup ./bin > /dev/null 2>&1&
   echo $!>/var/run/paymex.pid
   ;;
stop)
   kill `cat /var/run/paymex.pid`
   rm /var/run/paymex.pid
   ;;
restart)
   $0 stop
   $0 start
   ;;
status)
   if [ -e /var/run/paymex.pid ]; then
      echo run.sh is running, pid=`cat /var/run/paymex.pid`
   else
      echo run.sh is NOT running
      exit 1
   fi
   ;;
*)
   echo "Usage: $0 {start|stop|status|restart}"
esac

exit 0