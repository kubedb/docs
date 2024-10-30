---
title: Updating SingleStore Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-updating-cluster
    name: Update Version OpsRequest
    parent: guides-sdb-updating
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of SingleStore Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `SingleStore` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [Cluster](/docs/guides/singlestore/clustering/overview/)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/singlestore/update-version/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare SingleStore Cluster

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

Now, we are going to deploy a `SingleStore` cluster database with version `8.5.30`.

### Deploy SingleStore cluster

In this section, we are going to deploy a SingleStore Cluster. Then, in the next section we will update the version of the database using `SingleStoreOpsRequest` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sample-sdb
  namespace: demo
spec:
  version: "8.5.30"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        storageClassName: "longhorn"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut

```

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/update-version/cluster/examples/sample-sdb.yaml
singlestore.kubedb.com/sample-sdb created
```

Now, wait until `sample-sdb` created has status `Ready`. i.e,

```bash
$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
sample-sdb   kubedb.com/v1alpha2   8.5.30    Ready    4m37s
```

We are now ready to apply the `SingleStoreOpsRequest` CR to update this database.

### update SingleStore Version

Here, we are going to update `SingleStore` cluster from `8.5.30` to `8.7.10`.

#### Create SingleStoreOpsRequest:

In order to update the database cluster, we have to create a `SingleStoreOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `SingleStoreOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-update-patch
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: sample-sdb
  updateVersion:
    targetVersion: "8.7.10"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `sample-sdb` SingleStore database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `8.7.10`.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/update-version/cluster/examples/sdbops-update.yaml
singlestoreopsrequest.ops.kubedb.com/sdb-update-patch created
```

#### Verify SingleStore version updated successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the image of `SingleStore` object and related `PetSets` and `Pods`.

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CR,

```bash
$ kubectl get sdbops -n demo 
NAME               TYPE            STATUS       AGE
sdb-update-patch   UpdateVersion   Successful   3m46s
```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded.

Now, we are going to verify whether the `SingleStore` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get sdb -n demo sample-sdb -o=jsonpath='{.spec.version}{"\n"}'
8.7.10

$ kubectl get petset -n demo sample-sdb-aggregator -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/singlestore-node:alma-8.7.10-95e2357384

$ kubectl get petset -n demo sample-sdb-leaf -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/singlestore-node:alma-8.7.10-95e2357384

$ kubectl get pods -n demo sample-sdb-aggregator-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/singlestore-node:alma-8.7.10-95e2357384

$ kubectl get pods -n demo sample-sdb-leaf-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/singlestore-node:alma-8.7.10-95e2357384
```

You can see from above, our `SingleStore` cluster database has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete sdb -n demo sample-sdb
$ kubectl delete singlestoreopsrequest -n demo sdb-update-patch 
```