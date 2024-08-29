---
title: Updating Memcached Standalone
menu:
  docs_{{ .version }}:
    identifier: mc-update-version
    name: Memcached
    parent: mc-update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of Memcached

This guide will show you how to use `KubeDB` Enterprise operator to update the version of `Memcached`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Memcached](/docs/guides/memcached/concepts/memcached.md)
  - [MemcachedOpsRequest](/docs/guides/memcached/concepts/memcached-opsrequest.md)
  - [updating Overview](/docs/guides/memcached/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Memcached](/docs/examples/Memcached) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Memcached Database

Now, we are going to deploy a `Memcached` standalone database with version `1.6.22`.

### Deploy Memcached:

In this section, we are going to deploy a Memcached database. Then, in the next section we will update the version of the database using `MemcachedOpsRequest` CRD. Below is the YAML of the `Memcached` CR that we are going to create,

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
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: WipeOut

```

Let's create the `Memcached` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/update-version/memcached.yaml
memcached.kubedb.com/memcd-quickstart created
```

Now, wait until `memcd-quickstart` created has status `Ready`. i.e,

```bash
$ kubectl get mc -n demo
NAME               VERSION   STATUS   AGE
memcd-quickstart   1.6.22    Ready    3m
```

We are now ready to apply the `MemcachedOpsRequest` CR to update this database.

### Update Memcached Version

Here, we are going to update `Memcached` sdatabase from `1.6.22` to `1.6.29`.

#### Create MemcachedOpsRequest:

In order to update the memcached database, we have to create a `MemcachedOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `MemcachedOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: update-memcd
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: memcd-quickstart
  updateVersion:
    targetVersion: 1.6.29
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `memcd-quickstart` Memcached database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `1.6.29`.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/update-version/opsrequest-version-update.yaml
memcachedopsrequest.ops.kubedb.com/update-memcd created
```

#### Verify Memcached version updated successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `Memcached` object and related `PetSets` and `Pods`.

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CR,

```bash
$ watch kubectl get memcachedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME           TYPE            STATUS       AGE
update-memcd   UpdateVersion   Successful   7m
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded.

Now, we are going to verify whether the `Memcached` and the related `PetSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get memcached -n demo memcd-quickstart -o=jsonpath='{.spec.version}{"\n"}'
1.6.29

$ kubectl get petset -n demo memcd-quickstart -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/memcached:1.6.29-alpine

$ kubectl get pods -n demo memcd-quickstart-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/memcached:1.6.29-alpine
```

You can see from above, our `Memcached` database has been updated with the new version. So, the UpdateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo mc/memcd-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
memcached.kubedb.com/memcd-quickstart patched

$ kubectl delete -n demo Memcached memcd-quickstart
memcached.kubedb.com "memcd-quickstart" deleted

$ kubectl delete -n demo Memcachedopsrequest update-standalone
memcachedopsrequest.ops.kubedb.com "update-memcd" deleted
```
