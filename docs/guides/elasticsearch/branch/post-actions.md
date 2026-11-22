---
title: Running Elasticsearch Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: elasticsearch-branch-post-actions
    name: Post-Actions
    parent: guides-elasticsearch-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Elasticsearch Branch Post-Actions

Post-actions prepare branch data before the Branch becomes ready. They can mask personal information, remove sensitive data, or create test fixtures. This example masks five customer documents and deletes a sensitive index, leaving the source unchanged.

## Execution and readiness

Each `postActions` entry creates a Job in the target namespace. This tutorial uses one entry and one Job, running masking first and sensitive-data removal second. Courier waits for target readiness and any configured credential reset before running the action; the Branch becomes `Ready` only after the Job succeeds.

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

Save the following as `elasticsearch-pa-source.yaml`. It creates a three-node combined cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: elasticsearch-pa-source
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
kubectl -n demo apply -f elasticsearch-pa-source.yaml
kubectl -n demo wait elasticsearch/elasticsearch-pa-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

HTTP authentication remains enabled. `enableSSL: false` uses HTTP for the client API; Elasticsearch transport traffic has its own encryption configuration.

## Seed and inspect the source data

Discover the auth Secret referenced by `elasticsearch-pa-source`:

```shell
kubectl -n demo get elasticsearch elasticsearch-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-pa-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret elasticsearch-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it elasticsearch-pa-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/customers" \
  -H 'Content-Type: application/json' \
  --data '{"mappings":{"properties":{"id":{"type":"integer"}}}}'
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/customers/_doc/1" \
  -H 'Content-Type: application/json' \
  --data '{"id":1,"full_name":"Customer 1","email":"customer1@example.com","phone":"+1-555-0001"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/customers/_doc/2" \
  -H 'Content-Type: application/json' \
  --data '{"id":2,"full_name":"Customer 2","email":"customer2@example.com","phone":"+1-555-0002"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/customers/_doc/3" \
  -H 'Content-Type: application/json' \
  --data '{"id":3,"full_name":"Customer 3","email":"customer3@example.com","phone":"+1-555-0003"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/customers/_doc/4" \
  -H 'Content-Type: application/json' \
  --data '{"id":4,"full_name":"Customer 4","email":"customer4@example.com","phone":"+1-555-0004"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/customers/_doc/5" \
  -H 'Content-Type: application/json' \
  --data '{"id":5,"full_name":"Customer 5","email":"customer5@example.com","phone":"+1-555-0005"}'
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/sensitive_tokens/_doc/1" \
  -H 'Content-Type: application/json' \
  --data '{"token":"sample-token-1"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-pa-source.demo.svc:9200/sensitive_tokens/_doc/2" \
  -H 'Content-Type: application/json' \
  --data '{"token":"sample-token-2"}'
```

Refresh makes the documents searchable; flush commits pending index changes before the example storage snapshot:

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-pa-source.demo.svc:9200/_refresh"
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-pa-source.demo.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Customer 1", "email": "customer1@example.com", "phone": "+1-555-0001"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com", "phone": "+1-555-0002"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com", "phone": "+1-555-0003"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com", "phone": "+1-555-0004"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com", "phone": "+1-555-0005"}}
    ]
  }
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-source.demo.svc:9200/sensitive_tokens/_count?filter_path=count&pretty"
```

```json
{"count": 2}
```

Run `exit` to return to your terminal.

## Create a Branch with post-actions

Save the following as `elasticsearch-pa-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: elasticsearch-pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Elasticsearch
      name: elasticsearch-pa-source
    namespace: demo
  target:
    namespace: dev
    name: elasticsearch-pa-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  postActions:
  - image: ghcr.io/appscode-images/elastic:8.18.8
    jobDefaults:
      imagePullPolicy: IfNotPresent
      backoffLimit: 3
      activeDeadlineSeconds: 900
      ttlSecondsAfterFinished: 3600
    jobTemplate:
      metadata:
        labels:
          branch-action: elasticsearch-pa-branch-prepare
      spec:
        args:
        - bash
        - -ec
        - |
          # First operation: mask each of the five known customer documents.
          for customer_id in 1 2 3 4 5; do
            curl --silent --show-error --fail-with-body --user "$ES_USERNAME:$ES_PASSWORD" \
              -X POST "http://$ES_HOST:$ES_PORT/customers/_update/$customer_id?refresh=true" \
              -H 'Content-Type: application/json' \
              --data '{"doc":{"full_name":"Masked Customer","email":"user@example.invalid","phone":"+1-555-0000"}}'
          done
          # Second operation: remove the sensitive index only after masking succeeds.
          curl --silent --show-error --fail-with-body --user "$ES_USERNAME:$ES_PASSWORD" \
            -X DELETE "http://$ES_HOST:$ES_PORT/sensitive_tokens?ignore_unavailable=true"
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f elasticsearch-pa-branch.yaml
kubectl -n demo wait branch/elasticsearch-pa-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch elasticsearch-pa-branch
```

Courier creates `elasticsearch/elasticsearch-pa-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

The image executes the supplied Bash arguments. Each `curl --fail-with-body` call updates one known customer document and waits for search visibility. `bash -ec` stops on a failed request, so index deletion runs only after all five updates succeed. The fixed masking values and index deletion with `ignore_unavailable=true` support repetition.

## Inspect the action Job

The Branch passes through `ActionsRunning`; the phase may finish before you see it. Find the Job and inspect its logs:

```shell
kubectl -n dev get jobs -l courier.kubedb.com/branch=elasticsearch-pa-branch,courier.kubedb.com/branch-namespace=demo
kubectl -n dev logs job/elasticsearch-pa-branch-post-action -c actions
kubectl -n dev get pods -l branch-action=elasticsearch-pa-branch-prepare
```

Expect Job status `Complete` with `1/1` completions. The logs show five document-update responses and index-deletion acknowledgment.

```shell
kubectl -n demo get branch elasticsearch-pa-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
kubectl -n demo get branch elasticsearch-pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

Expect `PostActionsCompleted=True` and `Ready=True`. Resource references identify the target auth Secret, cloned PVCs, and action Job.

## Verify that only the branch was changed

### Inspect the branch

Discover the auth Secret referenced by `elasticsearch-pa-branch`:

```shell
kubectl -n dev get elasticsearch elasticsearch-pa-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-pa-branch-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret elasticsearch-pa-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it elasticsearch-pa-branch-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-branch.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-branch.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"}},
      {"_source": {"id": 2, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"}},
      {"_source": {"id": 3, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"}},
      {"_source": {"id": 4, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"}},
      {"_source": {"id": 5, "full_name": "Masked Customer", "email": "user@example.invalid", "phone": "+1-555-0000"}}
    ]
  }
}
```

```shell
curl --silent --show-error --user elastic \
  -X GET "http://elasticsearch-pa-branch.dev.svc:9200/sensitive_tokens"
```

```json
{"error":{"type":"index_not_found_exception"},"status":404}
```

The response excerpt shows the removed index.

Run `exit` to return to your terminal.

### Inspect the source

Discover the auth Secret referenced by `elasticsearch-pa-source`:

```shell
kubectl -n demo get elasticsearch elasticsearch-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-pa-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret elasticsearch-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it elasticsearch-pa-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
```

```json
{
  "hits": {
    "hits": [
      {"_source": {"id": 1, "full_name": "Customer 1", "email": "customer1@example.com", "phone": "+1-555-0001"}},
      {"_source": {"id": 2, "full_name": "Customer 2", "email": "customer2@example.com", "phone": "+1-555-0002"}},
      {"_source": {"id": 3, "full_name": "Customer 3", "email": "customer3@example.com", "phone": "+1-555-0003"}},
      {"_source": {"id": 4, "full_name": "Customer 4", "email": "customer4@example.com", "phone": "+1-555-0004"}},
      {"_source": {"id": 5, "full_name": "Customer 5", "email": "customer5@example.com", "phone": "+1-555-0005"}}
    ]
  }
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-pa-source.demo.svc:9200/sensitive_tokens/_count?filter_path=count&pretty"
```

```json
{"count": 2}
```

Run `exit` to return to your terminal.

The source keeps all five original customer records and both sensitive tokens.

## Connection environment

| Variable | Value used by the action |
|---|---|
| `ES_HOST` | Target Service |
| `ES_PORT` | 9200 |
| `ES_USERNAME` | Username from the target auth Secret |
| `ES_PASSWORD` | Password from the target auth Secret |

See the common [action fields and Job defaults](/docs/operatormanual/branching/post-actions.md#post-action-field-reference). The action image must execute its supplied arguments; the Branch template does not expose a container `command` override.

## Use post-actions with refresh or another cluster

[Refresh](/docs/guides/elasticsearch/branch/refresh.md) runs the action against every new copy before readiness. Inspect each generation’s Job using the Courier labels above; older Jobs can be removed during cleanup. For [cross-cluster branching](/docs/guides/elasticsearch/branch/cross-cluster/index.md), the Job runs on the target cluster and receives the target’s credentials.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch elasticsearch-pa-branch --wait=true --timeout=30m
kubectl -n demo delete elasticsearch elasticsearch-pa-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.
