---
title: Standalone Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: volume-expansion-guide-standalone
    name: Standalone
    parent: ms-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServer Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a SQL Server Standalone.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
    - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
    - [Volume Expansion Overview](/docs/guides/mssqlserver/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Expand Volume of MSSQLServer

Here, we are going to deploy a  `MSSQLServer` standalone using a supported version by `KubeDB` operator. Then we are going to apply `MSSQLServerOpsRequest` to expand its volume. The process of expanding MSSQLServer `standalone` is same as MSSQLServer Availability Group cluster.

### Prepare MSSQLServer Standalone

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  2d
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   3m25s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   3m19s
```

We can see from the output that `longhorn (default)` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We will use this storage class.


Now, we are going to deploy a `MSSQLServer` in `AvailabilityGroup` Mode with version `2022-cu12`.

### Deploy MSSQLServer Standalone

First, an issuer needs to be created, even if TLS is not enabled for SQL Server. The issuer will be used to configure the TLS-enabled Wal-G proxy server, which is required for the SQL Server backup and restore operations.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
- Create a secret using the certificate files we have just generated,
```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```
Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/standalone/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

In this section, we are going to deploy a MSSQLServer Standalone with 1GB volume. Then, in the next section we will expand its volume to 2GB using `MSSQLServerOpsRequest` CRD. Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-standalone
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation
          resources:
            requests:
              cpu: "500m"
              memory: "1.5Gi"
            limits:
              memory: "2Gi"
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MSSQLServer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/volume-expansion/mssql-standalone.yaml
mssqlserver.kubedb.com/mssql-standalone created
```

Now, wait until `mssql-standalone` has status `Ready`. i.e,

```bash
$ kubectl get ms -n demo mssql-standalone
NAME               VERSION     STATUS   AGE
mssql-standalone   2022-cu12   Ready    5m
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo mssql-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-7e7ed996-b682-4d84-8450-4c06fe92b11f   1Gi        RWO            Delete           Bound    demo/data-mssql-standalone-0   longhorn       <unset>                          5m29s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `MSSQLServerOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the MSSQLServer Standalone.

#### Create MSSQLServerOpsRequest

In order to expand the volume of the database, we have to create a `MSSQLServerOpsRequest` CR with our desired volume size. Below is the YAML of the `MSSQLServerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: mops-volume-exp-std
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: mssql-standalone
  volumeExpansion:
    mode: "Offline" # Online
    mssqlserver: 2Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `mssql-standalone` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.mssqlserver` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`). Storageclass `longhorn` supports `Offline` volume expansion.

> **Note:** If the Storageclass you are using support `Online` Volume Expansion, Try Online volume expansion by using `spec.volumeExpansion.mode:"Online"`.

During `Online` VolumeExpansion KubeDB expands volume without deleting the pods, it directly updates the underlying PVC. And for Offline volume expansion, the database is paused. The Pods are deleted and PVC is updated. Then the database Pods are recreated with updated PVC.


Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/example/mssqlserver/volume-expansion/mops-volume-exp-std.yaml
mssqlserveropsrequest.ops.kubedb.com/mops-volume-exp-std created
```

#### Verify MSSQLServer volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `MSSQLServer` object and related `PetSet` and `Persistent Volumes`.

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CR,

```bash
$ kubectl get mssqlserveropsrequest -n demo
NAME                  TYPE              STATUS       AGE
mops-volume-exp-std   VolumeExpansion   Successful   9m
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe msops -n demo mops-volume-exp-std
Name:         mops-volume-exp-std
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-13T05:13:02Z
  Generation:          1
  Resource Version:    2235347
  UID:                 e9d3733d-f472-406f-ade4-c92eea771d61
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mssql-standalone
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:         Offline
    Mssqlserver:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-11-13T05:13:02Z
    Message:               MSSQLServer ops-request has started to expand volume of mssqlserver nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-11-13T05:13:40Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2024-11-13T05:13:40Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2024-11-13T05:13:50Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-11-13T05:19:00Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-11-13T05:13:55Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2024-11-13T05:13:55Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-11-13T05:19:00Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-11-13T05:19:00Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2024-11-13T05:19:20Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-11-13T05:19:20Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-11-13T05:19:25Z
    Message:               running mssql server; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningMssqlServer
    Last Transition Time:  2024-11-13T05:19:45Z
    Message:               successfully updated node PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2024-11-13T05:19:50Z
    Message:               successfully reconciled the MSSQLServer resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-13T05:19:55Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-11-13T05:19:55Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-11-13T05:19:55Z
    Message:               Successfully completed volumeExpansion for MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo mssql-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-7e7ed996-b682-4d84-8450-4c06fe92b11f   2Gi        RWO            Delete           Bound    demo/data-mssql-standalone-0   longhorn       <unset>                          26m
```

The above output verifies that we have successfully expanded the volume of the MSSQLServer Standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:


```bash
$ kubectl patch -n demo ms/mssql-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
mssqlserver.kubedb.com/mssql-standalone patched

$ kubectl delete -n demo mssqlserver mssql-standalone
mssqlserver.kubedb.com "mssql-standalone" deleted

$ kubectl delete -n demo mssqlserveropsrequest mops-volume-exp-std
mssqlserveropsrequest.ops.kubedb.com "mops-volume-exp-std" deleted

kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- [Backup and Restore](/docs/guides/mssqlserver/backup/overview/index.md) MSSQLServer databases using KubeStash.