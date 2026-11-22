---
title: Provision a DC-DR Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-provision
    name: Provision
    parent: guides-postgres-distributed
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Provision a DC-DR Postgres

This page creates a PostgreSQL database that runs two nodes in `spoke-1` (DC A) and two in `spoke-2` (DC B), with the hub as the Arbiter data center. It then writes data, and checks that the data reaches the other data center and that applications can connect from any cluster.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/distributed/provision/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/provision/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Before You Begin

- Complete [Install KubeDB and dr-controlplane](/docs/guides/postgres/distributed/install/index.md). All three `dc-health-*` Leases must be fresh.
- You should be familiar with the [Postgres](/docs/guides/postgres/concepts/postgres.md) CRD and with [Concepts](/docs/guides/postgres/distributed/concepts/index.md).
- The database namespace (`demo`) must exist in all three clusters and be an application namespace of the slice.

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
```

Pick a `PostgresVersion` from the KubeDB catalog. This guide uses `17.2`, which runs `pg-coordinator` `v0.52.0`:

```bash
$ kubectl --kubeconfig $HUB get postgresversion 17.2
```

## Step 1: Create the failover group and the PlacementPolicy

### FailoverGroup

Every database follows one primary-DC Lease, its *failover scope*, and all databases in a scope fail over together. The `FailoverGroup` object itself is optional: the Lease is derived from the PlacementPolicy's `failoverGroupRef` alone, and the object adds ordering between groups (`spec.dependsOn`) and a status view.

> **Warning:** Databases whose PlacementPolicy has no `failoverPolicy.failoverGroupRef` all share the default scope, the Lease `primary-dc`. A switchover or failover of any one of them moves **all** of them to the other data center together, including a switchover you request on only one database. Use one FailoverGroup per database, or per group of databases that must always move together (for example an application's databases that must stay in the same data center). KubeDB sets the condition `DCDRFailoverScopeShared=True` on every database whose scope is shared.

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: FailoverGroup
metadata:
  name: pg-docs
spec: {}
```

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/provision/yamls/failovergroup.yaml
failovergroup.apps.k8s.appscode.com/pg-docs created
```

### PlacementPolicy

The PlacementPolicy, on the hub, decides which pod ordinal runs in which data center, and which data centers take part in failover:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: pg-docs
spec:
  clusterSpreadConstraint:
    distributionRules:
    - clusterName: spoke-1
      replicaIndices: [0, 1]
      role: Member
    - clusterName: spoke-2
      replicaIndices: [2, 3]
      role: Member
    - clusterName: hub
      replicaIndices: []
      role: Arbiter
    failoverPolicy:
      mode: TwoDC
      failoverGroupRef:
        name: pg-docs
    slice:
      projectNamespace: kubeslice-demo-distributed-postgres
      sliceName: demo-slice
```

- `distributionRules[].clusterName` is the OCM ManagedCluster name, which is the data center name.
- `replicaIndices` lists the global pod ordinals that run in that data center. Together they must cover `0` to `spec.replicas - 1` of the Postgres.
- `role: Member` marks a data-bearing data center that can become primary. `role: Arbiter` marks a data center that holds no data (empty `replicaIndices`) and only breaks ties.
- `failoverPolicy.mode: TwoDC` requires exactly two Members and at least one Arbiter or Witness; the PlacementPolicy is rejected otherwise.
- `failoverPolicy.failoverGroupRef.name` selects the Lease `primary-dc-pg-docs`. Without it, the database follows the shared Lease `primary-dc`.
- `slice` names the KubeSlice project namespace and slice.

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/provision/yamls/placementpolicy.yaml
placementpolicy.apps.k8s.appscode.com/pg-docs created
```

## Step 2: Create the Postgres

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-docs
  namespace: demo
  annotations:
    dr.kubedb.com/enabled: "true"
spec:
  version: "17.2"
  distributed: true
  replicas: 4
  standbyMode: Hot
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  replication:
    walLimitPolicy: WALKeepSize
    walKeepSize: 1536
    maxSlotWALKeepSize: 8192
  arbiter:
    resources:
      requests:
        cpu: 100m
        memory: 192Mi
        storage: 1Gi
      limits:
        memory: 256Mi
  podTemplate:
    spec:
      podPlacementPolicy:
        name: pg-docs
      containers:
      - name: postgres
        resources:
          requests:
            cpu: 250m
            memory: 512Mi
          limits:
            memory: 1Gi
      - name: pg-coordinator
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            memory: 256Mi
  deletionPolicy: WipeOut
```

- `metadata.annotations["dr.kubedb.com/enabled"]: "true"` together with `spec.distributed: true` makes this a DC-DR database. Without the annotation, `distributed: true` creates a single Raft group spread across clusters, with no cross data center failover.
- `spec.replicas` is the total over all data centers (2 + 2).
- `spec.replication` keeps enough WAL on the primary (`walKeepSize` and `maxSlotWALKeepSize`, in MB) for the other data center to catch up with a cheap rewind after an outage instead of a full re-seed. Size it for the longest data center outage you want to absorb at your write rate.
- `spec.arbiter` sizes the arbiter pod that a data center with an even node count gets. The arbiter keeps a small Raft volume, so give it a `storage` request.
- `spec.podTemplate.spec.podPlacementPolicy` names the PlacementPolicy from Step 1.

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/provision/yamls/postgres.yaml
postgres.kubedb.com/pg-docs created
```

## Step 3: Wait for Ready, Steady and protected

Provisioning takes several minutes: KubeDB creates the per data center PetSets on the hub, the PetSet controller delivers the pods to the spokes, the first data center bootstraps a primary, and the other data center clones it and starts streaming across the slice. Expect about 7 to 15 minutes.

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -w
NAME      VERSION   STATUS   AGE
pg-docs   17.2      Ready    14m
```

`Ready` alone is not enough for a DC-DR database. Check the DC-DR status as well:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{.status.phase} {.status.disasterRecovery.phase} active={.status.disasterRecovery.activeDC} protected={.status.disasterRecovery.protected}{"\n"}{range .status.disasterRecovery.dataCenters[*]}  {.clusterName} {.role} healthy={.healthy} writable={.writable} lag={.lagBytes}{"\n"}{end}'
Ready Steady active=spoke-2 protected=true
  hub Arbiter healthy=true writable= lag=
  spoke-1 Member healthy=true writable= lag=0
  spoke-2 Member healthy=true writable=true lag=
```

The database is ready for use when:

- the phase is `Ready` and the DR phase is `Steady`;
- `protected` is `true`;
- every data center, the Arbiter included, is `healthy`;
- exactly one data center is `writable`.

The first active data center is whichever Member's agent acquired the new Lease first; here it is `spoke-2`. The full status:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{.status.disasterRecovery}' | python3 -m json.tool
{
    "activeDC": "spoke-2",
    "dataCenters": [
        {
            "clusterName": "hub",
            "healthy": true,
            "lagObservedAt": "2026-10-23T04:07:05Z",
            "role": "Arbiter"
        },
        {
            "clusterName": "spoke-1",
            "crossDCStreamer": "pg-docs-spoke-1-0",
            "healthy": true,
            "lagBytes": 0,
            "lagObservedAt": "2026-10-23T04:07:05Z",
            "leader": "pg-docs-spoke-1-0",
            "role": "Member"
        },
        {
            "clusterName": "spoke-2",
            "healthy": true,
            "leader": "pg-docs-spoke-2-1",
            "role": "Member",
            "writable": true,
            "writableObservedAt": "2026-10-23T04:07:05Z"
        }
    ],
    "lastTransitionTime": "2026-10-23T03:53:05Z",
    "phase": "Steady",
    "protected": true,
    "protectionMessage": "data center \"spoke-1\" is streaming 0 bytes behind, within the 16777216 byte budget"
}
```

`crossDCStreamer` is the pod in the standby data center that streams from the active primary; it is the promotion target if the active data center is lost.

## Step 4: Inspect what was created

On the hub, KubeDB created one PetSet per data center (`pg-docs-<dc>`), one PlacementPolicy per data center (`<policy>-<dc>`, here `pg-docs-<dc>`), plus an arbiter PetSet for each data center with an even node count:

```bash
$ kubectl --kubeconfig $HUB -n demo get petset
NAME                      DESIRED   READY   AGE
pg-docs-spoke-1           2         2       8m40s
pg-docs-spoke-1-arbiter   1         1       5m52s
pg-docs-spoke-2           2         2       11m
pg-docs-spoke-2-arbiter   1         1       8m39s

$ kubectl --kubeconfig $HUB get placementpolicy
NAME              AGE
pg-docs           14m
pg-docs-spoke-1   8m44s
pg-docs-spoke-2   11m

$ kubectl --kubeconfig $HUB get failovergroup pg-docs
NAME      LEASE HOLDER   ACTIVE DC   PHASE    AGE
pg-docs   spoke-2        spoke-2     Steady   14m
```

The pods run on the spokes. The hub runs none:

```bash
$ kubectl --kubeconfig $S1 -n demo get pods -l app.kubernetes.io/instance=pg-docs -L kubedb.com/role
NAME                        READY   STATUS    RESTARTS   AGE     ROLE
pg-docs-spoke-1-0           3/3     Running   0          8m43s   standby
pg-docs-spoke-1-1           3/3     Running   0          7m35s   standby
pg-docs-spoke-1-arbiter-0   2/2     Running   0          5m54s   arbiter

$ kubectl --kubeconfig $S2 -n demo get pods -l app.kubernetes.io/instance=pg-docs -L kubedb.com/role
NAME                        READY   STATUS    RESTARTS       AGE     ROLE
pg-docs-spoke-2-0           3/3     Running   3 (3m3s ago)   11m     standby
pg-docs-spoke-2-1           3/3     Running   3 (3m3s ago)   10m     primary
pg-docs-spoke-2-arbiter-0   2/2     Running   2 (3m3s ago)   8m42s   arbiter
```

Exactly one pod in all clusters carries `kubedb.com/role=primary`. (The restarts shown on `spoke-2` come from a restart of that spoke VM while this example was recorded, not from provisioning.)

The objects reach each spoke as OCM ManifestWorks, which live in that cluster's namespace on the hub:

```bash
$ kubectl --kubeconfig $HUB -n spoke-1 get manifestwork | grep pg-docs
data-pg-docs-spoke-1-0           8m51s
data-pg-docs-spoke-1-1           7m43s
data-pg-docs-spoke-1-arbiter-0   6m3s
demo-pg-docs                     11m
demo-pg-docs-auth                11m
demo-pg-docs-pods-spoke-1        8m54s
demo-pg-docs-pods-spoke-2        11m
demo-pg-docs-rbac                11m
demo-pg-docs-serviceexports      11m
demo-pg-docs-standby             11m
pg-docs-spoke-1-0                8m51s
pg-docs-spoke-1-1                7m43s
pg-docs-spoke-1-arbiter-0        6m3s
```

The topology controller created the scope's Lease on the hub, and each agent mirrors its holder into a marker ConfigMap on its own cluster:

```bash
$ kubectl --kubeconfig $HUB -n dc-failover get lease primary-dc-pg-docs -o yaml
apiVersion: coordination.k8s.io/v1
kind: Lease
metadata:
  annotations:
    dr.open-cluster-management.io/member-dcs: spoke-1,spoke-2
    dr.open-cluster-management.io/scope: pg-docs
  labels:
    app.kubernetes.io/managed-by: dr-controlplane
  name: primary-dc-pg-docs
  namespace: dc-failover
spec:
  acquireTime: "2026-10-23T03:52:58.982123Z"
  holderIdentity: spoke-2
  leaseDurationSeconds: 45
  leaseTransitions: 1
  renewTime: "2026-10-23T04:07:18.430083Z"

$ kubectl --kubeconfig $S1 -n dc-failover get cm primary-dc-pg-docs -o jsonpath='{.data}'
{"activeDC":"spoke-2","quiesce":"","renewTime":"2026-10-23T04:07:13Z"}
```

The client Service `pg-docs` and the per-pod Services are exported on the slice:

```bash
$ kubectl --kubeconfig $S2 -n demo get serviceexport
NAME                SLICE        INGRESS   SERVICEPORT(S)   PORT(S)             ENDPOINTS   STATUS   ALIAS
pg-docs             demo-slice                              5432/TCP,2379/TCP   1           READY    ["pg-docs.demo.svc.cluster.local","pg-docs.demo.svc.slice.local"]
pg-docs-spoke-1-0   demo-slice                              5432/TCP                        READY    ["pg-docs-spoke-1-0.pg-docs-pods-spoke-1.demo.svc.cluster.local","pg-docs-spoke-1-0.pg-docs-pods-spoke-1.demo.svc.slice.local"]
pg-docs-spoke-1-1   demo-slice                              5432/TCP                        READY    ["pg-docs-spoke-1-1.pg-docs-pods-spoke-1.demo.svc.cluster.local","pg-docs-spoke-1-1.pg-docs-pods-spoke-1.demo.svc.slice.local"]
pg-docs-spoke-2-0   demo-slice                              5432/TCP            1           READY    ["pg-docs-spoke-2-0.pg-docs-pods-spoke-2.demo.svc.cluster.local","pg-docs-spoke-2-0.pg-docs-pods-spoke-2.demo.svc.slice.local"]
pg-docs-spoke-2-1   demo-slice                              5432/TCP            1           READY    ["pg-docs-spoke-2-1.pg-docs-pods-spoke-2.demo.svc.cluster.local","pg-docs-spoke-2-1.pg-docs-pods-spoke-2.demo.svc.slice.local"]
pg-docs-standby     demo-slice                              5432/TCP            1           READY    ["pg-docs-standby.demo.svc.cluster.local","pg-docs-standby.demo.svc.slice.local"]
```

## Step 5: Write data and see it in the other data center

Open a session on the primary. It runs in `spoke-2`, so use that cluster's kubeconfig:

```bash
$ kubectl --kubeconfig $S2 -n demo exec -it pg-docs-spoke-2-1 -c postgres -- psql -U postgres
```

Create a table, insert rows and update one:

```sql
postgres=# CREATE TABLE inventory (id serial PRIMARY KEY, item text NOT NULL, qty int NOT NULL, updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE
postgres=# INSERT INTO inventory (item, qty) VALUES ('disk', 10), ('cpu', 4), ('memory', 32);
INSERT 0 3
postgres=# UPDATE inventory SET qty = 12, updated_at = now() WHERE item = 'disk';
UPDATE 1
postgres=# SELECT id, item, qty FROM inventory ORDER BY id;
 id |  item  | qty
----+--------+-----
  1 | disk   |  12
  2 | cpu    |   4
  3 | memory |  32
(3 rows)
```

The primary streams to its local standby and, across the slice, to the other data center's cross-DC streamer, through the slot `dcdr_spoke1`:

```sql
postgres=# SELECT application_name, client_addr, state, sync_state, pg_wal_lsn_diff(pg_current_wal_lsn(), replay_lsn) AS lag_bytes FROM pg_stat_replication;
 application_name  | client_addr |   state   | sync_state | lag_bytes
-------------------+-------------+-----------+------------+-----------
 pg-docs-spoke-2-0 | 10.42.0.101 | streaming | async      |         0
 pg-docs-spoke-1-0 | 10.1.16.23  | streaming | async      |         0
(2 rows)

postgres=# SELECT slot_name, active FROM pg_replication_slots;
  slot_name  | active
-------------+--------
 dcdr_spoke1 | t
(1 row)
```

`10.1.16.23` is a slice address: `pg-docs-spoke-1-0` connects from DC A over the slice.

Read the same table from the other data center. The cross-DC streamer in `spoke-1` has the rows, and refuses writes:

```bash
$ kubectl --kubeconfig $S1 -n demo exec pg-docs-spoke-1-0 -c postgres -- psql -U postgres -c "SELECT pg_is_in_recovery();" -c "SELECT id, item, qty FROM inventory ORDER BY id;"
 pg_is_in_recovery
-------------------
 t
(1 row)

 id |  item  | qty
----+--------+-----
  1 | disk   |  12
  2 | cpu    |   4
  3 | memory |  32
(3 rows)

$ kubectl --kubeconfig $S1 -n demo exec pg-docs-spoke-1-0 -c postgres -- psql -U postgres -c "INSERT INTO inventory (item, qty) VALUES ('gpu', 1);"
ERROR:  cannot execute INSERT in a read-only transaction
command terminated with exit code 1
```

In the standby data center, the cross-DC streamer receives from the primary through the slice, and the other pod cascades from it:

```bash
$ for p in pg-docs-spoke-1-0 pg-docs-spoke-1-1; do kubectl --kubeconfig $S1 -n demo exec $p -c postgres -- psql -U postgres -tAc "SELECT '$p', status, sender_host FROM pg_stat_wal_receiver;"; done
pg-docs-spoke-1-0|streaming|pg-docs.demo.svc.slice.local
pg-docs-spoke-1-1|streaming|pg-docs-spoke-1-0.pg-docs-pods-spoke-1.demo.svc
```

## Step 6: Connect an application

Applications connect to `pg-docs.demo.svc.slice.local:5432`, from any cluster on the slice. This name always points to the current primary. The credentials are in the Secret `demo/pg-docs-auth`, which KubeDB also delivers to every spoke.

The application must run in a namespace that is an application namespace of the slice, so that KubeSlice injects its sidecar into the pod. [psql-client.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/provision/yamls/psql-client.yaml) runs a `psql` client in `demo` on spoke-1, the data center that is **not** active:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: psql-client
  namespace: demo
spec:
  containers:
  - name: psql
    image: postgres:17.2-alpine
    command: ["sleep", "infinity"]
    env:
    - name: PGHOST
      value: pg-docs.demo.svc.slice.local
    - name: PGUSER
      valueFrom:
        secretKeyRef:
          name: pg-docs-auth
          key: username
    - name: PGPASSWORD
      valueFrom:
        secretKeyRef:
          name: pg-docs-auth
          key: password
```

```bash
$ kubectl --kubeconfig $S1 apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/provision/yamls/psql-client.yaml
pod/psql-client created

$ kubectl --kubeconfig $S1 -n demo get pod psql-client
NAME          READY   STATUS    RESTARTS   AGE
psql-client   2/2     Running   0          17s
```

The second container is the KubeSlice sidecar. The pod reaches the slice once its `nsm0` interface is up:

```bash
$ kubectl --kubeconfig $S1 -n demo exec psql-client -c psql -- ip -4 addr show nsm0
310: nsm0@if309: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    inet 10.1.16.26/32 scope global nsm0
       valid_lft forever preferred_lft forever
```

Connect through the slice name and write:

```bash
$ kubectl --kubeconfig $S1 -n demo exec psql-client -c psql -- psql -d postgres \
    -c "SELECT inet_server_addr(), pg_is_in_recovery();" \
    -c "INSERT INTO inventory (item, qty) VALUES ('nic', 2);" \
    -c "SELECT id, item, qty FROM inventory ORDER BY id;"
 inet_server_addr | pg_is_in_recovery
------------------+-------------------
 10.1.32.5        | f
(1 row)

INSERT 0 1
 id |  item  | qty
----+--------+-----
  1 | disk   |  12
  2 | cpu    |   4
  3 | memory |  32
  4 | nic    |   2
(4 rows)
```

The client in DC A wrote to the primary in DC B over the slice. After every move of the primary, the slice name needs about 30 s to point to the new primary. Make your application retry failed connections.

## Cleaning up

Keep `pg-docs` and `psql-client` for the next page, [Switchover and failover](/docs/guides/postgres/distributed/switchover-failover/index.md). To delete the database and its policy now, see [Teardown](/docs/guides/postgres/distributed/teardown/index.md#delete-a-dc-dr-database).

## Next Steps

- [Switchover and failover](/docs/guides/postgres/distributed/switchover-failover/index.md).
- [Horizontal scaling](/docs/guides/postgres/distributed/scaling/index.md).
