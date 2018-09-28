Integrate with Kubernetes
-------

A simple demo to deploy ServiceCenter Cluster in Kubernetes.
ServiceCenter supports two deploy modes: `Platform Registration` and `Self Registration`

## Requirements

1. There is a Kubernetes cluster.
1. Already install `kubectl` and `helm client` in your local machine.
1. (Optional) Already deploy helm tiller on Kubernetes.

## Platform Registration

The platform-registration indicates that the ServiceCenter automatically accesses `kubernetes` cluster,
and micro-service instances can discover service and endpoints information through
the ServiceCenter.

Notes: After deployment, it only create ServiceCenter cluster in the `default` namespace.

#### Use Kubectl

You can use the command `kubectl apply` to deploy ServiceCenter cluster.

```bash
cd ${PROJECT_ROOT}/examples/infrastructures/k8s
kubectl apply -f <(helm template --name servicecomb --namespace default service-center/)
```

#### Use Helm Install

You can also use the helm commands to deploy ServiceCenter cluster if 
you already deploy helm tiller.

```bash
cd ${PROJECT_ROOT}/examples/infrastructures/k8s
helm install --name servicecomb --namespace default service-center/
```

## Self Registration

The self-registration representational ServiceCenter receives and 
processes registration requests from micro-service instances and 
stores instance information in `etcd`.

Notes: After deployment, it create ServiceCenter cluster and etcd cluster in the `default` namespace.

#### Use Kubectl

You can use the command `kubectl apply` to deploy ServiceCenter cluster.

```bash
cd ${PROJECT_ROOT}/examples/infrastructures/k8s
# install etcd cluster
kubectl apply -f <(helm template --name coreos --namespace default etcd/)
# install sc cluster
kubectl apply -f <(helm template --name servicecomb --namespace default \
 --set sc.discovery.type="etcd" \
 --set sc.discovery.uris="http://coreos-etcd-client:2379" \
 service-center/)
```

#### Use Helm Install

You can also use the helm commands to deploy ServiceCenter cluster if
you already deploy helm tiller.

```bash
cd ${PROJECT_ROOT}/examples/infrastructures/k8s
# install etcd cluster
helm install --name coreos --namespace default etcd/
# install sc cluster
helm install --name servicecomb --namespace default \
 --set sc.discovery.type="etcd" \
 --set sc.discovery.uris="http://coreos-etcd-client:2379" \
 service-center/
```

## Confirm the deploy is ok

By default, the ServiceCenter frontend use `NodePort` service type to deploy in Kubernetes.

1. You can execute the command `kubectl get pod`, to check all pods are running.
1. You can also point your browser to `http://${NODE}:30103` to view the dashboard of ServiceCenter.
1. (Recommended) You can use [`scctl`](/scctl) tool to list micro-service information.

```bash
# ./scctl get svc --addr http://servicecomb-service-center:30100 -owide
  DOMAIN  |                  NAME               |            APPID        | VERSIONS | ENV | FRAMEWORK  |        ENDPOINTS         | AGE  
+---------+-------------------------------------+-------------------------+----------+-----+------------+--------------------------+-----+
  default | servicecomb-service-center-frontend | service-center-frontend | 0.0.1    |     | Kubernetes | http://172.0.1.101:30103 | 2m   
  default | servicecomb-service-center          | service-center          | 0.0.1    |     | Kubernetes | http://172.0.1.102:30100 | 2m
```

## Clean up

If you use the kubectl to deploy, take deploy mode `platform registration` as example.

```bash
cd ${PROJECT_ROOT}/examples/infrastructures/k8s
kubectl delete -f <(helm template --name servicecomb --namespace default service-center/)
```

If you use helm tiller to deploy, take deploy mode `platform registration` as example.

```bash
cd ${PROJECT_ROOT}/k8s
helm delete --purge servicecomb
```