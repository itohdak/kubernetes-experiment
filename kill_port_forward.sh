ps aux | grep port-forward | grep -v grep | awk '{ print "kill -9", $2 }' | sh
