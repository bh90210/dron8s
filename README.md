# Dron8s

Yet another Kubernetes plugin for Drone using [dynamic](https://pkg.go.dev/k8s.io/client-go@v0.19.2/dynamic) [Server Side Apply](https://kubernetes.io/docs/reference/using-api/api-concepts/#server-side-apply) to achieve `kubectl apply -f` parity for your CI-CD pipelines.

## Features
* Create resources if they do not exist/update if they do
* Can handle multiple yaml configs in one file
* Can handle most resource types<sup>1</sup>
* In-cluster/Out-of-cluster use
* Easy set up, simple usage

_<sup>1</sup>Dron8s uses [client-go@v0.19.2](https://github.com/kubernetes/client-go/tree/v0.19.2). While most common Kubernetes API will work with your cluster's version, some features will not. For more information check the [compatibility matrix](https://github.com/kubernetes/client-go#compatibility-matrix)._

# [in-cluster](https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration) use

In-cluster use is intented to only work along [Kubernetes Runner](https://docs.drone.io/runner/kubernetes/overview/) with in-cluster deployment scope. That is your pipelines can only create/patch resources within the cluster Kubernetes Runner is running.

## Prerequisites 
You need to manually create a `clusterrolebinding` to allow cluster resource manipulation from Drone server.

Assuming you installed Drone/Kubernetes Runner using [Drone provided Helm charts](https://github.com/drone/charts/tree/master/charts) run:
```bash
$ kubectl create clusterrolebinding dron8s --clusterrole=cluster-admin --serviceaccount=drone:default
```
_If you opted for manual installation you have to replace the `--serviceaccount` flag with the correct service name you used (ie. `--serviceaccount=drone-ci:default`)._


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
$ kubectl delete clusterrolebinding dron8s
```

# [out-of-cluster](https://github.com/kubernetes/client-go/tree/master/examples/out-of-cluster-client-configuration) use

For out-of-cluster use you can choose whichever [runner](https://docs.drone.io/runner/overview/) you prefer but you need to provide you cluster's `kubeconfig` via a secret.

## Prerequisites 
Create a secret with the contents of kubeconfig.

_NOTE: You can always use Vault or AWS Secrets etc. But for this example I only show [Per Repository](https://docs.drone.io/secret/repository/), [Kubernetes Secrets](https://docs.drone.io/secret/external/kubernetes/) & [Encrypted](https://docs.drone.io/secret/encrypted/)._

**1. Per Repository (GUI)**

Copy the contents of your `~/.kube/config` in Drone's Secret Value field:

![Imgur](https://imgur.com/Cx9h3Xx.jpg)

### Per Repository Secret Pipe Example

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


**2. Kubenrnetes Secrets (Kubectl)**

Before using this type of secret you first need to manually create your secrets via `kubectl`

```
$
```

### Kubernetes Secret Pipe Example

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
get:
  path: kubernetes
  name: kubeconfig
```


**3. Encrypted (Drone)**
### Encrypted Secret Pipe Example

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
data: ZXVDFHSfiy5vzdvvZWRSEdIRlloamRmaW9saGJkc0vsVSDVsvsd97vsdvkpgu8n9yecrHFRDSeiorncsafASEVTBkyNjM0OTUxOTA1NDQ1NTQ2
```


# Developing

You need to have [Go](https://golang.org/doc/install) and Docker installed on your system.

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

_For more information see Drone's [Plugin Documentation](https://docs.drone.io/plugins/tutorials/golang/)._

