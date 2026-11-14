---
title: Running Redis Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: redis-branch-post-actions
    name: Post-Actions
    parent: guides-redis-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Redis Branch Post-Actions

Post-actions prepare branch data before the Branch becomes ready. They can mask personal information, remove sensitive data, or create test fixtures. This example masks five customer hashes and deletes a sensitive key, leaving the source unchanged.

## Execution and readiness

Each `postActions` entry creates a Job in the target namespace. This tutorial uses one entry and one Job, running masking first and sensitive-data removal second. Courier waits for target readiness and any configured credential reset before running the action; the Branch becomes `Ready` only after the Job succeeds.

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

Save the following as `redis-pa-source.yaml`. It creates a three-node Redis source managed by three Sentinels:

```yaml
apiVersion: kubedb.com/v1
kind: RedisSentinel
metadata:
  name: redis-pa-source-sentinel
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
  name: redis-pa-source
  namespace: demo
spec:
  version: 7.4.0
  replicas: 3
  mode: Sentinel
  sentinelRef:
    name: redis-pa-source-sentinel
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
kubectl -n demo apply -f redis-pa-source.yaml
kubectl -n demo wait redissentinel/redis-pa-source-sentinel --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo wait redis/redis-pa-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

Discover the auth Secret referenced by `redis-pa-source`:

```shell
kubectl -n demo get redis redis-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-pa-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret redis-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it redis-pa-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-pa-source.demo.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```redis
HSET customer:1 full_name "Customer 1" email "customer1@example.com" phone "+1-555-0001"
HSET customer:2 full_name "Customer 2" email "customer2@example.com" phone "+1-555-0002"
HSET customer:3 full_name "Customer 3" email "customer3@example.com" phone "+1-555-0003"
HSET customer:4 full_name "Customer 4" email "customer4@example.com" phone "+1-555-0004"
HSET customer:5 full_name "Customer 5" email "customer5@example.com" phone "+1-555-0005"
HSET sensitive_tokens token1 sample-token-1 token2 sample-token-2
WAIT 2 5000
```

`WAIT` should return `2`, confirming that both replicas acknowledged the writes from this client connection. If it returns less, wait for replication and repeat `WAIT 2 5000` before continuing.

```redis
HMGET customer:1 full_name email phone
HMGET customer:2 full_name email phone
HMGET customer:3 full_name email phone
HMGET customer:4 full_name email phone
HMGET customer:5 full_name email phone
```

```text
1) "Customer 1"
2) "customer1@example.com"
3) "+1-555-0001"
1) "Customer 2"
2) "customer2@example.com"
3) "+1-555-0002"
1) "Customer 3"
2) "customer3@example.com"
3) "+1-555-0003"
1) "Customer 4"
2) "customer4@example.com"
3) "+1-555-0004"
1) "Customer 5"
2) "customer5@example.com"
3) "+1-555-0005"
```

```redis
HMGET sensitive_tokens token1 token2
```

```text
1) "sample-token-1"
2) "sample-token-2"
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
HMGET customer:1 full_name email phone
HMGET customer:2 full_name email phone
HMGET customer:3 full_name email phone
HMGET customer:4 full_name email phone
HMGET customer:5 full_name email phone
```

```text
1) "Customer 1"
2) "customer1@example.com"
3) "+1-555-0001"
1) "Customer 2"
2) "customer2@example.com"
3) "+1-555-0002"
1) "Customer 3"
2) "customer3@example.com"
3) "+1-555-0003"
1) "Customer 4"
2) "customer4@example.com"
3) "+1-555-0004"
1) "Customer 5"
2) "customer5@example.com"
3) "+1-555-0005"
```

If those values are still stale, wait for replication and repeat the reads. Once they match, write this member’s RDB snapshot:

```redis
SAVE
```

```text
OK
```

Run `QUIT`, then `exit` to return to your terminal. This member must remain synchronized until Courier takes its snapshot.

## Create a Branch with post-actions

Save the following as `redis-pa-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-pa-source
    namespace: demo
  target:
    namespace: dev
    name: redis-pa-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  postActions:
  - image: ghcr.io/appscode-images/redis:7.4.0-bookworm
    jobDefaults:
      imagePullPolicy: IfNotPresent
      backoffLimit: 3
      activeDeadlineSeconds: 900
      ttlSecondsAfterFinished: 3600
    jobTemplate:
      metadata:
        labels:
          branch-action: redis-pa-branch-prepare
      spec:
        # First operation: mask the five customer hashes.
        # Second operation: remove sensitive tokens after masking succeeds.
        args: [redis-cli, -e, -h, "$(REDIS_HOST)", -p, "$(REDIS_PORT)",
               --no-auth-warning, EVAL,
               "for i = 1, 5 do redis.call('HSET', 'customer:' .. i, 'full_name', 'Masked Customer', 'email', 'user@example.invalid', 'phone', '+1-555-0000') end redis.call('DEL', 'sensitive_tokens') return 5", "0"]
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f redis-pa-branch.yaml
kubectl -n demo wait branch/redis-pa-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch redis-pa-branch
```

Courier creates `redis/redis-pa-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

The Redis image executes `redis-cli`. It reads `REDISCLI_AUTH` supplied by Courier, while Kubernetes expands the host and port references. The Lua script updates the five known hashes before deleting the sensitive key; a command error stops the script. `-e` makes a Redis command error produce a failing process exit code. The fixed updates and `DEL` are safe to repeat.

## Inspect the action Job

The Branch passes through `ActionsRunning`; the phase may finish before you see it. Find the Job and inspect its logs:

```shell
kubectl -n dev get jobs -l courier.kubedb.com/branch=redis-pa-branch,courier.kubedb.com/branch-namespace=demo
kubectl -n dev logs job/redis-pa-branch-post-action -c actions
kubectl -n dev get pods -l branch-action=redis-pa-branch-prepare
```

Expect Job status `Complete` with `1/1` completions. The script returns `5`.

```shell
kubectl -n demo get branch redis-pa-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
kubectl -n demo get branch redis-pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

Expect `PostActionsCompleted=True` and `Ready=True`. Resource references identify the target auth Secret, cloned PVCs, and action Job.

## Verify that only the branch was changed

### Inspect the branch

Discover the auth Secret referenced by `redis-pa-branch`:

```shell
kubectl -n dev get redis redis-pa-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-pa-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret redis-pa-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it redis-pa-branch-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-pa-branch.dev.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HMGET customer:1 full_name email phone
HMGET customer:2 full_name email phone
HMGET customer:3 full_name email phone
HMGET customer:4 full_name email phone
HMGET customer:5 full_name email phone
```

```text
1) "Masked Customer"
2) "user@example.invalid"
3) "+1-555-0000"
1) "Masked Customer"
2) "user@example.invalid"
3) "+1-555-0000"
1) "Masked Customer"
2) "user@example.invalid"
3) "+1-555-0000"
1) "Masked Customer"
2) "user@example.invalid"
3) "+1-555-0000"
1) "Masked Customer"
2) "user@example.invalid"
3) "+1-555-0000"
```

```redis
EXISTS sensitive_tokens
```

```text
(integer) 0
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

### Inspect the source

Discover the auth Secret referenced by `redis-pa-source`:

```shell
kubectl -n demo get redis redis-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-pa-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret redis-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it redis-pa-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-pa-source.demo.svc -p 6379 --askpass
```

Enter Redis commands at its prompt. Connections through the source Service reach its writable primary even when the pod you opened is a replica.

```redis
HMGET customer:1 full_name email phone
HMGET customer:2 full_name email phone
HMGET customer:3 full_name email phone
HMGET customer:4 full_name email phone
HMGET customer:5 full_name email phone
```

```text
1) "Customer 1"
2) "customer1@example.com"
3) "+1-555-0001"
1) "Customer 2"
2) "customer2@example.com"
3) "+1-555-0002"
1) "Customer 3"
2) "customer3@example.com"
3) "+1-555-0003"
1) "Customer 4"
2) "customer4@example.com"
3) "+1-555-0004"
1) "Customer 5"
2) "customer5@example.com"
3) "+1-555-0005"
```

```redis
HMGET sensitive_tokens token1 token2
```

```text
1) "sample-token-1"
2) "sample-token-2"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

The source keeps all five original customer records and both sensitive tokens.

## Connection environment

| Variable | Value used by the action |
|---|---|
| `REDIS_HOST` | Target Service |
| `REDIS_PORT` | 6379 |
| `REDIS_USERNAME` | Username from the target auth Secret |
| `REDISCLI_AUTH` | Target password, read automatically by redis-cli |
| `REDIS_PASSWORD` | The same target password for other clients |

See the common [action fields and Job defaults](/docs/operatormanual/branching/post-actions.md#post-action-field-reference). The action image must execute its supplied arguments; the Branch template does not expose a container `command` override.

## Use post-actions with refresh or another cluster

[Refresh](/docs/guides/redis/branch/refresh.md) runs the action against every new copy before readiness. Inspect each generation’s Job using the Courier labels above; older Jobs can be removed during cleanup. For [cross-cluster branching](/docs/guides/redis/branch/cross-cluster/index.md), the Job runs on the target cluster and receives the target’s credentials.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch redis-pa-branch --wait=true --timeout=30m
kubectl -n demo delete redis redis-pa-source --wait=true --timeout=30m
kubectl -n demo delete redissentinel redis-pa-source-sentinel --wait=true --timeout=15m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
