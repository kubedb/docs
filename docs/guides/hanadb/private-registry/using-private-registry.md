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

KubeDB supports private Docker registries. This tutorial shows how to run a KubeDB-managed HanaDB instance using private images.

## Before You Begin

Prepare a Kubernetes cluster and configure `kubectl` to communicate with it. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Private Docker Registry

- Prepare a private Docker [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).

- Push the required HanaDB images into your private registry. For HanaDB, push the database, coordinator, and exporter images from the active `HanaDBVersion` entries.

  ```bash
  $ kubectl get hanadbversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,COORDINATOR_IMAGE:.spec.coordinator.image,EXPORTER_IMAGE:.spec.exporter.image,DEPRECATED:.spec.deprecated
  NAME     VERSION   DB_IMAGE                                               COORDINATOR_IMAGE                          EXPORTER_IMAGE                           DEPRECATED
  2.0.82   2.0.82    docker.io/saplabs/hanaexpress:2.00.082.00.20250528.1   ghcr.io/kubedb/hanadb-coordinator:v0.4.0   ghcr.io/kubedb/hanadb-exporter:1.0.0     <none>
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

## Install the KubeDB Operator

Install the [KubeDB operator](/docs/setup/README.md) in your cluster and configure it to use the private registry that hosts the required images.

## Create a HanaDBVersion

Create a `HanaDBVersion` object that points to images in your private registry. Replace `PRIVATE_REGISTRY` with your registry address.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: HanaDBVersion
metadata:
  name: "2.0.82"
spec:
  coordinator:
    image: PRIVATE_REGISTRY/hanadb-coordinator:v0.4.0
  db:
    image: PRIVATE_REGISTRY/hanaexpress:2.00.082.00.20250528.1
  exporter:
    image: PRIVATE_REGISTRY/hanadb-exporter:1.0.0
  securityContext:
    runAsGroup: 79
    runAsUser: 12000
  updateConstraints:
    allowlist:
    - 2.0.82
  version: "2.0.82"
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/private-registry/pvt-hanadbversion.yaml
hanadbversion.catalog.kubedb.com/2.0.82 configured
```

## Deploy HanaDB from Private Registry

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-private-registry
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/private-registry/pvt-reg-hanadb.yaml
hanadb.kubedb.com/hanadb-private-registry created
```

Check that the HanaDB pod is running:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=hanadb-private-registry"
NAME                         READY   STATUS    RESTARTS   AGE
hanadb-private-registry-0    1/1     Running   0          3m
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/hanadb-private-registry -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/hanadb-private-registry

kubectl delete ns demo
```
