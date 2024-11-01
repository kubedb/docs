---
title: Vertical Scaling MSSQLServer
menu:
  docs_{{ .version }}:
    identifier: ms-scaling-vertical-ag-cluster
    name: Availability Group (HA Cluster)
    parent: ms-scaling-vertical
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale SQL Server Availability Group (HA Cluster)

This guide will show you how to use `kubeDB-Ops-Manager` to update the resources of a SQL Server Availability Group Cluster.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).


- You should be familiar with the following `KubeDB` concepts:
    - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
    - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/mssqlserver/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mssqlserver/scaling/vertical-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/scaling/vertical-scaling) directory of [kubedb/doc](https://github.com/kubedb/docs) repository.

### Apply Vertical Scaling on MSSQLServer Availability Group Cluster

Here, we are going to deploy a `MSSQLServer` instance using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

**Find supported MSSQLServer Version:**

When you have installed `KubeDB`, it has created `MSSQLServerVersion` CR for all supported `MSSQLServer` versions. Let's check the supported MSSQLServer versions,

```bash
$ kubectl get mssqlserverversion
NAME        VERSION   DB_IMAGE                                                DEPRECATED   AGE
2022-cu12   2022      mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04                3d21h
2022-cu14   2022      mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04                3d21h
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `MSSQLServer`. You can use any non-deprecated version. Here, we are going to create a mssqlserver using non-deprecated `MSSQLServer` version `2022-cu12`.


At first, we need to create an Issuer/ClusterIssuer which will be used to generate the certificate used for TLS configurations.

#### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
-
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

**Deploy MSSQLServer Availability Group Cluster:**

In this section, we are going to deploy a MSSQLServer instance. Then, in the next section, we will update the resources of the database server using vertical scaling.
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
              cpu: 1
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/scaling/vertical-scaling/mssql-ag-cluster.yaml
mssqlserver.kubedb.com/mssql-ag-cluster created
```


**Check mssqlserver Ready to Scale:**

`KubeDB` watches for `MSSQLServer` objects using Kubernetes API. When a `MSSQLServer` object is created, `KubeDB` will create a new PetSet, Services, and Secrets, etc.
Now, watch `MSSQLServer` is going to be in `Running` state and also watch `PetSet` and its pod is created and going to be in `Running` state,


```bash
$ watch kubectl get ms,petset,pods -n demo
Every 2.0s: kubectl get ms,petset,pods -n demo

NAME                                      VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-ag-cluster   2022-cu12   Ready    4m40s

NAME                                            AGE
petset.apps.k8s.appscode.com/mssql-ag-cluster   3m57s

NAME                     READY   STATUS    RESTARTS   AGE
pod/mssql-ag-cluster-0   2/2     Running   0          3m57s
pod/mssql-ag-cluster-1   2/2     Running   0          3m51s
pod/mssql-ag-cluster-2   2/2     Running   0          3m46s
```

Let's check pod's `mssql` container's resources, `mssql` container is the first container So it's index will be 0.

```bash
$ kubectl get pod -n demo mssql-ag-cluster-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
$ kubectl get pod -n demo mssql-ag-cluster-1 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
} 
$ kubectl get pod -n demo mssql-ag-cluster-2 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
```

Now, We are ready to apply a vertical scale on this mssqlserver database.

#### Vertical Scaling

Here, we are going to update the resources of the mssqlserver to meet up with the desired resources after scaling.

**Create MSSQLServerOpsRequest:**

In order to update the resources of your database, you have to create a `MSSQLServerOpsRequest` CR with your desired resources for scaling. Below is the YAML of the `MSSQLServerOpsRequest` CR that we are going to create,


```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: mops-vscale-ag-cluster
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mssql-ag-cluster
  verticalScaling:
    mssqlserver:
      resources:
        requests:
          memory: "1.7Gi"
          cpu: "700m"
        limits:
          cpu: 2
          memory: "4Gi"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `mssql-ag-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mssqlserver` specifies the expected `mssql` container resources after scaling.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/scaling/vertical-scaling/mops-vscale-ag-cluster.yaml
mssqlserveropsrequest.ops.kubedb.com/mops-vscale-ag-cluster created
```

**Verify MSSQLServer resources updated successfully:**

If everything goes well, `KubeDB-Ops-Manager` will update the resources of the PetSet's `Pod` containers. After a successful scaling process is done, the `KubeDB-Ops-Manager` updates the resources of the `MSSQLServer` object.

First, we will wait for `MSSQLServerOpsRequest` to be successful. Run the following command to watch `MSSQLServerOpsRequest` CR,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo mops-vscale-ag-cluster
Every 2.0s: kubectl get mssqlserveropsrequest -n demo mops-vscale-ag-cluster

NAME                     TYPE              STATUS       AGE
mops-vscale-ag-cluster   VerticalScaling   Successful   7m17s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest`, we will see that the mssqlserver resources are updated.

```bash
$ kubectl describe mssqlserveropsrequest -n demo mops-vscale-ag-cluster
Name:         mops-vscale-ag-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-10-24T14:13:05Z
  Generation:          1
  Resource Version:    747632
  UID:                 ed3c5cbc-e74e-46ba-b243-143a6007ac36
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mssql-ag-cluster
  Type:    VerticalScaling
  Vertical Scaling:
    Mssqlserver:
      Resources:
        Limits:
          Cpu:     2
          Memory:  4Gi
        Requests:
          Cpu:     700m
          Memory:  1.7Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-24T14:13:05Z
    Message:               MSSQLServer ops-request has started to vertically scaling the MSSQLServer nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-24T14:13:08Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-24T14:13:08Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-24T14:13:13Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-0
    Last Transition Time:  2024-10-24T14:13:13Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-0
    Last Transition Time:  2024-10-24T14:13:48Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-0
    Last Transition Time:  2024-10-24T14:13:53Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-1
    Last Transition Time:  2024-10-24T14:13:53Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-1
    Last Transition Time:  2024-10-24T14:14:28Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-1
    Last Transition Time:  2024-10-24T14:14:33Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T14:14:33Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T14:15:08Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-2
    Last Transition Time:  2024-10-24T14:15:13Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-24T14:15:13Z
    Message:               Successfully completed the VerticalScaling for MSSQLServer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age    From                         Message
  ----     ------                                                                ----   ----                         -------
  Normal   Starting                                                              7m46s  KubeDB Ops-manager Operator  Start processing for MSSQLServerOpsRequest: demo/mops-vscale-ag-cluster
  Normal   Starting                                                              7m46s  KubeDB Ops-manager Operator  Pausing MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                            7m46s  KubeDB Ops-manager Operator  Successfully paused MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: mops-vscale-ag-cluster
  Normal   UpdatePetSets                                                         7m43s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0             7m38s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0           7m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-0  7m33s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-0
  Warning  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0   7m3s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1             6m58s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
  Warning  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1           6m58s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
  Warning  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-1  6m53s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-1
  Warning  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1   6m23s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
  Warning  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2             6m18s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2           6m18s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Warning  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-2  6m13s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:mssql-ag-cluster-2
  Warning  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2   5m43s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
  Normal   RestartPods                                                           5m38s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                              5m38s  KubeDB Ops-manager Operator  Resuming MSSQLServer database: demo/mssql-ag-cluster
  Normal   Successful                                                            5m38s  KubeDB Ops-manager Operator  Successfully resumed MSSQLServer database: demo/mssql-ag-cluster for MSSQLServerOpsRequest: mops-vscale-ag-cluster
```

Now, we are going to verify whether the resources of the mssqlserver instance has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo mssql-ag-cluster-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "2",
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "1825361100800m"
  }
}
$ kubectl get pod -n demo mssql-ag-cluster-1 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "2",
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "1825361100800m"
  }
}
$ kubectl get pod -n demo mssql-ag-cluster-2 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "2",
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "1825361100800m"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the MSSQLServer.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mssqlserver -n demo mssql-ag-cluster
kubectl delete mssqlserveropsrequest -n demo mops-vscale-ag-cluster
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```



