# User Guide
This instructions will lead you to getting start with using Servicecomb-service-center-istio

## 1. Install dependencies
This tool can be used both inside a k8s cluster and a standalone service running on a VM.

For both ways you have to install dependencies first.

### 1.1 Install Kubernetes Cluster
You can follow K8S [installation instruction](https://kubernetes.io/docs/setup/) to install a K8S cluster

### 1.2 Install Istio
Follow this [instruction](https://istio.io/latest/docs/setup/getting-started/) to install istio

**note: the instruction is just a show case of how to install and use istio, if you want to use it in production, you have to use a production ready installation profile**

### 1.3 Install Istio DNS
As any Servicecomb service center service will be translated to Serviceentry in K8S, while Kubernetes provides DNS resolution for Kubernetes Services out of the box, any custom ServiceEntrys will not be recognized. In addition to capturing application traffic, Istio can also capture DNS requests to improve the performance and usability of your mesh

Use the following command to install istio DNS:
```
cat <<EOF | istioctl install -y -f -
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
spec:
  meshConfig:
    defaultConfig:
      proxyMetadata:
        # Enable basic DNS proxying
        ISTIO_META_DNS_CAPTURE: "true"
        # Enable automatic address allocation, optional
        ISTIO_META_DNS_AUTO_ALLOCATE: "true"
EOF
```

### 1.4 Install Servicecomb service center
Servicecomb service center could be installed in K8S or on VM. 
Install Servicecomb service center follow this [instruction](https://github.com/apache/servicecomb-service-center/blob/master/README.md)

## 2 Install Servicecomb-service-center-istio
### 2.1 Building
You don’t need to build from source to use Servicecomb-service-center-istio (binaries in apache nexus ), but if you want to try out the latest and greatest, Servicecomb-service-center-istio can be easily built.
```
go build -o servicecomb-service-center-istio cmd/main.go
```

### 2.2 Building docker image
```
docker build -t servicecomb-service-center-istio:dev .
```

### 2.2 Run on VM
```
./Servicecomb-service-center-istio --sc-addr=?SERVICE_CENTER_ADDRESS --kube-config=?KUBE_CONFIG_FILE_PATH
```

### 2.3 Run in K8S
```
# make sure you modified the input args in the deployment.yaml file first, specify you service center address
kubectl apply -f manifest/deployment.yaml
```

### 2.4 Input parameters
![image](imgs/cmd.png)