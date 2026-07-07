---
title: Reconfigure Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-cluster
    name: Cluster
    parent: neo4j-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Neo4j

This guide shows how to reconfigure a Neo4j database using `Neo4jOpsRequest`.
We will:

1. Deploy Neo4j with default configuration.
2. Check current values from Neo4j using `cypher-shell`.
3. Create a custom config secret.
4. Apply a `Reconfigure` OpsRequest.
5. Verify changed values and show `applyConfig` precedence over `configSecret`.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured.
- Install KubeDB Community and Enterprise operators following [the setup guide](/docs/setup/README.md).
- Review [Neo4j](/docs/guides/neo4j/concepts/neo4j.md), [OpsRequest](/docs/guides/neo4j/concepts/opsrequest.md), and [Reconfigure Overview](/docs/guides/neo4j/reconfigure/overview.md).

```bash
kubectl create ns demo
```
namespace/demo created

## Prepare Database

First, deploy Neo4j with default configuration (no custom config secret):

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
EOF
```
neo4j.kubedb.com/neo4j-test created

```bash
kubectl wait --for=condition=Ready neo4j/neo4j-test -n demo --timeout=600s
```
neo4j.kubedb.com/neo4j-test condition met

## Check Current Settings (Before Reconfigure)

Before applying the reconfigure request, connect with `cypher-shell` and check current values:

```bash
PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -it -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS"
```

These are the current values before running the reconfigure OpsRequest.

Check `db.query` settings:

```bash
neo4j@neo4j> SHOW SETTINGS
             YIELD name, value
             WHERE name STARTS WITH 'db.query'
             RETURN name, value
             ORDER BY name
             LIMIT 10;
+-------------------------------------------+
| name                        | value       |
+-------------------------------------------+
| "db.query.default_language" | "CYPHER_25" |
+-------------------------------------------+

1 row
ready to start consuming query after 12 ms, results consumed after another 1 ms
```

Check `db.checkpoint` settings:

```bash
neo4j@neo4j> SHOW SETTINGS
             YIELD name, value
             WHERE name STARTS WITH 'db'
             RETURN name, value
             ORDER BY name
             LIMIT 3;
+--------------------------------------------+
| name                          | value      |
+--------------------------------------------+
| "db.checkpoint"               | "PERIODIC" |
| "db.checkpoint.interval.time" | "15m"      |
| "db.checkpoint.interval.tx"   | "100000"   |
+--------------------------------------------+

3 rows
ready to start consuming query after 11 ms, results consumed after another 1 ms
```

Check `server.jvm` settings:

```bash
neo4j@neo4j> SHOW SETTINGS
             YIELD name, value
             WHERE name STARTS WITH 'server.jvm'
             RETURN name, value
             ORDER BY name
             LIMIT 3;
+---------------------------------+
| name                    | value |
+---------------------------------+
| "server.jvm.additional" | NULL  |
+---------------------------------+

1 row
ready to start consuming query after 13 ms, results consumed after another 2 ms
```

## Create Custom Configuration Secret

Create `custom-config` with your provided values:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: custom-config
  namespace: demo
stringData:
  db.query.default_language: "CYPHER_5"
  db.checkpoint.interval.time: "20m"
  server.jvm.additional: |-
    -XX:+UseG1GC
    -XX:-OmitStackTraceInFastThrow
```

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: custom-config
  namespace: demo
stringData:
  db.query.default_language: "CYPHER_5"
  db.checkpoint.interval.time: "20m"
  server.jvm.additional: |-
    -XX:+UseG1GC
    -XX:-OmitStackTraceInFastThrow
EOF
```
secret/custom-config created

## Reconfigure Request

Now apply this `Neo4jOpsRequest` exactly:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: neo4j-test
  configuration:
    configSecret:
      name: custom-config
    applyConfig:
      db.checkpoint.interval.time: "25m"
  timeout: 5m
  apply: IfReady
```

Here,

- `configSecret` provides base custom settings.
- `applyConfig` applies inline overrides.
- If the same key exists in both places, `applyConfig` takes precedence.

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: neo4j-test
  configuration:
    configSecret:
      name: custom-config
    applyConfig:
      db.checkpoint.interval.time: "25m"
  timeout: 5m
  apply: IfReady
EOF
```
neo4jopsrequest.ops.kubedb.com/reconfigure created

```bash
kubectl wait --for=jsonpath='{.status.phase}'=Successful neo4jopsrequest/reconfigure -n demo --timeout=600s
```
neo4jopsrequest.ops.kubedb.com/reconfigure condition met

## Verify Reconfiguration

Check OpsRequest status:

```bash
kubectl get neo4jopsrequest -n demo reconfigure
```
NAME          TYPE          STATUS       AGE
reconfigure   Reconfigure   Successful   2m5s

Now run the same three queries again and confirm updated values:

```bash
PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -it -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS"
```

Check `db.query` settings:

```bash
neo4j@neo4j> SHOW SETTINGS
             YIELD name, value
             WHERE name STARTS WITH 'db.query'
             RETURN name, value
             ORDER BY name
             LIMIT 10;
+-------------------------------------------+
| name                        | value       |
+-------------------------------------------+
| "db.query.default_language" | "CYPHER_5" |
+-------------------------------------------+

1 row
ready to start consuming query after 12 ms, results consumed after another 1 ms
```

Check `db.checkpoint` settings (updated):

```bash
neo4j@neo4j> SHOW SETTINGS
             YIELD name, value
             WHERE name STARTS WITH 'db'
             RETURN name, value
             ORDER BY name
             LIMIT 3;
+--------------------------------------------+
| name                          | value      |
+--------------------------------------------+
| "db.checkpoint"               | "PERIODIC" |
| "db.checkpoint.interval.time" | "25m"      |
| "db.checkpoint.interval.tx"   | "100000"   |
+--------------------------------------------+

3 rows
ready to start consuming query after 11 ms, results consumed after another 1 ms
```

Check `server.jvm` settings (updated):

```bash
neo4j@neo4j> SHOW SETTINGS
                          YIELD name, value
                          WHERE name STARTS WITH 'server.jvm'
                          RETURN name, value
                          ORDER BY name
                          LIMIT 3;
+-----------------------------------------------------------+
| name                    | value                           |
+-----------------------------------------------------------+
| "server.jvm.additional" | "-XX:+UseG1GC                   |
|                         \ -XX:-OmitStackTraceInFastThrow" |
+-----------------------------------------------------------+

1 row
ready to start consuming query after 80 ms, results consumed after another 7 ms
```

From the output:

- `db.query.default_language` comes from `custom-config` secret.
- `db.checkpoint.interval.time` is `25m` from `applyConfig`, not `20m` from secret.
- `server.jvm.additional` is now set from `custom-config` (it was `NULL` before reconfigure).
- This confirms `applyConfig` has higher precedence than `configSecret` for overlapping keys.

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo reconfigure
```
neo4jopsrequest.ops.kubedb.com "reconfigure" deleted

```bash
kubectl delete secret -n demo custom-config
```
secret "custom-config" deleted

```bash
kubectl patch -n demo neo4j/neo4j-test -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
```
neo4j.kubedb.com/neo4j-test patched

```bash
kubectl delete -n demo neo4j/neo4j-test
```
neo4j.kubedb.com "neo4j-test" deleted

```bash
kubectl delete ns demo
```
namespace "demo" deleted
