# Integrate with Kubernetes

A simple demo to deploy ServiceCenter Cluster in Kubernetes.
ServiceCenter supports two deploy modes: `Platform Registration` and `Client Side Registration`

## Requirements

1. There is a Kubernetes cluster.
1. Already install `kubectl` and `helm client` in your local machine.
1. (Optional) Already deploy helm tiller on Kubernetes.

## Platform Registration

The platform registration indicates that the ServiceCenter automatically accesses `kubernetes` cluster,
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

## Client Side Registration

The client-side registration representational ServiceCenter receives and 
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
    --set sc.discovery.clusters="http://coreos-etcd-client:2379" \
    --set sc.registry.enabled=true \
    --set sc.registry.type="etcd" \
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
    --set sc.discovery.clusters="http://coreos-etcd-client:2379" \
    --set sc.registry.enabled=true \
    --set sc.registry.type="etcd" \
    service-center/
```

## Confirm the deploy is ok

By default, the ServiceCenter frontend use `NodePort` service type to deploy in Kubernetes.

1. You can execute the command `kubectl get pod`, to check all pods are running.
1. You can also point your browser to `http://${NODE}:30103` to view the dashboard of ServiceCenter.
1. (Recommended) You can use [scctl](https://github.com/apache/servicecomb-service-center/tree/master/scctl) tool to list micro-service information.

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

## Helm Configuration Values

- **Service Center (sc)**
    + **deployment** (bool: true) Deploy this component or not.
    + **service**
        - **`type`** (string: "ClusterIP") The kubernetes service type.
        - **`externalPort`** (int16: 30100) The external access port. If the type is `ClusterIP`,
        it is set to the access port of the kubernetes service, and if the type
        is `NodePort`, it is set to the listening port of the node.
    + **discovery**
        - **`type`** (string: "aggregate") The Service Center discovery type.
        This can also be set to `etcd` or `servicecenter`. `aggregate` let Service Center merge the
        discovery sources and applications can discover microservices from these through using
        Service Center HTTP API. `etcd` let Service Center start with client registration mode, all the
        microservices information comes from application self registration. `servicecenter` let Service
        Center manage multiple Service Center clusters at the same time. It can be applied to multiple
        datacenters scenarios.
        - **`aggregate`** (string: "k8s,etcd") The discovery sources of aggregation, only enabled if `type`
        is set to `aggregate`. Different discovery sources are merged together by commas(,),
        indicating that the Service Center will aggregate service information through these
        sources. Now support these scenarios: `k8s,etcd`(for managing services from multiple platforms),
        `k8s,servicecenter`(for accessing distinct kubernetes clusters).
        - **`clusters`** (string: "sc-0=http://127.0.0.1:2380") The cluster address managed by Service Center.
        If `type` is set to `etcd`, its format is `http(s)://{etcd-1},http(s)://{etcd-2}`. If `type` is
        set to other value, its format is `{cluster name 1}=http(s)://{cluster-1-1},http(s)://{cluster-1-2},{cluster-2}=http(s)://{cluster-2-1}`
    + **registry**
        - **`enabled`** (bool: false) Register Service Center itself or not.
        - **`type`** (string: "embeded_etcd") The class of backend storage provider, this decide how
        Service Center store the microservices information. `embeded_etcd` let Service Center store data
        in local file system, it means distributed file system is need if you deploy high availability
        Service Center. `etcd` let Service Center store data in existing etcd cluster, then Service Center
        could be a stateless service. `builin` disabled the storage.
        - **`name`** (string: "sc-0") The Service Center cluster name, only enabled if `type` is set to
        `embeded_etcd` or `etcd`.
        - **`addr`** (string: "http://127.0.0.1:2380") The backend storage provider address. This value
        should be a part of `sc.discovery.clusters` value.

- **UI (frontend)**
    + **deployment** (bool: true) Deploy this component of not.
    + **service**
        - **`type`** (string: "NodePort") The kubernetes service type.
        - **`externalPort`** (int16: 30103) The external access port. If the type is `ClusterIP`,
        it is set to the access port of the kubernetes service, and if the type
        is `NodePort`, it is set to the listening port of the node.