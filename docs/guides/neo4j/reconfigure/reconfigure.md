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

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured.
- Install KubeDB Community and Enterprise operators following [the setup guide](/docs/setup/README.md).
- Review [Neo4j](/docs/guides/neo4j/concepts/neo4j.md), [OpsRequest](/docs/guides/neo4j/concepts/opsrequest.md), and [Reconfigure Overview](/docs/guides/neo4j/reconfigure/overview.md).

```bash
kubectl create ns demo
```

## Prepare Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  configuration:
    secretName: neo4j-config
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

## Create Custom Configuration Secret

Use a Secret like this for `spec.configuration.configSecret.name`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: custom-config
  namespace: demo
stringData:
  dbms.logs.query.enabled: "INFO"
  dbms.logs.query.parameter_logging: "true"
  server.jvm.additional: |-
    -XX:+UseG1GC
    -XX:-OmitStackTraceInFastThrow
    -XX:+AlwaysPreTouch
    -XX:+UnlockExperimentalVMOptions
    -XX:+TrustFinalNonStaticFields
    -XX:+DisableExplicitGC
    -Djdk.nio.maxCachedBufferSize=1024
    -Dio.netty.tryReflectionSetAccessible=true
    -Djdk.tls.ephemeralDHKeySize=2048
    -Djdk.tls.rejectClientInitiatedRenegotiation=true
    -XX:FlightRecorderOptions=stackdepth=256
    -XX:+UnlockDiagnosticVMOptions
    -XX:+DebugNonSafepoints
    --add-opens=java.base/java.nio=ALL-UNNAMED
    --add-opens=java.base/java.io=ALL-UNNAMED
    --add-opens=java.base/sun.nio.ch=ALL-UNNAMED
    -Dlog4j2.disable.jmx=true
```

## Reconfigure Request

Apply a `Neo4jOpsRequest` with `configSecret` and inline `applyConfig`:

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
      server.metrics.enabled: "false"
  timeout: 5m
  apply: IfReady
```

```bash
$ cat <<'EOF' | kubectl apply -f -
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
      server.metrics.enabled: "false"
  timeout: 5m
  apply: IfReady
EOF
neo4jopsrequest.ops.kubedb.com/reconfigure created

$ kubectl wait --for=jsonpath='{.status.phase}'=Successful neo4jopsrequest/reconfigure -n demo --timeout=600s
neo4jopsrequest.ops.kubedb.com/reconfigure condition met
```

## Verify Reconfiguration

This request was applied successfully:

```bash
$ kubectl get neo4jopsrequest -n demo reconfigure
NAME          TYPE          STATUS       AGE
reconfigure   Reconfigure   Successful   2m5s
```

Check that the updated configuration is visible from Neo4j using `cypher-shell`:

```bash
$ PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS" "SHOW SETTINGS YIELD name, value WHERE name STARTS WITH 'server.metrics' RETURN name, value ORDER BY name LIMIT 3;"
name, value
"server.metrics.enabled", "false"
"server.metrics.csv.enabled", "false"
"server.metrics.csv.interval", "5000"
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo reconfigure
kubectl patch -n demo neo4j/neo4j-test -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/neo4j-test
kubectl delete ns demo
```
