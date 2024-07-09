---
title: Postgres HA Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: pg-volume-expansion-ha
    name: HA
    parent: pg-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Postgres HA Cluster Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Postgres HA cluster.

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/volume-expansion/ha-cluster/yamls](/docs/guides/postgres/volume-expansion/ha-cluster/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Expand Volume of HA Cluster

Here, we are going to deploy a `Postgres` High Availability cluster using a supported version by `KubeDB` operator. Then we are going to apply `PostgresOpsRequest` to expand its volume.

### Prepare Postgres HA Cluster Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER               RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
linode-block-storage  linodebs.csi.linode.com   Delete          Immediate           true                   5m
```

We can see the output from the `linode-block-storage` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Postgres` HA cluster database with version `13.13`.

#### Deploy Postgres HA Cluster

In this section, we are going to deploy a Postgres HA database with 10GB volume. Then, in the next section we will expand its volume to 12GB using `PostgresOpsRequest` CRD. Below is the YAML of the `Postgres` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-ha-cluster
  namespace: demo
spec:
  version: "13.13"
  replicas: 3
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/volume-expansion/ha-cluster/yamls/pg-ha-cluster.yaml
postgres.kubedb.com/pg-ha-cluster created
```

Now, wait until `pg-ha-cluster` has status `Ready`. i.e,

```bash
$ kubectl get pg -n demo
NAME            VERSION   STATUS   AGE
pg-ha-cluster   13.13     Ready    3m6s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo pg-ha-cluster -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"10Gi"

$ kubectl get pv -n demo
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS           REASON   AGE
pvc-037525b1de294233   10Gi       RWO            Delete           Bound    demo/data-pg-ha-cluster-0           linode-block-storage            4m24s
pvc-3bd05d8b36c84c0a   10Gi       RWO            Delete           Bound    demo/data-pg-ha-cluster-1           linode-block-storage            3m2s
pvc-f03277c318c44029   10Gi       RWO            Delete           Bound    demo/data-pg-ha-cluster-2           linode-block-storage            3m35s
```

You can see the petset has 10GB storage, and the capacity of the persistent volume is also 10GB.

We are now ready to apply the `PostgresOpsRequest` CR to expand the volume of this HA Cluster. 

### Volume Expansion

Here, we are going to expand the volume of the Postgres database.

#### Create PostgresOpsRequest

In order to expand the volume of the database, we have to create a `PostgresOpsRequest` CR with our desired volume size for HA cluster. Below is the YAML of the `PostgresOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pgops-vol-exp-ha-cluster
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: pg-ha-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Online
    postgres: 12Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `pg-ha-cluster` Postgres database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.postgres` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode(only `Online`). 

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/volume-expansion/ha-cluster/yamls/vol-exp-ha-cluster.yaml
postgresopsrequest.ops.kubedb.com/pgops-vol-exp-ha-cluster created
```

#### Verify Postgres HA Cluster volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Postgres` object and related `PetSet` and `Persistent Volume`.

Let's wait for `PostgresOpsRequest` to be `Successful`. Run the following command to watch `PostgresOpsRequest` CR,

```bash
$ kubectl get postgresopsrequest -n demo
NAME                       TYPE              STATUS       AGE
pgops-vol-exp-ha-cluster   VolumeExpansion   Successful   105s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe postgresopsrequest pgops-vol-exp-ha-cluster -n demo
Name:         pgops-vol-exp-ha-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
Metadata:
  Creation Timestamp:  2024-03-15T05:12:53Z
  Generation:          1
  Resource Version:    73874
  UID:                 4388eacc-4bf6-4ca4-90a2-cb8b4293b9a5
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pg-ha-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Postgres:  12Gi
    Mode:     Online
Status:
  Conditions:
    Last Transition Time:  2024-03-15T05:12:53Z
    Message:               Postgres ops request is expanding volume of database
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2024-03-15T05:13:11Z
    Message:               Online Volume Expansion performed successfully in Postgres pods for PostgresDBOpsRequest: demo/pgops-vol-exp-ha-cluster
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-03-15T05:13:16Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-03-15T05:13:17Z
    Message:               Successfully Expanded Volume.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                    Age    From                         Message
  ----    ------                    ----   ----                         -------
  Normal  PauseDatabase             2m26s  KubeDB Ops-manager Operator  Pausing Postgres demo/pg-ha-cluster
  Normal  PauseDatabase             2m26s  KubeDB Ops-manager Operator  Successfully paused Postgres demo/pg-ha-cluster
  Normal  VolumeExpansion           2m8s   KubeDB Ops-manager Operator  Online Volume Expansion performed successfully in Postgres pods for PostgresDBOpsRequest: demo/pgops-vol-exp-ha-cluster
  Normal  ResumeDatabase            2m8s   KubeDB Ops-manager Operator  Resuming PostgreSQL demo/pg-ha-cluster
  Normal  ResumeDatabase            2m8s   KubeDB Ops-manager Operator  Successfully resumed PostgreSQL demo/pg-ha-cluster
  Normal  PauseDatabase             2m8s   KubeDB Ops-manager Operator  Pausing Postgres demo/pg-ha-cluster
  Normal  PauseDatabase             2m8s   KubeDB Ops-manager Operator  Successfully paused Postgres demo/pg-ha-cluster
  Normal  ReadyPetSets         2m3s   KubeDB Ops-manager Operator  PetSet is recreated
  Normal  ResumeDatabase            2m3s   KubeDB Ops-manager Operator  Resuming PostgreSQL demo/pg-ha-cluster
  Normal  ResumeDatabase            2m3s   KubeDB Ops-manager Operator  Successfully resumed PostgreSQL demo/pg-ha-cluster
  Normal  Successful                2m2s   KubeDB Ops-manager Operator  Successfully Expanded Volume
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the `pg-ha-cluster` has expanded to meet the desired state, Let's check that particular petset,

```bash
$ kubectl get sts -n demo pg-ha-cluster -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"12Gi"

$ kubectl get pv -n demo
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS           REASON   AGE
pvc-037525b1de294233   10Gi       RWO            Delete           Bound    demo/data-pg-ha-cluster-0           linode-block-storage            16m
pvc-3bd05d8b36c84c0a   12Gi       RWO            Delete           Bound    demo/data-pg-ha-cluster-1           linode-block-storage            14m
pvc-f03277c318c44029   10Gi       RWO            Delete           Bound    demo/data-pg-ha-cluster-2           linode-block-storage            15m
```

The above output verifies that we have successfully expanded the volume of the Postgres HA cluster database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete pg -n demo pg-ha-cluster
postgres.kubedb.com "pg-ha-cluster" deleted

$ kubectl delete postgresopsrequest -n demo pgops-vol-exp-ha-cluster
postgresopsrequest.ops.kubedb.com "pgops-vol-exp-ha-cluster" deleted
```
