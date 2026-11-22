---
title: Branch Redis Across Clusters
menu:
  docs_{{ .version }}:
    identifier: redis-branch-cross-cluster
    name: Cross-Cluster
    parent: guides-redis-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch Redis Across Clusters

Cross-cluster branching creates an independent Redis deployment on another Kubernetes cluster from a point-in-time snapshot of an existing deployment. Use it to prepare development, testing, or data analysis environments on a separate cluster using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `RedisVersion` `7.4.0` is available in your installed catalog.
- Ensure the RedisSentinel CRD is installed; Sentinel uses the same `RedisVersion` catalog entry. Use Ceph RBD storage with copy-on-write cloning and a compatible VolumeSnapshotClass.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source Redis pods, three Sentinel pods, the target database, and any action Jobs.
- Have existing OCM registration for managed clusters `source` and `target`, with Courier agents enabled and kubectl contexts `hub`, `source`, and `target`.
- Both clusters must access **the same network-reachable Ceph backend**, with compatible CSI identities, pool settings, and credentials. Matching StorageClass names alone do not provide shared snapshots.

The source uses `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the tutorial names and create namespaces if they do not already exist:

```shell
kubectl --context source create namespace demo
kubectl --context target create namespace dev
```

Check the existing registration and storage classes:

```shell
kubectl --context hub get managedclusters source target
kubectl --context source get storageclass ceph-rbd
kubectl --context target get storageclass ceph-rbd
```

## Deploy the source database

Save the following as `redis-source.yaml`. It creates a three-node Redis source managed by three Sentinels:

```yaml
apiVersion: kubedb.com/v1
kind: RedisSentinel
metadata:
  name: redis-source-sentinel
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
  name: redis-source
  namespace: demo
spec:
  version: 7.4.0
  replicas: 3
  mode: Sentinel
  sentinelRef:
    name: redis-source-sentinel
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
kubectl --context source -n demo apply -f redis-source.yaml
kubectl --context source -n demo wait redissentinel/redis-source-sentinel --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl --context source -n demo wait redis/redis-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

Discover the auth Secret referenced by `redis-source`:

```shell
kubectl --context source -n demo get redis redis-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context source -n demo get secret redis-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it redis-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-source.demo.svc -p 6379 --askpass
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

## Create the cross-cluster Branch

Save the following as `redis-cross.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-cross
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-source
    namespace: demo
  target:
    namespace: dev
    name: redis-cross
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
    clusterName: target
  deletionPolicy: Delete
```

```shell
kubectl --context source -n demo apply -f redis-cross.yaml
kubectl --context source -n demo wait branch/redis-cross --for=condition=Ready --timeout=60m
kubectl --context source -n demo get branch redis-cross
```

The Branch is created on `source`. Courier sends the target resource manifests through the OCM hub; the target KubeDB provisioner creates the database using the delivered cloned PVCs. `target.clusterName` is the OCM ManagedCluster name.

```shell
kubectl --context hub -n source get branchworks
kubectl --context hub -n target get manifestworks
kubectl --context target -n dev get redis redis-cross
```

## Verify the copied data

Discover the auth Secret referenced by `redis-cross`:

```shell
kubectl --context target -n dev get redis redis-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-cross-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context target -n dev get secret redis-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it redis-cross-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-cross.dev.svc -p 6379 --askpass
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

### Inspect target resources

```shell
kubectl --context target -n dev get pods,pvc -l app.kubernetes.io/instance=redis-cross
```

Read and write checks below verify the target is usable independently of the source.

## Verify independent writes

### Change the branch

Discover the auth Secret referenced by `redis-cross`:

```shell
kubectl --context target -n dev get redis redis-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-cross-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context target -n dev get secret redis-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it redis-cross-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-cross.dev.svc -p 6379 --askpass
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

### Change and inspect the source

Discover the auth Secret referenced by `redis-source`:

```shell
kubectl --context source -n demo get redis redis-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context source -n demo get secret redis-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it redis-source-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-source.demo.svc -p 6379 --askpass
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

The source retains the original customer `5`; the branch edit did not reach it.

### Check the branch again

Discover the auth Secret referenced by `redis-cross`:

```shell
kubectl --context target -n dev get redis redis-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
redis-cross-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context target -n dev get secret redis-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it redis-cross-0 -c redis -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
redis-cli -h redis-cross.dev.svc -p 6379 --askpass
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
1) "Branch-only Customer 5"
2) "customer5@example.com"
```

Run `QUIT` to leave `redis-cli`, then `exit` to leave the pod shell.

The branch retains its own customer `5` and the original customer `1`. The source update did not reach it.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl --context source -n demo delete branch redis-cross --wait=true --timeout=30m
kubectl --context source -n demo delete redis redis-source --wait=true --timeout=30m
kubectl --context source -n demo delete redissentinel redis-source-sentinel --wait=true --timeout=15m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/redis/branch/refresh.md).
- [Reset the branch root password](/docs/guides/redis/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/redis/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/redis/branch/customization/index.md).
