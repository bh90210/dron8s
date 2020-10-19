# Dron8s

Yet another Kubernetes plugin for dynamic [Server Side Apply](https://kubernetes.io/docs/reference/using-api/api-concepts/#server-side-apply). 

explain what dynamic SSA usage means from the perspective of dron8s
https://ymmt2005.hatenablog.com/entry/2020/04/14/An_example_of_using_dynamic_client_of_k8s.io/client-go#Go-client-libraries

NOTE: client-go @ HEAD, resources depent on cluster version/client-go version

# [in-cluster](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration) use

In-cluster use is intented to only work along [`Kubernetes Runner`](https://docs.drone.io/runner/kubernetes/overview/) with in-cluster deployment scope. That is your pipelines can only create/patch resources within the cluster `Kubernetes Runner` is running.

## Prerequisites 
```
kubectl create clusterrolebinding drone-runner-cluster-admin --clusterrole=cluster-admin --serviceaccount=drone:default
```

## Example 
```
kind: pipeline
type: kubernetes
name: dron8s-example


steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
```

# [out-of-cluster](https://github.com/kubernetes/client-go/tree/master/examples/out-of-cluster-client-configuration) use

```
kind: pipeline
type: kubernetes
name: dron8s-example


steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
    kubeconfig:
        from_secret:
            secret
```