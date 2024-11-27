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

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgbouncer](/docs/examples/pgbouncer) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

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
  Creation Timestamp:  2024-11-27T09:40:03Z
  Generation:          1
  Resource Version:    41823
  UID:                 a53940fd-4d2d-4b4b-8ef1-0419dfbce660
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pb-update
  Type:    UpdateVersion
  Update Version:
    Target Version:  1.23.1
Status:
  Conditions:
    Last Transition Time:  2024-11-27T09:40:03Z
    Message:               Controller has started to Progress with UpdateVersion of PgBouncerOpsRequest: demo/pgbouncer-version-update
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2024-11-27T09:40:08Z
    Message:               Successfully updated Petset resource
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-27T09:40:13Z
    Message:               get pod; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-update-0
    Last Transition Time:  2024-11-27T09:40:13Z
    Message:               evict pod; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pb-update-0
    Last Transition Time:  2024-11-27T09:40:18Z
    Message:               check replica func; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckReplicaFunc--pb-update-0
    Last Transition Time:  2024-11-27T09:40:18Z
    Message:               check pod ready; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--pb-update-0
    Last Transition Time:  2024-11-27T09:40:48Z
    Message:               check pg bouncer running; ConditionStatus:True; PodName:pb-update-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPgBouncerRunning--pb-update-0
    Last Transition Time:  2024-11-27T09:40:53Z
    Message:               Restarting all pods performed successfully in PgBouncer: demo/pb-update for PgBouncerOpsRequest: pgbouncer-version-update
    Observed Generation:   1
    Reason:                RestartPodsSucceeded
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-27T09:41:04Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-11-27T09:41:04Z
    Message:               Successfully version updated
    Observed Generation:   1
    Reason:                VersionUpdate
    Status:                True
    Type:                  VersionUpdate
    Last Transition Time:  2024-11-27T09:41:04Z
    Message:               Controller has successfully completed  with UpdateVersion of PgBouncerOpsRequest: demo/pgbouncer-version-update
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age   From                         Message
  ----     ------                                                                ----  ----                         -------
  Normal   Starting                                                              114s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pgbouncer-version-update
  Normal   Starting                                                              114s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-update
  Normal   Successful                                                            114s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-update for PgBouncerOpsRequest: pgbouncer-version-update
  Warning  get pod; ConditionStatus:True; PodName:pb-update-0                    104s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-update-0
  Warning  evict pod; ConditionStatus:True; PodName:pb-update-0                  104s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pb-update-0
  Warning  check replica func; ConditionStatus:True; PodName:pb-update-0         99s   KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pb-update-0
  Warning  check pod ready; ConditionStatus:True; PodName:pb-update-0            99s   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pb-update-0
  Warning  check pg bouncer running; ConditionStatus:False; PodName:pb-update-0  89s   KubeDB Ops-manager Operator  check pg bouncer running; ConditionStatus:False; PodName:pb-update-0
  Warning  check replica func; ConditionStatus:True; PodName:pb-update-0         89s   KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pb-update-0
  Warning  check pod ready; ConditionStatus:True; PodName:pb-update-0            89s   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pb-update-0
  Warning  check replica func; ConditionStatus:True; PodName:pb-update-0         79s   KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pb-update-0
  Warning  check pod ready; ConditionStatus:True; PodName:pb-update-0            79s   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pb-update-0
  Warning  check replica func; ConditionStatus:True; PodName:pb-update-0         69s   KubeDB Ops-manager Operator  check replica func; ConditionStatus:True; PodName:pb-update-0
  Warning  check pod ready; ConditionStatus:True; PodName:pb-update-0            69s   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:pb-update-0
  Warning  check pg bouncer running; ConditionStatus:True; PodName:pb-update-0   69s   KubeDB Ops-manager Operator  check pg bouncer running; ConditionStatus:True; PodName:pb-update-0
  Normal   Successful                                                            64s   KubeDB Ops-manager Operator  Restarting all pods performed successfully in PgBouncer: demo/pb-update for PgBouncerOpsRequest: pgbouncer-version-update
  Normal   Starting                                                              53s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-update
  Normal   Successful                                                            53s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-update
  Normal   Successful                                                            53s   KubeDB Ops-manager Operator  Controller has Successfully updated the version of PgBouncer database: demo/pb-update
```

Now, we are going to verify whether the `PgBouncer` and the related `PetSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get pb -n demo pb-update -o=jsonpath='{.spec.version}{"\n"}'                                                                                          
1.23.1

$ kubectl get petset -n demo pb-update -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                               
ghcr.io/kubedb/pgbouncer:1.23.1@sha256:9829a24c60938ab709fe9e039fecd9f0019354edf4e74bfd9e62bb2203e945ee

$ kubectl get pods -n demo pb-update-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           
ghcr.io/kubedb/pgbouncer:1.23.1@sha256:9829a24c60938ab709fe9e039fecd9f0019354edf4e74bfd9e62bb2203e945ee
```

You can see from above, our `PgBouncer` has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pb -n demo pb-update
kubectl delete pgbounceropsrequest -n demo pgbouncer-version-update
```