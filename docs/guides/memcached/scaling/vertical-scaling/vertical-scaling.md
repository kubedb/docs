---
title: Vertical Scaling Memcached
menu:
  docs_{{ .version }}:
    identifier: mc-vertical-scaling
    name: Vertical Scaling
    parent: vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Memcached

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a Memcached database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Memcached](/docs/guides/memcached/concepts/memcached.md)
  - [MemcachedOpsRequest](/docs/guides/memcached/concepts/memcached-opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/memcached/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/memcached](/docs/examples/memcached) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Memcahced

Here, we are going to deploy a  `Memcahced` database using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Memcahced Database

Now, we are going to deploy a `Memcached` database with version `1.6.22`.

### Deploy Memcahced

In this section, we are going to deploy a Memcached database. Then, in the next section we will update the resources of the database using `MemcachedOpsRequest` CRD. Below is the YAML of the `Memcached` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 100m
              memory: 128Mi
            requests:
              cpu: 100m
              memory: 128Mi
  deletionPolicy: WipeOut
```

Let's create the `Memcached` CR we have shown above, 

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/scaling/memcached-vertical.yaml
memcached.kubedb.com/memcd-quickstart created
```

Now, wait until `memcd-quickstart` has status `Ready`. i.e. ,

```bash
$ kubectl get memcached -n demo
NAME               VERSION   STATUS   AGE
memcd-quickstart   1.6.22    Ready    5m
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo memcd-quickstart-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "100m",
    "memory": "128Mi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "128Mi"
  }
}
```

We can see from the above output that there are some default resources set by the operator. And the scheduler will choose the best suitable node to place the container of the Pod.

We are now ready to apply the `MemcachedOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the database to meet the desired resources after scaling.

#### Create MemcahedOpsRequest

In order to update the resources of the database, we have to create a `MemcachedOpsRequest` CR with our desired resources. Below is the YAML of the `MemcachedOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcached-mc
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: memcd-quickstart
  verticalScaling:
    memcached:
      resources:
        requests:
          memory: "400Mi"
          cpu: "500m"
        limits:
          memory: "400Mi"
          cpu: "500m"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `memcd-quickstart` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.memcached` specifies the desired resources after scaling.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/scaling/vertical-scaling.yaml
memcachedopsrequest.ops.kubedb.com/memcached-mc created
```

#### Verify Memcached Database resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `Memcached` object and related `PetSets` and `Pods`.

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CR,

```bash
$ watch kubectl get memcachedopsrequest -n demo
NAME                  TYPE                STATUS       AGE
memcached-mc          VerticalScaling     Successful   5m
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded. 
Now, we are going to verify from the Pod yaml whether the resources of the Memcached database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo memcd-quickstart-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "400Mi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the Memcached database.

## Cleaning up

To clean up the Kubernetes resources created by this turorial, run:

```bash

$ kubectl patch -n demo mc/memcached-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
memcached.kubedb.com/memcd-quickstart patched

$ kubectl delete -n demo memcached memcd-quickstart
memcached.kubedb.com "memcd-quickstart" deleted

$ kubectl delete memcachedopsrequest -n demo memcached-mc
memcachedopsrequest.ops.kubedb.com "memcached-mc" deleted
```