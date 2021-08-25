kubectl -n istio-system port-forward svc/prometheus 9090:9090 > /dev/null 2>&1 &
kubectl -n istio-system port-forward svc/grafana 3000:3000 > /dev/null 2>&1 &
kubectl port-forward svc/locust-master 8089:8089 > /dev/null 2>&1 &
