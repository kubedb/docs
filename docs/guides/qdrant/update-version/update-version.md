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

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/update-version](/docs/examples/qdrant/update-version) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Version Updating on Qdrant

Here, we are going to deploy a `Qdrant` instance using a supported version by `KubeDB` provisioner. Then we are going to apply update-ops-request on it.

### Prepare Qdrant

At first, we are going to deploy a Qdrant instance using a supported `Qdrant` version. In the next two sections, we are going to find out the supported versions and version update constraints.

**Find supported QdrantVersion:**

When you have installed `KubeDB`, it has created `QdrantVersion` CR for all supported `Qdrant` versions. Let's check the supported versions:

```bash
$ kubectl get qdrantversion
NAME     VERSION   DB_IMAGE                                       DEPRECATED   AGE
1.15.4   1.15.4    docker.io/qdrant/qdrant:v1.15.4-unprivileged                24d
1.16.2   1.16.2    docker.io/qdrant/qdrant:v1.16.2-unprivileged                24d
1.17.0   1.17.0    docker.io/qdrant/qdrant:v1.17.0-unprivileged                24d
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `Qdrant`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `QdrantVersion` for the `Qdrant` instance that we will update from this version to another. In the next section, we are going to verify version update constraints.

**Check update Constraints:**

Qdrant supports rolling version updates. You can update from any currently running version to a newer patch or minor version. Major version jumps should follow the Qdrant upstream upgrade notes. For example, you can update directly from `1.16.2` to `1.17.0`.

Let's get one of the `qdrantversion` YAMLs:

```bash
$ kubectl get qdrantversion 1.17.0 -o yaml
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
    image: docker.io/qdrant/qdrant:v1.17.0-unprivileged
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
  version: "1.16.2"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/update-version/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

**Wait for the database to be ready:**

`KubeDB` operator watches for `Qdrant` objects using Kubernetes API. When a `Qdrant` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc.

Now, watch `Qdrant` is going to `Running` state and also watch `PetSet` and its pod is created and going to `Running` state:

```bash
$ watch -n 3 kubectl get qdrant -n demo
Every 3.0s: kubectl get qdrant -n demo

NAME              VERSION   STATUS   AGE
qdrant-sample     1.16.2    Ready    3m42s

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
1.16.2

$ kubectl get petset -n demo qdrant-sample -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
docker.io/qdrant/qdrant:v1.16.2-unprivileged

$ kubectl get pod -n demo qdrant-sample-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
docker.io/qdrant/qdrant:v1.16.2-unprivileged

We are ready to apply version updating on this `Qdrant` instance.

### UpdateVersion

Here, we are going to update `Qdrant` from version `1.16.2` to `1.17.0`.

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
    targetVersion: "1.17.0"
  databaseRef:
    name: qdrant-sample
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `qdrant-sample` Qdrant database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version `1.17.0` after updating.
- `spec.timeout` specifies the timeout for the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#spectimeout)).
- `spec.apply` specifies when to apply the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#specapply)).

Let's create the `QdrantOpsRequest` cr we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/update-version/ops-request.yaml
qdrantopsrequest.ops.kubedb.com/qdops-update-version created
```

#### Verify Qdrant version updated successfully

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
Metadata:
  Creation Timestamp:  2026-05-15T05:36:46Z
  Generation:          1
  Resource Version:    3356440
  UID:                 156625b9-5f36-465e-aa88-ff0e5452c7ba
Spec:
  Apply:  IfReady
  Database Ref:
    Name:       qdrant-sample
  Max Retries:  1
  Timeout:      5m
  Type:         UpdateVersion
  Update Version:
    Target Version:  1.17.0
Status:
  Conditions:
    Last Transition Time:  2026-05-15T05:36:46Z
    Message:               Qdrant ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2026-05-15T05:37:01Z
    Message:               successfully reconciled the Qdrant with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-05-15T05:38:10Z
    Message:               Successfully Restarted Qdrant nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2026-05-15T05:37:15Z
    Message:               get pod; ConditionStatus:True; PodName:qdrant-sample-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--qdrant-sample-0
    Last Transition Time:  2026-05-15T05:37:17Z
    Message:               evict pod; ConditionStatus:True; PodName:qdrant-sample-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--qdrant-sample-0
    Last Transition Time:  2026-05-15T05:37:45Z
    Message:               running pod; ConditionStatus:True; PodName:qdrant-sample-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--qdrant-sample-0
    Last Transition Time:  2026-05-15T05:37:30Z
    Message:               get pod; ConditionStatus:True; PodName:qdrant-sample-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--qdrant-sample-1
    Last Transition Time:  2026-05-15T05:37:31Z
    Message:               evict pod; ConditionStatus:True; PodName:qdrant-sample-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--qdrant-sample-1
    Last Transition Time:  2026-05-15T05:37:45Z
    Message:               running pod; ConditionStatus:True; PodName:qdrant-sample-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--qdrant-sample-1
    Last Transition Time:  2026-05-15T05:37:50Z
    Message:               get pod; ConditionStatus:True; PodName:qdrant-sample-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--qdrant-sample-2
    Last Transition Time:  2026-05-15T05:37:52Z
    Message:               evict pod; ConditionStatus:True; PodName:qdrant-sample-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--qdrant-sample-2
    Last Transition Time:  2026-05-15T05:38:05Z
    Message:               running pod; ConditionStatus:True; PodName:qdrant-sample-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--qdrant-sample-2
    Last Transition Time:  2026-05-15T05:38:11Z
    Message:               Successfully updated Qdrant version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age   From                         Message
  ----     ------                                                      ----  ----                         -------
  Normal   Starting                                                    86s   KubeDB Ops-manager Operator  Pausing Qdrant database demo/qdrant-sample
  Normal   Successful                                                  86s   KubeDB Ops-manager Operator  Successfully paused Qdrant database: demo/qdrant-sample for QdrantOpsRequest: qdops-update-version
  Normal   UpdatePetSets                                               74s   KubeDB Ops-manager Operator  successfully reconciled the Qdrant with updated version
  Warning  get pod; ConditionStatus:True; PodName:qdrant-sample-0      60s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:qdrant-sample-0
  Warning  evict pod; ConditionStatus:True; PodName:qdrant-sample-0    58s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:qdrant-sample-0
  Normal   RestartPods                                                 5s    KubeDB Ops-manager Operator  Successfully Restarted Qdrant nodes
  Normal   Starting                                                    4s    KubeDB Ops-manager Operator  Resuming Qdrant database: demo/qdrant-sample
  Normal   Successful                                                  4s    KubeDB Ops-manager Operator  Successfully resumed Qdrant database: demo/qdrant-sample for QdrantOpsRequest: qdops-update-version
```

Now, we are going to verify whether the `Qdrant`, `PetSet` and its `Pod` have updated with the new image. Let's check:

```bash
$ kubectl get qdrant -n demo qdrant-sample -o=jsonpath='{.spec.version}{"\n"}'
1.17.0

$ kubectl get petset -n demo qdrant-sample -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
docker.io/qdrant/qdrant:v1.17.0-unprivileged

$ kubectl get pod -n demo qdrant-sample-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
docker.io/qdrant/qdrant:v1.17.0-unprivileged
```

You can see above that our `Qdrant` has been updated with the new version. It verifies that we have successfully updated our Qdrant instance.

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted

$ kubectl delete qdrantopsrequest -n demo qdops-update-version
qdrantopsrequest.ops.kubedb.com "qdops-update-version" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```