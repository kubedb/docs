---
title: Run HanaDB using Private Registry
menu:
  docs_{{ .version }}:
    identifier: hanadb-using-private-registry
    name: Quickstart
    parent: hanadb-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Private Docker Registry

KubeDB supports using private Docker registries. This tutorial will show you how to run KubeDB managed HanaDB database using private Docker images.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Private Docker Registry

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).

- Push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For HanaDB, push the `DB_IMAGE` of the following HanaDBVersions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get hanadbversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,DEPRECATED:.spec.deprecated
  NAME    VERSION   DB_IMAGE                   DEPRECATED
  2.0     2.0       kubedb/hanadb:2.0          <none>
  ```

## Create ImagePullSecret

Run the following command to create an image pull secret for your private Docker registry:

```bash
$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

## Install KubeDB Operator

Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that it points to the DOCKER_REGISTRY you wish to pull images from.

## Create HanaDBVersion CRD

Create a HanaDBVersion CRD specifying images from your private registry. Replace `PRIVATE_REGISTRY` with your private registry.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: HanaDBVersion
metadata:
  name: "2.0"
spec:
  coordinator:
    image: PRIVATE_REGISTRY/hanadb-coordinator:v0.4.0
  db:
    image: PRIVATE_REGISTRY/hanadb:2.0
  exporter:
    image: PRIVATE_REGISTRY/hanadb-exporter:1.0.0
  version: "2.0.82"
```

```bash
$ kubectl apply -f pvt-hanadbversion.yaml
hanadbversion.catalog.kubedb.com/2.0 created
```

## Deploy HanaDB from Private Registry

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: pvt-reg-hanadb
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/private-registry/pvt-reg-hanadb.yaml
hanadb.kubedb.com/pvt-reg-hanadb created
```

Check that the HanaDB is in Running state:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=pvt-reg-hanadb"
NAME                READY   STATUS    RESTARTS   AGE
pvt-reg-hanadb-0    1/1     Running   0          3m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/pvt-reg-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/pvt-reg-hanadb

kubectl delete ns demo
```
