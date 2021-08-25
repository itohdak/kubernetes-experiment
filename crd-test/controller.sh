#/bin/bash

# 制御ループ
while true
do
  # 現在の状態取得
  name=$(kubectl get fcrd -o jsonpath='{.items[0].metadata.name}')
  message=$(printf "%q" $(kubectl get fcrd ${name} -o jsonpath='{.spec.message}'))
  podnum=$(echo $(kubectl get pod ${name} --no-headers --ignore-not-found | wc -l))

  # firstcrdリソースが存在する場合
  if [[ -n ${name} ]] ; then
    # podが存在しない場合
    if [[ ${podnum} -eq 0 ]] ; then
      # myfirstcrリソースが存在するのにPodが存在しないので是正する
      echo "creating ${name} pod"
      cat <<YAML | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: ${name}
spec:
  initContainers:
  - name: init
    image: busybox
    args:
    - /bin/sh
    - -c
    - echo ${message} > /test/index.html
    volumeMounts:
    - mountPath: /test
      name: shared-data
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - mountPath: /usr/share/nginx/html/
      name: shared-data
  volumes:
    - name: shared-data
      emptyDir: {}
YAML
    # podが存在する場合
    else
      # 現在のPodが表示するメッセージを取得する
      podmessage=$(printf "%q" $(kubectl exec ${name} -- curl -s http://localhost))
      if [[ "${podmessage}" != "${message}" ]] ; then
      # メッセージがmyfirstcrリソースの定義と異なるのでPodを削除する
        echo "deleting ${name} pod"
        kubectl delete pod ${name}
      fi
    fi
  fi
  # firstcrdリソースが存在しない場合
  if [[ -z ${name}  ]] ; then
    # firstcrdリソースが存在しないのにPodが存在するので是正する
    if [[ ${podnum} -ne 0 ]] ; then
      kubectl delete pod --all
    fi
  fi
  sleep 15
done
