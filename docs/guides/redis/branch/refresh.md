---
title: Refreshing a Redis Branch
menu:
  docs_{{ .version }}:
    identifier: redis-branch-refresh
    name: Refresh
    parent: guides-redis-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a Redis Branch

Set `spec.schedule.cron` to periodically replace branch data with a fresh source snapshot. This guide deploys its own source and scheduled branch, changes both sides, and verifies what a completed refresh preserves and replaces.

## What happens during refresh

1. Courier advances the refresh generation and takes fresh source snapshots.
2. It imports snapshot references into the target namespace, pauses the target, and replaces its cloned volumes.
3. KubeDB restarts the target using the new volumes.
4. After the target and any configured preparation complete, the Branch returns to `Ready`; Courier releases the previous generation’s snapshots.

**Refresh discards branch-only writes and temporarily interrupts the target.** It replaces data rather than merging changes. The source remains independent. Configured [post-actions](/docs/guides/redis/branch/post-actions.md) run again on each refresh. Redis’s [password-reset option](/docs/guides/redis/branch/reset-root-password.md) gives the branch its own credential during provisioning; refresh does not require a new rotation request or guarantee another password.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `RedisVersion` `7.4.0` is available in your installed catalog.
- Ensure the RedisSentinel CRD is installed; Sentinel uses the same `RedisVersion` catalog entry. Use Ceph RBD storage with copy-on-write cloning and a compatible VolumeSnapshotClass.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source Redis pods, three Sentinel pods, the target database, and any action Jobs.

The source uses `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the tutorial names and create namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `redis-refresh-source.yaml`. It creates a three-node Redis source managed by three Sentinels:

```yaml
apiVersion: kubedb.com/v1
kind: RedisSentinel
metadata:
  name: redis-refresh-source-sentinel
  namespace: demo
spec:
  version: 7.4.0
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
---
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-refresh-source
  namespace: demo
spec:
  version: 7.4.0
  replicas: 3
  mode: Sentinel
  sentinelRef:
    name: redis-refresh-source-sentinel
    namespace: demo
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            memory: 1Gi
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```shell
kubectl -n demo apply -f redis-refresh-source.yaml
kubectl -n demo wait redissentinel/redis-refresh-source-sentinel --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo wait redis/redis-refresh-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

Discover the auth Secret referenced by `redis-refresh-source`:

```shell
kubectl -n demo get redis redis-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret redis-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it redis-refresh-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-refresh-source.demo.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```redis
HSET customer:1 full_name "Customer 1" email "customer1@example.com"
HSET customer:2 full_name "Customer 2" email "customer2@example.com"
HSET customer:3 full_name "Customer 3" email "customer3@example.com"
HSET customer:4 full_name "Customer 4" email "customer4@example.com"
HSET customer:5 full_name "Customer 5" email "customer5@example.com"
WAIT 2 5000
```

`WAIT` should return `2`, confirming that both replicas acknowledged the writes from this client connection. If it returns less, wait for replication and repeat `WAIT 2 5000` before continuing.

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to return to the pod’s Bash shell; keep that shell open.

### Persist the source member used for the snapshot

Courier snapshots the source’s ordinal-0 data volume for this Sentinel example. Confirm that replication has reached it and persist its expected data before creating the Branch or waiting for refresh.

In the source pod’s Bash shell, connect directly to its local member rather than through the primary Service:

```shell
redis-cli -h 127.0.0.1 -p 6379 --askpass
```

Enter the source password, then check all five customers. The values must match the source results above:

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

If those values are still stale, wait for replication and repeat the reads. Once they match, write this member’s RDB snapshot:

```redis
SAVE
```

```text
OK
```

Run `QUIT`, then `exit` to return to your terminal. This member must remain synchronized until Courier takes its snapshot.

## Create a scheduled Branch

The example runs on ten-minute clock boundaries, such as `10:00`, `10:10`, and `10:20`; it is not a ten-minute timer from creation.

Save the following as `redis-refresh-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-refresh-source
    namespace: demo
  target:
    namespace: dev
    name: redis-refresh-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  schedule:
    cron: '*/10 * * * *'
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f redis-refresh-branch.yaml
kubectl -n demo wait branch/redis-refresh-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch redis-refresh-branch
```

Courier creates `redis/redis-refresh-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

### Check the initial copy

Discover the auth Secret referenced by `redis-refresh-branch`:

```shell
kubectl -n dev get redis redis-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret redis-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it redis-refresh-branch-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-refresh-branch.dev.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

### Record the current generation and schedule

```shell
kubectl -n demo get branch redis-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT:.status.nextRefreshTime'
```

Record the generation and successful completion timestamp before making the following edits. Initial generation `0` can appear as `<none>`. Make both edits before the next scheduled run; if one starts while you are editing, finish the edits and wait for a subsequent generation. Runs that are still active can cause scheduled ticks to be skipped.

## Change source and branch data

### Update the source

Discover the auth Secret referenced by `redis-refresh-source`:

```shell
kubectl -n demo get redis redis-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret redis-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it redis-refresh-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-refresh-source.demo.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HSET customer:1 full_name "Updated Customer 1"
WAIT 2 5000
```

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Updated Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

Confirm `WAIT` returned `2`. Run `QUIT` to return to the pod’s Bash shell.

### Persist the source member used for the snapshot

Courier snapshots the source’s ordinal-0 data volume for this Sentinel example. Confirm that replication has reached it and persist its expected data before creating the Branch or waiting for refresh.

In the source pod’s Bash shell, connect directly to its local member rather than through the primary Service:

```shell
redis-cli -h 127.0.0.1 -p 6379 --askpass
```

Enter the source password, then check all five customers. The values must match the source results above:

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Updated Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

If those values are still stale, wait for replication and repeat the reads. Once they match, write this member’s RDB snapshot:

```redis
SAVE
```

```text
OK
```

Run `QUIT`, then `exit` to return to your terminal. This member must remain synchronized until Courier takes its snapshot.

### Make a branch-only change

Discover the auth Secret referenced by `redis-refresh-branch`:

```shell
kubectl -n dev get redis redis-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret redis-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it redis-refresh-branch-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-refresh-branch.dev.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HSET customer:5 full_name "Branch-only Customer 5"
```

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Branch-only Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

## Wait for the next completed refresh

```shell
kubectl -n demo get branch redis-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' --watch
```

Illustrative output:

```text
NAME                   PHASE        GENERATION   LAST-SUCCESS
redis-refresh-branch   Ready        <none>       2026-11-11T10:02:00Z
redis-refresh-branch   Refreshing   1            2026-11-11T10:02:00Z
redis-refresh-branch   Ready        1            2026-11-11T10:18:00Z
```

Wait for **a later generation, phase `Ready`, and an advanced `LAST-SUCCESS` timestamp**. An initial `Ready` row or an increased generation alone does not establish completion. Press **Ctrl+C** to stop watching. Recovery duration depends on the engine and available resources.

```shell
kubectl -n demo get branch redis-refresh-branch -o yaml
```

Inspect `status.history`, `lastSuccessfulRefreshTime`, and `snapshot.generation`. For this Sentinel-source example, `snapshot.strategy` is `VolumeSnapshot`, using the source’s ordinal-0 member. `FRESHNESS` measures time since successful completion, not the source snapshot’s age.

## Verify the refreshed data

### Check the branch

Open a new pod/client session after recovery; the previous session may have been disconnected.

Discover the auth Secret referenced by `redis-refresh-branch`:

```shell
kubectl -n dev get redis redis-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret redis-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it redis-refresh-branch-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-refresh-branch.dev.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Updated Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

Customer `1` has the source update, customer `5` has its original value, and five records remain. The branch-only edit was discarded.

### Confirm the source remains unchanged by refresh

Discover the auth Secret referenced by `redis-refresh-source`:

```shell
kubectl -n demo get redis redis-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret redis-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it redis-refresh-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-refresh-source.demo.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HMGET customer:1 full_name email
HMGET customer:2 full_name email
HMGET customer:3 full_name email
HMGET customer:4 full_name email
HMGET customer:5 full_name email
```

```text
1) "Updated Customer 1"
2) "customer1@example.com"
1) "Customer 2"
2) "customer2@example.com"
1) "Customer 3"
2) "customer3@example.com"
1) "Customer 4"
2) "customer4@example.com"
1) "Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

The source retains its own update and never receives the branch-only change.

## Apply the same strategy across clusters

Use the [cross-cluster tutorial](/docs/guides/redis/branch/cross-cluster/index.md) and add the same schedule to its Branch. Manage the original Branch on the source cluster, inspect the database on the target, and wait for a later completed generation before checking data.

See the common [schedule and history fields](/docs/operatormanual/branching/refresh.md#schedule-and-history-field-reference) and [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) for configuration details.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch redis-refresh-branch --wait=true --timeout=30m
kubectl -n demo delete redis redis-refresh-source --wait=true --timeout=30m
kubectl -n demo delete redissentinel redis-refresh-source-sentinel --wait=true --timeout=15m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
