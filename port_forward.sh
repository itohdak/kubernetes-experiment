LOG=./log/port_forward.log
ERROR_LOG=./log/port_forward_error.log
kubectl -n istio-system port-forward svc/prometheus 9090:9090 > $LOG 2> $ERROR_LOG &
kubectl -n istio-system port-forward svc/grafana 3000:3000 > $LOG 2> $ERROR_LOG &
kubectl port-forward svc/locust-master 8089:8089 > $LOG 2> $ERROR_LOG &
kubectl port-forward svc/locust-master-loadtestmanager-sample 8089:8089 > $LOG 2> $ERROR_LOG &
