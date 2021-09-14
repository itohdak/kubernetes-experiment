# Prerequisite

## pull git repos
```
git clone https://github.com/itohdak/kubernetes-experiment.git # clone this repo
cd kubernetes-experiment

git clone -b dev https://github.com/itohdak/microservices-demo.git
git clone -b dev https://github.com/itohdak/locust-experiments.git
# git clone https://github.com/prometheus-operator/kube-prometheus.git
```


## set env
```
PROJECT_ID=[your-project-id]
ZONE=asia-northeast1
CLUSTER_NAME=[your-cluster-name]
```
### example
```
PROJECT_ID=charming-scarab-316315
ZONE=asia-northeast1
CLUSTER_NAME=onlineboutique
```


# Tips (GKE)

## before starting
```
gcloud init
gcloud services enable container.googleapis.com
```
### ref
- [gcloud install](https://cloud.google.com/sdk/docs/quickstart-linux?hl=ja)


## create GKE cluster
```
gcloud container clusters create ${CLUSTER_NAME} \
    --project=${PROJECT_ID} --zone=${ZONE} \
    --node-locations ${ZONE}-a,${ZONE}-b \
    --machine-type=e2-standard-4 --num-nodes=2
```

## scale in / out
```
gcloud container clusters resize $CLUSTER_NAME --size 0 --zone $ZONE
gcloud container clusters resize $CLUSTER_NAME --size 2 --zone $ZONE
```


## accessing gcloud from local using kubectl
```
gcloud container clusters get-credentials ${CLUSTER_NAME} --zone ${ZONE}
```
### ref
- [enable kubectl](https://qiita.com/oguogura/items/c4f73dbcf0c73e25ec9a)


## install istio
```
istioctl install --set profile=demo -y
kubectl label namespace default istio-injection=enabled
```
### if istioctl is not installed
```
curl -L https://istio.io/downloadIstio | sh -
cd istio-1.10.1
echo export PATH=$PWD/bin:'$PATH' >> ~/.bashrc
exec bash
```


## taint node for locust
```
LOCUST_NODE=`kubectl get nodes -ojsonpath='{.items[].metadata.name}'`
kubectl taint node $LOCUST_NODE run=locust:NoSchedule
kubectl label nodes $LOCUST_NODE app=locust
```


## deploy
```
kubectl apply -f ./microservices-demo/release/kubernetes-manifests.yaml
kubectl apply -f ./microservices-demo/release/istio-manifests.yaml
```


## get access path
```
INGRESS_HOST="$(kubectl -n istio-system get service istio-ingressgateway \
   -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"
echo "$INGRESS_HOST"
```
You can now access the sample application on `http://$INGRESS_HOST`.

## Prometheus
### without helm
```
git clone https://github.com/prometheus-operator/kube-prometheus.git
cd kube-prometheus
kubectl create -f manifests/setup
until kubectl get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done
kubectl create -f manifests/
```
#### access prometheus/grafana via port-forwarding
```
kubectl --namespace monitoring port-forward svc/prometheus-k8s 9090
kubectl --namespace monitoring port-forward svc/grafana 3000
```
|key|value|
----|----
|user|admin|
|password|admin|

### with helm
```
kubectl create ns monitoring
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add stable https://charts.helm.sh/stable
helm repo update
helm -n monitoring install prometheus-operator prometheus-community/kube-prometheus-stack
```
[install prometheus-stack via heml](https://qiita.com/MetricFire/items/1f15b6f1237ade0ce0d9)
[helm install error](https://stackoverflow.com/questions/64226913/install-prometheus-operator-doesnt-work-on-aws-ec2-keeps-produce-error-failed)

#### access grafana via port-forwarding
```
kubectl -n monitoring port-forward svc/prometheus-operator-kube-p-prometheus 9090:9090
kubectl -n monitoring port-forward svc/prometheus-operator-grafana 3000:80
```
user: admin, password: prom-operator


# locust
```
kubectl apply -f ./locust-experiments/kubernetes
# deploys
# - locust-cm.yaml
# - scripts-cm.yaml
# - master-deployment.yaml
# - service.yaml
# - slave-deployment.yaml

# curl "http://localhost:8089/swarm" -X POST -H "Content-Type: application/x-www-form-urlencoded" --data "locust_count=200&hatch_rate=1"
curl "http://localhost:8089/swarm" -X POST -H "Content-Type: application/x-www-form-urlencoded; charset=UTF-8" --data "user_count=200&spawn_rate=1"
```

### load test
```
kubectl port-forward svc/locust-master 8089:8089 
```

# Installation
## Flagger
[ref](https://docs.flagger.app/install/flagger-install-on-kubernetes)
```
$ helm repo add flagger https://flagger.app
"flagger" has been added to your repositories

$ kubectl apply -f https://raw.githubusercontent.com/fluxcd/flagger/main/artifacts/flagger/crd.yaml
customresourcedefinition.apiextensions.k8s.io/canaries.flagger.app created
customresourcedefinition.apiextensions.k8s.io/metrictemplates.flagger.app created
customresourcedefinition.apiextensions.k8s.io/alertproviders.flagger.app created

$ helm upgrade -i flagger flagger/flagger \
 --namespace=istio-system \
 --set crd.create=false \
 --set meshProvider=istio \
 --set metricsServer=http://prometheus:9090
Release "flagger" does not exist. Installing it now.
NAME: flagger
LAST DEPLOYED: Sat Jul 17 15:32:11 2021
NAMESPACE: istio-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
Flagger installed
```


### Metrics check in Flagger
[ref](https://github.com/fluxcd/flagger/blob/55de241f48b241143825ac4d52f9ddf5fa2ba797/pkg/controller/scheduler_metrics.go#L131-L229)


## Golang
[ref](https://qiita.com/notchi/items/5f76b2f77cff39eca4d8)
```
$ wget https://dl.google.com/go/go1.16.6.linux-amd64.tar.gz
--2021-07-17 15:51:45--  https://dl.google.com/go/go1.16.6.linux-amd64.tar.gz
Resolving dl.google.com (dl.google.com)... 2404:6800:4004:81d::200e, 172.217.175.46
Connecting to dl.google.com (dl.google.com)|2404:6800:4004:81d::200e|:443... connected.
HTTP request sent, awaiting response... 200 OK
Length: 129049323 (123M) [application/x-gzip]
Saving to: ‘go1.16.6.linux-amd64.tar.gz’

go1.16.6.linux-amd64.tar.gz               100%[===================================================================================>] 123.07M  70.8MB/s    in 1.7s    

2021-07-17 15:51:47 (70.8 MB/s) - ‘go1.16.6.linux-amd64.tar.gz’ saved [129049323/129049323]

$ sudo tar -C /usr/local -xzf go1.16.6.linux-amd64.tar.gz

$ echo export PATH=/usr/local/go/bin:'$PATH' >> ~/.bashrc
```


### get Prometheus exported metrics with golang
```
$ go get github.com/prometheus/client_golang/api
$ go get github.com/prometheus/client_golang/api/prometheus/v1
```


## Helm
```
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```



