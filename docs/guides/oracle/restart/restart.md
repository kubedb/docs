---
title: Restart Oracle
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-restart-details
    name: Restart Oracle
    parent: guides-oracle-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Oracle

KubeDB supports restarting the Oracle database via an `OracleOpsRequest`. Restarting is useful if some pods are stuck in some phase, or you want to apply some changes (e.g. mounted secrets/configs) by recreating the pods. This tutorial will show you how to restart an Oracle database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/restart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/restart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. Create an Oracle Container Registry token, if you haven't created one already, by following the instructions in the guide below: [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important) Make sure the `orclcred` secret exists in the `demo` namespace before deploying.

## Deploy Oracle

In this section, we are going to deploy an Oracle standalone database using KubeDB. Below is the YAML of the `Oracle` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sa-sample
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/restart/standalone-minimal.yaml
```
oracle.kubedb.com/oracle-sa-sample created

Now, wait until `oracle-sa-sample` has status `Ready`. i.e,

```bash
kubectl get oracle -n demo
```
NAME               VERSION   MODE         STATUS   AGE
oracle-sa-sample   21.3.0    Standalone   Ready    8m49s

## Apply Restart opsRequest

In order to restart the database, we have to create an `OracleOpsRequest` CR. Below is the YAML of the `OracleOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: standalone-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: oracle-sa-sample
  timeout: 30m
```

Here,

- `spec.type` specifies the type of the OpsRequest. In this case, it is `Restart` to restart the database.
- `spec.databaseRef.name` refers to the `Oracle` database `oracle-sa-sample` in the `demo` namespace.
- `spec.timeout` is the time the operator waits for each restart step to complete.

Let's create the `OracleOpsRequest` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/restart/standalone-restart.yaml
```
oracleopsrequest.ops.kubedb.com/standalone-restart created

Now the Ops-manager operator will restart the pods one by one (in a reconciliation-safe rolling manner). Let's wait until the `OracleOpsRequest` becomes `Successful`,

```bash
kubectl get oracleopsrequest -n demo
```
NAME                 TYPE      STATUS       AGE
standalone-restart   Restart   Successful   2m

We can see from the above output that the `OracleOpsRequest` has succeeded. Let's check the details with `kubectl describe`,

```bash
kubectl describe oracleopsrequest -n demo standalone-restart
```
Name:         standalone-restart
Namespace:    demo
...
Status:
  Conditions:
    Last Transition Time:  2026-06-22T18:53:29Z
    Message:               Oracle ops-request has started to restart Oracle nodes
    Observed Generation:   1
    Reason:                Restart
    Status:                True
    Type:                  Restart
    Last Transition Time:  2026-06-22T18:54:21Z
    Message:               Successfully Restarted Oracle nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-22T18:53:41Z
    Message:               get pod; ConditionStatus:True; PodName:oracle-sa-sample-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--oracle-sa-sample-0
    Last Transition Time:  2026-06-22T18:53:42Z
    Message:               evict pod; ConditionStatus:True; PodName:oracle-sa-sample-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--oracle-sa-sample-0
    Last Transition Time:  2026-06-22T18:54:16Z
    Message:               running pod; ConditionStatus:True; PodName:oracle-sa-sample-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--oracle-sa-sample-0
    Last Transition Time:  2026-06-22T18:54:16Z
    Message:               Pod oracle-sa-sample-0 restarted and healthy
    Observed Generation:   1
    Status:                True
    Type:                  RestartedPod--oracle-sa-sample-0
    Last Transition Time:  2026-06-22T18:54:22Z
    Message:               Controller has successfully restarted the Oracle replicas
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason       Age   From                         Message
  ----     ------       ----  ----                         -------
  Normal   Starting     66s   KubeDB Ops-manager Operator  Pausing Oracle database demo/oracle-sa-sample
  Normal   Successful   66s   KubeDB Ops-manager Operator  Successfully paused Oracle database: demo/oracle-sa-sample for OracleOpsRequest: standalone-restart
  Warning  get pod; ConditionStatus:True; PodName:oracle-sa-sample-0       56s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:oracle-sa-sample-0
  Warning  evict pod; ConditionStatus:True; PodName:oracle-sa-sample-0     55s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:oracle-sa-sample-0
  Warning  running pod; ConditionStatus:False; PodName:oracle-sa-sample-0  51s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False; PodName:oracle-sa-sample-0
  Warning  running pod; ConditionStatus:True; PodName:oracle-sa-sample-0   45s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:oracle-sa-sample-0

After the ops request succeeds, the database pod is freshly recreated. Oracle then re-opens the existing database; while it is opening, the `Oracle` object may briefly report the `Critical` phase before settling back to `Ready`,

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=oracle-sa-sample
```
NAME                 READY   STATUS    RESTARTS   AGE
oracle-sa-sample-0   1/1     Running   0          24s

```bash
kubectl get oracle -n demo oracle-sa-sample
```
NAME               VERSION   MODE         STATUS   AGE
oracle-sa-sample   21.3.0    Standalone   Ready    12m

## Restarting a DataGuard cluster

The same `OracleOpsRequest` works for a DataGuard cluster — just point `spec.databaseRef.name` at the DataGuard database. A DataGuard cluster (`mode: DataGuard`, `replicas: 3`) consists of 3 database pods (each running an `oracle` and an `oracle-coordinator` container) plus a single observer pod:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=oracle-dg-sample -L kubedb.com/role
```
NAME                          READY   STATUS    RESTARTS   AGE   ROLE
oracle-dg-sample-0            2/2     Running   0          18m   primary
oracle-dg-sample-1            2/2     Running   0          18m   standby
oracle-dg-sample-2            2/2     Running   0          18m   standby
oracle-dg-sample-observer-0   1/1     Running   0          18m

```bash
kubectl get svc -n demo -l app.kubernetes.io/instance=oracle-dg-sample
```
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
oracle-dg-sample           ClusterIP   10.43.232.240   <none>        1521/TCP   18m
oracle-dg-sample-pods      ClusterIP   None            <none>        1521/TCP   18m
oracle-dg-sample-standby   ClusterIP   10.43.50.35     <none>        1521/TCP   18m

Here, the `kubedb.com/role` label marks pod `oracle-dg-sample-0` as the `primary` and the other two as `standby`. The `oracle-dg-sample` service always routes to the live primary, while `oracle-dg-sample-standby` routes to the read-only standbys.

To restart the cluster, create the following `OracleOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: dataguard-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: oracle-dg-sample
  timeout: 30m
```

For a DataGuard cluster the operator restarts the pods one at a time (a reconciliation-safe rolling restart). The KubeDB coordinator keeps the `kubedb.com/role` label (`primary`/`standby`) in sync, and the observer drives Fast-Start Failover, so a standby is promoted if the primary pod is restarted — the primary service (`oracle-dg-sample`) always points at the live primary.

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/restart/dataguard-restart.yaml
```
oracleopsrequest.ops.kubedb.com/dataguard-restart created

```bash
kubectl get oracleopsrequest -n demo dataguard-restart
```
NAME                TYPE      STATUS       AGE
dataguard-restart   Restart   Successful   2m42s

The `kubectl describe` output shows the pods being evicted and restarted one at a time (the standby pods `oracle-dg-sample-1` and `oracle-dg-sample-2`, then the primary), each verified healthy before moving to the next,

```bash
kubectl describe oracleopsrequest -n demo dataguard-restart
```
...
Status:
  Conditions:
    Last Transition Time:  2026-06-22T20:58:51Z
    Message:               Oracle ops-request has started to restart Oracle nodes
    Reason:                Restart
    Status:                True
    Type:                  Restart
    Last Transition Time:  2026-06-22T20:59:38Z
    Message:               Pod oracle-dg-sample-1 restarted and healthy
    Status:                True
    Type:                  RestartedPod--oracle-dg-sample-1
    Last Transition Time:  2026-06-22T21:00:18Z
    Message:               Pod oracle-dg-sample-2 restarted and healthy
    Status:                True
    Type:                  RestartedPod--oracle-dg-sample-2
    Last Transition Time:  2026-06-22T21:01:13Z
    Message:               Successfully Restarted Oracle nodes
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-22T21:01:13Z
    Message:               Controller has successfully restarted the Oracle replicas
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Phase:                   Successful

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo standalone-restart dataguard-restart
kubectl patch -n demo oracle/oracle-sa-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo oracle-sa-sample
kubectl patch -n demo oracle/oracle-dg-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo oracle-dg-sample
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Monitor your Oracle database with KubeDB using [Prometheus operator](/docs/guides/oracle/monitoring/using-prometheus-operator.md).
- Learn how to [reconfigure](/docs/guides/oracle/reconfigure/reconfigure.md) an Oracle database.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

> ## ⚠️ Legal Notice
>
> Oracle® and Oracle Database® are registered trademarks of Oracle Corporation.
> KubeDB is not affiliated with, endorsed by, or sponsored by Oracle Corporation.
>
> KubeDB provides only orchestration and management tooling for Kubernetes.
> It does not distribute, bundle, ship, or include any Oracle Database software or binaries.
>
> Users must provide their own Oracle container images and hold valid Oracle licenses.
> Users are solely responsible for compliance with Oracle’s licensing terms, including all rules regarding containers, Docker, and Kubernetes environments.
>
> KubeDB makes no representations or warranties regarding Oracle licensing compliance.
