---
title: Refreshing a Elasticsearch Branch
menu:
  docs_{{ .version }}:
    identifier: elasticsearch-branch-refresh
    name: Refresh
    parent: guides-elasticsearch-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a Elasticsearch Branch

Set `spec.schedule.cron` to periodically replace branch data with a fresh source snapshot. This guide deploys its own source and scheduled branch, changes both sides, and verifies what a completed refresh preserves and replaces.

## What happens during refresh

1. Courier advances the refresh generation and takes fresh source snapshots.
2. It imports snapshot references into the target namespace, pauses the target, and replaces its cloned volumes.
3. KubeDB restarts the target using the new volumes.
4. After the target and any configured preparation complete, the Branch returns to `Ready`; Courier releases the previous generation’s snapshots.

**Refresh discards branch-only writes and temporarily interrupts the target.** It replaces data rather than merging changes. The source remains independent. If configured, [password reset](/docs/guides/elasticsearch/branch/reset-root-password.md) and [post-actions](/docs/guides/elasticsearch/branch/post-actions.md) run again on each refresh.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `ElasticsearchVersion` `xpack-8.18.8` is available in your installed catalog.
- Use Ceph RBD storage with copy-on-write cloning, compatible VolumeSnapshot and VolumeGroupSnapshot classes, and the corresponding CSI APIs, controllers, and driver support.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source and three target database pods, each requesting 2Gi of memory, plus operators and action Jobs.

The source uses `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the tutorial names and create namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `elasticsearch-refresh-source.yaml`. It creates a three-node combined cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: elasticsearch-refresh-source
  namespace: demo
spec:
  version: xpack-8.18.8
  replicas: 3
  enableSSL: false
  podTemplate:
    spec:
      containers:
      - name: elasticsearch
        resources:
          requests:
            cpu: 500m
            memory: 2Gi
          limits:
            memory: 2Gi
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
kubectl -n demo apply -f elasticsearch-refresh-source.yaml
kubectl -n demo wait elasticsearch/elasticsearch-refresh-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

HTTP authentication remains enabled. `enableSSL: false` uses HTTP for the client API; Elasticsearch transport traffic has its own encryption configuration.

## Seed and inspect the source data

Discover the auth Secret referenced by `elasticsearch-refresh-source`:

```shell
kubectl -n demo get elasticsearch elasticsearch-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret elasticsearch-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it elasticsearch-refresh-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-refresh-source.demo.svc:9200/customers" \
  -H 'Content-Type: application/json' \
  --data '{"mappings":{"properties":{"id":{"type":"integer"}}}}'
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-refresh-source.demo.svc:9200/customers/_doc/1" \
  -H 'Content-Type: application/json' \
  --data '{"id":1,"full_name":"Customer 1","email":"customer1@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-refresh-source.demo.svc:9200/customers/_doc/2" \
  -H 'Content-Type: application/json' \
  --data '{"id":2,"full_name":"Customer 2","email":"customer2@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-refresh-source.demo.svc:9200/customers/_doc/3" \
  -H 'Content-Type: application/json' \
  --data '{"id":3,"full_name":"Customer 3","email":"customer3@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-refresh-source.demo.svc:9200/customers/_doc/4" \
  -H 'Content-Type: application/json' \
  --data '{"id":4,"full_name":"Customer 4","email":"customer4@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-refresh-source.demo.svc:9200/customers/_doc/5" \
  -H 'Content-Type: application/json' \
  --data '{"id":5,"full_name":"Customer 5","email":"customer5@example.com"}'
```

Refresh makes the documents searchable; flush commits pending index changes before the example storage snapshot:

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-refresh-source.demo.svc:9200/_refresh"
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-refresh-source.demo.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}}
    ]
  }
}
```

Run `exit` to return to your terminal.

## Create a scheduled Branch

The example runs on ten-minute clock boundaries, such as `10:00`, `10:10`, and `10:20`; it is not a ten-minute timer from creation.

Save the following as `elasticsearch-refresh-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: elasticsearch-refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Elasticsearch
      name: elasticsearch-refresh-source
    namespace: demo
  target:
    namespace: dev
    name: elasticsearch-refresh-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  schedule:
    cron: '*/10 * * * *'
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f elasticsearch-refresh-branch.yaml
kubectl -n demo wait branch/elasticsearch-refresh-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch elasticsearch-refresh-branch
```

Courier creates `elasticsearch/elasticsearch-refresh-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

### Check the initial copy

Discover the auth Secret referenced by `elasticsearch-refresh-branch`:

```shell
kubectl -n dev get elasticsearch elasticsearch-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret elasticsearch-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it elasticsearch-refresh-branch-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-branch.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-branch.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}}
    ]
  }
}
```

Run `exit` to return to your terminal.

### Record the current generation and schedule

```shell
kubectl -n demo get branch elasticsearch-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT:.status.nextRefreshTime'
```

Record the generation and successful completion timestamp before making the following edits. Initial generation `0` can appear as `<none>`. Make both edits before the next scheduled run; if one starts while you are editing, finish the edits and wait for a subsequent generation. Runs that are still active can cause scheduled ticks to be skipped.

## Change source and branch data

### Update the source

Discover the auth Secret referenced by `elasticsearch-refresh-source`:

```shell
kubectl -n demo get elasticsearch elasticsearch-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret elasticsearch-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it elasticsearch-refresh-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-refresh-source.demo.svc:9200/customers/_update/1?refresh=true" \
  -H 'Content-Type: application/json' \
  --data '{"doc":{"full_name":"Updated Customer 1"}}'
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-refresh-source.demo.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}}
    ]
  }
}
```

Run `exit` to return to your terminal.

### Make a branch-only change

Discover the auth Secret referenced by `elasticsearch-refresh-branch`:

```shell
kubectl -n dev get elasticsearch elasticsearch-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret elasticsearch-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it elasticsearch-refresh-branch-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-refresh-branch.dev.svc:9200/customers/_update/5?refresh=true" \
  -H 'Content-Type: application/json' \
  --data '{"doc":{"full_name":"Branch-only Customer 5"}}'
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-refresh-branch.dev.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-branch.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-branch.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}},
      {"_source": {"id": 5, "full_name": "Branch-only Customer 5", "email": "customer5@example.com"}}
    ]
  }
}
```

Run `exit` to return to your terminal.

## Wait for the next completed refresh

```shell
kubectl -n demo get branch elasticsearch-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' --watch
```

Illustrative output:

```text
NAME                   PHASE        GENERATION   LAST-SUCCESS
elasticsearch-refresh-branch Ready        <none>       2026-11-11T10:02:00Z
elasticsearch-refresh-branch Refreshing   1            2026-11-11T10:02:00Z
elasticsearch-refresh-branch Ready        1            2026-11-11T10:18:00Z
```

Wait for **a later generation, phase `Ready`, and an advanced `LAST-SUCCESS` timestamp**. An initial `Ready` row or an increased generation alone does not establish completion. Press **Ctrl+C** to stop watching. Recovery duration depends on the engine and available resources.

```shell
kubectl -n demo get branch elasticsearch-refresh-branch -o yaml
```

Inspect `status.history`, `lastSuccessfulRefreshTime`, and `snapshot.generation`. With the group-snapshot prerequisite above, `snapshot.strategy` is `VolumeGroupSnapshot`; per-volume fallback can occur if no compatible group class is available. `FRESHNESS` measures time since successful completion, not the source snapshot’s age.

## Verify the refreshed data

### Check the branch

Open a new pod/client session after recovery; the previous session may have been disconnected.

Discover the auth Secret referenced by `elasticsearch-refresh-branch`:

```shell
kubectl -n dev get elasticsearch elasticsearch-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-refresh-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret elasticsearch-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it elasticsearch-refresh-branch-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-branch.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-branch.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}}
    ]
  }
}
```

Run `exit` to return to your terminal.

Customer `1` has the source update, customer `5` has its original value, and five records remain. The branch-only edit was discarded.

### Confirm the source remains unchanged by refresh

Discover the auth Secret referenced by `elasticsearch-refresh-source`:

```shell
kubectl -n demo get elasticsearch elasticsearch-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-refresh-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret elasticsearch-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it elasticsearch-refresh-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-refresh-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Updated Customer 1", "email": "customer1@example.com"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}}
    ]
  }
}
```

Run `exit` to return to your terminal.

The source retains its own update and never receives the branch-only change.

## Apply the same strategy across clusters

Use the [cross-cluster tutorial](/docs/guides/elasticsearch/branch/cross-cluster/index.md) and add the same schedule to its Branch. Manage the original Branch on the source cluster, inspect the database on the target, and wait for a later completed generation before checking data.

See the common [schedule and history fields](/docs/operatormanual/branching/refresh.md#schedule-and-history-field-reference) and [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) for configuration details.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch elasticsearch-refresh-branch --wait=true --timeout=30m
kubectl -n demo delete elasticsearch elasticsearch-refresh-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
