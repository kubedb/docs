---
title: Restart MSSQLServer
menu:
  docs_{{ .version }}:
    identifier: ms-restart-guide
    name: Restart MSSQLServer
    parent: ms-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart MSSQLServer

KubeDB supports restarting the MSSQLServer via a MSSQLServerOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MSSQLServer

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

In this section, we are going to deploy a MSSQLServer database using KubeDB.

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
              value: Evaluation # Change it 
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/restart/mssqlserver-ag-cluster.yaml
mssqlserver.kubedb.com/mssqlserver-ag-cluster created
```

Check the database is provisioned successfully
```bash
$ kubectl get ms -n demo mssqlserver-ag-cluster
NAME                     VERSION     STATUS   AGE
mssqlserver-ag-cluster   2022-cu12   Ready    4m
```


## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: mssqlserver-ag-cluster
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the MSSQLServer database.  The db should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields can be found [here](/docs/guides/mssqlserver/concepts/opsrequest.md)

> Note: The method of restarting the standalone & cluster mode db is exactly same as above. All you need, is to specify the corresponding MSSQLServer name in `spec.databaseRef.name` section.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/restart/msops-restart.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-restart created
```

Now the Ops-manager operator will first restart the general secondary pods and lastly will restart the Primary pod of the database.

```shell
$ kubectl get msops -n demo msops-restart
NAME            TYPE      STATUS       AGE
msops-restart   Restart   Successful   5m23s

$ kubectl get msops -n demo msops-restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"MSSQLServerOpsRequest","metadata":{"annotations":{},"name":"msops-restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"mssqlserver-ag-cluster"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-10-25T06:58:21Z"
  generation: 1
  name: msops-restart
  namespace: demo
  resourceVersion: "771141"
  uid: 9e531521-c369-4ce4-983f-a3dafd90cb8a
spec:
  apply: Always
  databaseRef:
    name: mssqlserver-ag-cluster
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-10-25T06:58:21Z"
    message: MSSQLServerOpsRequest has started to restart MSSQLServer nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-10-25T06:58:45Z"
    message: get pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-0
    observedGeneration: 1
    status: "True"
    type: GetPod--mssqlserver-ag-cluster-0
  - lastTransitionTime: "2024-10-25T06:58:45Z"
    message: evict pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--mssqlserver-ag-cluster-0
  - lastTransitionTime: "2024-10-25T06:59:20Z"
    message: check pod running; ConditionStatus:True; PodName:mssqlserver-ag-cluster-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--mssqlserver-ag-cluster-0
  - lastTransitionTime: "2024-10-25T06:59:25Z"
    message: get pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-1
    observedGeneration: 1
    status: "True"
    type: GetPod--mssqlserver-ag-cluster-1
  - lastTransitionTime: "2024-10-25T06:59:25Z"
    message: evict pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--mssqlserver-ag-cluster-1
  - lastTransitionTime: "2024-10-25T07:00:00Z"
    message: check pod running; ConditionStatus:True; PodName:mssqlserver-ag-cluster-1
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--mssqlserver-ag-cluster-1
  - lastTransitionTime: "2024-10-25T07:00:05Z"
    message: get pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-2
    observedGeneration: 1
    status: "True"
    type: GetPod--mssqlserver-ag-cluster-2
  - lastTransitionTime: "2024-10-25T07:00:05Z"
    message: evict pod; ConditionStatus:True; PodName:mssqlserver-ag-cluster-2
    observedGeneration: 1
    status: "True"
    type: EvictPod--mssqlserver-ag-cluster-2
  - lastTransitionTime: "2024-10-25T07:00:40Z"
    message: check pod running; ConditionStatus:True; PodName:mssqlserver-ag-cluster-2
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--mssqlserver-ag-cluster-2
  - lastTransitionTime: "2024-10-25T07:00:45Z"
    message: Successfully restarted MSSQLServer nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-10-25T07:00:45Z"
    message: Controller has successfully restart the MSSQLServer replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

We can see that, the database is ready after restarting the pods  
```bash
$ kubectl get ms -n demo mssqlserver-ag-cluster
NAME                     VERSION     STATUS   AGE
mssqlserver-ag-cluster   2022-cu12   Ready    14m
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mssqlserveropsrequest -n demo msops-restart
kubectl delete mssqlserver -n demo mssqlserver-ag-cluster
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete secret -n demo mssqlserver-ca
kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/mssqlserver/backup/overview/index.md) MSSQLServer database using KubeStash.
- Want to set up MSSQLServer cluster? Check how to [Configure SQL Server Availability Group Cluster](/docs/guides/mssqlserver/clustering/ag_cluster.md)
- Detail concepts of [MSSQLServer Object](/docs/guides/mssqlserver/concepts/mssqlserver.md).

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
