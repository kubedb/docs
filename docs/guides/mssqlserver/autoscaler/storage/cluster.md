---
title: MSSQLServer Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: ms-storage-autoscaling-cluster
    name: Cluster
    parent: ms-storage-autoscaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a MSSQLServer Availability Group Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a MSSQLServer Availability Group Cluster.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
  - [MSSQLServerAutoscaler](/docs/guides/mssqlserver/concepts/autoscaler.md)
  - [Storage Autoscaling Overview](/docs/guides/mssqlserver/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Storage Autoscaling MSSQLServer Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  4d21h
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   2d20h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   2d20h
```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `MSSQLServer` cluster using a supported version by `KubeDB` operator. Then we are going to apply `MSSQLServerAutoscaler` to set up autoscaling.

#### Deploy MSSQLServer Cluster

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/ag-cluster/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

Now, we are going to deploy a MSSQLServer cluster database with version `2022-cu12`. Then, in the next section we will set up autoscaling for this database using `MSSQLServerAutoscaler` CRD. Below is the YAML of the `MSSQLServer` CR that we are going to create,

> If you want to autoscale MSSQLServer `Standalone`, Just deploy a [standalone](/docs/guides/mssqlserver/clustering/standalone.md) sql server instance using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-ag-cluster
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
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
              cpu: "600m"
              memory: "1.6Gi"
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MSSQLServer` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/autoscaler/storage/mssqlserver-ag-cluster.yaml
mssqlserver.kubedb.com/mssqlserver-ag-cluster created
```

Now, wait until `mssqlserver-ag-cluster` has status `Ready`. i.e,

```bash
$ kubectl get mssqlserver -n demo
NAME                     VERSION     STATUS   AGE
mssqlserver-ag-cluster   2022-cu12   Ready    4m
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo mssqlserver-ag-cluster -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-1497dd6d-9cbd-467a-8e0c-c3963ce09e1b   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-1   longhorn       <unset>                          8m
pvc-37a7bc8d-2c04-4eb4-8e53-e610fd1daaf5   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-0   longhorn       <unset>                          8m
pvc-817866af-5277-4d51-8d81-434e8ec1c442   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-2   longhorn       <unset>                          8m
```

You can see the petset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `MSSQLServerAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a `MSSQLServerAutoscaler` Object.

#### Create MSSQLServerAutoscaler Object

In order to set up storage autoscaling for this database cluster, we have to create a `MSSQLServerAutoscaler` CRO with our desired configuration. Below is the YAML of the `MSSQLServerAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MSSQLServerAutoscaler
metadata:
  name: ms-as-storage
  namespace: demo
spec:
  databaseRef:
    name: mssqlserver-ag-cluster
  storage:
    mssqlserver:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
      expansionMode: "Offline"
      upperBound: "100Gi"
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `mssqlserver-ag-cluster` database.
- `spec.storage.mssqlserver.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.mssqlserver.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.mssqlserver.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- `spec.storage.mssqlserver.expansionMode` specifies the expansion mode of volume expansion `MSSQLServerOpsRequest` created by `MSSQLServerAutoscaler`, `longhorn` supports offline volume expansion so here `expansionMode` is set as "Offline".

Let's create the `MSSQLServerAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/autoscaler/storage/ms-as-storage.yaml
mssqlserverautoscaler.autoscaling.kubedb.com/ms-as-storage created
```

#### Storage Autoscaling is set up successfully

Let's check that the `mssqlserverautoscaler` resource is created successfully,

```bash
$ kubectl get mssqlserverautoscaler -n demo
NAME            AGE
ms-as-storage   17s


$ kubectl describe mssqlserverautoscaler ms-as-storage -n demo
Name:         ms-as-storage
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MSSQLServerAutoscaler
Metadata:
  Creation Timestamp:  2024-11-01T09:39:54Z
  Generation:          1
  Resource Version:    922388
  UID:                 1e239b31-c6c8-4e2c-8cf6-2b95a88b9d45
Spec:
  Database Ref:
    Name:  mssqlserver-ag-cluster
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Mssqlserver:
      Expansion Mode:  Offline
      Scaling Rules:
        Applies Upto:     
        Threshold:        50pc
      Scaling Threshold:  50
      Trigger:            On
      Upper Bound:        100Gi
      Usage Threshold:    60
Events:                   <none>
```

So, the `mssqlserverautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see storage autoscaling.

Lets exec into the database pod and fill the database volume(`/var/opt/mssql/`) using the following commands:

```bash
$ kubectl exec -it -n demo mssqlserver-ag-cluster-0 -c mssql -- bash
mssql@mssqlserver-ag-cluster-0:/$ df -h /var/opt/mssql
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-37a7bc8d-2c04-4eb4-8e53-e610fd1daaf5  974M  274M  685M  29% /var/opt/mssql

mssql@mssqlserver-ag-cluster-0:/$ dd if=/dev/zero of=/var/opt/mssql/file.img bs=120M count=5
5+0 records in
5+0 records out
629145600 bytes (629 MB, 600 MiB) copied, 6.09315 s, 103 MB/s
mssql@mssqlserver-ag-cluster-0:/$ df -h /var/opt/mssql
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-37a7bc8d-2c04-4eb4-8e53-e610fd1daaf5  974M  874M   85M  92% /var/opt/mssql
```

So, from the above output we can see that the storage usage is 92%, which exceeded the `usageThreshold` 60%.

Let's watch the `mssqlserveropsrequest` in the demo namespace to see if any `mssqlserveropsrequest` object is created. After some time you'll see that a `mssqlserveropsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.


```bash
$ watch kubectl get mssqlserveropsrequest -n demo
NAME                                  TYPE              STATUS        AGE
msops-mssqlserver-ag-cluster-8m7l5s   VolumeExpansion   Progressing   2m20s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get mssqlserveropsrequest -n demo
NAME                                  TYPE              STATUS       AGE
msops-mssqlserver-ag-cluster-8m7l5s   VolumeExpansion   Successful   17m
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-mssqlserver-ag-cluster-8m7l5s 
Name:         msops-mssqlserver-ag-cluster-8m7l5s
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mssqlserver-ag-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mssqlservers.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-01T09:40:05Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MSSQLServerAutoscaler
    Name:                  ms-as-storage
    UID:                   1e239b31-c6c8-4e2c-8cf6-2b95a88b9d45
  Resource Version:        924068
  UID:                     d0dfbe3d-4f0f-43ec-bdff-6d9f3fa96516
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mssqlserver-ag-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:         Offline
    Mssqlserver:  1531054080
Status:
  Conditions:
    Last Transition Time:  2024-11-01T09:40:05Z
    Message:               MSSQLServer ops-request has started to expand volume of mssqlserver nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-11-01T09:40:13Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2024-11-01T09:40:13Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2024-11-01T09:40:23Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-11-01T09:46:48Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-11-01T09:40:28Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2024-11-01T09:40:28Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-11-01T09:41:03Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-11-01T09:41:03Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2024-11-01T09:48:33Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-11-01T09:42:48Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-11-01T09:42:53Z
    Message:               running mssql server; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningMssqlServer
    Last Transition Time:  2024-11-01T09:48:58Z
    Message:               successfully updated node PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2024-11-01T09:49:03Z
    Message:               successfully reconciled the MSSQLServer resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-01T09:49:03Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-11-01T09:49:03Z
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
$ kubectl get petset -n demo mssqlserver-ag-cluster -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1531054080"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-2ff83356-1bbc-44ab-99f1-025e3690a471   1462Mi     RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-2   longhorn       <unset>                          15m
pvc-a5cc0ae9-2c8d-456c-ace2-fc4fafc6784f   1462Mi     RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-1   longhorn       <unset>                          16m
pvc-e8ab47a4-17a6-45fb-9f39-e71a03498ab5   1462Mi     RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-0   longhorn       <unset>                          16m
```

The above output verifies that we have successfully autoscaled the volume of the MSSQLServer cluster database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mssqlserver -n demo mssqlserver-ag-cluster
kubectl delete mssqlserverautoscaler -n demo ms-as-storage
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```
