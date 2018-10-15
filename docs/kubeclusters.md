Access Distinct Clusters
-------

## ServiceCenter Aggregate Architecture

In the [`Multiple Datacenters`](/docs/multidcs.md) article, we introduce
the aggregation architecture of service center. In fact, this aggregation
architecture of service center can be applied not only to the scene deployed 
in multiple datacenters, but also to the scene services data aggregation in 
multiple kubernetes clusters.

![architecture](/docs/kubeclusters.PNG)

The service centers deployed in distinct kubernetes clusters can communicate 
with each other, sync the services data from other kubernetes clusters.
Applications can discover services from different the kubernetes cluster through
using the service center HTTP API.
**It solve the problem of isolation between kubernetes clusters**.

## Quick Start

Let's assume you want to install 2 clusters of Service-Center in different Kubernetes clusters with following details.

| Cluster | Kubernetes | namespace  | Node        |  
| :-----: | :--------: | :--------: | :---------: |  
| sc1     | k1         | default    | 10.12.0.1   |   
| sc2     | k2         | default    | 10.12.0.2   | 

To facilitate deployment, we will publish the service address of the service center in [`NodePort`] mode.

##### Deploy the Service Center

Using helm to deploy the service center to kubernetes here, the instructions for specific `values` can be referred to
[`here`](/examples/infrastructures/k8s/README.md#helm-configuration-values).

Take deployment to kubernetes cluster 1 as an example.
```bash
# login the k1 kubernetes master node to deploy sc1
git clone git@github.com:apache/incubator-servicecomb-service-center.git
cd examples/infrastructures/k8s
helm install --name k1 \
    --set sc.discovery.clusters="sc2=http://10.12.0.2:30100" \
    --set sc.discovery.aggregate="k8s\,servicecenter" \
    --set sc.registry.type="buildin" \
    --set sc.service.type=NodePort \
    service-center/
```
Notes: To deploy Service Center in kuberbetes cluster 2, you can repeat the
above steps and just change the `sc.discovery.clusters` value to 
`sc1=http://10.12.0.1:30100`.

##### Confirm the service is OK

We recommend that you use [`scctl`](/scctl/README.md), and using
[`cluster command`](/scctl/pkg/plugin/README.md#cluster-options)
which makes it very convenient to verify OK.

```bash
# check the sc1 cluster api
scctl --addr http://10.12.0.3:30100 get cluster
#   CLUSTER |        ENDPOINTS         
# +---------+-------------------------+
#   sc2     | http://10.12.0.2:30100

# check the sc2 cluster api
scctl --addr http://10.12.0.3:30100 get cluster
#   CLUSTER |        ENDPOINTS         
# +---------+-------------------------+
#   sc1     | http://10.12.0.1:30100
```



