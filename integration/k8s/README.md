# Integrate with Kubernetes

A simple demo to deploy ServiceCenter Cluster in Kubernetes.

## Quick Start

### Requirements

1. There is a Kubernetes cluster.
1. Already install `kubectl` and `helm client` in your local machine.
1. (Optional) Already deploy helm tiller on Kubernetes.

### 1. Use Kubectl

You can use the command `kubectl apply` to deploy ServiceCenter cluster.

```bash
cd ${PROJECT_ROOT}/integration/k8s
kubectl apply -f <(helm template --name servicecomb --namespace default service-center/)
```

### 2. Use Helm Install

You can also use the helm commands to deploy ServiceCenter cluster if you already deploy helm tiller.

```bash
cd ${PROJECT_ROOT}/integration/k8s
helm install --name servicecomb --namespace default service-center/
```

### Confirm the deploy is ok

By default, the ServiceCenter frontend use `NodePort` service type to deploy in Kubernetes.

1. You can execute the command `kubectl get pod`, to check all pods are running.
1. You can also point your browser to `http://${NODE}:30103` to view the dashboard of ServiceCenter.

## Clean up

If you use the kubectl to deploy

```bash
cd ${PROJECT_ROOT}/integration/k8s
kubectl delete -f <(helm template --name servicecomb --namespace default service-center/)
```

If you use helm tiller to deploy

```bash
cd ${PROJECT_ROOT}/k8s
helm delete --purge servicecomb
```