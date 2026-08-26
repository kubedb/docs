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

Create an isolated namespace and apply the example Neo4j resource:

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/autoscaler/neo4j.yaml
neo4j.kubedb.com/neo4j-autoscale created
```

Wait for the cluster to become ready:

```bash
$ kubectl get neo4j -n demo neo4j-autoscale -w
NAME              VERSION     STATUS   AGE
neo4j-autoscale   2025.12.1   Ready    3m
```

The Neo4j container initially requests and limits `500m` CPU and `1Gi` memory:

```bash
$ kubectl get pod -n demo neo4j-autoscale-0 \
    -o jsonpath='{.spec.containers[?(@.name=="neo4j")].resources}' | jq .
{
  "limits": {"cpu": "500m", "memory": "1Gi"},
  "requests": {"cpu": "500m", "memory": "1Gi"}
}
```

## Create a Sample Graph

Read the generated admin password, create an application database, and insert users connected by `FOLLOWS` relationships:

```bash
$ PASS=$(kubectl get secret -n demo neo4j-autoscale-auth \
    -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -u neo4j -p "$PASS" \
    "CREATE DATABASE appdb IF NOT EXISTS WAIT"

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "CREATE CONSTRAINT user_id IF NOT EXISTS
     FOR (u:User) REQUIRE u.id IS UNIQUE"

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "UNWIND range(1,10000) AS i
     CREATE (:User {id: i, name: 'user-' + toString(i)})"

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "UNWIND range(1,9999) AS i
     MATCH (a:User {id: i}), (b:User {id: i + 1})
     CREATE (a)-[:FOLLOWS]->(b)"
```

Verify the initial graph:

```bash
$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "MATCH (u:User) RETURN count(u) AS users"
users
10000
```

## Create the Neo4jAutoscaler

The example policy permits recommendations from `600m` to `2` CPU and from `1200Mi` to `2Gi` memory:

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
        memory: 1200Mi
      maxAllowed:
        cpu: "2"
        memory: 2Gi
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
       RETURN count(v)" >/dev/null
  done
```

After the `podLifeTimeThreshold` has passed, inspect the recommendation:

```bash
$ kubectl get neo4jautoscaler -n demo neo4j-compute-autoscaler \
    -o jsonpath='{.status.vpas[*].recommendation.containerRecommendations}' | jq .
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
```

The precise recommendation depends on observed usage, but it will not be below `600m` CPU and `1200Mi` memory or above `2` CPU and `2Gi` memory.

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
- If no OpsRequest is created, check `podLifeTimeThreshold`, `resourceDiffPercentage`, and the Autoscaler conditions.
- If an operation remains pending, describe it with `kubectl describe neo4jopsrequest -n demo <name>` and check Ops Manager logs.

## Cleaning Up

```bash
$ kubectl delete neo4jautoscaler -n demo neo4j-compute-autoscaler
$ kubectl delete neo4j -n demo neo4j-autoscale
$ kubectl delete namespace demo
```
