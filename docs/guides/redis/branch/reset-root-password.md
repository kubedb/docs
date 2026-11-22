---
title: Resetting a Redis Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: redis-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-redis-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a Redis Branch Root Password

Set `spec.resetRootPassword: true` to give the branch its own administrator password. Separating branch credentials reduces the risk of exposing or reusing production credentials in development and test environments. The source keeps its original credential.

This tutorial deploys its own source and branch, compares their passwords, and verifies authentication through their Services.

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

Save the following as `redis-reset-source.yaml`. It creates a three-node Redis source managed by three Sentinels:

```yaml
apiVersion: kubedb.com/v1
kind: RedisSentinel
metadata:
  name: redis-reset-source-sentinel
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
  name: redis-reset-source
  namespace: demo
spec:
  version: 7.4.0
  replicas: 3
  mode: Sentinel
  sentinelRef:
    name: redis-reset-source-sentinel
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
kubectl -n demo apply -f redis-reset-source.yaml
kubectl -n demo wait redissentinel/redis-reset-source-sentinel --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo wait redis/redis-reset-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Record the source password

Record the source password before creating the Branch. You will compare it with the source value after branching and use it for the rejection check.

Discover the auth Secret referenced by `redis-reset-source`:

```shell
kubectl -n demo get redis redis-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-reset-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret redis-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

## Seed and inspect the source data

Use the source password recorded above.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it redis-reset-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-reset-source.demo.svc -p 6379 --askpass
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

## Create a Branch with password reset enabled

Save the following as `redis-reset-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-reset-source
    namespace: demo
  target:
    namespace: dev
    name: redis-reset-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  resetRootPassword: true
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f redis-reset-branch.yaml
kubectl -n demo wait branch/redis-reset-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch redis-reset-branch
```

Courier creates `redis/redis-reset-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

Check the Branch conditions:

```shell
kubectl -n demo get branch redis-reset-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Expect `RootPasswordReset=True` and `Ready=True`.

## Compare source and branch passwords

Discover the Secret names:

```shell
kubectl -n demo get redis redis-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get redis redis-reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-reset-source-auth
redis-reset-branch-auth
```

Use the returned names if they differ. Read each password separately:

```shell
kubectl -n demo get secret redis-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
source-example-password
```

```shell
kubectl -n dev get secret redis-reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
branch-example-password
```

These strings are illustrative. The source password must match the value recorded before branching; the branch password must differ.

## Verify authentication through the Services

Open the source pod’s Bash shell from your terminal:

```shell
kubectl -n demo exec -it redis-reset-source-0 -c redis -- bash
```

### Authenticate with the branch password

Inside the pod, connect to the **branch** Service and enter the decoded **branch password** when prompted:

```shell
redis-cli -h redis-reset-branch.dev.svc -p 6379 --askpass
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
1) "Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to return to the pod shell.

### Confirm the source password is rejected by the branch

Repeat this connection to the **same branch Service**, entering the **original source password** at the prompt:

```shell
redis-cli -h redis-reset-branch.dev.svc -p 6379 --askpass
```

```text
AUTH failed: WRONGPASS invalid username-password pair or user is disabled.
```

Expect an authentication rejection; a DNS failure or connection timeout does not establish password isolation.

If `redis-cli` leaves a prompt open after the failed authentication, run `QUIT` before the next connection.

### Confirm the source credential still works

Connect to the **source** Service and enter the **original source password**:

```shell
redis-cli -h redis-reset-source.demo.svc -p 6379 --askpass
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
1) "Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

## Use password reset with other branch features

Redis applies the branch credential independently of the copied data. Refresh does not need a RotateAuth request and does not guarantee a new password on every generation.

Configured [post-actions](/docs/guides/redis/branch/post-actions.md) receive target credentials after password preparation. For [cross-cluster branching](/docs/guides/redis/branch/cross-cluster/index.md), inspect credentials and any rotation requests on the target cluster. See the common [password-reset fields](/docs/operatormanual/branching/reset-root-password.md#field-reference) and [refresh tutorial](/docs/guides/redis/branch/refresh.md).

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch redis-reset-branch --wait=true --timeout=30m
kubectl -n demo delete redis redis-reset-source --wait=true --timeout=30m
kubectl -n demo delete redissentinel redis-reset-source-sentinel --wait=true --timeout=15m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
