---
title: Branch Elasticsearch Across Clusters
menu:
  docs_{{ .version }}:
    identifier: elasticsearch-branch-cross-cluster
    name: Cross-Cluster
    parent: guides-elasticsearch-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch Elasticsearch Across Clusters

Cross-cluster branching creates an independent Elasticsearch deployment on another Kubernetes cluster from a point-in-time snapshot of an existing deployment. Use it to prepare development, testing, or data analysis environments on a separate cluster using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `ElasticsearchVersion` `xpack-8.18.8` is available in your installed catalog.
- Use Ceph RBD storage with copy-on-write cloning, compatible VolumeSnapshot and VolumeGroupSnapshot classes, and the corresponding CSI APIs, controllers, and driver support.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source and three target database pods, each requesting 2Gi of memory, plus operators and action Jobs.
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

Save the following as `elasticsearch-source.yaml`. It creates a three-node combined cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: elasticsearch-source
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
kubectl --context source -n demo apply -f elasticsearch-source.yaml
kubectl --context source -n demo wait elasticsearch/elasticsearch-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

HTTP authentication remains enabled. `enableSSL: false` uses HTTP for the client API; Elasticsearch transport traffic has its own encryption configuration.

## Seed and inspect the source data

Discover the auth Secret referenced by `elasticsearch-source`:

```shell
kubectl --context source -n demo get elasticsearch elasticsearch-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context source -n demo get secret elasticsearch-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it elasticsearch-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-source.demo.svc:9200/customers" \
  -H 'Content-Type: application/json' \
  --data '{"mappings":{"properties":{"id":{"type":"integer"}}}}'
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-source.demo.svc:9200/customers/_doc/1" \
  -H 'Content-Type: application/json' \
  --data '{"id":1,"full_name":"Customer 1","email":"customer1@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-source.demo.svc:9200/customers/_doc/2" \
  -H 'Content-Type: application/json' \
  --data '{"id":2,"full_name":"Customer 2","email":"customer2@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-source.demo.svc:9200/customers/_doc/3" \
  -H 'Content-Type: application/json' \
  --data '{"id":3,"full_name":"Customer 3","email":"customer3@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-source.demo.svc:9200/customers/_doc/4" \
  -H 'Content-Type: application/json' \
  --data '{"id":4,"full_name":"Customer 4","email":"customer4@example.com"}'
curl --silent --show-error --fail-with-body --user elastic \
  -X PUT "http://elasticsearch-source.demo.svc:9200/customers/_doc/5" \
  -H 'Content-Type: application/json' \
  --data '{"id":5,"full_name":"Customer 5","email":"customer5@example.com"}'
```

Refresh makes the documents searchable; flush commits pending index changes before the example storage snapshot:

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-source.demo.svc:9200/_refresh"
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-source.demo.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
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

## Create the cross-cluster Branch

Save the following as `elasticsearch-cross.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: elasticsearch-cross
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Elasticsearch
      name: elasticsearch-source
    namespace: demo
  target:
    namespace: dev
    name: elasticsearch-cross
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
    clusterName: target
  deletionPolicy: Delete
```

```shell
kubectl --context source -n demo apply -f elasticsearch-cross.yaml
kubectl --context source -n demo wait branch/elasticsearch-cross --for=condition=Ready --timeout=60m
kubectl --context source -n demo get branch elasticsearch-cross
```

The Branch is created on `source`. Courier sends the target resource manifests through the OCM hub; the target KubeDB provisioner creates the database using the delivered cloned PVCs. `target.clusterName` is the OCM ManagedCluster name.

```shell
kubectl --context hub -n source get branchworks
kubectl --context hub -n target get manifestworks
kubectl --context target -n dev get elasticsearch elasticsearch-cross
```

## Verify the copied data

Discover the auth Secret referenced by `elasticsearch-cross`:

```shell
kubectl --context target -n dev get elasticsearch elasticsearch-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-cross-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context target -n dev get secret elasticsearch-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it elasticsearch-cross-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-cross.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-cross.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
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

### Inspect target resources

```shell
kubectl --context target -n dev get pods,pvc -l app.kubernetes.io/instance=elasticsearch-cross
```

Discover the auth Secret referenced by `elasticsearch-cross`:

```shell
kubectl --context target -n dev get elasticsearch elasticsearch-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-cross-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context target -n dev get secret elasticsearch-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it elasticsearch-cross-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-cross.dev.svc:9200/_cluster/health?pretty"
```

Expect three nodes and an operational cluster. Its `cluster_name` and node endpoints belong to the branch.

Run `exit` to return to your terminal.

## Verify independent writes

### Change the branch

Discover the auth Secret referenced by `elasticsearch-cross`:

```shell
kubectl --context target -n dev get elasticsearch elasticsearch-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-cross-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context target -n dev get secret elasticsearch-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it elasticsearch-cross-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-cross.dev.svc:9200/customers/_update/5?refresh=true" \
  -H 'Content-Type: application/json' \
  --data '{"doc":{"full_name":"Branch-only Customer 5"}}'
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-cross.dev.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-cross.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-cross.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
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

### Change and inspect the source

Discover the auth Secret referenced by `elasticsearch-source`:

```shell
kubectl --context source -n demo get elasticsearch elasticsearch-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context source -n demo get secret elasticsearch-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it elasticsearch-source-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-source.demo.svc:9200/customers/_update/1?refresh=true" \
  -H 'Content-Type: application/json' \
  --data '{"doc":{"full_name":"Updated Customer 1"}}'
curl --silent --show-error --fail-with-body --user elastic \
  -X POST "http://elasticsearch-source.demo.svc:9200/_flush"
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-source.demo.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-source.demo.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
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

The source retains the original customer `5`; the branch edit did not reach it.

### Check the branch again

Discover the auth Secret referenced by `elasticsearch-cross`:

```shell
kubectl --context target -n dev get elasticsearch elasticsearch-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
elasticsearch-cross-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl --context target -n dev get secret elasticsearch-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it elasticsearch-cross-0 -c elasticsearch -- bash
```

Run the following `curl` commands inside the pod shell. `--user elastic` prompts for the decoded password on each request. Requests use the database Service.

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-cross.dev.svc:9200/customers/_count?filter_path=count&pretty"
```

```json
{
  "count": 5
}
```

```shell
curl --silent --show-error --fail-with-body --user elastic \
  -X GET "http://elasticsearch-cross.dev.svc:9200/customers/_search?size=5&sort=id:asc&filter_path=hits.hits._source&pretty"
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

The branch retains its own customer `5` and the original customer `1`. The source update did not reach it.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl --context source -n demo delete branch elasticsearch-cross --wait=true --timeout=30m
kubectl --context source -n demo delete elasticsearch elasticsearch-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/elasticsearch/branch/refresh.md).
- [Reset the branch root password](/docs/guides/elasticsearch/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/elasticsearch/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/elasticsearch/branch/customization/index.md).
