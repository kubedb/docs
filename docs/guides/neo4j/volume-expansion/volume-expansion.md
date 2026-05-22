---
title: Expand Neo4j Volume
menu:
  docs_{{ .version }}:
    identifier: neo4j-volume-expansion-cluster
    name: Cluster
    parent: neo4j-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> 🆕 New to KubeDB? Start with the [quickstart guide](/docs/README.md) before continuing here.

# Neo4j Volume Expansion with KubeDB

Volume expansion increases the persistent storage size allocated to Neo4j pods without downtime. KubeDB delegates the resize to your storage provisioner through `Neo4jOpsRequest`.

This guide walks you through:
- Deploying a Neo4j cluster with Longhorn storage
- Checking current PVC sizes
- Applying an online volume expansion
- Verifying the new disk size is available to the database

> **Longhorn requirement:** This guide uses the [Longhorn](https://longhorn.io/) storage class, which supports online volume expansion out of the box. Ensure Longhorn is installed in your cluster before proceeding:
> ```bash
> kubectl get storageclass longhorn
> ```

---

## How It Works

KubeDB uses `Neo4jOpsRequest` with `type: VolumeExpansion` to resize persistent volumes. Under the hood it:

1. Patches the `PersistentVolumeClaim` size on each pod
2. Signals Longhorn to perform the resize online — no pod restart needed
3. Verifies the new capacity is reflected inside the pod
4. Marks the operation `Successful` once all PVCs match the target size

**Online vs Offline mode:**

| Mode | When to use |
|---|---|
| `Online` | Pod keeps running during resize. Longhorn and most cloud CSI drivers support this. |
| `Offline` | Pod is stopped, volume is resized, then pod restarts. Use when your storage class does not support online expansion. |

> **Important:** Volume expansion is **irreversible** — you cannot shrink a PVC after expanding it.

---

## Prerequisites

| Requirement | Details |
|---|---|
| KubeDB installed | Provisioner and Ops-manager operators running |
| Longhorn installed | Storage class `longhorn` available in the cluster |
| `kubectl` configured | With permissions to create namespaces and resources |

---

## Step 1 — Set Up the Namespace

```bash
kubectl create ns demo
```

---

## Step 2 — Deploy Neo4j with Longhorn Storage

Save this as `neo4j.yaml`. Note that `storageClassName: longhorn` is set so that PVCs are provisioned by Longhorn and support online expansion:

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
    storageClassName: longhorn
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
neo4j-test   2025.12.1   Ready    3m
```

---

## Step 3 — Check Current PVC Size

Confirm each pod has a `2Gi` volume provisioned by Longhorn:

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=neo4j-test
```

Expected output:

```bash
NAME                STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-neo4j-test-0   Bound    ...      2Gi        RWO            longhorn       3m
data-neo4j-test-1   Bound    ...      2Gi        RWO            longhorn       3m
data-neo4j-test-2   Bound    ...      2Gi        RWO            longhorn       3m
```

---

## Step 4 — Apply the Volume Expansion OpsRequest

Save this as `neo4j-volume-expansion.yaml`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: neo4j-test
  volumeExpansion:
    mode: "Online"
    server: "4Gi"
```

Apply it and wait for completion:

```bash
kubectl apply -f neo4j-volume-expansion.yaml

kubectl wait --for=jsonpath='{.status.phase}'=Successful \
  neo4jopsrequest/neo4j-volume-expansion \
  -n demo --timeout=600s
```

Expected output:

```
neo4jopsrequest.ops.kubedb.com/neo4j-volume-expansion condition met
```

### What Happens During This Step

Once you apply the OpsRequest, KubeDB Ops-manager begins the expansion:

1. The OpsRequest moves to `Progressing` phase
2. KubeDB patches each PVC's `spec.resources.requests.storage` to `4Gi`
3. Longhorn detects the PVC change and resizes the volume online — pods keep running
4. KubeDB verifies the new capacity is visible inside each container
5. Once all PVCs reflect the new size, the OpsRequest moves to `Successful`

You can watch live progress with:

```bash
kubectl get neo4jopsrequest -n demo neo4j-volume-expansion -w
```

---

## Step 5 — Verify the New Volume Size

```bash
kubectl get neo4jopsrequest -n demo neo4j-volume-expansion

kubectl get pvc -n demo -l app.kubernetes.io/instance=neo4j-test
```

Expected output:

```
NAME                     TYPE              STATUS       AGE
neo4j-volume-expansion   VolumeExpansion   Successful   2m10s

NAME                STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-neo4j-test-0   Bound    ...      4Gi        RWO            longhorn       8m
data-neo4j-test-1   Bound    ...      4Gi        RWO            longhorn       8m
data-neo4j-test-2   Bound    ...      4Gi        RWO            longhorn       8m
```

> ✅ All PVCs now show `4Gi` — the expansion is complete.

You can also verify the available disk from inside a container:

```bash
kubectl exec -n demo neo4j-test-0 -- df -h /data
```

---

## Understanding the OpsRequest Fields

| Field | Description |
|---|---|
| `spec.type` | `VolumeExpansion` — identifies this as a storage resize operation |
| `spec.databaseRef.name` | Name of the target `Neo4j` resource |
| `spec.volumeExpansion.mode` | `Online` (no restart) or `Offline` (pod stopped during resize) |
| `spec.volumeExpansion.server` | Target size for each Neo4j server PVC — must be larger than the current size |

---

## Troubleshooting

If this OpsRequest does not finish, first inspect the affected PVC and then check the `kubedb-ops-manager` operator logs for the exact error. For a shared checklist, see the [Neo4j Ops Request Overview](/docs/guides/neo4j/concepts/opsrequest.md#troubleshooting).

**OpsRequest stays in `Progressing` — PVC capacity does not change**

First, confirm the PVC resize was actually sent:

```bash
kubectl describe pvc -n demo data-neo4j-test-0
```

Look for a `FileSystemResizePending` condition. If it is there, the volume backend accepted the resize but the filesystem inside the pod has not updated yet. If it has not happened after a few minutes, restart the pod:

```bash
kubectl delete pod -n demo neo4j-test-0
```

**`Longhorn volume driver not found` or PVC stays in `Pending`**

Verify the storage backend and the storage class:

```bash
kubectl get storageclass longhorn
kubectl get pods -n longhorn-system
```

If the storage class is missing, install Longhorn and re-deploy the Neo4j cluster.

**OpsRequest moves to `Failed` with error `volume expansion not supported`**

This means the storage class does not have `allowVolumeExpansion: true`. Check it and patch the storage class if needed:

```bash
kubectl get storageclass longhorn -o jsonpath='{.allowVolumeExpansion}'
kubectl patch storageclass longhorn -p '{"allowVolumeExpansion": true}'
```

Then re-apply the OpsRequest.

**OpsRequest moves to `Failed` — read the exact reason**

Use the OpsRequest status and the `kubedb-ops-manager` logs together:

```bash
kubectl get neo4jopsrequest -n demo neo4j-volume-expansion -o jsonpath='{.status.conditions}' | jq .
kubectl logs -n <kubedb-namespace> -l app.kubernetes.io/name=kubedb-ops-manager --tail=50
```

---

## Cleanup

```bash
kubectl delete neo4jopsrequest -n demo neo4j-volume-expansion
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```

---

## Next Steps

- [Neo4j Vertical Scaling](/docs/guides/neo4j/scaling/vertical-scaling/) — adjust CPU and memory
- [Neo4j Horizontal Scaling](/docs/guides/neo4j/scaling/horizontal-scaling/) — add or remove cluster members
- [Neo4j Version Upgrade](/docs/guides/neo4j/update-version/) — roll to a newer Neo4j release