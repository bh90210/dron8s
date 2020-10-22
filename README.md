# Dron8s

Yet another Kubernetes plugin for Drone using [dynamic](https://pkg.go.dev/k8s.io/client-go@v0.19.2/dynamic) [Server Side Apply](https://kubernetes.io/docs/reference/using-api/api-concepts/#server-side-apply) to achieve `kubectl apply -f` parity for your CI-CD pipelines.

## Features
* Create resources if they do not exist/update if they do
* Can handle multiple yaml configs in one file
* Can handle most resource types<sup>1</sup>
* In-cluster/Out-of-cluster use
* Easy set up, simple usage, well documented

_<sup>1</sup>Dron8s uses [client-go@v0.19.2](https://github.com/kubernetes/client-go/tree/v0.19.2). While most common Kubernetes API will work with your cluster's version, some features will not. For more information check the [compatibility matrix](https://github.com/kubernetes/client-go#compatibility-matrix)._

# [in-cluster](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration) use

In-cluster use is intented to only work along [Kubernetes Runner](https://docs.drone.io/runner/kubernetes/overview/) with in-cluster deployment scope. That is your pipelines can only `apply` resources within the cluster Kubernetes Runner is running.

## Prerequisites 
You need to manually create a `clusterrolebinding` resource [to allow cluster edit access](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) for Drone.

Assuming you installed Drone/Kubernetes Runner using [Drone provided Helm charts](https://github.com/drone/charts/tree/master/charts) run:
```bash
$ kubectl create clusterrolebinding dron8s --clusterrole=edit --serviceaccount=drone:default --namespace=drone
```
_If you opted for manual installation you have to replace the `--serviceaccount` and/or `--namespace` flag with the correct service/namespace name you used (ie. `--serviceaccount=drone-ci:default --namespace=default`)._


### In-cluster Pipe Example 
```yaml
kind: pipeline
type: kubernetes
name: dron8s-in-cluster-example

steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
```

## Uninstall

You need to manually delete the `clusterrolebinding` created as prerequisite. Run:

```bash
$ kubectl delete clusterrolebinding dron8s --namespace=drone
```

# [out-of-cluster](https://github.com/kubernetes/client-go/tree/master/examples/out-of-cluster-client-configuration) use

For out-of-cluster use you can choose whichever [runner](https://docs.drone.io/runner/overview/) you prefer but you need to provide you cluster's `kubeconfig` via a secret.

## Prerequisites 
Create a secret with the contents of kubeconfig.

_NOTE: You can always use Vault or AWS Secrets etc. But for this example I only show [Kubernetes Secrets](https://docs.drone.io/secret/external/kubernetes/) & [Encrypted](https://docs.drone.io/secret/encrypted/)._

## **1. Per Repository Secrets (GUI)**

Copy the contents of your `~/.kube/config` in Drone's Secret Value field and name the secret `kubeconfig`:

![Imgur](https://imgur.com/Cx9h3Xx.jpg)

### Per Repository Secrets - Docker Runner Pipe Example

```yaml
kind: pipeline
type: docker
name: dron8s-out-of-cluster-example

steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
    kubeconfig:
        from_secret: kubeconfig
```
## Uninstall

Delete the `secret` containing kubeconfig.

![Imgur](https://imgur.com/nyxIlxY.jpg)

## **2. Kubernetes Secrets (Kubectl)**

_In order to use this type of secret you have to install `Kubernetes Secrets` [Helm Chart](https://github.com/drone/charts/tree/master/charts/drone-kubernetes-secrets).
Furthermore the assumption is that you use `Kubernetes Runner` with out-of-cluster scope. 
That is a scenario where your CI/CD exists in cluster **a** and you apply configurations in cluster **b**. For in-cluster usage you do not need `Kubernetes Secrets` or secrets at all. See <a href="#in-cluster-use">in-cluster use</a>._

Before using Kubernetes Secrets in your pipeline you first need to manually create your secrets via `kubectl`. In our case we need to create a secret out of `~/.kube/config`. Run:

```bash
$ kubectl create secret generic dron8s --from-file=kubeconfig=$HOME/.kube/config --namespace=drone
```

### Kubernetes Secrets - Kubernetes Runner Pipe Example

```yaml
kind: pipeline
type: kubernetes
name: dron8s-out-of-cluster-example

steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
    kubeconfig:
        from_secret: kubeconfig
---
kind: secret
name: kubeconfig
get:
  path: dron8s
  name: kubeconfig
```

## Uninstall

Delete the `secret` containing kubeconfig. Run:

```bash
$ kubectl delete secret dron8s-kubeconfig --namespace=drone
```

## **3. Encrypted (Drone CLI)**

In order to use this method you need to have Drone CLI [installed](https://docs.drone.io/cli/install/) and [configured](https://docs.drone.io/cli/configure/) on your machine.

To generate the secret run:
```bash
$ drone encrypt user/repositry $(printf “%s” “$(<~/.kube/config)”)
```
where `user` is your real username and `repository` the name of the repository that you are creating the secret for.

Copy the output of your terminal to `data` field inside kubeconfig secret.

### Encrypted Secret - Docker Runner Pipe Example

```yaml
kind: pipeline
type: docker
name: dron8s-out-of-cluster-example

steps:
- name: dron8s
  image: bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
    kubeconfig:
        from_secret: kubeconfig
---
kind: secret
name: kubeconfig
data: ZGDJTGfiy5vzdvvZWRSEdIRlloamRmaW9saGJkc0vsVSDVs[...]
```


# Developing

You need to have [Go](https://golang.org/doc/install) and [Docker](https://docs.docker.com/get-docker/) installed on your system.

If you wish you may clone the repo and directly edit `.drone.yaml` as everything you need for the build is right there.

Otherwise:

```bash
$ git clone github.com/bh90210/dron8s
$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dron8s
$ docker build -t {yourusername}/dron8s .
$ docker push {yourusername}/dron8s
```
To use your own repo inside Drone pipelines just change the `image` field to `{yourusername}/dron8s`
```yaml
kind: pipeline
type: docker
name: default

steps:
- name: dron8s
  image: {yourusername}/dron8s
  settings:
    yaml: ./config
```
_Replace `{yourusername}` with your actual Docker Hub (or other registry) username._

_For more information see Drone's [Go Plugin Documentation](https://docs.drone.io/plugins/tutorials/golang/)._

# Contributing 

Any code improvements, updates, documentation spelling corrections etc are _always_ very welcome.

It is a very simple project so just clone the master branch, edit it and open a PR.