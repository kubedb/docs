---
title: Postgres Standalone Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: pg-volume-expansion-standalone
    name: Standalone
    parent: pg-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Postgres Standalone Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Postgres standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/postgres/volume-expansion/Overview/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/volume-expansion/standalone/yamls](/docs/guides/postgres/volume-expansion/standalone/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Expand Volume of Standalone Database

Here, we are going to deploy a `Postgres` standalone using a supported version by `KubeDB` operator. Then we are going to apply `PostgresOpsRequest` to expand its volume.

### Prepare Postgres Standalone Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER               RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
linode-block-storage  linodebs.csi.linode.com   Delete          Immediate           true                   13m
```

We can see the output from the `linode-block-storage` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Postgres` standalone database with version `13.13`.

#### Deploy Postgres standalone

In this section, we are going to deploy a Postgres standalone database with 10GB volume. Then, in the next section we will expand its volume to 12GB using `PostgresOpsRequest` CRD. Below is the YAML of the `Postgres` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: pg-standalone
  namespace: demo
spec:
  version: "13.13"
  replicas: 1
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: "linode-block-storage"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Let's create the `Postgres` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/volume-expansion/standalone/yamls/pg-standalone.yaml
postgres.kubedb.com/pg-standalone created
```

Now, wait until `pg-standalone` has status `Ready`. i.e,

```bash
$ kubectl get pg -n demo
NAME            VERSION    STATUS    AGE
pg-standalone   13.13      Ready     3m47s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo pg-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"10Gi"

$ kubectl get pv -n demo
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                       STORAGECLASS          REASON    AGE
pvc-7a8a538d017a4f32   10Gi       RWO            Delete           Bound    demo/data-pg-standalone-0   linode-block-storage  <unset>   7m
```

You can see the petset has 10GB storage, and the capacity of the persistent volume is also 10GB.

We are now ready to apply the `PostgresOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the standalone database.

#### Create PostgresOpsRequest

In order to expand the volume of the database, we have to create a `PostgresOpsRequest` CR with our desired volume size. Below is the YAML of the `PostgresOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pgops-vol-exp
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: pg-standalone
  type: VolumeExpansion
  volumeExpansion:
    mode: Online
    postgres: 12Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `pg-stanalone` Postgres database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.postgres` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode(`Online` or `Offline`) 
 
> Note: If the Storageclass doesn’t support `Online` volume expansion, Try offline volume expansion by using spec.volumeExpansion.mode:"Offline".

During `Online` VolumeExpansion KubeDB expands volume without pausing database object, it directly updates the underlying PVC. And for `Offline` volume expansion, the database is paused. The Pods are deleted and PVC is updated. Then the database Pods are recreated with updated PVC.

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/volume-expansion/standalone/yamls/vol-exp-standalone.yaml
postgresopsrequest.ops.kubedb.com/pgops-vol-exp created
```

#### Verify Postgres Standalone volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Postgres` object and related `PetSet` and `Persistent Volume`.

Let's wait for `PostgresOpsRequest` to be `Successful`. Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ kubectl get postgresopsrequest -n demo
NAME            TYPE              STATUS       AGE
pgops-vol-exp   VolumeExpansion   Successful   10m
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe postgresopsrequest pgops-vol-exp -n demo
Name:         pgops-vol-exp
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
Metadata:
  Creation Timestamp:  2024-03-14T09:04:06Z
  Generation:          1
  Resource Version:    8621
  UID:                 54256467-7bc1-42f5-b0e4-4bf64337b9a0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pg-standalone
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:      Online
    Postgres:  12Gi
Status:
  Conditions:
    Last Transition Time:  2024-03-14T09:04:19Z
    Message:               Postgres ops request is expanding volume of database
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2024-03-14T09:05:12Z
    Message:               Online Volume Expansion performed successfully in Postgres pods for PostgresDBOpsRequest: demo/pgops-vol-exp
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-03-14T09:06:08Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-03-14T09:06:52Z
    Message:               Successfully Expanded Volume.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason             Age   From                         Message
  ----    ------             ----  ----                         -------
  Normal  PauseDatabase      12m   KubeDB Ops-manager Operator  Pausing Postgres demo/pg-standalone
  Normal  PauseDatabase      12m   KubeDB Ops-manager Operator  Successfully paused Postgres demo/pg-standalone
  Normal  VolumeExpansion    11m   KubeDB Ops-manager Operator  Online Volume Expansion performed successfully in Postgres pods for PostgresDBOpsRequest: demo/pgops-vol-exp
  Normal  ResumeDatabase     11m   KubeDB Ops-manager Operator  Resuming PostgreSQL demo/pg-standalone
  Normal  ResumeDatabase     11m   KubeDB Ops-manager Operator  Successfully resumed PostgreSQL demo/pg-standalone
  Normal  PauseDatabase      11m   KubeDB Ops-manager Operator  Pausing Postgres demo/pg-standalone
  Normal  PauseDatabase      11m   KubeDB Ops-manager Operator  Successfully paused Postgres demo/pg-standalone
  Normal  ReadyPetSets  10m   KubeDB Ops-manager Operator  PetSet is recreated
  Normal  Successful         10m   KubeDB Ops-manager Operator  Successfully Expanded Volume
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the standalone database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo pg-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"12Gi"

$ kubectl get pv -n demo
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                       STORAGECLASS           REASON   AGE
pvc-7a8a538d017a4f32   12Gi       RWO            Delete           Bound    demo/data-pg-standalone-0   linode-block-storage   <unset>  3m8s
```

The above output verifies that we have successfully expanded the volume of the Postgres standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete pg -n demo pg-standalone
postgres.kubedb.com "pg-standalone" deleted

$ kubectl delete postgresopsrequest -n demo pgops-vol-exp
postgresopsrequest.ops.kubedb.com "pgops-vol-exp" deleted
```
