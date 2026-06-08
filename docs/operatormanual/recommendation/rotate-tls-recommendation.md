---
title: Rotate TLS Recommendation
menu:
  docs_{{ .version }}:
    identifier: rotate-tls-recommendation
    name: Rotate TLS
    parent: recommendation
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate TLS Recommendation

TLS certificate rotation is essential for security, compliance, and operational continuity. Letting a certificate expire takes the database offline; ignoring rotation leaves key material exposed for years; falling behind on cryptographic standards eventually breaks compatibility with modern clients. To make rotation reliable and timely, KubeDB ships the `ReconfigureTLS` OpsRequest — and the Ops-manager generates a `Recommendation` to trigger it automatically as certificates approach expiry.

KubeDB raises a TLS rotation recommendation when **at least one** certificate of the database satisfies one of:

- Lifespan is **more than one month** and **less than one month** of life remains.
- Lifespan is **less than one month** and **less than one third** of life remains.

> **Note:** Recommendations work across most KubeDB-managed databases. The walkthrough below uses Elasticsearch as a concrete example.

## Before you begin

- KubeDB and the **Supervisor** are installed (`--set supervisor.enabled=true`).
- [`cert-manager`](https://cert-manager.io/docs/installation/) v1.0.0 or later is installed in the cluster. **TLS rotation only works when certificates are provisioned by cert-manager** — the operator-provisioned path does not support rotation.
- `kubectl` is configured against the cluster.

## Create an Issuer (or ClusterIssuer)

We will create a self-signed CA and an `Issuer` to back the demo. In production, follow [cert-manager's CA guide](https://cert-manager.io/docs/configuration/ca/) or your existing PKI.

Generate a CA with openssl:

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ./ca.key -out ./ca.crt \
    -subj "/CN=es/O=kubedb"
```

Create a TLS secret in the target namespace:

```bash
$ kubectl create secret tls es-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

And an `Issuer` that uses it:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: es-issuer
  namespace: demo
spec:
  ca:
    secretName: es-ca
```

```bash
$ kubectl apply -f issuer.yaml
issuer.cert-manager.io/es-issuer created
```

## Deploy Elasticsearch with short-lived certificates

To make rotation observable inside this walkthrough we use **short** certificate durations (`1h` and `2h10m`). In production you would use real values such as `2160h` (90 days).

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-tls
  namespace: demo
spec:
  deletionPolicy: WipeOut
  version: xpack-9.1.9
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: es-issuer
    certificates:
      - alias: client
        duration: 1h
      - alias: http
        duration: 2h10m
```

Wait until the cluster reports `Ready`:

```bash
$ kubectl get elasticsearch,pods -n demo
NAME                              VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-tls   xpack-9.1.9   Ready    4m39s

NAME           READY   STATUS    RESTARTS   AGE
pod/es-tls-0   1/1     Running   0          4m34s
pod/es-tls-1   1/1     Running   0          4m28s
```

## A rotate-tls Recommendation appears

Once one of the certificates crosses the threshold (about two-thirds of its lifespan, given the very short durations we set), the Ops-manager creates a `Recommendation`. With the durations above, the first one shows up roughly 35–40 minutes after the cluster comes up — and another every time another certificate nears expiry.

```bash
$ kubectl get recommendation -n demo
NAME                                             STATUS      OUTDATED   AGE
es-tls-x-elasticsearch-x-rotate-tls-w3j40x       Succeeded   false      37m
```

The Recommendation name follows the pattern `<DB-name>-x-<DB-type>-x-<recommendation-type>-<random-suffix>`. Let's look at the full manifest:

```yaml
$ kubectl get recommendation -n demo es-tls-x-elasticsearch-x-rotate-tls-w3j40x -oyaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  creationTimestamp: "2026-06-08T16:09:56Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: es-tls
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: rotate-tls
  name: es-tls-x-elasticsearch-x-rotate-tls-w3j40x
  namespace: demo
spec:
  backoffLimit: 10
  deadline: "2026-06-08T16:24:47Z"
  description: Recommending TLS certificate rotation,es-tls-client-cert Certificate
    is going to be expire on 2026-06-08 16:29:47 +0000 UTC
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: ElasticsearchOpsRequest
    metadata:
      name: rotate-tls
      namespace: demo
    spec:
      databaseRef:
        name: es-tls
      tls:
        rotateCertificates: true
      type: ReconfigureTLS
    status: {}
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-tls
status:
  approvalStatus: Approved
  approvedWindow:
    window: Immediate
  conditions:
  - lastTransitionTime: "2026-06-08T16:24:56Z"
    message: OpsRequest is successfully created
    reason: SuccessfullyCreatedOperation
    status: "True"
    type: SuccessfullyCreatedOperation
  - lastTransitionTime: "2026-06-08T16:26:56Z"
    message: OpsRequest is successfully executed
    reason: SuccessfullyExecutedOperation
    status: "True"
    type: SuccessfullyExecutedOperation
  createdOperationRef:
    name: es-tls-1780935896-rotate-tls-auto
  failedAttempt: 0
  observedGeneration: 1
  outdated: false
  parallelism: Namespace
  phase: Succeeded
  reason: SuccessfullyExecutedOperation
```

What this manifest tells you:

- `spec.description` — the rotation is recommended because the `es-tls-client-cert` certificate expires at `2026-06-08T16:29:47Z`.
- `spec.deadline` (`2026-06-08T16:24:47Z`) — Supervisor will auto-approve and execute at or before this time so the rotation happens **before** the certificate expires.
- `spec.operation` — an `ElasticsearchOpsRequest` of type `ReconfigureTLS` with `tls.rotateCertificates: true`.
- Notice that `spec.requireExplicitApproval` is **not set**. TLS rotation defaults to automatic approval — letting a certificate expire is far worse than running a rotation unattended.
- `status.approvalStatus: Approved`, `status.approvedWindow.window: Immediate` — Supervisor approved itself when the deadline arrived and chose immediate execution.
- The two conditions show the OpsRequest was created and then executed successfully.
- `status.createdOperationRef.name` — the resulting `ElasticsearchOpsRequest`.

## Watching the OpsRequest

```bash
$ kubectl get elasticsearchopsrequest -n demo es-tls-1780935896-rotate-tls-auto
NAME                                TYPE             STATUS       AGE
es-tls-1780935896-rotate-tls-auto   ReconfigureTLS   Successful   25m
```

The `ReconfigureTLS` operation re-issues the affected certificates via cert-manager, reloads each Elasticsearch pod in a controlled order, and finishes with no client-visible downtime when the cluster has more than one replica.

You can re-check the Recommendation status as JSON:

```bash
$ kubectl get recommendation es-tls-x-elasticsearch-x-rotate-tls-w3j40x \
     -n demo -o json | jq '.status'
```

You will see `phase: Succeeded` and `reason: SuccessfullyExecutedOperation`.

## Rejecting a recommendation

If you ever need to skip a rotation (for example, because you're about to swap issuers), reject it:

```bash
$ kubectl patch recommendation es-tls-x-elasticsearch-x-rotate-tls-w3j40x \
     -n demo \
     --type merge \
     --subresource='status' \
     -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/es-tls-x-elasticsearch-x-rotate-tls-w3j40x patched
```

## Automating execution with a maintenance window

Auto-approval is great, but `Immediate` execution can still be disruptive at the wrong hour. Pair a `MaintenanceWindow` with an `ApprovalPolicy` to keep rotations off-peak.

### 1. Define a MaintenanceWindow

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

See [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for the complete reference.

### 2. Auto-approve with an ApprovalPolicy

This policy auto-approves `ReconfigureTLS` (and any other Elasticsearch ops) and binds them to the window above.

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

See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for the complete reference.

### 3. Or use a cluster-wide default

If you would rather set one default for the whole cluster, replace the namespace-scoped `MaintenanceWindow` with a `ClusterMaintenanceWindow` and point `maintenanceWindowRef.kind` at it. See [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md).

> **Important:** Make sure the configured window is long enough — and frequent enough — to land **before** the certificate deadline. If your TLS lifespan is short and the window is weekly, the deadline can pass before the next window opens, and the Supervisor will rotate immediately to keep the cluster healthy.

---

For the complete field reference for Recommendation, see [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md).
