---
title: Updating PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-updating-pgbouncer
    name: updatingPgBouncer
    parent: pb-updating
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of PgBouncer

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `PgBouncer`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/pgbouncer/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgbouncer](/docs/examples/pgbouncer) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare PgBouncer

Now, we are going to deploy a `PgBouncer` with version `1.18.0`.

### Deploy PgBouncer:

In this section, we are going to deploy a PgBouncer. Then, in the next section we will update the version  using `PgBouncerOpsRequest` CRD. Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-update
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    poolMode: session
    port: 5432
    reservePoolSize: 5
    maxClientConnections: 87
    defaultPoolSize: 2
    minPoolSize: 1
    authType: md5
  deletionPolicy: WipeOut
```

Let's create the `PgBouncer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/update-version/pb-update.yaml
pgbouncer.kubedb.com/pb-update created
```

Now, wait until `pb-update` created has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
 NAME        TYPE                  VERSION   STATUS   AGE
 pb-update   kubedb.com/v1         1.18.0    Ready    26s
```

We are now ready to apply the `PgBouncerOpsRequest` CR to update this PgBouncer.

### update PgBouncer Version

Here, we are going to update `PgBouncer` from `1.18.0` to `1.23.1`.

#### Create PgBouncerOpsRequest:

In order to update the PgBouncer, we have to create a `PgBouncerOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `PgBouncerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-version-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: pb-update
  updateVersion:
    targetVersion: 1.23.1
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `pb-update` PgBouncer.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our PgBouncer.
- `spec.updateVersion.targetVersion` specifies the expected version of the PgBouncer `1.23.1`.


Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/update-version/pbops-update.yaml
pgbounceropsrequest.ops.kubedb.com/pgbouncer-version-update created
```

#### Verify PgBouncer version updated successfully :

If everything goes well, `KubeDB` Ops-manager operator will update the image of `PgBouncer` object and related `PetSets` and `Pods`.

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CR,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                      TYPE                STATUS       AGE
pgbouncer-version-update  UpdateVersion       Successful   93s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to update the PgBouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pgbouncer-version-update
Name:         pgbouncer-version-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T06:31:58Z
  Generation:          1
  Resource Version:    51165
  UID:                 1409aec6-3a25-4b2b-90fe-02e5d8b1e8c1
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pb-update
  Type:    UpdateVersion
  Update Version:
    Target Version:  1.18.0
Status:
  Conditions:
    Last Transition Time:  2024-07-17T06:31:58Z
    Message:               PgBouncer ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2024-07-17T06:32:01Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-17T06:32:07Z
    Message:               successfully reconciled the PgBouncer with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-17T06:32:52Z
    Message:               Successfully Restarted PgBouncer pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-17T06:32:12Z
    Message:               get pod; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-update-0
    Last Transition Time:  2024-07-17T06:32:12Z
    Message:               evict pod; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pb-update-0
    Last Transition Time:  2024-07-17T06:32:47Z
    Message:               check pod running; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pb-update-0
    Last Transition Time:  2024-07-17T06:32:52Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-17T06:32:52Z
    Message:               Successfully updated PgBouncer version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age    From                         Message
  ----     ------                                                         ----   ----                         -------
  Normal   Starting                                                       2m55s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pgbouncer-version-update
  Normal   Starting                                                       2m55s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-update
  Normal   Successful                                                     2m55s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-update for PgBouncerOpsRequest: pgbouncer-version-update
  Normal   UpdatePetSets                                                  2m46s  KubeDB Ops-manager Operator  successfully reconciled the PgBouncer with updated version
  Normal   get pod; ConditionStatus:True; PodName:pb-update-0             2m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-update-0
  Normal   evict pod; ConditionStatus:True; PodName:pb-update-0           2m41s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pb-update-0
  Normal   check pod running; ConditionStatus:False; PodName:pb-update-0  2m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pb-update-0
  Normal   check pod running; ConditionStatus:True; PodName:pb-update-0   2m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pb-update-0
  Normal   RestartPods                                                    2m1s   KubeDB Ops-manager Operator  Successfully Restarted PgBouncer pods
  Normal   Starting                                                       2m1s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-update
  Normal   Successful                                                     2m1s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-update for PgBouncerOpsRequest: pgbouncer-version-update
```

Now, we are going to verify whether the `PgBouncer` and the related `PetSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get pb -n demo pb-update -o=jsonpath='{.spec.version}{"\n"}'                                                                                          
1.23.1

$ kubectl get petset -n demo pb-update -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                               
mongo:4.0.5

$ kubectl get pods -n demo mg-standalone-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           
ghcr.io/appscode-images/pgbouncer:1.23.1@sha256:2697fcad9e11bdc704f6ae0fba85c4451c6b0243140aaaa33e719c3af548bda1
```

You can see from above, our `PgBouncer` has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pb -n demo pb-update
kubectl delete pgbounceropsrequest -n demo pgbouncer-version-update
```