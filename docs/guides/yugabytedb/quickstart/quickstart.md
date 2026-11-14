---
title: YugabyteDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-quickstart-overview
    name: Overview
    parent: yugabytedb-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running YugabyteDB

This tutorial shows how to run a replication factor 3 YugabyteDB universe on Kubernetes with KubeDB, spread across three zones, and connect to it with `ysqlsh`.

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs). Every command and manifest here was run on a kind cluster with one control-plane node and three worker nodes labelled `topology.kubernetes.io/zone=zone-a|zone-b|zone-c`.

## Before You Begin

You need a Kubernetes cluster, `kubectl` configured for it, and the KubeDB operator (with the YugabyteDB feature enabled) installed following the steps [here](/docs/setup/README.md). You also need:

- The kubeops.dev **PetSet** controller and its `PlacementPolicy` CRD. They are installed together with KubeDB.
- Nodes labelled with `topology.kubernetes.io/zone`. KubeDB reads the zone of each pod from its node. Check with:

```bash
$ kubectl get nodes -L topology.kubernetes.io/zone
NAME                   STATUS   ROLES           AGE   VERSION   ZONE
yb-dev-control-plane   Ready    control-plane   8h    v1.34.0
yb-dev-worker          Ready    <none>          8h    v1.34.0   zone-a
yb-dev-worker2         Ready    <none>          8h    v1.34.0   zone-b
yb-dev-worker3         Ready    <none>          8h    v1.34.0   zone-c
```

- [cert-manager](https://cert-manager.io/) is optional. It is only needed for [TLS](/docs/guides/yugabytedb/tls/overview/).

To keep things isolated, this tutorial uses a namespace called `demo`:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Find Available YugabyteDBVersion

The installer creates a `YugabyteDBVersion` for every supported version, so there is nothing to apply:

```bash
$ kubectl get ybversion
NAME       VERSION    DB_IMAGE                                        DEPRECATED   AGE
2025.1.4   2025.1.4   docker.io/yugabytedb/yugabyte:2025.1.4.0-b103   false        5h39m
2025.2.8   2025.2.8   docker.io/yugabytedb/yugabyte:2025.2.8.0-b60    false        5h39m
```

This tutorial uses `2025.2.8`.

## Create a PlacementPolicy

A `PlacementPolicy` tells the PetSet controller how to spread the pods of a role. This one puts at most a skew of one pod per zone, which gives one master and one tserver in each zone for a replication factor 3 universe:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: spread-across-zones
spec:
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: DoNotSchedule
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/quickstart/placementpolicy.yaml
placementpolicy.apps.k8s.appscode.com/spread-across-zones created
```

> Important: with `whenUnsatisfiable: DoNotSchedule` on the zone constraint, the PetSet controller also adds a required one-pod-per-zone anti-affinity. A role then cannot have more pods than zones, so scaling the tservers from 3 to 6 over three zones leaves the new pods `Pending`. If you plan to run more tservers than zones, use `whenUnsatisfiable: ScheduleAnyway` on the zone constraint. Existing pods keep their old affinity until they are recreated (for example by a [Restart](/docs/guides/yugabytedb/restart/restart.md)).

## Create a YugabyteDB

```yaml
apiVersion: kubedb.com/v1alpha2
kind: YugabyteDB
metadata:
  name: yb-demo
  namespace: demo
spec:
  version: 2025.2.8
  replicationFactor: 3
  topology:
    master:
      podTemplate:
        spec:
          podPlacementPolicy:
            name: spread-across-zones
          containers:
          - name: yb-master
            resources:
              requests: {cpu: 500m, memory: 1Gi}
              limits: {memory: 1Gi}
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 2Gi}}
    tserver:
      replicas: 3
      podTemplate:
        spec:
          podPlacementPolicy:
            name: spread-across-zones
          containers:
          - name: yb-tserver
            resources:
              requests: {cpu: "1", memory: 2Gi}
              limits: {memory: 2Gi}
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 5Gi}}
  deletionPolicy: WipeOut
```

Here,

- `spec.version` is the name of the `YugabyteDBVersion`.
- `spec.replicationFactor: 3` creates 3 masters and keeps 3 replicas of every tablet.
- `spec.topology.master` and `spec.topology.tserver` configure the two roles. KubeDB creates one PetSet for each, `yb-demo-master` and `yb-demo-tserver`. Both use the PlacementPolicy created above.
- The container `resources` set the CPU and memory. YugabyteDB derives `--memory_limit_hard_bytes` from the memory limit (85% for masters, 70% for tservers).
- `spec.topology.*.storage` is the PVC template of each pod. The `StorageClass` is the default one of the cluster. YugabyteDB rejects writes when a data disk has less than 5% free space, so leave headroom.
- `spec.deletionPolicy: WipeOut` removes everything, including PVCs and Secrets, when the object is deleted.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/quickstart/yugabytedb.yaml
yugabytedb.kubedb.com/yb-demo created

$ kubectl get yb -n demo yb-demo -w
NAME      VERSION    STATUS         AGE
yb-demo   2025.2.8   Provisioning   5s
yb-demo   2025.2.8   Ready          2m8s
```

In the test cluster the universe became `Ready` in 94 seconds (images already pulled) up to about two minutes. The first reconcile creates the PetSets, the Services, the `yb-demo-auth` and `yb-demo-config` Secrets and the AppBinding. If the status stays `Provisioning`, run `kubectl describe yb -n demo yb-demo` and check the conditions and events.

## Verify the universe

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=yb-demo -o custom-columns=NAME:.metadata.name,ZONE:.metadata.annotations.yugabytedb\\.kubedb\\.com/zone,ROLE:.metadata.labels.kubedb\\.com/role,NODE:.spec.nodeName
NAME                ZONE     ROLE      NODE
yb-demo-master-0    zone-b   master    yb-dev-worker2
yb-demo-master-1    zone-c   master    yb-dev-worker3
yb-demo-master-2    zone-a   master    yb-dev-worker
yb-demo-tserver-0   zone-b   tserver   yb-dev-worker2
yb-demo-tserver-1   zone-c   tserver   yb-dev-worker3
yb-demo-tserver-2   zone-a   tserver   yb-dev-worker
```

There is one master and one tserver in each zone. The conditions of a healthy universe:

```bash
$ kubectl get yb -n demo yb-demo -o jsonpath='{range .status.conditions[*]}{.type}={.status} {end}'
ProvisioningStarted=True ReplicaReady=True MastersReady=True TServersReady=True PlacementSet=True InternalUsersSynced=True Ready=True AcceptingConnection=True Provisioned=True
```

KubeDB also created the PVCs, two Services, the AppBinding and the Secrets. The Services:

```bash
$ kubectl get svc -n demo
NAME            TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                                                 AGE
yb-demo         ClusterIP   10.96.99.45   <none>        5433/TCP                                                2m8s
yb-demo-pods    ClusterIP   None          <none>        7100/TCP,7000/TCP,9100/TCP,9000/TCP,5433/TCP,9042/TCP   2m8s
```

- `yb-demo` is the client Service (YSQL, port 5433). It selects only tserver pods that carry the label `kubedb.com/auth-bootstrapped=true`, which KubeDB sets after the password has been changed. Until then the Service has no endpoints, so the default `yugabyte` password is never reachable through it.
- `yb-demo-pods` is the headless governing Service used for pod-to-pod addressing.

## Connect to YugabyteDB

The first reconcile creates the Secret `yb-demo-auth` with the user `yugabyte` and a random password:

```bash
$ kubectl get secret -n demo yb-demo-auth -o jsonpath='{.data.username}' | base64 -d
yugabyte
$ kubectl get secret -n demo yb-demo-auth -o jsonpath='{.data.password}' | base64 -d
<random password>
```

Run `ysqlsh` from a tserver pod, against the client Service:

```bash
$ PW=$(kubectl get secret -n demo yb-demo-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo yb-demo-tserver-0 -c yb-tserver -- env PGPASSWORD="$PW" ysqlsh -h yb-demo.demo.svc -U yugabyte -c "select version()" -c "create table t(id int primary key, v text)" -c "insert into t values (1,'hello')" -c "select * from t"
 PostgreSQL 15.12-YB-2025.2.8.0-b0 on x86_64-pc-linux-gnu, compiled by clang version 19.1.0 ... 64-bit
(1 row)

CREATE TABLE
INSERT 0 1
 id |   v
----+-------
  1 | hello
(1 row)
```

To use an interactive shell, use `kubectl exec -it ... -- env PGPASSWORD="$PW" ysqlsh -h yb-demo.demo.svc -U yugabyte`.

Note: YCQL (port 9042) is off by default; set `spec.ycql.enabled: true` at creation to start it (that path was unit tested but not run in a cluster).

## Cleaning up

Because `deletionPolicy` is `WipeOut`, deleting the object also removes the PVCs and the Secrets:

```bash
$ kubectl delete yb -n demo yb-demo
yugabytedb.kubedb.com "yb-demo" deleted

$ kubectl delete ns demo
```

The `PlacementPolicy` is cluster scoped and can be reused by other databases; delete it with `kubectl delete placementpolicy spread-across-zones` if you no longer need it.

Deletion policies, as observed: `Halt` keeps PVCs and Secrets (pods, PetSets, Services and AppBinding are removed); `Delete` removes PetSets, Services and PVCs and keeps the auth Secret; `WipeOut` removes everything.

## Next Steps

- Understand the roles and placement in [Clustering](/docs/guides/yugabytedb/clustering/architecture-overview.md).
- Learn about the [YugabyteDB CRD](/docs/guides/yugabytedb/concepts/yugabytedb.md).
- Try day-2 operations: [Restart](/docs/guides/yugabytedb/restart/restart.md), [Scaling](/docs/guides/yugabytedb/scaling/vertical-scaling/scale-vertically/), [Update Version](/docs/guides/yugabytedb/update-version/versionupgrading/).
- Enable [TLS](/docs/guides/yugabytedb/tls/overview/) and [monitoring](/docs/guides/yugabytedb/monitoring/overview.md).
