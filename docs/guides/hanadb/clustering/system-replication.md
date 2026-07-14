---
title: HanaDB System Replication
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-clustering-system-replication
    name: System Replication
    parent: guides-hanadb-clustering
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB System Replication

KubeDB can run SAP HANA in a multi-node **System Replication** cluster: a primary node serves
read/write traffic and one or more secondary nodes replicate from it. This guide deploys a System
Replication cluster and inspects its replication state.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/clustering](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/clustering) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- A System Replication cluster runs multiple full HANA instances and can take **30–60 minutes** to
  become `Ready`. Make sure your cluster has enough capacity.
- Create a namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create a System Replication Cluster

Set `spec.topology.mode: SystemReplication`. The `systemReplication` block controls the HANA
replication and operation modes:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: Durable
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  podTemplate:
    spec:
      containers:
      - name: hanadb
        resources:
          requests:
            cpu: "1500m"
            memory: "8Gi"
          limits:
            cpu: "4"
            memory: "14Gi"
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/clustering/system-replication.yaml
hanadb.kubedb.com/hanadb-cluster created
```

Here,

- `spec.replicas` is the total number of HANA nodes. It must be `>= 2`.
- `spec.topology.systemReplication.replicationMode` is one of `sync`, `syncmem`, `async`, or `fullsync`.
- `spec.topology.systemReplication.operationMode` is one of `logreplay`, `delta_datashipping`, or
  `logreplay_readaccess`. With `logreplay_readaccess`, the secondary accepts read-only queries and
  KubeDB creates a dedicated `secondary-hanadb-cluster` Service.
- When `spec.replicas` is even, KubeDB also creates a small **arbiter** pod that acts as a raft
  tie-breaker so the cluster can always elect a primary.

## Verify the Cluster

Wait until `hanadb-cluster` is `Ready`:

```bash
$ kubectl get hanadb.kubedb.com -n demo hanadb-cluster
NAME             VERSION   STATUS   AGE
hanadb-cluster   2.0.82    Ready    12m
```

KubeDB labels each pod with its role (`kubedb.com/role`). Note the `hanadb-cluster-arbiter-0` pod —
because `spec.replicas` is even (2), KubeDB added an arbiter as the raft tie-breaker:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=hanadb-cluster -L kubedb.com/role
NAME                       READY   STATUS    RESTARTS   AGE     ROLE
hanadb-cluster-0           2/2     Running   0          12m     primary
hanadb-cluster-1           2/2     Running   0          12m     secondary
hanadb-cluster-arbiter-0   1/1     Running   0          5m33s   arbiter
```

The Services route traffic to the primary and (read-only) secondary:

```bash
$ kubectl get svc -n demo -l app.kubernetes.io/instance=hanadb-cluster
NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)               AGE
hanadb-cluster             ClusterIP   10.43.65.245    <none>        39017/TCP             12m
hanadb-cluster-pods        ClusterIP   None            <none>        39001/TCP,39017/TCP   12m
hanadb-cluster-secondary   ClusterIP   10.43.219.128   <none>        39017/TCP             12m
```

Here `hanadb-cluster` always points at the primary, `hanadb-cluster-secondary` at the read-only
secondary (created because `operationMode` is `logreplay_readaccess`), and `hanadb-cluster-pods` is the
governing headless Service.

## Inspect System Replication State

Identify the primary pod (role `primary`) and inspect HANA's System Replication status:

```bash
$ kubectl exec -n demo hanadb-cluster-0 -c hanadb -- /bin/sh -lc \
  'source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbnsutil -sr_state'
System Replication State
~~~~~~~~~~~~~~~~~~~~~~~~
online: true
mode: primary
operation mode: primary
site id: 1
site name: SITE_hanadb-cluster-0
is source system: true
is secondary/consumer system: false
has secondaries/consumers attached: true
is a takeover active: false
Site Mappings:
~~~~~~~~~~~~~~
SITE_hanadb-cluster-0 (primary/primary)
    |---SITE_hanadb-cluster-1 (sync/logreplay_readaccess)
done.
```

The HANA SystemReplication status confirms the secondary is connected and `ACTIVE`. (HANA maps the
`fullsync` replication mode to `SYNC` plus the full-sync option, so the runtime mode reads `SYNC`.)

```bash
$ HANA_PASSWORD="$(kubectl get secret hanadb-cluster-auth -n demo -o jsonpath='{.data.password}' | base64 -d)"
$ kubectl exec -n demo hanadb-cluster-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p '$HANA_PASSWORD' \
  \"SELECT SITE_NAME, SECONDARY_SITE_NAME, REPLICATION_MODE, REPLICATION_STATUS FROM SYS.M_SERVICE_REPLICATION\""
SITE_NAME,SECONDARY_SITE_NAME,REPLICATION_MODE,REPLICATION_STATUS
"SITE_hanadb-cluster-0","SITE_hanadb-cluster-1","SYNC","ACTIVE"
1 row selected
```

## Day-2 Operations

Once the cluster is running you can:

- [Restart](/docs/guides/hanadb/restart/restart.md) it (primary restarted last),
- [Vertically scale](/docs/guides/hanadb/scaling/vertical-scaling/vertical-scaling.md) the nodes,
- Enable [TLS](/docs/guides/hanadb/tls/overview.md), and
- [Migrate storage](/docs/guides/hanadb/storage-migration/storage-migration.md).

## Cleaning Up

```bash
$ kubectl delete hanadb.kubedb.com -n demo hanadb-cluster
$ kubectl delete ns demo
```

## Next Steps

- Detailed concepts of the [HanaDB object](/docs/guides/hanadb/concepts/hanadb.md).
- Review the [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md).
