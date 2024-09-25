---
title: Horizontal Scaling Memcached
menu:
  docs_{{ .version }}:
    identifier: mc-horizontal-scaling
    name: Horizontal Scaling
    parent: horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Memcached

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Memcached` database replicas.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Memcached](/docs/guides/memcached/concepts/memcached.md)
  - [MemcachedOpsRequest](/docs/guides/memcached/concepts/memcached-opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/memcached/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/memcached](/docs/examples/memcached) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Memcached

Here, we are going to deploy a `Memcached` database using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Memcached Database

Now, we are going to deploy a `Memcached` database with version `1.6.22`.

### Deploy Memcached Database

In this section, we are going to deploy a Memcached database. Then, in the next section we will update the resources of the database using `MemcachedOpsRequest` CRD. Below is the YAML of the `Memcached` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 3
  version: "1.6.22"
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: WipeOut
```

Let's create the `Memcached` CR we have shown above, 

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/scaling/memcached-horizontal.yaml
memcached.kubedb.com/memcd-quickstart created
```

Now, wait until `memcd-quickstart` has status `Ready`. i.e. ,

```bash
$ kubectl get memcached -n demo
NAME               VERSION   STATUS   AGE
memcd-quickstart   1.6.22    Ready    5m
```

Let's check the number of replicas this database has from the Memcached object

```bash
$ kubectl get memcached -n demo memcd-quickstart -o json | jq '.spec.replicas'
3
```

We are now ready to apply the `MemcachedOpsRequest` CR to update the resources of this database.

### Horizontal Scaling

Here, we are going to scale up the replicas of the memcached database to meet the desired resources after scaling.

#### Create MemcachedOpsRequest

In order to  scale up the replicas of the memcached database, we have to create a `MemcachedOpsRequest` CR with our desired number of replicas. Below is the YAML of the `MemcachedOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-horizontal-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: memcd-quickstart
  horizontalScaling:
    replicas: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `memcd-quickstart` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/scaling/horizontal-scaling.yaml
memcachedopsrequest.ops.kubedb.com/memcd-horizontal-up created
```

#### Verify Memcached resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `Memcached` object and related `PetSets`.

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CR,

```bash
$ watch kubectl get memcachedopsrequest -n demo memcd-horizontal-up
NAME                  TYPE                STATUS       AGE
memcd-horizontal-up   HorizontalScaling   Successful   3m
```

Now, we are going to verify if the number of replicas the memcached database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get memcached -n demo memcd-quickstart -o json | jq '.spec.replicas'
5
```

The above output verifies that we have successfully scaled up the replicas of the Memcached database.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash

$ kubectl patch -n demo mc/memcd-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
memcached.kubedb.com/memcd-quickstart patched

$ kubectl delete -n demo memcached memcd-quickstart
memcached.kubedb.com "memcd-quickstart" deleted

$ kubectl delete -n demo memcachedopsrequest memcd-horizontal-up 
memcachedopsrequest.ops.kubedb.com "memcd-horizontal-up" deleted
```