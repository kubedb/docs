---
title: Dynamic Neo4j Configuration
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-dynamic
    name: Dynamic Configuration
    parent: neo4j-reconfigure
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Dynamic Neo4j Configuration

This guide applies a dynamic setting through a `Reconfigure` OpsRequest and verifies that the database Pods are not replaced.

## Before You Begin

- Install KubeDB Provisioner and Ops-manager versions that support dynamic reconfiguration, following the [setup guide](/docs/setup/README.md).
- Start with a Ready `Neo4j` database named `neo4j-test` in namespace `demo`. You can use the database from the [Neo4j reconfiguration guide](/docs/guides/neo4j/reconfigure/reconfigure.md).
- Read the [reconfiguration overview](/docs/guides/neo4j/reconfigure/overview.md) for restart rules.
- Ensure no other operation is changing the database during this example.

## Record the Current Pods

Save the Pod names and UIDs before applying the request. The selector below is shared by the database Pods managed by KubeDB:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test \
  -o custom-columns=NAME:.metadata.name,UID:.metadata.uid --no-headers | sort > /tmp/neo4j-pods-before.txt
cat /tmp/neo4j-pods-before.txt
```

## Check Whether the Setting Is Dynamic

Connect to a server and inspect the setting:

```bash
PASS=$(kubectl get secret -n demo neo4j-test-auth -o jsonpath='{.data.password}' | base64 -d)
kubectl exec -n demo neo4j-test-0 -- cypher-shell -u neo4j -p "$PASS" \
  "SHOW SETTINGS YIELD name, value, isDynamic WHERE name = 'db.checkpoint.interval.time' RETURN name, value, isDynamic;"
```

Proceed when `isDynamic` is `true` and the current value differs from `30m`. Otherwise, choose a different value or a setting that the running version reports as dynamic. The operator reads this metadata from the running server rather than maintaining a fixed list.

## Apply the Dynamic Configuration

Save the following manifest as `neo4j-reconfigure-dynamic.yaml`. It is also available in [the examples directory](/docs/examples/neo4j/reconfigure/neo4j-reconfigure-dynamic.yaml).

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-reconfigure-dynamic
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: neo4j-test
  configuration:
    restart: "false"
    applyConfig:
      db.checkpoint.interval.time: "30m"
  timeout: 5m
  apply: IfReady
```

`restart: "false"` makes this request fail if the change requires a restart. Omit it or use `restart: "auto"` to allow the operator to restart Pods when a change cannot be applied live. Use `restart: "true"` to request a restart explicitly.

```bash
kubectl apply -f neo4j-reconfigure-dynamic.yaml
kubectl wait -n demo neo4jopsrequest/neo4j-reconfigure-dynamic \
  --for=jsonpath='{.status.phase}'=Successful --timeout=600s
```

If the wait times out, inspect the request before proceeding:

```bash
kubectl describe neo4jopsrequest -n demo neo4j-reconfigure-dynamic
```

## Verify Live Application

Check the mode selected by the operator:

```bash
kubectl get neo4jopsrequest -n demo neo4j-reconfigure-dynamic \
  -o jsonpath='{range .status.conditions[?(@.type=="ReconfigureMode")]}{.reason}{"\n"}{.message}{"\n"}{end}'
```

For a successful dynamic update, the reason is `Live`. A no-change request also uses `Live`, with a message explaining that no setting changed.

Check the effective value on every server. For the three-server database from the reconfiguration guide:

```bash
for POD in neo4j-test-0 neo4j-test-1 neo4j-test-2; do
  echo "$POD"
  kubectl exec -n demo "$POD" -- cypher-shell -u neo4j -p "$PASS" \
    "SHOW SETTINGS YIELD name, value, isDynamic WHERE name = 'db.checkpoint.interval.time' RETURN name, value, isDynamic;"
done
```

Each server should report `30m`. Neo4j may normalize the displayed value. KubeDB applies the change directly to every server with `dbms.setConfigValue` and verifies that the servers agree. It also updates the stored configuration, so the new setting is loaded after a future Pod replacement.

Compare the Pod UIDs with the saved values:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-test \
  -o custom-columns=NAME:.metadata.name,UID:.metadata.uid --no-headers | sort > /tmp/neo4j-pods-after.txt
diff -u /tmp/neo4j-pods-before.txt /tmp/neo4j-pods-after.txt
```

An empty diff confirms that the Pods were not replaced during the operation. You can also inspect container restart counts with `kubectl get pods -n demo`.

## Clean Up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-reconfigure-dynamic
```

Deleting the OpsRequest does not undo the configuration change. Use another reconfiguration request to restore the previous value. Removing a known setting from the user configuration requires a restart, even if its value can be changed dynamically.
