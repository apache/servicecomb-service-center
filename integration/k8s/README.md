Integrate with Kubernetes
-------

A simple demo to deploy ServiceCenter Cluster in Kubernetes.
ServiceCenter supports two deploy modes: `Self Registration` and `Platform Registration`

## Requirements

1. There is a Kubernetes cluster.
1. Already install `kubectl` and `helm client` in your local machine.
1. (Optional) Already deploy helm tiller on Kubernetes.

## Self Registration

The self-registration representational ServiceCenter receives and 
processes registration requests from micro-service instances and 
stores instance information in `etcd`.

Notes: After deployment, it create ServiceCenter cluster and etcd cluster in the `default` namespace.

#### Use Kubectl

You can use the command `kubectl apply` to deploy ServiceCenter cluster.

```bash
cd ${PROJECT_ROOT}/integration/k8s
kubectl apply -f <(helm template --name servicecomb --namespace default service-center/)
```

#### Use Helm Install

You can also use the helm commands to deploy ServiceCenter cluster if
you already deploy helm tiller.

## Platform Registration

The platform-registration indicates that the ServiceCenter automatically accesses `kubernetes` cluster,
and micro-service instances can discover service and endpoints information through
the ServiceCenter.

Notes: After deployment, it only create ServiceCenter cluster in the `default` namespace.

#### Use Kubectl

You can use the command `kubectl apply` to deploy ServiceCenter cluster.

```bash
cd ${PROJECT_ROOT}/integration/k8s
kubectl apply -f <(helm template --name servicecomb --namespace default service-center-k8s/)
```

#### Use Helm Install

You can also use the helm commands to deploy ServiceCenter cluster if 
you already deploy helm tiller.


```bash
cd ${PROJECT_ROOT}/integration/k8s
helm install --name servicecomb --namespace default service-center-k8s/
```

## Confirm the deploy is ok

By default, the ServiceCenter frontend use `NodePort` service type to deploy in Kubernetes.

1. You can execute the command `kubectl get pod`, to check all pods are running.
1. You can also point your browser to `http://${NODE}:30103` to view the dashboard of ServiceCenter.
1. (Recommended) You can use [`scctl`](/scctl) tool to list micro-service information.

```bash
# ./scctl get svc --addr http://servicecomb-service-center-k8s:30100 -owide
  DOMAIN  |                  NAME                   |            APPID            | VERSIONS | ENV | FRAMEWORK  |        ENDPOINTS         | AGE  
+---------+-----------------------------------------+-----------------------------+----------+-----+------------+--------------------------+-----+
  default | servicecomb-service-center-k8s-frontend | service-center-k8s-frontend | 0.0.1    |     | Kubernetes | http://172.0.1.101:30103 | 2m   
  default | servicecomb-service-center-k8s          | service-center-k8s          | 0.0.1    |     | Kubernetes | http://172.0.1.102:30100 | 2m
```

## Clean up

If you use the kubectl to deploy, take deploy mode `self registration` as example.

```bash
cd ${PROJECT_ROOT}/integration/k8s
kubectl delete -f <(helm template --name servicecomb --namespace default service-center/)
```

If you use helm tiller to deploy, take deploy mode `self registration` as example.

```bash
cd ${PROJECT_ROOT}/k8s
helm delete --purge servicecomb
```