---
title: Updating MSSQLServer Availability Group
menu:
  docs_{{ .version }}:
    identifier: ms-updating-ag-cluster
    name: Availability Group (HA Cluster)
    parent: mssql-updating
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of MSSQLServer Availability Group

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `MSSQLServer` Availability Group Cluster.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [Availability Group Cluster](/docs/guides/mssqlserver/clustering/ag_cluster.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/mssqlserver/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mssqlserver](/docs/examples/mssqlserver/update-version) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare MSSQLServer Availability Group Database

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

Now, we are going to deploy a `MSSQLServer` availability group with version `2022-cu12`.

### Deploy MSSQLServer AG Cluster

In this section, we are going to deploy a MSSQLServer availability group cluster. Then, in the next section we will update the version of the database using `MSSQLServerOpsRequest` CR. 

Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-ag-cluster
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
              memory: "2Gi"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MSSQLServer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/update-version/mssql-ag-cluster.yaml
mssqlserver.kubedb.com/mssql-ag-cluster created
```

Now, wait until `mssql-ag-cluster` created has status `Ready`. i.e,

```bash
$ kubectl get ms -n demo     
NAME               VERSION     STATUS     AGE
mssql-ag-cluster   2022-cu12   Ready      4m
```

We are now ready to apply the `MSSQLServerOpsRequest` CR to update this database.

### Update MSSQLServer Version

Here, we are going to update `MSSQLServer` ag cluster from `2022-cu12` to `2022-cu14`.

#### Create MSSQLServerOpsRequest:

In order to update the version of the mssql server ag cluster database, we have to create a `MSSQLServerOpsRequest` CR with your desired version that is supported by `KubeDB`. 
Below is the YAML of the `MSSQLServerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-update-ag-cluster
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mssql-ag-cluster
  updateVersion:
    targetVersion: 2022-cu14
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mssql-ag-cluster` MSSQLServer database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `2022-cu14`.
- Have a look [here](/docs/guides/mssqlserver/concepts/opsrequest.md#spectimeout) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/update-version/msops-update-ag-cluster .yaml
mssqlserveropsrequest.ops.kubedb.com/msops-update-ag-cluster created
```

#### Verify MSSQLServer version updated successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the image of `MSSQLServer` object and related `PetSets` and `Pods`.

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CR,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo
NAME                      TYPE            STATUS       AGE
msops-update-ag-cluster   UpdateVersion   Successful   2m33s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-update-ag-cluster
Name:         msops-update-ag-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-12T13:38:04Z
  Generation:          1
  Resource Version:    333751
  UID:                 f4977424-0ddc-40bc-a3b7-bc69f5ac6e82
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mssql-ag-cluster
  Timeout:  5m
  Type:     UpdateVersion
  Update Version:
    Target Version:  2022-cu14
Status:
  Conditions:
    Last Transition Time:  2024-11-12T13:38:05Z
    Message:               MSSQLServer ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2024-11-12T13:38:28Z
    Message:               successfully reconciled the MSSQLServer with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-12T13:38:33Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-12T13:38:33Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-12T13:39:08Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-0
    Last Transition Time:  2024-11-12T13:39:13Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-12T13:39:13Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-12T13:39:48Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-1
    Last Transition Time:  2024-11-12T13:39:53Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-12T13:39:53Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-12T13:40:28Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-2
    Last Transition Time:  2024-11-12T13:40:33Z
    Message:               Successfully Restarted MSSQLServer pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-12T13:40:33Z
    Message:               Successfully updated MSSQLServer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-11-12T13:40:33Z
    Message:               Successfully updated MSSQLServer version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age    From                         Message
  ----     ------                                                                ----   ----                         -------
  Normal   Starting                                                              7m49s  KubeDB Ops-manager Operator  Start processing for MSSQLServerOpsRequest: demo/msops-update-ag-cluster
  Normal   Starting                                                              7m49s  KubeDB Ops-manager Operator  Pausing MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                            7m49s  KubeDB Ops-manager Operator  Successfully paused MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: msops-update-ag-cluster
  Normal   UpdatePetSets                                                         7m26s  KubeDB Ops-manager Operator  successfully reconciled the MSSQLServer with updated version
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0             7m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0           7m21s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-0  7m16s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-0
  Warning  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0   6m46s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1             6m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
  Warning  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1           6m41s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
  Warning  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-1  6m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-1
  Warning  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1   6m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2             6m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2           6m1s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-2  5m56s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-2
  Warning  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2   5m26s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Normal   RestartPods                                                           5m21s  KubeDB Ops-manager Operator  Successfully Restarted MSSQLServer pods
  Normal   Starting                                                              5m21s  KubeDB Ops-manager Operator  Resuming MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                            5m21s  KubeDB Ops-manager Operator  Successfully resumed MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: msops-update-ag-cluster
```

Now, we are going to verify whether the `MSSQLServer` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get ms -n demo mssql-ag-cluster -o=jsonpath='{.spec.version}{"\n"}'
2022-cu14

$ kubectl get petset -n demo mssql-ag-cluster -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04@sha256:c1aa8afe9b06eab64c9774a4802dcd032205d1be785b1fd51e1c0151e7586b74

$ kubectl get pods -n demo mssql-ag-cluster-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04
```

You can see from above, our `MSSQLServer` ag database cluster has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ms -n demo mssql-ag-cluster
kubectl delete msops -n demo msops-update-ag-cluster
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```