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
  internalAuth:
    endpointCert:
      issuerRef:
        apiGroup: cert-manager.io
        name: mssqlserver-ca-issuer
        kind: Issuer
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
mssqlserver-ag-cluster   2022-cu12   Ready    89m
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo mssqlserver-ag-cluster -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-1497dd6d-9cbd-467a-8e0c-c3963ce09e1b   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-1   longhorn       <unset>                          88m
pvc-37a7bc8d-2c04-4eb4-8e53-e610fd1daaf5   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-0   longhorn       <unset>                          88m
pvc-817866af-5277-4d51-8d81-434e8ec1c442   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-2   longhorn       <unset>                          88m
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
  Creation Timestamp:  2024-10-28T12:27:31Z
  Generation:          1
  Resource Version:    494502
  UID:                 834cf122-b68c-41a5-924f-8b7fc09f4ed5
Spec:
  Database Ref:
    Name:  mssqlserver-ag-cluster
  Storage:
    Mssqlserver:
      Expansion Mode:     Offline
      Scaling Threshold:  50
      Trigger:            On
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
NAME                         TYPE              STATUS        AGE
msops-mssqlserver-ag-cluster-xojkua   VolumeExpansion   Progressing   15s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get mssqlserveropsrequest -n demo
NAME                         TYPE              STATUS       AGE
msops-mssqlserver-ag-cluster-xojkua   VolumeExpansion   Successful   97s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-mssqlserver-ag-cluster-xojkua
Name:         msops-mssqlserver-ag-cluster-xojkua
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mssqlserver-ag-cluster
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mssqlservers.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2022-01-14T06:13:10Z
  Generation:          1
  Managed Fields: ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MSSQLServerAutoscaler
    Name:                  ms-as-storage
    UID:                   4f45a3b3-fc72-4d04-b52c-a770944311f6
  Resource Version:        25557
  UID:                     90763a49-a03f-407c-a233-fb20c4ab57d7
Spec:
  Database Ref:
    Name:  mssqlserver-ag-cluster
  Type:    VolumeExpansion
  Volume Expansion:
    Mariadb:  1594884096
Status:
  Conditions:
    Last Transition Time:  2022-01-14T06:13:10Z
    Message:               Controller has started to Progress the MSSQLServerOpsRequest: demo/mops-mssqlserver-ag-cluster-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Volume Expansion performed successfully in MSSQLServer pod for MSSQLServerOpsRequest: demo/mops-mssqlserver-ag-cluster-xojkua
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Controller has successfully expand the volume of MSSQLServer demo/mops-mssqlserver-ag-cluster-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Start processing for MSSQLServerOpsRequest: demo/mops-mssqlserver-ag-cluster-xojkua
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Pausing MSSQLServer databse: demo/mssqlserver-ag-cluster
  Normal  Successful  2m58s  KubeDB Enterprise Operator  Successfully paused MSSQLServer database: demo/mssqlserver-ag-cluster for MSSQLServerOpsRequest: mops-mssqlserver-ag-cluster-xojkua
  Normal  Successful  103s   KubeDB Enterprise Operator  Volume Expansion performed successfully in MSSQLServer pod for MSSQLServerOpsRequest: demo/mops-mssqlserver-ag-cluster-xojkua
  Normal  Starting    103s   KubeDB Enterprise Operator  Updating MSSQLServer storage
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully Updated MSSQLServer storage
  Normal  Starting    103s   KubeDB Enterprise Operator  Resuming MSSQLServer database: demo/mssqlserver-ag-cluster
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully resumed MSSQLServer database: demo/mssqlserver-ag-cluster
  Normal  Successful  103s   KubeDB Enterprise Operator  Controller has Successfully expand the volume of MSSQLServer: demo/mssqlserver-ag-cluster
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo mssqlserver-ag-cluster -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   2Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-2   topolvm-provisioner            23m
pvc-4a509b05-774b-42d9-b36d-599c9056af37   2Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-0   topolvm-provisioner            24m
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   2Gi        RWO            Delete           Bound    demo/data-mssqlserver-ag-cluster-1   topolvm-provisioner            23m
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
