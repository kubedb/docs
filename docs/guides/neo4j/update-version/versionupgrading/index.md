---
title: Upgrade Neo4j Version
menu:
  docs_{{ .version }}:
    identifier: neo4j-version-upgrading
    name: Version Upgrading
    parent: neo4j-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> 🆕 New to KubeDB? Start with the [quickstart guide](/docs/README.md) before continuing here.

# Neo4j Version Upgrade with KubeDB

Upgrading Neo4j involves restarting pods with an updated container image. KubeDB handles the rolling upgrade automatically through `Neo4jOpsRequest`, ensuring the cluster stays available throughout.

This guide walks you through:
- Deploying a Neo4j cluster
- Checking the current version
- Applying a version upgrade request
- Verifying the cluster is running the new version

---

## How It Works

KubeDB uses `Neo4jOpsRequest` with `type: UpdateVersion` to upgrade Neo4j. Under the hood it:

1. Resolves the target image from the `Neo4jVersion` catalog
2. Performs a rolling restart, updating one pod at a time
3. Waits for each pod to become ready before proceeding to the next
4. Marks the operation `Successful` once all pods are on the new version

> **Note:** During the rollout, Neo4j may briefly show a `Critical` status before converging back to `Ready`. This is expected — the cluster remains available because only one pod restarts at a time.

---

## Prerequisites

| Requirement | Details |
|---|---|
| KubeDB installed | Provisioner and Ops-manager operators running |
| `kubectl` configured | With permissions to create namespaces and resources |

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
  version: "2025.10.1"
  replicas: 3
  storage:
    resources:
      requests:
        storage: 2Gi
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

Apply it and wait for the cluster to become ready:

```bash
kubectl apply -f neo4j.yaml

kubectl get neo4j -n demo neo4j-test -w
```

Wait until `STATUS` shows `Ready` before proceeding.

```
NAME         VERSION     STATUS   AGE
neo4j-test   2025.10.1   Ready    3m
```

To see all available upgrade targets before proceeding:

```bash
kubectl get neo4jversions
```

---

## Step 3 — Apply the Version Upgrade OpsRequest

Save this as `neo4j-update-version.yaml`, setting `targetVersion` to the version you want:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: neo4j-test
  updateVersion:
    targetVersion: "2025.12.1"
```

Apply it and wait for completion:

```bash
kubectl apply -f neo4j-update-version.yaml

kubectl wait --for=jsonpath='{.status.phase}'=Successful \
  neo4jopsrequest/neo4j-update-version \
  -n demo --timeout=600s
```

Expected output:

```
neo4jopsrequest.ops.kubedb.com/neo4j-update-version condition met
```

### What Happens During This Step

Once you apply the OpsRequest, KubeDB Ops-manager begins the upgrade:

1. The OpsRequest moves to `Progressing` phase
2. KubeDB resolves the new container image from the `Neo4jVersion` catalog
3. Pods are updated one at a time — the cluster stays available throughout
4. Neo4j may briefly report `Critical` status between pod restarts — this is normal
5. Once all pods are running the new image and are healthy, the OpsRequest moves to `Successful`

You can watch live progress with:

```bash
kubectl get neo4jopsrequest -n demo neo4j-update-version -w
```

---

## Step 4 — Verify the Upgrade

```bash
kubectl get neo4jopsrequest -n demo neo4j-update-version

kubectl get neo4j -n demo neo4j-test
```

Expected output:

```
NAME                   TYPE            STATUS       AGE
neo4j-update-version   UpdateVersion   Successful   2m57s

NAME         VERSION     STATUS   AGE
neo4j-test   2025.12.1   Ready    21m
```

> ✅ The `VERSION` column now shows `2025.12.1` and `STATUS` is `Ready` — the upgrade is complete.

---

## Understanding the OpsRequest Fields

| Field | Description |
|---|---|
| `spec.type` | `UpdateVersion` — identifies this as a version upgrade operation |
| `spec.databaseRef.name` | Name of the target `Neo4j` resource |
| `spec.updateVersion.targetVersion` | The `Neo4jVersion` name to upgrade to (must exist in the catalog) |

---

## Troubleshooting

If this OpsRequest does not finish, first inspect the affected pod and then check the `kubedb-ops-manager` operator logs for the exact error. For a shared checklist, see the [Neo4j Ops Request Overview](/docs/guides/neo4j/concepts/opsrequest.md#troubleshooting).

**OpsRequest stays in `Progressing` and never completes**

A pod is likely not becoming ready after its image was updated. Find which pod is stuck and inspect its events and logs:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test
kubectl describe pod -n demo neo4j-test-0
kubectl logs -n demo neo4j-test-0
```

Common causes include image pull failures, insufficient node resources, or a version mismatch.

**`targetVersion` not found**

The version must exist as a `Neo4jVersion` catalog entry in your cluster. If the version you want is missing from the list, your KubeDB catalog is outdated:

```bash
kubectl get neo4jversions | grep 2025
```

Update your KubeDB operator to get the latest catalog, or check the [supported versions list](/docs/guides/neo4j/README.md).

**Neo4j stuck in `Critical` and does not recover**

If the new version fails to start cleanly, inspect the pod logs for startup errors and then check the `kubedb-ops-manager` logs:

```bash
kubectl logs -n demo neo4j-test-0
kubectl logs -n demo neo4j-test-1
kubectl logs -n <kubedb-namespace> -l app.kubernetes.io/name=kubedb-ops-manager --tail=50
```

**OpsRequest moves to `Failed`**

Read the failure condition directly, then use the operator logs for the detailed message:

```bash
kubectl get neo4jopsrequest -n demo neo4j-update-version -o jsonpath='{.status.conditions}' | jq .
kubectl logs -n <kubedb-namespace> -l app.kubernetes.io/name=kubedb-ops-manager --tail=50
```

---

## Cleanup

```bash
kubectl delete neo4jopsrequest -n demo neo4j-update-version
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```

---

## Next Steps

- [Neo4j Vertical Scaling](/docs/guides/neo4j/scaling/vertical-scaling/) — adjust CPU and memory
- [Neo4j Horizontal Scaling](/docs/guides/neo4j/scaling/horizontal-scaling/) — add or remove cluster members
- [Neo4j Volume Expansion](/docs/guides/neo4j/volume-expansion/) — increase persistent storage