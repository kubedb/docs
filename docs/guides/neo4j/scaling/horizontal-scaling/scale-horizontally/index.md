---
title: Scale Neo4j Horizontally
menu:
  docs_{{ .version }}:
    identifier: neo4j-scale-horizontally
    name: Scale Horizontally
    parent: neo4j-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> 🆕 New to KubeDB? Start with the [quickstart guide](/docs/README.md) before continuing here.

# Neo4j Horizontal Scaling with KubeDB

Horizontally scaling a Neo4j cluster means adding or removing cluster members (pods) at runtime — without downtime and without data loss. KubeDB handles the orchestration automatically through `Neo4jOpsRequest`.

This guide walks you through:
- Seeding a user database with test data
- Scaling **up** from 3 → 5 members
- Scaling **down** from 5 → 4 members
- Verifying cluster topology and data integrity at each step

---

## How It Works

KubeDB uses the `Neo4jOpsRequest` custom resource to perform horizontal scaling operations. Under the hood it:

1. Adds (or drains) StatefulSet replicas
2. Enables / deallocates the new/removed Neo4j server in the cluster
3. Waits for database reallocation to complete before marking the operation `Successful`

You verify the result using two complementary Cypher views:

| Cypher View | What It Shows |
|---|---|
| `SHOW DATABASE <name>` | Database allocation status, primary/secondary count |
| `SHOW SERVERS` | Which servers host which databases |

> **Tip (Neo4j 2025.x):** Always use **both** views together for a reliable picture of cluster topology after a scale event.

---

## Prerequisites

| Requirement | Details |
|---|---|
| KubeDB installed | Operator running in your cluster |
| Neo4j instance | `status.phase=Ready` |
| `kubectl` access | With permissions to the `demo` namespace |

---

## Step 1 — Set Up the Namespace

```bash
kubectl create ns demo
```

---

## Step 2 — Deploy Neo4j

Apply the example manifest and wait for the cluster to become ready:

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml

kubectl get neo4j -n demo neo4j-test -w
```

Wait until `STATUS` shows `Ready` before proceeding.

---

## Step 3 — Create a Database and Seed Data

Open a shell into pod `neo4j-test-0` and run the following Cypher commands to:
- Create a new database called `appdb`
- Insert 2,000 test `User` nodes
- Verify the count

```bash
# Retrieve the admin password
PASS=$(kubectl get secret -n demo neo4j-test-auth \
  -o jsonpath='{.data.password}' | base64 -d)

# Create the database
kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -u neo4j -p "$PASS" \
  "CREATE DATABASE appdb IF NOT EXISTS WAIT"

# Seed 2,000 User nodes
kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -d appdb -u neo4j -p "$PASS" \
  "UNWIND range(1,2000) AS i CREATE (:User {id:i, name:'user-'+toString(i)})"

# Confirm the count
kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -d appdb -u neo4j -p "$PASS" \
  "MATCH (u:User) RETURN count(u) AS totalUsers"
```

Expected output:

```
totalUsers
2000
```

---

## Step 4 — Scale Up (3 → 5 Members)

Apply the scale-up ops request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: neo4j-test
  horizontalScaling:
    server: 5
    reallocate:
      strategy: "incremental"
      batchSize: 1
```

```bash
$ cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: neo4j-test
  horizontalScaling:
    server: 5
    reallocate:
      strategy: "incremental"
      batchSize: 1
EOF
neo4jopsrequest.ops.kubedb.com/neo4j-horizontal-scale-up created

$ kubectl wait \
  --for=jsonpath='{.status.phase}'=Successful \
  neo4jopsrequest/neo4j-horizontal-scale-up \
  -n demo --timeout=900s
neo4jopsrequest.ops.kubedb.com/neo4j-horizontal-scale-up condition met
```

### Verify the Scale-Up

Run these commands to confirm the cluster now has 5 members and that databases have been reallocated:

```bash
$ kubectl get neo4jopsrequest -n demo neo4j-horizontal-scale-up
NAME                        TYPE                STATUS       AGE
neo4j-horizontal-scale-up   HorizontalScaling   Successful   16s

$ kubectl get neo4j -n demo neo4j-test -o jsonpath='{.spec.replicas}{"\n"}'
5

$ kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test
NAME            READY   STATUS    RESTARTS   AGE
neo4j-test-0    1/1     Running   0          ...
neo4j-test-1    1/1     Running   0          ...
neo4j-test-2    1/1     Running   0          ...
neo4j-test-3    1/1     Running   0          ...
neo4j-test-4    1/1     Running   0          ...

$ PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS" \
  "SHOW DATABASE appdb YIELD name, currentStatus, currentPrimariesCount, currentSecondariesCount RETURN name, currentStatus, currentPrimariesCount, currentSecondariesCount"
name, currentStatus, currentPrimariesCount, currentSecondariesCount
"appdb", "online", 2, 0
"appdb", "online", 2, 0

$ kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS" \
  "SHOW SERVERS YIELD name, state, health, hosting RETURN name, state, health, hosting ORDER BY name"
name, state, health, hosting
"neo4j-test-0", "Enabled", "Available", ["neo4j", "system"]
"neo4j-test-1", "Enabled", "Available", ["neo4j", "system"]
"neo4j-test-2", "Enabled", "Available", ["appdb", "system"]
"neo4j-test-3", "Enabled", "Available", ["appdb", "system"]
"neo4j-test-4", "Enabled", "Available", ["system"]
```

> **What to look for:**
> - All 5 pods are `Running`
> - `appdb` shows `currentStatus: online`
> - `SHOW SERVERS` shows new servers hosting databases (reallocation is complete)

---

## Step 5 — Scale Down (5 → 3 Members)

Apply the scale-down ops request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: neo4j-test
  horizontalScaling:
    server: 3
    reallocate:
      strategy: "full"
```

```bash
$ cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: neo4j-test
  horizontalScaling:
    server: 3
    reallocate:
      strategy: "full"
EOF
neo4jopsrequest.ops.kubedb.com/neo4j-horizontal-scale-down created

$ kubectl wait \
  --for=jsonpath='{.status.phase}'=Successful \
  neo4jopsrequest/neo4j-horizontal-scale-down \
  -n demo --timeout=900s
neo4jopsrequest.ops.kubedb.com/neo4j-horizontal-scale-down condition met
```

### Verify the Scale-Down

```bash
$ kubectl get neo4jopsrequest -n demo neo4j-horizontal-scale-down
NAME                          TYPE                STATUS       AGE
neo4j-horizontal-scale-down   HorizontalScaling   Successful   37s

$ kubectl get neo4j -n demo neo4j-test -o jsonpath='{.spec.replicas}{"\n"}'
3

$ kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test
NAME            READY   STATUS    RESTARTS   AGE
neo4j-test-0    1/1     Running   0          ...
neo4j-test-1    1/1     Running   0          ...
neo4j-test-2    1/1     Running   0          ...

$ PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -n demo neo4j-test-0 -- \
  cypher-shell -d appdb -u neo4j -p "$PASS" \
  "MATCH (u:User) RETURN count(u) AS totalUsers"
totalUsers
2000

$ kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS" \
  "SHOW DATABASE appdb YIELD name, currentStatus, currentPrimariesCount, currentSecondariesCount RETURN name, currentStatus, currentPrimariesCount, currentSecondariesCount"
name, currentStatus, currentPrimariesCount, currentSecondariesCount
"appdb", "online", 2, 0
"appdb", "online", 2, 0


$ kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS" \
  "SHOW SERVERS YIELD name, state, health, hosting RETURN name, state, health, hosting ORDER BY name"
name, state, health, hosting
"neo4j-test-0", "Enabled", "Available", ["neo4j", "system"]
"neo4j-test-1", "Enabled", "Available", ["appdb", "neo4j", "system"]
"neo4j-test-2", "Enabled", "Available", ["appdb", "system"]
```

> ✅ The `totalUsers: 2000` result confirms **no data was lost** during the scale-down. The database remained online and queryable throughout.

---

## Understanding the Output

### `SHOW DATABASE` Fields

| Field | Description |
|---|---|
| `name` | Database name |
| `currentStatus` | `online` means the database is healthy and accepting queries |
| `currentPrimariesCount` | Number of primary copies currently allocated |
| `currentSecondariesCount` | Number of secondary (read replica) copies |

> One row appears per server that hosts the database. Seeing `"appdb", "online", 2, 0` twice means two servers each confirm they hold a primary copy.

### `SHOW SERVERS` Fields

| Field | Description |
|---|---|
| `name` | Pod name (maps directly to Kubernetes pod) |
| `state` | `Enabled` = active cluster member |
| `health` | `Available` = healthy and reachable |
| `hosting` | List of databases currently allocated to this server |

---

## Troubleshooting

**Ops request stuck in `Progressing`**

Check the KubeDB operator logs and Neo4j pod events:

```bash
kubectl describe neo4jopsrequest -n demo <ops-request-name>
kubectl logs -n <kubedb-namespace> <kubedb-operator-pod>
```

**Database shows `offline` after scaling**

Neo4j may need time to reallocate. Wait a few seconds and re-run `SHOW DATABASE`. If it persists, check pod readiness:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test
```

**Password retrieval fails**

Ensure the secret name matches your Neo4j instance name:

```bash
kubectl get secrets -n demo | grep neo4j
```

---

## Cleanup

Remove all resources created in this guide:

```bash
kubectl delete neo4jopsrequest -n demo \
  neo4j-horizontal-scale-up \
  neo4j-horizontal-scale-down

kubectl delete neo4j -n demo neo4j-test

kubectl delete ns demo
```

---

## Next Steps

- [Neo4j Vertical Scaling](/docs/guides/neo4j/scaling/vertical-scaling/) — adjust CPU and memory
- [Neo4j TLS Configuration](/docs/guides/neo4j/tls/) — enable encrypted connections
- [Neo4j Backup & Restore](/docs/guides/neo4j/backup/) — schedule automated backups
