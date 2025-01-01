---
title: Elasticsearch Version Update Recommendation
menu:
  docs_{{ .version }}:
    identifier: es-version-update-recommendation
    name: Version Update Recommendation
    parent: es-recommendation-elasticsearch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Version Update Recommendation

Database versions often need to be updated due to several reasons. Older database versions may have vulnerabilities that hackers can exploit. New versions often include optimizations for query execution, indexing, and storage mechanisms. Modern databases frequently introduce new features, such as better data types, improved indexing methods, or advanced analytics capabilities. Database vendors release patches and updates to address these issues and introduce new features.

`Recommendation` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative recommendation for KubeDB managed databases like [Elasticsearch](https://www.elastic.co/products/elasticsearch) and [OpenSearch](https://opensearch.org/) in a Kubernetes native way. KubeDB generates Elasticsearch/Opensearch Version Update recommendation regarding three particular cases. 

1. There's been an update in the current version image
2. There's a new major/minor version available
3. There's a version available with patch fix

Let's go through a demo to see version update recommendations being generated. First, get the available Elasticsearch versions provided by KubeDB.

```bash
$ kubectl get elasticsearchversions | grep xpack
xpack-6.8.23        6.8.23    ElasticStack   ghcr.io/appscode-images/elastic:6.8.23                        17h
xpack-7.13.4        7.13.4    ElasticStack   ghcr.io/appscode-images/elastic:7.13.4                        17h
xpack-7.14.2        7.14.2    ElasticStack   ghcr.io/appscode-images/elastic:7.14.2                        17h
xpack-7.16.3        7.16.3    ElasticStack   ghcr.io/appscode-images/elastic:7.16.3                        17h
xpack-7.17.15       7.17.15   ElasticStack   ghcr.io/appscode-images/elastic:7.17.15                       17h
xpack-7.17.23       7.17.23   ElasticStack   ghcr.io/appscode-images/elastic:7.17.23                       17h
xpack-7.17.25       7.17.25   ElasticStack   ghcr.io/appscode-images/elastic:7.17.25                       16h
xpack-8.11.1        8.11.1    ElasticStack   ghcr.io/appscode-images/elastic:8.11.1                        17h
xpack-8.11.4        8.11.4    ElasticStack   ghcr.io/appscode-images/elastic:8.11.4                        17h
xpack-8.13.4        8.13.4    ElasticStack   ghcr.io/appscode-images/elastic:8.13.4                        17h
xpack-8.14.1        8.14.1    ElasticStack   ghcr.io/appscode-images/elastic:8.14.1                        17h
xpack-8.14.3        8.14.3    ElasticStack   ghcr.io/appscode-images/elastic:8.14.3                        17h
xpack-8.15.0        8.15.0    ElasticStack   ghcr.io/appscode-images/elastic:8.15.0                        17h
xpack-8.15.4        8.15.4    ElasticStack   ghcr.io/appscode-images/elastic:8.15.4                        16h
xpack-8.16.0        8.16.0    ElasticStack   ghcr.io/appscode-images/elastic:8.16.0                        16h
xpack-8.2.3         8.2.3     ElasticStack   ghcr.io/appscode-images/elastic:8.2.3                         17h
xpack-8.5.3         8.5.3     ElasticStack   ghcr.io/appscode-images/elastic:8.5.3                         17h
xpack-8.6.2         8.6.2     ElasticStack   ghcr.io/appscode-images/elastic:8.6.2                         17h
xpack-8.8.2         8.8.2     ElasticStack   ghcr.io/appscode-images/elastic:8.8.2                         17h
```

Let's deploy an Elasticsearch cluster with version `xpack-8.15.0`. We are going to create a cluster topology with 2 master nodes, 3 data nodes and 2 ingest node. We also have to provide an available storageclass for each of the node types.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: elastic
  namespace: es
spec:
  version: xpack-8.15.0
  storageType: Durable
  deletionPolicy: WipeOut
  topology:
    master:
      replicas: 2
      storage:
        storageClassName: "local-path"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 3
      storage:
        storageClassName: "local-path"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 2
      storage:
        storageClassName: "local-path"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Wait for a while till elasicsearch cluster gets into `Ready` state. Required time depends on image pulling and node's physical specifications.

```bash
$ kubectl get es elastic -n es -w
NAME      VERSION        STATUS         AGE
elastic   xpack-8.15.0   Provisioning   98s
elastic   xpack-8.15.0   Provisioning   5m43s
elastic   xpack-8.15.0   Provisioning   8m7s
.
.
.
elastic   xpack-8.15.0   Ready          10m
elastic   xpack-8.15.0   Ready          10m
```

Once elastic instance is `Ready`, a `Recommendation` instance will be automatically generated by KubeDB `Ops-Manager` controller.

```bash
NAME                                              STATUS    OUTDATED   AGE
elastic-x-elasticsearch-x-update-version-2juuee   Pending   false      10m
```

The `Recommendation` custom resource will be named as `<DB-name>-x-<DB type>-x-<Recommendation typer>-<random hash>`. Initially, the KubeDB `Supervisor` controller will mark the `Status` of this object to `Pending`. Let's check the complete Recommendation custom resource manifest:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: xpack-8.15.0
  creationTimestamp: "2025-01-29T12:06:43Z"
  generation: 5
  labels:
    app.kubernetes.io/instance: elastic
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
    kubedb.com/version-update-recommendation-type: major-minor
  name: elastic-x-elasticsearch-x-update-version-2juuee
  namespace: es
  resourceVersion: "783271"
  uid: 3026d740-64fd-4ac4-8f33-2bd305ab0e69
spec:
  backoffLimit: 5
  description: Latest Major/Minor version is available. Recommending version Update
    from xpack-8.15.0 to xpack-8.16.0.
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: ElasticsearchOpsRequest
    metadata:
      name: update-version
      namespace: es
    spec:
      databaseRef:
        name: elastic
      type: UpdateVersion
      updateVersion:
        targetVersion: xpack-8.16.0
    status: {}
  recommender:
    name: kubedb-ops-manager
  requireExplicitApproval: true
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase ==
      'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: elastic
  vulnerabilityReport:
    message: 'ImageScanRequest phase is not Current: timed out waiting for the condition'
    status: Failure
status:
  approvalStatus: Pending
  failedAttempt: 0
  outdated: false
  parallelism: Namespace
  phase: Pending
  reason: WaitingForApproval
```

In the generated Recommendation you will find a description, targeted db object, recommended operation or Ops-Request manifest, current status of the recommendation etc. Let's just focus on the recommendation description first.

```bash
$ kubectl get recommendation -n es elastic-x-elasticsearch-x-update-version-2juuee -o jsonpath='{.spec.description}'
Latest Major/Minor version is available. Recommending version Update from xpack-8.15.0 to xpack-8.16.0.
```

The recommendation says current version `xpack-8.15.0` should be latest upgradable version `xpack-8.16.0`. You can also find the recommended operation which is a `ElasticsearchOpsRequest` of `UpdateVersion` type in this case.

```bash
$ kubectl get recommendation -n es elastic-x-elasticsearch-x-update-version-2juuee -o jsonpath='{.spec.operation}' | yq -y
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: update-version
  namespace: es
spec:
  databaseRef:
    name: elastic
  type: UpdateVersion
  updateVersion:
    targetVersion: xpack-8.16.0
status: {}
```

Let's check the status part of this recommendation.

```bash
$ kubectl get recommendation -n es elastic-x-elasticsearch-x-update-version-2juuee -o jsonpath='{.status}' | yq -y
approvalStatus: Pending
failedAttempt: 0
outdated: false
parallelism: Namespace
phase: Pending
reason: WaitingForApproval
```

Now, This recommendation can be approved and operation can be executed immediately by setting `ApprovalStatus` to `Approved` and Setting `approvedWindow` to `Immediate`. You can approve this easily through Appscode UI or edit it manually. Also, You can use kubectl cli for this - 

```bash
$ kubectl patch Recommendation elastic-x-elasticsearch-x-update-version-2juuee \
                  -n es \
                  --type merge \
                  --subresource='status' \
                  -p '{"status":{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"}}}'
recommendation.supervisor.appscode.com/elastic-x-elasticsearch-x-update-version-2juuee patched
```

Now, check the status part again. You will find a condition have appeared which says `OpsRequest is successfully created`. 

```bash
$ kubectl get recommendation -n es elastic-x-elasticsearch-x-update-version-2juuee -o jsonpath='{.status}' | yq -y
approvalStatus: Approved
approvedWindow:
  window: Immediate
conditions:
  - lastTransitionTime: '2025-01-29T13:01:40Z'
    message: OpsRequest is successfully created
    reason: SuccessfullyCreatedOperation
    status: 'True'
    type: SuccessfullyCreatedOperation
createdOperationRef:
  name: elastic-1738155700-update-version-auto
failedAttempt: 0
outdated: false
parallelism: Namespace
phase: InProgress
reason: StartedExecutingOperation
```

You will find an `ElasticsearchOpsRequest` custom resource have been created and, it is updating the `elastic` cluster version to `xpack-8.16.0` with negligible downtime. Let's wait for it to reach `Successfull` status.

```bash
$ kubectl get elasticsearchopsrequest -n es elastic-1738155700-update-version-auto -w
NAME                                     TYPE            STATUS        AGE
elastic-1738155700-update-version-auto   UpdateVersion   Progressing   3m12s
elastic-1738155700-update-version-auto   UpdateVersion   Progressing   3m34s
.
.
elastic-1738155700-update-version-auto   UpdateVersion   Successful    11m
```

Let's recheck the recommendation for one last time. We should find that `.status.phase` has been marked as `Succeeded`. 

```bash
$ kubectl get recommendation -n es elastic-x-elasticsearch-x-update-version-2juuee
NAME                                              STATUS      OUTDATED   AGE
elastic-x-elasticsearch-x-update-version-2juuee   Succeeded   false      78m
```

Finally, You can check `elastic` cluster version now, which should be upgraded to version `xpack-8.16.0`.

```bash
$ kubectl get es elastic -n es
NAME      VERSION        STATUS   AGE
elastic   xpack-8.16.0   Ready    85m
```

You may not want to do trigger recommended operations manually. Rather, trigger them autonomously in a preferred schedule when infrastructure is idle or traffic rate is at the lowest. For this purpose, You can create a `MaintenanceWindow` custom resource where you can set your desired schedule/period for triggering these recommended operations automatically. Here's a sample one:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: elastic-maintenance
  namespace: es
spec:
  timezone: Asia/Dhaka
  days:
    Wednesday:
      - start: 5:40AM
        end: 7:00PM
  dates:
    - start: 2025-01-25T00:00:18Z
      end: 2025-01-25T23:41:18Z
```

You can now create a `ApprovalPolicy` custom resource to refer this `MaintenanceWindow` for particular DB type. Following is a sample `ApprovalPolicy` for any `Elasticsearch` custom resource deployed in `es` namespace. This `ApprovalPolicy` custom resource is referring to the `elastic-maintenance` MaintenanceWindow created in the same namespace. You can also create `ClusterMaintenanceWindow` instead which is effective for cluster-wide operations and refer it here. The following ApprovalPolicy will trigger recommended operations when referred maintenance window timeframe is reached. 

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: es-policy
  namespace: es
maintenanceWindowRef:
  name: elastic-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
    operations:
      - group: ops.kubedb.com
        kind: ElasticsearchOpsRequest
```

Lastly, If you want to reject a recommendation, you can just set `ApprovalStatus` to `Rejected` in the recommendation status section. Here's how you can do it using kubectl cli.

```bash
$ kubectl patch Recommendation elastic-x-elasticsearch-x-update-version-2juuee \
                  -n es \
                  --type merge \
                  --subresource='status' \
                  -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/elastic-x-elasticsearch-x-update-version-2juuee patched
```

## Next Steps

- Learn about [backup & restore](/docs/guides/elasticsearch/backup/stash/overview/index.md) Elasticsearch database using Stash.
- Learn how to configure [Elasticsearch Topology Cluster](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
