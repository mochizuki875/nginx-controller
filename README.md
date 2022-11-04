# nginx-controller

[mochizuki875/nginx-controller-basic](https://github.com/mochizuki875/nginx-controller-basic)をベースに機能拡張を行なったController。

```
$ k get ng
NAME      REPLICAS   SERVICE_NAME      CLUSTER-IP       EXTERNAL-IP
nginx-1   3          service-nginx-1   10.102.147.239   192.168.2.162

$ k get all -l controller=nginx-1
NAME                                  READY   STATUS    RESTARTS   AGE
pod/deploy-nginx-1-6fb898c576-ml5xm   1/1     Running   0          11s
pod/deploy-nginx-1-6fb898c576-rr6fq   1/1     Running   0          11s
pod/deploy-nginx-1-6fb898c576-xznvq   1/1     Running   0          11s

NAME                      TYPE           CLUSTER-IP       EXTERNAL-IP     PORT(S)        AGE
service/service-nginx-1   LoadBalancer   10.102.147.239   192.168.2.162   80:31584/TCP   11s

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/deploy-nginx-1   3/3     3            3           11s

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/deploy-nginx-1-6fb898c576   3         3         3       11s
```

## Environment

```
$ kubebuilder version
Version: main.version{KubeBuilderVersion:"3.7.0", KubernetesVendor:"1.24.1", GitCommit:"3bfc84ec8767fa760d1771ce7a0cb05a9a8f6286", BuildDate:"2022-09-20T17:21:57Z", GoOs:"darwin", GoArch:"amd64"}
$ go version
go version go1.19 darwin/amd64
$ kind version
kind v0.15.0 go1.19 darwin/amd64
```