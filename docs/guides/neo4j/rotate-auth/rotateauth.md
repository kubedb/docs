---
title: Rotate Auth of Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-rotate-auth-cluster
    name: Cluster
    parent: neo4j-rotate-auth
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> 🆕 New to KubeDB? Start with the [quickstart guide](/docs/README.md) before continuing here.

# Neo4j Auth Rotation with KubeDB

Auth rotation replaces the Neo4j admin password without downtime. KubeDB updates the Kubernetes Secret and rotates the credential inside the running cluster automatically through `Neo4jOpsRequest`.

This guide covers two rotation modes:

| Mode | When to use |
|---|---|
| **KubeDB-generated** | Let KubeDB create a new random password |
| **User-provided** | Supply your own password via a Kubernetes Secret |

---

## How It Works

KubeDB uses `Neo4jOpsRequest` with `type: RotateAuth` to rotate credentials. Under the hood it:

1. Generates a new password — or reads one from `spec.authentication.secretRef`
2. Updates the Kubernetes auth Secret
3. Calls the Neo4j password-change API on the running cluster
4. Marks the operation `Successful` once all pods accept the new credential

---

## Prerequisites

| Requirement | Details |
|---|---|
| KubeDB installed | Provisioner and Ops-manager operators running |
| Neo4j instance | `neo4j-test` in namespace `demo`, `status.phase=Ready` |
| `kubectl` configured | With permissions to the `demo` namespace |

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

---

## Mode 1 — KubeDB-Generated Password

In this mode, KubeDB generates a new random password and stores it back in the auth Secret.

### Step 1 — Record the Current Password Hash

Capture the current base64-encoded password so you can confirm it changes after rotation.

```bash
BEFORE_B64=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}')
echo "Before: $BEFORE_B64"
```

### Step 2 — Apply the RotateAuth Request

Save this as `neo4j-rotate-auth.yaml`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: neo4j-test
  timeout: 5m
  apply: IfReady
```

Apply it and wait for completion:

```bash
kubectl apply -f neo4j-rotate-auth.yaml

kubectl wait --for=jsonpath='{.status.phase}'=Successful \
  neo4jopsrequest/neo4j-rotate-auth \
  -n demo --timeout=900s
```

Expected output:

```
neo4jopsrequest.ops.kubedb.com/neo4j-rotate-auth condition met
```

### Step 3 — Verify the Password Changed

```bash
kubectl get neo4jopsrequest -n demo neo4j-rotate-auth

AFTER_B64=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}')
[ "$BEFORE_B64" != "$AFTER_B64" ] && echo "password_changed=true" || echo "password_changed=false"

NEW_PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)
kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$NEW_PASS" "RETURN 'auth-ok' AS status"
```

Expected output:

```
NAME                TYPE         STATUS       AGE
neo4j-rotate-auth   RotateAuth   Successful   37s

password_changed=true

status
"auth-ok"
```

> ✅ The password in the Secret has changed and Neo4j accepts the new credential.

---

## Mode 2 — User-Provided Password

In this mode, you supply a Kubernetes Secret containing your chosen password. KubeDB reads it and rotates Neo4j to use it.

### Step 1 — Create the Auth Secret

```bash
kubectl create secret generic external-neo4j-auth \
  -n demo \
  --from-literal=username=neo4j \
  --from-literal=password='Neo4j@12345' \
  --dry-run=client -o yaml | kubectl apply -f -
```

> **Password requirements:** Neo4j requires the password to differ from the username and be at least 8 characters.

### Step 2 — Apply the RotateAuth Request

Save this as `neo4j-rotate-auth-user.yaml`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: neo4j-test
  authentication:
    secretRef:
      kind: Secret
      name: external-neo4j-auth
  timeout: 5m
  apply: IfReady
```

Apply it and wait for completion:

```bash
kubectl apply -f neo4j-rotate-auth-user.yaml

kubectl wait --for=jsonpath='{.status.phase}'=Successful \
  neo4jopsrequest/neo4j-rotate-auth-user \
  -n demo --timeout=900s
```

Expected output:

```
neo4jopsrequest.ops.kubedb.com/neo4j-rotate-auth-user condition met
```

### Step 3 — Verify Login with the New Password

```bash
kubectl get neo4jopsrequest -n demo neo4j-rotate-auth-user

kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -u neo4j -p 'Neo4j@12345' "RETURN 'user-auth-ok' AS status"
```

Expected output:

```
NAME                     TYPE         STATUS       AGE
neo4j-rotate-auth-user   RotateAuth   Successful   33s

status
"user-auth-ok"
```

> ✅ Neo4j now accepts the password from `external-neo4j-auth`.

---

## Understanding the OpsRequest Fields

| Field | Description |
|---|---|
| `spec.type` | `RotateAuth` — identifies this as a credential rotation operation |
| `spec.databaseRef.name` | Name of the target `Neo4j` resource |
| `spec.authentication.secretRef` | *(Optional)* Secret with `username` and `password` keys. Omit to let KubeDB generate a password. |
| `spec.timeout` | How long KubeDB waits before marking the operation failed |
| `spec.apply` | `IfReady` — only proceed if the database is in a healthy state |

---

## Troubleshooting

**Ops request stuck in `Progressing`**

The password change may be waiting on a pod that is not ready:

```bash
kubectl describe neo4jopsrequest -n demo neo4j-rotate-auth
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test
```

**Login fails after rotation**

Re-read the Secret and retry — the pod may have cached the old credential:

```bash
PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)
kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS" "RETURN 1"
```

**`apply: IfReady` causes the request to be skipped**

The database must be in `Ready` state before the operation runs. Check its status:

```bash
kubectl get neo4j -n demo neo4j-test
```

---

## Cleanup

```bash
kubectl delete neo4jopsrequest -n demo neo4j-rotate-auth neo4j-rotate-auth-user
kubectl delete secret -n demo external-neo4j-auth
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```

---

## Next Steps

- [Neo4j TLS Configuration](/docs/guides/neo4j/tls/) — encrypt traffic between clients and the cluster
- [Neo4j Vertical Scaling](/docs/guides/neo4j/scaling/vertical-scaling/) — adjust CPU and memory
- [Neo4j Version Upgrade](/docs/guides/neo4j/update-version/) — roll to a newer Neo4j release