---
title: Rotate Auth Recommendation
menu:
  docs_{{ .version }}:
    identifier: rotate-auth-recommendation
    name: Rotate Auth
    parent: recommendation
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Authentication Rotate Recommendation

Rotating authentication secrets in database management is vital to mitigate security risks, such as credential leakage or unauthorized access, and to comply with regulatory requirements. Regular rotation limits the exposure of compromised credentials, reduces the risk of insider threats, and enforces updated security policies like stronger passwords or algorithms. It also ensures operational resilience by testing the rotation process and revoking stale or unused credentials. KubeDB provides `RotateAuth` which reduces manual errors, and strengthens database security with minimal effort. KubeDB Ops-manager generates Recommendation for rotating authentication secrets via this OpsRequest.

> Note: We provide support for `Recommendation` across most database systems. Below is an example demonstrating how recommendations are applied for the `Elasticsearch` database.

`Recommendation` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative recommendation for KubeDB managed databases like [Elasticsearch](https://www.elastic.co/products/elasticsearch) and [OpenSearch](https://opensearch.org/) in a Kubernetes native way. The recommendation will only be created if `.spec.authSecret.rotateAfter` is set. KubeDB generates Elasticsearch/Opensearch Rotate Auth recommendation regarding two particular cases.

1. AuthSecret lifespan is more than one month and, less than one month remaining till expiry
2. AuthSecret lifespan is less than one month and, less than one third of lifespan remaining till expiry

Let's go through a demo to see `RotateAuth` recommendations being generated. First, get the available versions provided by KubeDB.

```bash
$  kubectl get elasticsearchversions
NAME                VERSION   DISTRIBUTION   DB_IMAGE                                                   DEPRECATED   AGE
opensearch-1.3.13   1.3.13    OpenSearch     ghcr.io/appscode-images/opensearch:1.3.13                               3m1s
opensearch-1.3.20   1.3.20    OpenSearch     ghcr.io/appscode-images/opensearch:1.3.20                               3m1s
opensearch-2.19.2   2.19.2    OpenSearch     ghcr.io/appscode-images/opensearch:2.19.2                               3m1s
opensearch-2.5.0    2.5.0     OpenSearch     ghcr.io/appscode-images/opensearch:2.5.0                                3m1s
opensearch-3.1.0    3.1.0     OpenSearch     ghcr.io/appscode-images/opensearch:3.1.0                                3m1s
opensearch-3.4.0    3.4.0     OpenSearch     ghcr.io/appscode-images/opensearch:3.4.0                                3m1s
searchguard-7.9.3   7.9.3     SearchGuard    docker.io/floragunncom/sg-elasticsearch:7.9.3-oss-47.1.0                3m1s
xpack-6.8.23        6.8.23    ElasticStack   ghcr.io/appscode-images/elastic:6.8.23                                  3m1s
xpack-7.17.15       7.17.15   ElasticStack   ghcr.io/appscode-images/elastic:7.17.15                                 3m1s
xpack-7.17.28       7.17.28   ElasticStack   ghcr.io/appscode-images/elastic:7.17.28                                 3m1s
xpack-8.17.10       8.17.10   ElasticStack   ghcr.io/appscode-images/elastic:8.17.10                                 3m1s
xpack-8.17.6        8.17.6    ElasticStack   ghcr.io/appscode-images/elastic:8.17.6                                  3m1s
xpack-8.18.2        8.18.2    ElasticStack   ghcr.io/appscode-images/elastic:8.18.2                                  3m1s
xpack-8.18.8        8.18.8    ElasticStack   ghcr.io/appscode-images/elastic:8.18.8                                  3m1s
xpack-8.19.9        8.19.9    ElasticStack   ghcr.io/appscode-images/elastic:8.19.9                                  3m1s
xpack-8.2.3         8.2.3     ElasticStack   ghcr.io/appscode-images/elastic:8.2.3                                   3m1s
xpack-8.5.3         8.5.3     ElasticStack   ghcr.io/appscode-images/elastic:8.5.3                                   3m1s
xpack-9.0.2         9.0.2     ElasticStack   ghcr.io/appscode-images/elastic:9.0.2                                   3m1s
xpack-9.0.8         9.0.8     ElasticStack   ghcr.io/appscode-images/elastic:9.0.8                                   3m1s
xpack-9.1.4         9.1.4     ElasticStack   ghcr.io/appscode-images/elastic:9.1.4                                   3m1s
xpack-9.1.9         9.1.9     ElasticStack   ghcr.io/appscode-images/elastic:9.1.9                                   3m1s
xpack-9.2.3         9.2.3     ElasticStack   ghcr.io/appscode-images/elastic:9.2.3                                   3m1s
```

Let's deploy a cluster with version `xpack-9.1.9`. First create a credential:

```bash
kubectl create secret generic es-auth -n demo \
                                                  --type=kubernetes.io/basic-auth \
                                                  --from-literal=username=elastic \
                                                  --from-literal=password=testpassword

secret/es-auth created
```

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-rarecommendation
  namespace: demo
spec:
  version: xpack-9.1.9
  authSecret:
     name: es-auth
     rotateAfter: 1h
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

Wait for a while till elasticsearch cluster gets into `Ready` state. Required time depends on image pulling and node's physical specifications.

```bash
$ kubectl get elasticsearch,pods -n demo
NAME                                              VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-rarecommendation      xpack-9.1.9   Ready    3m24s

NAME                           READY   STATUS    RESTARTS   AGE
pod/es-rarecommendation-0      1/1     Running   0          3m20s
pod/es-rarecommendation-1      1/1     Running   0          113s
pod/es-rarecommendation-2      1/1     Running   0          104s
```

Since, `.spec.authSecret.rotateAfter` is set as `1h`, it is expected that the recommendation engine will generate a rotate-auth recommendation at least after 40 minutes (two-third of lifespan) of the authsecret creation. Once generated you will get a similar recommendation as follows.

```bash
$ kubectl get recommendation -n demo
NAME                                                          STATUS      OUTDATED   AGE
es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b      Succeeded   true       73m
es-rarecommendation-x-elasticsearch-x-rotate-auth-hbtptb      Succeeded   true       153m
es-rarecommendation-x-elasticsearch-x-rotate-auth-zuh67o      Succeeded   false      11m
es-rarecommendation-x-elasticsearch-x-update-version-jdz0h1   Pending     false      152m
```

The `Recommendation` custom resource will be named as `<DB-name>-x-<DB type>-x-<Recommendation type>-<random hash>`. Initially, the KubeDB `Supervisor` controller will mark the `Status` of this object to `Pending`. Let's check the complete Recommendation custom resource manifest:

```yaml
$ kubectl get recommendation -n demo es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b -oyaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  creationTimestamp: "2026-06-05T10:23:51Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: es-rarecommendation
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: rotate-auth
  name: es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b
  namespace: demo
  resourceVersion: "15975"
  uid: 7ec40fc4-3291-4322-b142-bb650232d0c8
spec:
  backoffLimit: 10
  deadline: "2026-06-05T09:54:05Z"
  description: Recommending AuthSecret rotation,es-auth AuthSecret needs to be rotated
    before 2026-06-05 10:04:05 +0000 UTC
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: ElasticsearchOpsRequest
    metadata:
      name: rotate-auth
      namespace: demo
    spec:
      databaseRef:
        name: es-rarecommendation
      type: RotateAuth
    status: {}
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase ==
      'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-rarecommendation
status:
  approvalStatus: Approved
  approvedWindow:
    window: Immediate
  conditions:
  - lastTransitionTime: "2026-06-05T10:23:51Z"
    message: OpsRequest is successfully created
    reason: SuccessfullyCreatedOperation
    status: "True"
    type: SuccessfullyCreatedOperation
  - lastTransitionTime: "2026-06-05T10:25:51Z"
    message: OpsRequest is successfully executed
    reason: SuccessfullyExecutedOperation
    status: "True"
    type: SuccessfullyExecutedOperation
  createdOperationRef:
    name: es-rarecommendation-1780655031-rotate-auth-auto
  failedAttempt: 0
  observedGeneration: 1
  outdated: true
  parallelism: Namespace
  phase: Succeeded
  reason: SuccessfullyExecutedOperation
```

In the `spec.operation` field, the recommendation suggests rotating the authentication secret of `es-rarecommendation`. The recommended operation is an `ElasticsearchOpsRequest` of `RotateAuth` type.

Notice that unlike version update recommendations, rotate-auth recommendations do **not** set `requireExplicitApproval`. This means KubeDB Supervisor automatically approves and executes the operation when the `deadline` is reached — ensuring authentication secrets are always rotated on time without requiring manual intervention.

In this case, the `deadline` (`"2026-06-05T09:54:05Z"`) had already passed by the time the recommendation was created (`"2026-06-05T10:23:51Z"`), so Supervisor immediately set `approvalStatus: Approved` with `approvedWindow: Immediate` and triggered the OpsRequest automatically.

You can also approve the recommendation manually before the deadline using kubectl CLI:

```bash
$ kubectl patch Recommendation es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b \
                  -n demo \
                  --type merge \
                  --subresource='status' \
                  -p '{"status":{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"}}}'
recommendation.supervisor.appscode.com/es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b patched
```

Now, check the status part again. You will find a condition have appeared which says `OpsRequest is successfully created`.

```bash
$ kubectl get recommendation -n demo es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b -o jsonpath='{.status}' | yq -y
approvalStatus: Approved
approvedWindow:
  window: Immediate
conditions:
  - lastTransitionTime: '2026-06-05T10:23:51Z'
    message: OpsRequest is successfully created
    reason: SuccessfullyCreatedOperation
    status: 'True'
    type: SuccessfullyCreatedOperation
createdOperationRef:
  name: es-rarecommendation-1780655031-rotate-auth-auto
failedAttempt: 0
outdated: true
parallelism: Namespace
phase: InProgress
reason: StartedExecutingOperation
```

You will find an `ElasticsearchOpsRequest` custom resource has been created and it is rotating the auth secret of `es-rarecommendation` cluster with negligible downtime. Let's wait for it to reach `Successful` status.

```bash
$ kubectl get elasticsearchopsrequest -n demo es-rarecommendation-1780655031-rotate-auth-auto -w
NAME                                              TYPE         STATUS        AGE
es-rarecommendation-1780655031-rotate-auth-auto   RotateAuth   Progressing   60s
es-rarecommendation-1780655031-rotate-auth-auto   RotateAuth   Progressing   114s
.
.
es-rarecommendation-1780655031-rotate-auth-auto   RotateAuth   Successful    5m
```

Let's recheck the recommendation for one last time. We should find that `.status.phase` has been marked as `Succeeded`.

```bash
$ kubectl get recommendation -n demo es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b
NAME                                                       STATUS      OUTDATED   AGE
es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b   Succeeded   true       78m
```

You may not want to trigger recommended operations manually. Rather, trigger them autonomously in a preferred schedule when infrastructure is idle or traffic rate is at the lowest. For this purpose, You can create a `MaintenanceWindow` custom resource where you can set your desired schedule/period for triggering these recommended operations automatically. See [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for detailed documentation. Here's a sample one:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: elastic-maintenance
  namespace: demo
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

You can now create a `ApprovalPolicy` custom resource to refer this `MaintenanceWindow` for particular DB type. See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for detailed documentation. Following is a sample `ApprovalPolicy` for any `Elasticsearch` custom resource deployed in `demo` namespace. This `ApprovalPolicy` custom resource is referring to the `elastic-maintenance` MaintenanceWindow created in the same namespace. You can also create `ClusterMaintenanceWindow` instead (see [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md)) which is effective for cluster-wide operations and refer it here. The following ApprovalPolicy will trigger recommended operations when referred maintenance window timeframe is reached.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: es-policy
  namespace: demo
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
$ kubectl patch Recommendation es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b \
                  -n demo \
                  --type merge \
                  --subresource='status' \
                  -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/es-rarecommendation-x-elasticsearch-x-rotate-auth-eonv4b patched
```

For complete reference on all Recommendation fields, phases, and status conditions, see [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md).
