---
title: Resetting a Elasticsearch Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: elasticsearch-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-elasticsearch-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a Elasticsearch Branch Root Password

Set `spec.resetRootPassword: true` to give the branch its own administrator password. Separating branch credentials reduces the risk of exposing or reusing production credentials in development and test environments. The source keeps its original credential.

This tutorial deploys its own source and branch, compares their passwords, and verifies authentication through their Services.

## Execution order

Users are stored in the cloned database data, so the target first starts with a copy of the source credential. KubeDB Ops Manager processes a `ElasticsearchOpsRequest` with type `RotateAuth`. Courier waits for success before completing password reset and reporting Branch readiness.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `ElasticsearchVersion` `xpack-8.18.8` is available in your installed catalog.
- Use Ceph RBD storage with copy-on-write cloning, compatible VolumeSnapshot and VolumeGroupSnapshot classes, and the corresponding CSI APIs, controllers, and driver support.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source and three target database pods, each requesting 2Gi of memory, plus operators and action Jobs.
- Install KubeDB Ops Manager and the `ElasticsearchOpsRequest` CRD; credential reset uses a target `RotateAuth` request.

The source uses `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the tutorial names and create namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `elasticsearch-reset-source.yaml`. It creates a three-node combined cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: elasticsearch-reset-source
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
kubectl -n demo apply -f elasticsearch-reset-source.yaml
kubectl -n demo wait elasticsearch/elasticsearch-reset-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

HTTP authentication remains enabled. `enableSSL: false` uses HTTP for the client API; Elasticsearch transport traffic has its own encryption configuration.

## Record the source password

Record the source password before creating the Branch. You will compare it with the source value after branching and use it for the rejection check.

Discover the auth Secret referenced by `elasticsearch-reset-source`:

```shell
kubectl -n demo get elasticsearch elasticsearch-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-reset-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret elasticsearch-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

## Seed and inspect the source data

Use the source password recorded above.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it elasticsearch-reset-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-reset-source.demo.svc:9200/customers" \
  -H 'Content-Type: application/json' \
  --data '{"mappings":{"properties":{"id":{"type":"integer"}}}}'
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-reset-source.demo.svc:9200/customers/_doc/1" \
  -H 'Content-Type: application/json' \
  --data '{"id":1,"full_name":"Customer 1","email":"customer1@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-reset-source.demo.svc:9200/customers/_doc/2" \
  -H 'Content-Type: application/json' \
  --data '{"id":2,"full_name":"Customer 2","email":"customer2@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-reset-source.demo.svc:9200/customers/_doc/3" \
  -H 'Content-Type: application/json' \
  --data '{"id":3,"full_name":"Customer 3","email":"customer3@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-reset-source.demo.svc:9200/customers/_doc/4" \
  -H 'Content-Type: application/json' \
  --data '{"id":4,"full_name":"Customer 4","email":"customer4@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-reset-source.demo.svc:9200/customers/_doc/5" \
  -H 'Content-Type: application/json' \
  --data '{"id":5,"full_name":"Customer 5","email":"customer5@example.com"}'
```

Refresh makes the documents searchable; flush commits pending index changes before the example storage snapshot:

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-reset-source.demo.svc:9200/_refresh"
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-reset-source.demo.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-reset-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-reset-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
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

## Create a Branch with password reset enabled

Save the following as `elasticsearch-reset-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: elasticsearch-reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Elasticsearch
      name: elasticsearch-reset-source
    namespace: demo
  target:
    namespace: dev
    name: elasticsearch-reset-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  resetRootPassword: true
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f elasticsearch-reset-branch.yaml
kubectl -n demo wait branch/elasticsearch-reset-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch elasticsearch-reset-branch
```

Courier creates `elasticsearch/elasticsearch-reset-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

Inspect the target rotation request:

```shell
kubectl -n dev get elasticsearchopsrequest elasticsearch-reset-branch-branch-reset-0 -o custom-columns='NAME:.metadata.name,TYPE:.spec.type,PHASE:.status.phase'
```

```text
NAME                                  TYPE         PHASE
elasticsearch-reset-branch-branch-reset-0   RotateAuth   Successful
```

Check the Branch conditions:

```shell
kubectl -n demo get branch elasticsearch-reset-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Expect `RootPasswordReset=True` and `Ready=True`.

## Compare source and branch passwords

Discover the Secret names:

```shell
kubectl -n demo get elasticsearch elasticsearch-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get elasticsearch elasticsearch-reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-reset-source-auth
elasticsearch-reset-branch-auth
```

Use the returned names if they differ. Read each password separately:

```shell
kubectl -n demo get secret elasticsearch-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
source-example-password
```

```shell
kubectl -n dev get secret elasticsearch-reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
branch-example-password
```

These strings are illustrative. The source password must match the value recorded before branching; the branch password must differ.

## Verify authentication through the Services

Open the source pod’s Bash shell from your terminal:

```shell
kubectl -n demo exec -it elasticsearch-reset-source-0 -c elasticsearch -- bash
```

### Authenticate with the branch password

Inside the pod, connect to the **branch** Service and enter the decoded **branch password** when prompted:

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-reset-branch.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{"count": 5}
```

### Confirm the source password is rejected by the branch

Repeat this connection to the **same branch Service**, entering the **original source password** at the prompt:

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-reset-branch.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```text
curl: (22) The requested URL returned error: 401
```

Expect an authentication rejection; a DNS failure or connection timeout does not establish password isolation.

### Confirm the source credential still works

Connect to the **source** Service and enter the **original source password**:

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-reset-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{"count": 5}
```

Run `exit` to return to your terminal.

## Use password reset with other branch features

On refresh, the source credential returns with the freshly cloned data, and Courier runs another target rotation before readiness. Re-read the target Secret after completion.

Configured [post-actions](/docs/guides/elasticsearch/branch/post-actions.md) receive target credentials after password preparation. For [cross-cluster branching](/docs/guides/elasticsearch/branch/cross-cluster/index.md), inspect credentials and any rotation requests on the target cluster. See the common [password-reset fields](/docs/operatormanual/branching/reset-root-password.md#field-reference) and [refresh tutorial](/docs/guides/elasticsearch/branch/refresh.md).

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch elasticsearch-reset-branch --wait=true --timeout=30m
kubectl -n demo delete elasticsearch elasticsearch-reset-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
