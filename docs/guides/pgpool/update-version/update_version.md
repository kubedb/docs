---
title: Updating Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-updating-pgpool
    name: updatingPgpool
    parent: pp-updating
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of Pgpool

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `Pgpool`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Pgpool](/docs/guides/pgpool/concepts/pgpool.md)
  - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/pgpool/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgpool](/docs/examples/pgpool) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare Pgpool

Now, we are going to deploy a `Pgpool` =with version `4.4.5`.

### Deploy Pgpool:

In this section, we are going to deploy a Pgpool. Then, in the next section we will update the version  using `PgpoolOpsRequest` CRD. Below is the YAML of the `Pgpool` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-update
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  deletionPolicy: WipeOut
```

Let's create the `Pgpool` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/update-version/pp-update.yaml
pgpool.kubedb.com/pp-update created
```

Now, wait until `pp-update` created has status `Ready`. i.e,

```bash
$ kubectl get pp -n demo
 NAME        TYPE                  VERSION   STATUS   AGE
 pp-update   kubedb.com/v1alpha2   4.4.5     Ready    26s
```

We are now ready to apply the `PgpoolOpsRequest` CR to update this Pgpool.

### update Pgpool Version

Here, we are going to update `Pgpool` from `4.4.5` to `4.5.0`.

#### Create PgpoolOpsRequest:

In order to update the Pgpool, we have to create a `PgpoolOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `PgpoolOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: pgpool-version-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: pp-update
  updateVersion:
    targetVersion: 4.5.0
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `pp-update` Pgpool.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our Pgpool.
- `spec.updateVersion.targetVersion` specifies the expected version of the Pgpool `4.5.0`.


Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/update-version/ppops-update.yaml
pgpoolopsrequest.ops.kubedb.com/pgpool-version-update created
```

#### Verify Pgpool version updated successfully :

If everything goes well, `KubeDB` Ops-manager operator will update the image of `Pgpool` object and related `PetSets` and `Pods`.

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CR,

```bash
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                      TYPE                STATUS       AGE
pgpool-version-update     UpdateVersion       Successful   93s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed to update the Pgpool.

```bash
$ kubectl describe pgpoolopsrequest -n demo pgpool-version-update
Name:         pgpool-version-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T06:31:58Z
  Generation:          1
  Resource Version:    51165
  UID:                 1409aec6-3a25-4b2b-90fe-02e5d8b1e8c1
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pp-update
  Type:    UpdateVersion
  Update Version:
    Target Version:  4.5.0
Status:
  Conditions:
    Last Transition Time:  2024-07-17T06:31:58Z
    Message:               Pgpool ops-request has started to update version
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
    Message:               successfully reconciled the Pgpool with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-17T06:32:52Z
    Message:               Successfully Restarted Pgpool pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-17T06:32:12Z
    Message:               get pod; ConditionStatus:True; PodName:pp-update-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pp-update-0
    Last Transition Time:  2024-07-17T06:32:12Z
    Message:               evict pod; ConditionStatus:True; PodName:pp-update-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pp-update-0
    Last Transition Time:  2024-07-17T06:32:47Z
    Message:               check pod running; ConditionStatus:True; PodName:pp-update-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pp-update-0
    Last Transition Time:  2024-07-17T06:32:52Z
    Message:               Successfully updated Pgpool
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-17T06:32:52Z
    Message:               Successfully updated Pgpool version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age    From                         Message
  ----     ------                                                         ----   ----                         -------
  Normal   Starting                                                       2m55s  KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/pgpool-version-update
  Normal   Starting                                                       2m55s  KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pp-update
  Normal   Successful                                                     2m55s  KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pp-update for PgpoolOpsRequest: pgpool-version-update
  Normal   UpdatePetSets                                                  2m46s  KubeDB Ops-manager Operator  successfully reconciled the Pgpool with updated version
  Normal   get pod; ConditionStatus:True; PodName:pp-update-0             2m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pp-update-0
  Normal   evict pod; ConditionStatus:True; PodName:pp-update-0           2m41s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pp-update-0
  Normal   check pod running; ConditionStatus:False; PodName:pp-update-0  2m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pp-update-0
  Normal   check pod running; ConditionStatus:True; PodName:pp-update-0   2m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pp-update-0
  Normal   RestartPods                                                    2m1s   KubeDB Ops-manager Operator  Successfully Restarted Pgpool pods
  Normal   Starting                                                       2m1s   KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pp-update
  Normal   Successful                                                     2m1s   KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pp-update for PgpoolOpsRequest: pgpool-version-update
```

Now, we are going to verify whether the `Pgpool` and the related `PetSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get pp -n demo pp-update -o=jsonpath='{.spec.version}{"\n"}'                                                                                          
4.5.0

$ kubectl get petset -n demo pp-update -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                               
ghcr.io/appscode-images/pgpool2:4.5.0

$ kubectl get pods -n demo pp-update-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           
ghcr.io/appscode-images/pgpool2:4.5.0@sha256:2697fcad9e11bdc704f6ae0fba85c4451c6b0243140aaaa33e719c3af548bda1
```

You can see from above, our `Pgpool` has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pp -n demo pp-update
kubectl delete pgpoolopsrequest -n demo pgpool-version-update
```