---
title: Update Qdrant Version
menu:
  docs_{{ .version }}:
    identifier: qdrant-update-version-ops
    name: Update Version
    parent: qdrant-update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update Qdrant Version

This guide will show you how to use `KubeDB` ops-manager operator to update the version of `Qdrant` cr.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/qdrant/update-version/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/qdrant/update-version/versionupgrading/yamls](/docs/guides/qdrant/update-version/versionupgrading/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Version Updating on Qdrant

Here, we are going to deploy a `Qdrant` instance using a supported version by `KubeDB` provisioner. Then we are going to apply update-ops-request on it.

#### Prepare Qdrant

At first, we are going to deploy a Qdrant instance using a supported `Qdrant` version. In the next two sections, we are going to find out the supported versions and version update constraints.

**Find supported QdrantVersion:**

When you have installed `KubeDB`, it has created `QdrantVersion` CR for all supported `Qdrant` versions. Let's check the supported versions:

```bash
$ kubectl get qdrantversion
NAME       VERSION   DB_IMAGE                      DEPRECATED   AGE
1.7.4      1.7.4     qdrant/qdrant:v1.7.4                       5m
1.8.4      1.8.4     qdrant/qdrant:v1.8.4                       5m
1.9.7      1.9.7     qdrant/qdrant:v1.9.7                       5m
1.10.1     1.10.1    qdrant/qdrant:v1.10.1                      5m
1.11.5     1.11.5    qdrant/qdrant:v1.11.5                      5m
1.12.6     1.12.6    qdrant/qdrant:v1.12.6                      5m
1.13.4     1.13.4    qdrant/qdrant:v1.13.4                      5m
1.14.1     1.14.1    qdrant/qdrant:v1.14.1                      5m
1.15.1     1.15.1    qdrant/qdrant:v1.15.1                      5m
1.16.1     1.16.1    qdrant/qdrant:v1.16.1                      5m
1.17.0     1.17.0    qdrant/qdrant:v1.17.0                      5m
1.18.0     1.18.0    qdrant/qdrant:v1.18.0                      5m
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `Qdrant`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `QdrantVersion` for the `Qdrant` instance that we will update from this version to another. In the next section, we are going to verify version update constraints.

**Check update Constraints:**

Qdrant supports rolling version updates. You can update from any currently running version to a newer patch or minor version. Major version jumps should follow the Qdrant upstream upgrade notes. For example, you can update directly from `1.17.0` to `1.18.0`.

Let's get one of the `qdrantversion` YAMLs:

```bash
$ kubectl get qdrantversion 1.17.0 -o yaml | kubectl neat
apiVersion: catalog.kubedb.com/v1alpha1
kind: QdrantVersion
metadata:
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
  name: "1.17.0"
spec:
  db:
    image: qdrant/qdrant:v1.17.0
  version: "1.17.0"
```

**Deploy Qdrant Instance:**

In this section, we are going to deploy a Qdrant instance. Then, in the next section, we will update the version of the database using `UpdateVersion`. Below is the YAML of the `Qdrant` cr that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` cr we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/update-version/versionupgrading/yamls/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

**Wait for the database to be ready:**

`KubeDB` operator watches for `Qdrant` objects using Kubernetes API. When a `Qdrant` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc.

Now, watch `Qdrant` is going to `Running` state and also watch `PetSet` and its pod is created and going to `Running` state:

```bash
$ watch -n 3 kubectl get qdrant -n demo
Every 3.0s: kubectl get qdrant -n demo

NAME              VERSION   STATUS   AGE
qdrant-sample     1.17.0    Ready    3m42s

$ watch -n 3 kubectl get petset -n demo qdrant-sample
Every 3.0s: kubectl get petset -n demo qdrant-sample

NAME              READY   AGE
qdrant-sample     3/3     4m17s

$ watch -n 3 kubectl get pod -n demo
Every 3.0s: kubectl get pods -n demo

NAME                READY   STATUS    RESTARTS   AGE
qdrant-sample-0     1/1     Running   0          4m55s
qdrant-sample-1     1/1     Running   0          4m12s
qdrant-sample-2     1/1     Running   0          3m38s
```

Let's verify the `Qdrant`, the `PetSet` and its `Pod` image version:

```bash
$ kubectl get qdrant -n demo qdrant-sample -o=jsonpath='{.spec.version}{"\n"}'
1.17.0

$ kubectl get petset -n demo qdrant-sample -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
qdrant/qdrant:v1.17.0

$ kubectl get pod -n demo qdrant-sample-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
qdrant/qdrant:v1.17.0
```

We are ready to apply version updating on this `Qdrant` instance.

#### UpdateVersion

Here, we are going to update `Qdrant` from version `1.17.0` to `1.18.0`.

**Create QdrantOpsRequest:**

To update the instance, you have to create a `QdrantOpsRequest` cr with your desired version that is supported by `KubeDB`. Below is the YAML of the `QdrantOpsRequest` cr that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-update-version
  namespace: demo
spec:
  type: UpdateVersion
  updateVersion:
    targetVersion: "1.18.0"
  databaseRef:
    name: qdrant-sample
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `qdrant-sample` Qdrant database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version `1.18.0` after updating.
- `spec.timeout` specifies the timeout for the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#spectimeout)).
- `spec.apply` specifies when to apply the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#specapply)).

Let's create the `QdrantOpsRequest` cr we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/update-version/versionupgrading/yamls/update_version.yaml
qdrantopsrequest.ops.kubedb.com/qdops-update-version created
```

**Verify Qdrant version updated successfully:**

If everything goes well, `KubeDB` ops-manager operator will update the image of `Qdrant`, `PetSet`, and its `Pod`.

At first, we will wait for `QdrantOpsRequest` to be successful. Run the following command to watch `QdrantOpsRequest` cr:

```bash
$ watch -n 3 kubectl get QdrantOpsRequest -n demo qdops-update-version
Every 3.0s: kubectl get QdrantOpsRequest -n demo qdops-update-version

NAME                     TYPE            STATUS       AGE
qdops-update-version     UpdateVersion   Successful   3m12s
```

We can see from the above output that the `QdrantOpsRequest` has succeeded. If we describe the `QdrantOpsRequest`, we shall see that the `Qdrant`, `PetSet`, and its `Pod` have updated with a new image.

```bash
$ kubectl describe QdrantOpsRequest -n demo qdops-update-version
Name:         qdops-update-version
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         QdrantOpsRequest
Spec:
  Database Ref:
    Name:  qdrant-sample
  Type:    UpdateVersion
  Update Version:
    Target Version:  1.18.0
Status:
  Conditions:
    Last Transition Time:  2026-05-01T10:00:00Z
    Message:               Qdrant ops request is updating version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2026-05-01T10:00:05Z
    Message:               Successfully updated PetSets update strategy type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-05-01T10:00:30Z
    Message:               Successfully updated pod images
    Observed Generation:   1
    Reason:                UpdatePetSetImage
    Status:                True
    Type:                  UpdatePetSetImage
    Last Transition Time:  2026-05-01T10:02:45Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason            Age    From                        Message
  ----    ------            ----   ----                        -------
  Normal  PauseDatabase     3m12s  KubeDB Enterprise Operator  Pausing Qdrant demo/qdrant-sample
  Normal  PauseDatabase     3m12s  KubeDB Enterprise Operator  Successfully paused Qdrant demo/qdrant-sample
  Normal  Updating          3m12s  KubeDB Enterprise Operator  Updating PetSets
  Normal  Updating          3m12s  KubeDB Enterprise Operator  Successfully Updated PetSets
  Normal  UpdatePetSetImage 1m10s  KubeDB Enterprise Operator  Successfully Updated pod images
  Normal  ResumeDatabase    1m10s  KubeDB Enterprise Operator  Resuming Qdrant demo/qdrant-sample
  Normal  ResumeDatabase    1m10s  KubeDB Enterprise Operator  Successfully resumed Qdrant demo/qdrant-sample
  Normal  Successful        1m10s  KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify whether the `Qdrant`, `PetSet` and its `Pod` have updated with the new image. Let's check:

```bash
$ kubectl get qdrant -n demo qdrant-sample -o=jsonpath='{.spec.version}{"\n"}'
1.18.0

$ kubectl get petset -n demo qdrant-sample -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
qdrant/qdrant:v1.18.0

$ kubectl get pod -n demo qdrant-sample-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
qdrant/qdrant:v1.18.0
```

You can see above that our `Qdrant` has been updated with the new version. It verifies that we have successfully updated our Qdrant instance.

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete QdrantOpsRequest -n demo qdops-update-version
kubectl delete ns demo
```