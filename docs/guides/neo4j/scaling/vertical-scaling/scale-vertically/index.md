---
title: Scale Neo4j Vertically
menu:
  docs_{{ .version }}:
    identifier: neo4j-scale-vertically
    name: Scale Vertically
    parent: neo4j-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> 🆕 New to KubeDB? Start with the [quickstart guide](/docs/README.md) before continuing here.

# Neo4j Vertical Scaling with KubeDB

Vertical scaling adjusts the CPU and memory **requests and limits** on running Neo4j pods without recreating the database or losing data. KubeDB handles the rollout automatically through `Neo4jOpsRequest`.

This guide walks you through:
- Deploying a Neo4j cluster
- Checking current pod resource allocation
- Applying a vertical scale-up
- Verifying the new resources are live

---

## How It Works

KubeDB uses `Neo4jOpsRequest` with `type: VerticalScaling` to update pod resources. Under the hood it:

1. Patches the StatefulSet container spec with the new CPU/memory values
2. Performs a rolling restart of Neo4j pods one at a time
3. Marks the operation `Successful` once all pods are running with the new resources

---

## Prerequisites

| Requirement | Details |
|---|---|
| KubeDB installed | Provisioner and Ops-manager operators running |
| `kubectl` configured | With permissions to create namespaces and resources |

See also: [Neo4j](/docs/guides/neo4j/concepts/neo4j.md) · [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md) · [Vertical Scaling Overview](/docs/guides/neo4j/scaling/vertical-scaling/overview.md)

---

## Step 1 — Set Up the Namespace

```bash
kubectl create ns demo
```

---

## Step 2 — Deploy Neo4j

Save this as `neo4j.yaml`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storage:
    resources:
      requests:
        storage: 2Gi
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
  podTemplate:
    spec:
      containers:
        - name: neo4j
          resources:
            requests:
              cpu: "250m"
              memory: "1Gi"
            limits:
              cpu: "500m"
              memory: "1Gi"
  deletionPolicy: WipeOut
```

Apply it and wait for the cluster to become ready:

```bash
kubectl apply -f neo4j.yaml
```

```bash
kubectl get neo4j -n demo neo4j-test -w
```

Wait until `STATUS` shows `Ready` before proceeding.

```bash
NAME         VERSION     STATUS   AGE
neo4j-test   2025.12.1   Ready    3m
```

---

## Step 3 — Check Current Pod Resources

Before scaling, record the existing CPU and memory values so you can confirm the change later.

```bash
kubectl get pod -n demo neo4j-test-0 \
  -o jsonpath='{.spec.containers[0].resources}' | jq .
```

Expected output:

```json
{
  "limits":   { "cpu": "500m",  "memory": "1Gi" },
  "requests": { "cpu": "250m",  "memory": "1Gi" }
}
```

---

## Step 4 — Apply the Vertical Scaling OpsRequest

The `Neo4jOpsRequest` below raises the CPU limit to `1500m` and memory to `4Gi`.

Save this as `neo4j-vertical-scale.yaml`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: neo4j-test
  verticalScaling:
    server:
      resources:
        requests:
          cpu: "700m"
          memory: "4Gi"
        limits:
          cpu: "1500m"
          memory: "4Gi"
```

Apply it and wait for completion:

```bash
kubectl apply -f neo4j-vertical-scale.yaml
```

```bash
kubectl wait --for=jsonpath='{.status.phase}'=Successful \
  neo4jopsrequest/neo4j-vertical-scale \
  -n demo --timeout=600s
```

Expected output:

```bash
neo4jopsrequest.ops.kubedb.com/neo4j-vertical-scale condition met
```

### What Happens During This Step

Once you apply the OpsRequest, KubeDB Ops-manager picks it up and begins the scaling process:

1. The OpsRequest moves to `Progressing` phase
2. KubeDB patches the StatefulSet with the new resource values
3. Pods are restarted one at a time — each must reach `Running` before the next restarts
4. Once all pods are up with the new resources, the OpsRequest moves to `Successful`

You can watch the live status with:

```bash
kubectl get neo4jopsrequest -n demo neo4j-vertical-scale -w
```

---

## Step 5 — Verify the New Resources

```bash
kubectl get neo4jopsrequest -n demo neo4j-vertical-scale
```

```bash
kubectl get pod -n demo neo4j-test-0 \
  -o jsonpath='{.spec.containers[0].resources}' | jq .
```

Expected output:

```
NAME                   TYPE              STATUS       AGE
neo4j-vertical-scale   VerticalScaling   Successful   101s

{
  "limits":   { "cpu": "1500m", "memory": "4Gi" },
  "requests": { "cpu": "700m",  "memory": "4Gi" }
}
```

> ✅ The values now match `spec.verticalScaling.server.resources` — the scaling is complete.

---

## Understanding the OpsRequest Fields

| Field | Description |
|---|---|
| `spec.type` | `VerticalScaling` — identifies this as a resource update operation |
| `spec.databaseRef.name` | Name of the target `Neo4j` resource |
| `spec.verticalScaling.server.resources.requests` | Minimum CPU/memory guaranteed to the pod |
| `spec.verticalScaling.server.resources.limits` | Maximum CPU/memory the pod is allowed to use |

> **Tip:** Always set `requests` equal to `limits` for Neo4j in production. This gives the pod a [Guaranteed QoS class](https://kubernetes.io/docs/concepts/workloads/pods/pod-qos/), preventing it from being evicted under memory pressure.

---

## Troubleshooting

If this OpsRequest does not finish, first inspect the affected pod and then check the `kubedb-ops-manager` operator logs for the exact error. For a shared checklist, see the [Neo4j Ops Request Overview](/docs/guides/neo4j/concepts/opsrequest.md#troubleshooting).

**OpsRequest stays in `Progressing` and never completes**

Check the pod that is being restarted and look for scheduling or resource issues:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test
```

```bash
kubectl describe pod -n demo neo4j-test-0
```

```bash
kubectl describe node <node-name> | grep -A 10 "Allocated resources"
```

**OpsRequest moves to `Failed`**

Read the failure condition and then inspect the `kubedb-ops-manager` logs:

```bash
kubectl get neo4jopsrequest -n demo neo4j-vertical-scale -o jsonpath='{.status.conditions}' | jq .
kubectl logs -n <kubedb-namespace> -l app.kubernetes.io/name=kubedb-ops-manager --tail=50
```

**Pod restarts repeatedly after scaling (`CrashLoopBackOff`)**

If Neo4j does not start with the new memory values, inspect the previous pod logs:

```bash
kubectl logs -n demo neo4j-test-0 --previous
```

**`jq` not installed**

Use jsonpath directly:

```bash
kubectl get neo4jopsrequest -n demo neo4j-vertical-scale -o jsonpath='{.status.conditions}'
```

---

## Cleanup

```bash
kubectl delete neo4jopsrequest -n demo neo4j-vertical-scale
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```

---

## Next Steps

- [Neo4j Horizontal Scaling](/docs/guides/neo4j/scaling/horizontal-scaling/) — add or remove cluster members
- [Neo4j Version Upgrade](/docs/guides/neo4j/update-version/) — roll to a newer Neo4j release
- [Neo4j Volume Expansion](/docs/guides/neo4j/volume-expansion/) — increase persistent storage