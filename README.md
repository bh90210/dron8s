<p align="center">
  <img width="26%" src="https://user-images.githubusercontent.com/22690219/119139719-02ab6580-ba4c-11eb-9dd4-fa810133c9f4.png" />
</p>

[![Build and publish docker image](https://github.com/bh90210/dron8s/actions/workflows/publish-image.yaml/badge.svg)](https://github.com/bh90210/dron8s/actions/workflows/publish-image.yaml) <img src=https://goreportcard.com/badge/github.com/bh90210/dron8s /> 

# Dron8s

Yet another Kubernetes plugin for Drone using [dynamic](https://pkg.go.dev/k8s.io/client-go@v0.19.2/dynamic) [Server Side Apply](https://kubernetes.io/docs/reference/using-api/api-concepts/#server-side-apply) to achieve `kubectl apply --server-side` parity for your CI-CD pipelines.

## Features
* Create resources if they do not exist/update if they do
* Can handle multiple yaml configs in one file
* Can handle most resource types<sup>1</sup>
* In-cluster/Out-of-cluster use
* Easy set up, simple usage, well documented
* Support variables

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
  image: ghcr.io/bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
```

### In-cluster Pipe Example With Variables

_for a full example see the [examples](https://github.com/bh90210/dron8s/tree/main/examples) folder_

```yaml
kind: pipeline
type: kubernetes
name: dron8s-in-cluster-example

steps:
- name: dron8s
  image: ghcr.io/bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
    # variables. Must be lowercase, Usage: {{.service_name}}
    service_name: myservice
```
And in your config:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{.service_name}}
spec:
...
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

_NOTE: You can always use Vault or AWS Secrets etc. But for this example I only show [Per Repository](https://docs.drone.io/secret/repository/),  [Kubernetes Secrets](https://docs.drone.io/secret/external/kubernetes/) & [Encrypted](https://docs.drone.io/secret/encrypted/)._

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
  image: ghcr.io/bh90210/dron8s:latest
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

Before using Kubernetes Secrets in your pipeline you first need to manually create your secrets via `kubectl`. In this case you need to create a secret out of `~/.kube/config`. Run:

```bash
$ kubectl create secret generic dron8s --from-file=kubeconfig=$HOME/.kube/config
```
_note that if you opted for different namespace than the default when installed `drone-kubernetes-secret` chart (`secretNamespace` & `KUBERNETES_NAMESPACE`) you need to also pass the appropriate `--namespace` flag to the above command_
### Kubernetes Secrets - Kubernetes Runner Pipe Example

```yaml
kind: pipeline
type: kubernetes
name: dron8s-out-of-cluster-example

steps:
- name: dron8s
  image: ghcr.io/bh90210/dron8s:latest
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
$ kubectl delete secret dron8s
```

## **3. Encrypted (Drone CLI)**

In order to use this method you need to have Drone CLI [installed](https://docs.drone.io/cli/install/) and [configured](https://docs.drone.io/cli/configure/) on your machine.

To generate the secret run:
```bash
$ drone encrypt user/repository @$HOME/.kube/config
```
where `user` is your real username and `repository` the name of the repository that you are creating the secret for.

Copy the output of your terminal to `data` field inside kubeconfig secret.

### Encrypted Secret - Exec Runner Pipe Example

```yaml
kind: pipeline
type: exec
name: dron8s-out-of-cluster-example

platform:
  os: linux
  arch: amd64

steps:
- name: dron8s
  image: ghcr.io/bh90210/dron8s:latest
  settings:
    yaml: ./config.yaml
    kubeconfig:
        from_secret: kubeconfig
---
kind: secret
name: kubeconfig
data: ZGDJTGfiy5vzdvvZWRSEdIRlloamRmaW9saGJkc0vsVSDVs[...]
```

# Field Manager

[When transferring ownership for `server-side-apply`](https://kubernetes.io/docs/reference/using-api/server-side-apply/#transferring-ownership) you will need to know the field manager of Dron8s ([as described on the relevant issue](https://github.com/bh90210/dron8s/issues/24)).

The field manager is `dron8s-plugin` and can be found in the [source code](https://github.com/bh90210/dron8s/blob/03fc616ea0bc8a612ee3ae1b95f9c4c2a385ffa2/main.go#L155) too.

# Known issues (and workarounds)

* If your resource contains `ports:` without specifically declaring `protocol: TCP`/`protocol: UDP` [you will probably get](https://github.com/bh90210/dron8s/issues/5) a similar error:
```log
failed to create typed patch object: .spec.template.spec.containers[name=].ports: element 0: associative list with keys has an element that omits key field "protocol"
```
The workaround is to simply define a protocol like so where applicable: 
```yaml
        ports:
          - protocol: TCP
            containerPort: 80
```
If it is not possible to alter the resource then maybe consider upgrading to Kubernetes v.0.20.0 where this bug is [hopefully resolved](https://github.com/kubernetes-sigs/structured-merge-diff/issues/130#issuecomment-706488157).

# Developing

You need to have [Go](https://golang.org/doc/install) and [Docker](https://docs.docker.com/get-docker/) installed on your system.

If you wish you may clone the repo and directly edit `.drone.yaml` as everything you need for the build is right there.

Otherwise:

```bash
$ git clone github.com/bh90210/dron8s
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
    yaml: ./config.yaml
```
_Replace `{yourusername}` with your actual Docker Hub (or other registry) username._

_For more information see Drone's [Go Plugin Documentation](https://docs.drone.io/plugins/tutorials/golang/)._

# Contributing 

Any code improvements, updates, documentation spelling corrections etc are _always_ very welcome.

It is a very simple project so just clone the master branch, edit it and open a PR.
