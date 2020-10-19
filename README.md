# Dron8s

Yet another Drone plugin for [in-cluster](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration) Kubernetes use. This plugin is intented to only work with [Kubernetes Runner](https://docs.drone.io/runner/kubernetes/overview/) with in-cluster deployment scope. That is your pipelines can only deploy in the same cluster that Kubernetes Runner is deployed.


```
kubectl create clusterrolebinding drone-runner-cluster-admin --clusterrole=cluster-admin --serviceaccount=drone:default
```