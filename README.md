# Dron8s

Yet another Kubernetes plugin for [in-cluster](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration) [Server Side Apply](https://kubernetes.io/docs/reference/using-api/api-concepts/#server-side-apply). The plugin is intented to only work with [Kubernetes Runner](https://docs.drone.io/runner/kubernetes/overview/) with in-cluster deployment scope. That is your pipelines can only create/patch resources only in the same cluster that Kubernetes Runner is running.

NOTE: client-go @ HEAD, resources depent on cluster version/client-go version
