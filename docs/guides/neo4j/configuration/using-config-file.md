---
title: Run Neo4j with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: neo4j-using-config-file
    name: Config File
    parent: neo4j-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Neo4j. This tutorial will show you how to use KubeDB to run Neo4j with custom configuration.

## Before You Begin

> Prerequisites: A running Kubernetes cluster with KubeDB installed. See the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md) if you need to set up your environment.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Overview

Neo4j supports configuration via key-value pairs. KubeDB uses `spec.configuration.secretName` to allow users to provide a custom configuration Secret. The operator merges these key-value entries into the Neo4j configuration and restarts the cluster automatically.

In this tutorial, we will configure `dbms.logs.query.enabled`, `dbms.logs.query.parameter_logging`, and JVM options via `server.jvm.additional`.

## Custom Configuration

KubeDB expects the custom configuration Secret to use `stringData` where each key is a Neo4j configuration property name and the corresponding value is its setting. Create the Secret:

```yaml
apiVersion: v1
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
kind: Secret
metadata:
  name: neo4j-configuration
  namespace: demo
```

```bash
$ kubectl apply -f neo4j-configuration-secret.yaml
secret/neo4j-configuration created
```

Verify the Secret was created:

```bash
$ kubectl get secret -n demo neo4j-configuration
NAME                 TYPE     DATA   AGE
neo4j-configuration  Opaque   3      10s
```

Now, create the Neo4j CRD specifying `spec.configuration.secretName`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: custom-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  configuration:
    secretName: neo4j-configuration
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/configuration/neo4j-configuration.yaml
neo4j.kubedb.com/custom-neo4j created
```

Now, wait for the Neo4j cluster to be ready:

```bash
$ kubectl get neo4j -n demo custom-neo4j -w
NAME           VERSION     STATUS   AGE
custom-neo4j   2025.12.1   Ready    3m
```

## Verify the Applied Configuration

To confirm the settings are active, connect to Neo4j via `cypher-shell` and run a `SHOW SETTINGS` query. First, get the default auth credentials:

```bash
$ kubectl get secret -n demo custom-neo4j-auth \
    -o jsonpath='{.data.password}' | base64 -d
<your-password>
```

Then exec into a Neo4j pod and run `cypher-shell`:

```bash
$ kubectl exec -it -n demo custom-neo4j-0 -- \
    cypher-shell -u neo4j -p <your-password> \
    "SHOW SETTINGS
     YIELD name, value
     WHERE name STARTS WITH 'dbms.logs.query'
     RETURN name, value
     ORDER BY name;"
```

Expected output:

```
+-----------------------------------------------------------+
| name                                   | value            |
+-----------------------------------------------------------+
| "dbms.logs.query.enabled"              | "INFO"           |
| "dbms.logs.query.parameter_logging"    | "true"           |
+-----------------------------------------------------------+

2 rows
```

You can also query other setting groups. For example, to check the Neo4j data directory paths:

```bash
$ kubectl exec -it -n demo custom-neo4j-0 -- \
    cypher-shell -u neo4j -p <your-password> \
    "SHOW SETTINGS
     YIELD name, value
     WHERE name STARTS WITH 'server.jvm.additional'
     RETURN name, value
     ORDER BY name
     LIMIT 3;"
```

Expected output:

```
+-----------------------------------------------------------------------------+
| name                    | value                                             |
+-----------------------------------------------------------------------------+
| "server.jvm.additional" | "-XX:+UseG1GC                                     |
|                         \ -XX:-OmitStackTraceInFastThrow                    |
|                         \ -XX:+AlwaysPreTouch                               |
|                         \ -XX:+UnlockExperimentalVMOptions                  |
|                         \ -XX:+TrustFinalNonStaticFields                    |
|                         \ -XX:+DisableExplicitGC                            |
|                         \ -Djdk.nio.maxCachedBufferSize=1024                |
|                         \ -Dio.netty.tryReflectionSetAccessible=true        |
|                         \ -Djdk.tls.ephemeralDHKeySize=2048                 |
|                         \ -Djdk.tls.rejectClientInitiatedRenegotiation=true |
|                         \ -XX:FlightRecorderOptions=stackdepth=256          |
|                         \ -XX:+UnlockDiagnosticVMOptions                    |
|                         \ -XX:+DebugNonSafepoints                           |
|                         \ --add-opens=java.base/java.nio=ALL-UNNAMED        |
|                         \ --add-opens=java.base/java.io=ALL-UNNAMED         |
|                         \ --add-opens=java.base/sun.nio.ch=ALL-UNNAMED      |
|                         \ -Dlog4j2.disable.jmx=true"                        |
+-----------------------------------------------------------------------------+
1 row
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo neo4j/custom-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/custom-neo4j

kubectl delete -n demo secret neo4j-configuration
kubectl delete ns demo
```
