---
title: Version Update Recommendation
menu:
  docs_{{ .version }}:
    identifier: version-update-recommendation
    name: Version Update
    parent: recommendation
    weight: 80
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Version Update Recommendation

Database versions often need to be updated. Older versions can carry CVEs that attackers exploit, newer versions include query/indexing/storage optimisations, and vendors keep shipping bug fixes and features. Staying on top of these upgrades is one of the most impactful and most easily forgotten maintenance tasks.

KubeDB watches the versions you actually have running and generates a `Recommendation` when it notices any of these:

1. The current version's container image has been updated.
2. A newer major or minor version is available.
3. A patch release is available for your current minor.

> **Note:** `Recommendation` works for most KubeDB-managed databases. The walkthrough below uses [Elasticsearch](/docs/guides/elasticsearch)  as a concrete example.
## Before you begin

- A Kubernetes cluster with **KubeDB** and the **Supervisor** installed. The easiest way is `--set supervisor.enabled=true` when installing KubeDB via Helm.
- `kubectl` configured to talk to the cluster.
- Familiarity with the [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md) page is helpful but not required.

## How a version-update recommendation flows

`Recommendation` is a Kubernetes `Custom Resource Definition` (CRD) that declares a maintenance action for a KubeDB-managed database like [Elasticsearch](https://www.elastic.co/products/elasticsearch) or [OpenSearch](https://opensearch.org/). The Ops-manager creates it; the Supervisor schedules and executes it.

Let's walk through a complete demo. First, list the Elasticsearch versions provided by KubeDB:

```bash
$ kubectl get elasticsearchversions | grep xpack
xpack-6.8.23        6.8.23    ElasticStack   ghcr.io/appscode-images/elastic:6.8.23                                  12d
xpack-7.17.15       7.17.15   ElasticStack   ghcr.io/appscode-images/elastic:7.17.15                                 12d
xpack-7.17.28       7.17.28   ElasticStack   ghcr.io/appscode-images/elastic:7.17.28                                 12d
xpack-8.17.10       8.17.10   ElasticStack   ghcr.io/appscode-images/elastic:8.17.10                                 12d
xpack-8.17.6        8.17.6    ElasticStack   ghcr.io/appscode-images/elastic:8.17.6                                  12d
xpack-8.18.2        8.18.2    ElasticStack   ghcr.io/appscode-images/elastic:8.18.2                                  12d
xpack-8.18.8        8.18.8    ElasticStack   ghcr.io/appscode-images/elastic:8.18.8                                  12d
xpack-8.19.9        8.19.9    ElasticStack   ghcr.io/appscode-images/elastic:8.19.9                                  12d
xpack-8.2.3         8.2.3     ElasticStack   ghcr.io/appscode-images/elastic:8.2.3                                   12d
xpack-8.5.3         8.5.3     ElasticStack   ghcr.io/appscode-images/elastic:8.5.3                                   12d
xpack-9.0.2         9.0.2     ElasticStack   ghcr.io/appscode-images/elastic:9.0.2                                   12d
xpack-9.0.8         9.0.8     ElasticStack   ghcr.io/appscode-images/elastic:9.0.8                                   12d
xpack-9.1.4         9.1.4     ElasticStack   ghcr.io/appscode-images/elastic:9.1.4                                   12d
xpack-9.1.9         9.1.9     ElasticStack   ghcr.io/appscode-images/elastic:9.1.9                                   12d
xpack-9.2.3         9.2.3     ElasticStack   ghcr.io/appscode-images/elastic:9.2.3                                   12d
```

We will deliberately deploy an older version, `xpack-9.1.9`, so KubeDB will recommend the upgrade to `xpack-9.2.3`:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-vurecommendation
  namespace: demo
spec:
  version: xpack-9.1.9
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Wait until the Elasticsearch cluster reports `Ready`. The required time depends on image pull speed and node specs.

```bash
$ kubectl get elasticsearch,pods -n demo
NAME                                           VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-vurecommendation   xpack-9.1.9   Ready    3m43s

NAME                        READY   STATUS    RESTARTS   AGE
pod/es-vurecommendation-0   1/1     Running   0          3m37s
pod/es-vurecommendation-1   1/1     Running   0          3m30s
pod/es-vurecommendation-2   1/1     Running   0          3m25s
```

Once the Elasticsearch instance is `Ready`, the KubeDB Ops-manager creates a `Recommendation` automatically. It can take a couple of minutes for the create-event to be reconciled.

```bash
$ kubectl get recommendation -n demo
NAME                                                          STATUS    OUTDATED   AGE
es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o   Pending   false      2m49s
```

The Recommendation name follows the pattern `<DB-name>-x-<DB-type>-x-<recommendation-type>-<random-suffix>`. Initially the Supervisor sets `status.phase: Pending`. Let's look at the full manifest:

```yaml
$ kubectl get recommendation -n demo es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o -oyaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: xpack-9.1.9
  creationTimestamp: "2026-06-08T16:44:12Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: es-vurecommendation
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
    kubedb.com/version-update-recommendation-type: major-minor
  name: es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o
  namespace: demo
spec:
  backoffLimit: 10
  description: Latest Major/Minor version is available. Recommending version Update
    from xpack-9.1.9 to xpack-9.2.3.
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: ElasticsearchOpsRequest
    metadata:
      name: update-version
      namespace: demo
    spec:
      databaseRef:
        name: es-vurecommendation
      type: UpdateVersion
      updateVersion:
        targetVersion: xpack-9.2.3
    status: {}
  recommender:
    name: kubedb-ops-manager
  requireExplicitApproval: true
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-vurecommendation
status:
  approvalStatus: Pending
  failedAttempt: 0
  outdated: false
  parallelism: Namespace
  phase: Pending
  reason: WaitingForApproval
```

What this manifest tells you:

- `spec.description` — a new Major/Minor version (`xpack-9.2.3`) is available; KubeDB recommends upgrading from `xpack-9.1.9`.
- `spec.operation` — the recommended action is an `ElasticsearchOpsRequest` of type `UpdateVersion` with `updateVersion.targetVersion: xpack-9.2.3`.
- `spec.requireExplicitApproval: true` — **version updates always require explicit human approval**, because an upgrade can involve breaking changes or compatibility concerns. Auto-approval via an `ApprovalPolicy` is intentionally bypassed.
- `status.approvalStatus: Pending` / `status.reason: WaitingForApproval` — nothing will run until you approve it.

## Approving the recommendation

Approve via the AppsCode UI, or with `kubectl`:

```bash
$ kubectl patch Recommendation es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o \
     -n demo \
     --type merge \
     --subresource='status' \
     -p '{"status":{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"}}}'
recommendation.supervisor.appscode.com/es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o patched
```

A new condition appears almost immediately confirming the OpsRequest was created:

```bash
$ kubectl get recommendation -n demo es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o -o jsonpath='{.status}'
{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"},"conditions":[{"lastTransitionTime":"2026-06-08T16:47:29Z","message":"OpsRequest is successfully created","reason":"SuccessfullyCreatedOperation","status":"True","type":"SuccessfullyCreatedOperation"}],"createdOperationRef":{"name":"es-vurecommendation-1780937248-update-version-auto"},"failedAttempt":0,"outdated":false,"parallelism":"Namespace","phase":"InProgress","reason":"StartedExecutingOperation"}
```

The Supervisor has now created an `ElasticsearchOpsRequest` and is upgrading the cluster to `xpack-9.2.3` with negligible downtime. The Supervisor will keep retrying on transient failures up to `spec.backoffLimit` attempts.

```bash
$ kubectl get elasticsearchopsrequest -n demo
NAME                                                 TYPE            STATUS       AGE
es-vurecommendation-1780937248-update-version-auto   UpdateVersion   Successful   2m39s
```

Once the OpsRequest succeeds, the Recommendation rolls into `Succeeded`:

```bash
$ kubectl get recommendation -n demo es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o
NAME                                                          STATUS      OUTDATED   AGE
es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o   Succeeded   false      5m55s
```

The Elasticsearch cluster is now on the target version:

```bash
$ kubectl get es es-vurecommendation -n demo
NAME                  VERSION       STATUS   AGE
es-vurecommendation   xpack-9.2.3   Ready    6m50s
```

## Rejecting a recommendation

If you do not want a recommendation to run, set its `approvalStatus` to `Rejected`:

```bash
$ kubectl patch Recommendation es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o \
     -n demo \
     --type merge \
     --subresource='status' \
     -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/es-vurecommendation-x-elasticsearch-x-update-version-t7dy9o patched
```

## Automating execution

Approving every recommendation by hand defeats the point. For routine maintenance you can have the Supervisor approve and run recommendations automatically — but only inside a window of your choosing (off-peak hours, weekends, etc.).

> **Note:** Because version updates set `requireExplicitApproval: true`, ApprovalPolicies do not auto-approve them. The pattern below applies to recommendation types where a human approval is not mandatory (e.g. TLS or auth secret rotation). It is shown here for reference because the same `MaintenanceWindow` resource is used for those flows.

### 1. Define a MaintenanceWindow

A `MaintenanceWindow` says **when** ops are allowed to run. See [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for the full reference.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: elastic-maintenance
  namespace: demo
spec:
  isDefault: true
  timezone: UTC
  days:
    Sunday:
      - start: "00:00"
        end: "04:00"
    Saturday:
      - start: "00:00"
        end: "04:00"
```

### 2. Auto-approve with an ApprovalPolicy

An `ApprovalPolicy` says **what gets auto-approved** and links to the window above. See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for the full reference.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: elasticsearch-policy
  namespace: demo
maintenanceWindowRef:
  kind: MaintenanceWindow
  name: elastic-maintenance
  namespace: demo
targets:
  - group: kubedb.com
    kind: Elasticsearch
    operations:
      - group: ops.kubedb.com
        kind: ElasticsearchOpsRequest
```

### 3. Or use a cluster-wide default

If you want one default schedule for the whole cluster, swap the namespace-scoped `MaintenanceWindow` for a `ClusterMaintenanceWindow` and point `maintenanceWindowRef.kind` at it. See [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md).

---

For the complete field reference for Recommendation, see [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md).
