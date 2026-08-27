---
title: Autoscale Neo4j Compute Resources
menu:
  docs_{{ .version }}:
    identifier: neo4j-compute-autoscaling-guide
    name: Autoscale Compute Resources
    parent: neo4j-compute-autoscaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscale Neo4j Compute Resources

This guide deploys a Neo4j cluster, loads a small social graph, and configures KubeDB to adjust CPU and memory automatically. After scaling, we query the graph again to verify that the data remains available.

## Before You Begin

| Requirement | Details |
|---|---|
| KubeDB | Provisioner, Ops Manager, and Autoscaler operators must be installed. |
| Metrics Server | Install [Kubernetes Metrics Server](https://github.com/kubernetes-sigs/metrics-server#installation) so the recommender can observe pod usage. |
| Storage | This example uses a `longhorn` StorageClass; substitute another available class if necessary. |
| Tools | `kubectl`, `jq`, and `base64` must be available locally. |

See [Neo4jAutoscaler](/docs/guides/neo4j/concepts/autoscaler.md) and the [compute autoscaling overview](/docs/guides/neo4j/autoscaler/compute/overview.md) for background.

## Deploy Neo4j

Create an isolated namespace. The examples in this guide use the `longhorn` StorageClass. If your cluster uses a different StorageClass, change `spec.storage.storageClassName` before applying the manifest.

```bash
$ kubectl create namespace demo
namespace/demo created
```

The following manifest creates a three-member Neo4j cluster. Neo4j requires at least `2Gi` of storage per member, so each pod receives its own `2Gi` persistent volume:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-autoscale
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  podTemplate:
    spec:
      containers:
        - name: neo4j
          resources:
            requests:
              cpu: 500m
              memory: 2Gi
            limits:
              cpu: 500m
              memory: 2Gi
  deletionPolicy: WipeOut
```

Here, `spec.version` selects an installed `Neo4jVersion`, `replicas: 3` creates a fault-tolerant cluster, and `storageType: Durable` preserves data across pod restarts. `deletionPolicy: WipeOut` removes the database-owned PVCs and credentials when the Neo4j resource is deleted, so use a safer deletion policy when retention is required.

Apply the same manifest from the examples directory:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/autoscaler/neo4j.yaml
neo4j.kubedb.com/neo4j-autoscale created
```

Wait for the cluster to become ready:

```bash
$ kubectl get neo4j -n demo neo4j-autoscale -w
NAME              VERSION     STATUS   AGE
neo4j-autoscale   2025.12.1   Ready    3m
```

The Neo4j container initially requests and limits `500m` CPU and `2Gi` memory. Neo4j needs enough memory for the JVM, page cache, and native allocations; limits that are too small can cause the process to be OOM-killed.

```bash
$ kubectl get pod -n demo neo4j-autoscale-0 \
    -o jsonpath='{.spec.containers[?(@.name=="neo4j")].resources}' | jq .
{
  "limits": {"cpu": "500m", "memory": "2Gi"},
  "requests": {"cpu": "500m", "memory": "2Gi"}
}
```

## Create a Sample Graph

Read the generated admin password, create an application database, and wait until it is online:

```bash
$ PASS=$(kubectl get secret -n demo neo4j-autoscale-auth \
    -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -u neo4j -p "$PASS" \
    "CREATE DATABASE appdb IF NOT EXISTS WAIT"
```

Create a uniqueness constraint, then use `MERGE` to load users and `FOLLOWS` relationships. These commands are safe to repeat because they do not create duplicate users or relationships:

```bash
$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "CREATE CONSTRAINT user_id IF NOT EXISTS
     FOR (u:User) REQUIRE u.id IS UNIQUE"

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "UNWIND range(1,10000) AS i
     MERGE (u:User {id: i})
     SET u.name = 'user-' + toString(i)"

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "UNWIND range(1,9999) AS i
     MATCH (a:User {id: i}), (b:User {id: i + 1})
     MERGE (a)-[:FOLLOWS]->(b)"
```

Verify the initial graph:

```bash
$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "MATCH (u:User) OPTIONAL MATCH (u)-[r:FOLLOWS]->()
     RETURN count(DISTINCT u) AS users, count(r) AS follows"
users, follows
10000, 9999
```

## Create the Neo4jAutoscaler

The example policy permits recommendations from `600m` to `2` CPU and from `2500Mi` to `4Gi` memory:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: Neo4jAutoscaler
metadata:
  name: neo4j-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: neo4j-autoscale
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
    maxRetries: 3
  compute:
    neo4j:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 2500Mi
      maxAllowed:
        cpu: "2"
        memory: 4Gi
      controlledResources:
        - cpu
        - memory
      containerControlledValues: RequestsAndLimits
```

Apply it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/autoscaler/compute/neo4j-compute-autoscaler.yaml
neo4jautoscaler.autoscaling.kubedb.com/neo4j-compute-autoscaler created
```

The minimum values are deliberately higher than the initial allocation, making this tutorial reproducible. In production, choose bounds based on workload requirements and capacity.

## Observe the Recommendation and Scaling

Run a read workload while the recommender gathers samples:

```bash
$ for i in $(seq 1 100); do
    kubectl exec -n demo neo4j-autoscale-0 -- \
      cypher-shell -d appdb -u neo4j -p "$PASS" \
      "MATCH (u:User)-[:FOLLOWS*1..3]->(v:User)
       RETURN count(v)" >/dev/null 2>&1 || true
  done
```

The Autoscaler may start the rolling resize while the loop is still running. The `|| true` allows this workload generator to continue past a temporary connection failure while a pod is being replaced. Application clients should use bounded retries with backoff for the same condition.

After the `podLifeTimeThreshold` has passed, inspect the recommendation:

```bash
$ kubectl get neo4jautoscaler -n demo neo4j-compute-autoscaler \
    -o jsonpath='{.status.vpas[*].recommendation.containerRecommendations}' | jq .
[
  {
    "containerName": "neo4j",
    "lowerBound": {"cpu": "600m", "memory": "2500Mi"},
    "target": {"cpu": "716m", "memory": "2500Mi"},
    "upperBound": {"cpu": "2", "memory": "4Gi"}
  }
]
```

KubeDB creates a `Neo4jOpsRequest` when the recommendation differs sufficiently from the current resources:

```bash
$ kubectl get neo4jopsrequest -n demo -w
NAME                              TYPE              STATUS       AGE
neoops-neo4j-autoscale-xxxxxx     VerticalScaling   Successful   2m
```

Verify that the pod allocation is now within the configured bounds:

```bash
$ kubectl get pod -n demo neo4j-autoscale-0 \
    -o jsonpath='{.spec.containers[?(@.name=="neo4j")].resources}' | jq .
{
  "limits": {"cpu": "600m", "memory": "2500Mi"},
  "requests": {"cpu": "600m", "memory": "2500Mi"}
}
```

The precise recommendation and completion time depend on observed usage and the available samples. The applied allocation will stay between `600m` and `2` CPU and between `2500Mi` and `4Gi` memory.

## Verify the Graph

Confirm that the database still contains the users and relationships after the scaling rollout:

```bash
$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "MATCH (u:User) OPTIONAL MATCH (u)-[r:FOLLOWS]->()
     RETURN count(DISTINCT u) AS users, count(r) AS follows"
users, follows
10000, 9999
```

## Troubleshooting

- If no recommendation appears, verify Metrics Server with `kubectl top pod -n demo` and wait for more samples.
- Check the Autoscaler's conditions with `kubectl describe neo4jautoscaler -n demo neo4j-compute-autoscaler` before changing its thresholds.
- If no OpsRequest is created, check `podLifeTimeThreshold`, `resourceDiffPercentage`, and the Autoscaler conditions.
- If an operation remains pending, describe it with `kubectl describe neo4jopsrequest -n demo <name>` and check Ops Manager logs.

## Cleaning Up

```bash
$ kubectl delete neo4jautoscaler -n demo neo4j-compute-autoscaler
$ kubectl delete neo4j -n demo neo4j-autoscale
$ kubectl delete namespace demo
```
