---
title: Run Neo4j using Private Registry
menu:
  docs_{{ .version }}:
    identifier: neo4j-using-private-registry
    name: Quickstart
    parent: neo4j-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Private Docker Registry

KubeDB supports using private Docker registries. This tutorial will show you how to run KubeDB managed Neo4j database using private Docker images.

## Before You Begin

> Prerequisites: A running Kubernetes cluster with KubeDB installed. See the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md) if you need to set up your environment.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Private Docker Registry

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).

- Push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For Neo4j, push the `DB_IMAGE` of the following Neo4jVersions, where `deprecated` is not true.

  ```bash
  $ kubectl get neo4jversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,DEPRECATED:.spec.deprecated
  NAME      VERSION   DB_IMAGE                       DEPRECATED
  2025.11.2 2025.11.2 kubedb/neo4j:2025.11.2         <none>
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

Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that it points to the `DOCKER_REGISTRY` you wish to pull images from.

## Create Neo4jVersion CRD

Create a Neo4jVersion CRD specifying images from your private registry. Replace `PRIVATE_REGISTRY` with your private registry.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: Neo4jVersion
metadata:
  name: "2025.12.1"
spec:
  db:
    image: PRIVATE_REGISTRY/neo4j:2025.12.1
  version: "2025.12.1"
```

```bash
$ kubectl apply -f pvt-neo4jversion.yaml
neo4jversion.catalog.kubedb.com/2025.12.1 created
```

## Deploy Neo4j from Private Registry

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: pvt-reg-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/private-registry/pvt-reg-neo4j.yaml
neo4j.kubedb.com/pvt-reg-neo4j created
```

Check that the Neo4j is in Running state:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=pvt-reg-neo4j"
NAME               READY   STATUS    RESTARTS   AGE
pvt-reg-neo4j-0    1/1     Running   0          3m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo neo4j/pvt-reg-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/pvt-reg-neo4j
kubectl delete ns demo
```
