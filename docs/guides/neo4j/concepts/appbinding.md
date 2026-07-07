---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: neo4j-appbinding-concepts
    name: AppBinding
    parent: neo4j-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition` (CRD) that points to an application endpoint and its access credentials.

If you deploy a Neo4j database using KubeDB, KubeDB automatically creates an `AppBinding` for that database. This object is used by tools like KubeStash to discover connection information and database credentials.

## AppBinding CRD Specification

Like other Kubernetes resources, an `AppBinding` has `TypeMeta`, `ObjectMeta`, and `Spec` sections. It does not have a `Status` section.

An `AppBinding` created by KubeDB for a Neo4j database looks like this:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: neo4j-test
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: neo4j-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Neo4j
    name: neo4j-test
    namespace: demo
  clientConfig:
    service:
      name: neo4j-test
      port: 7687
      scheme: neo4j
  secret:
    name: neo4j-test-auth
  type: kubedb.com/Neo4j
  version: 2025.12.1-enterprise
```

Here, we describe the important sections of this AppBinding.

### spec.type

`spec.type` identifies the app type represented by this AppBinding.

Format: `<group>/<kind>`.

For Neo4j managed by KubeDB, it is typically:

- `kubedb.com/Neo4j`

### spec.appRef

`spec.appRef` points back to the source database object that owns this binding.

For Neo4j, this includes:

- `apiGroup: kubedb.com`
- `kind: Neo4j`
- `name: <neo4j-name>`
- `namespace: <namespace>`

### spec.secret

`spec.secret` references the Secret that stores credentials required to connect to the database. The Secret must be in the same namespace.

For Neo4j, KubeDB-generated auth secret typically contains:

| Key | Usage |
|-----|-------|
| `username` | Neo4j user name |
| `password` | Password for that user |

### spec.clientConfig

`spec.clientConfig` defines how clients should connect to the target database.

For in-cluster Neo4j deployments, KubeDB sets `spec.clientConfig.service`.

#### spec.clientConfig.service

- `name`: Kubernetes Service name for the database.
- `port`: Service port used for client connection (Neo4j Bolt is commonly `7687`).
- `scheme`: Connection scheme (for example, `neo4j`).

## Verify AppBinding

You can inspect the generated AppBinding with:

```bash
kubectl get appbinding -n demo neo4j-test -o yaml
```

## Next Steps

- Read the [Neo4j CRD concept](/docs/guides/neo4j/concepts/neo4j.md).
- Learn Neo4j operations from [Neo4j OpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).
- Run the [Neo4j quickstart](/docs/guides/neo4j/quickstart/quickstart.md).

