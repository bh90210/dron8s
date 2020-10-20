# Dron8s

Yet another Kubernetes plugin for Drone using dynamic [Server Side Apply](https://kubernetes.io/docs/reference/using-api/api-concepts/#server-side-apply) to achieve `kubectl apply -f` parity.

## Features
* Creates a resources if it does not exist/updates if it does
* Can handle multiple yaml configs in one file
* Can handle most resource types<sup>1</sup>
* In-cluster/Out-of-cluster usage

_<sup>1</sup> This plugin use client-go@v0.19.2, while common Kubernetes API will work, some features will not. For more information check the [compatibility matrix](https://github.com/kubernetes/client-go#compatibility-matrix)._

# [in-cluster](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration) use

In-cluster use is intented to only work along [Kubernetes Runner](https://docs.drone.io/runner/kubernetes/overview/) with in-cluster deployment scope. That is your pipelines can only create/patch resources within the cluster Kubernetes Runner is running.

## Prerequisites 
You need to manually create a `clusterrolebinding` to allow cluster resource manipulation from Drone server.

Assuming you installed Drone/Kubernetes Runner using Drone provided Helm charts run:
```
kubectl create clusterrolebinding dron8s --clusterrole=cluster-admin --serviceaccount=drone:default
```
_If you opted for manual installation you have to replace the `--serviceaccount` flag with the correct service name you used (ie. `--serviceaccount=drone-ci:default`)._


## Example 
```
kind: pipeline
type: kubernetes
name: dron8s-in-cluster-example


steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
```

# [out-of-cluster](https://github.com/kubernetes/client-go/tree/master/examples/out-of-cluster-client-configuration) use

For out-of-cluster use you can choose whichever [runner](https://docs.drone.io/runner/overview/) you prefer but you need to provide you cluster's `kubeconfig` via a secret.

## Prerequisites 
Create a secret with the contents of kubeconfig

1. gui
2. kubenrnetes secrets
3. encrypted

## Example 
```
kind: pipeline
type: docker
name: dron8s-out-of-cluster-example


steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
    kubeconfig:
        from_secret:
            secret
```
