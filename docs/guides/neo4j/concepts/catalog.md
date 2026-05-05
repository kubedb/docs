---
title: Neo4jVersion CRD
menu:
  docs_{{ .version }}:
    identifier: neo4j-catalog-concepts
    name: Neo4jVersion
    parent: neo4j-concepts-neo4j
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4jVersion

## What is Neo4jVersion

`Neo4jVersion` is the catalog CRD that provides image and version metadata for KubeDB-managed Neo4j.

The value of `Neo4j.spec.version` must correspond to a valid `Neo4jVersion` resource.

## Neo4jVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: Neo4jVersion
metadata:
  name: 2025.11.2
spec:
  db:
    image: docker.io/library/neo4j:2025.11.2-enterprise
  securityContext:
    runAsUser: 7474
  version: 2025.11.2-enterprise
```

## Key fields

- `metadata.name` is used in `Neo4j.spec.version`.
- `spec.version` identifies the Neo4j release represented.
- `spec.db.image` defines the image for Neo4j pods.
- `spec.deprecated` signals if the version should be avoided.

## Next Steps

- Read the [Neo4j CRD concept](/docs/guides/neo4j/concepts/neo4j.md).
- Run the [Neo4j quickstart](/docs/guides/neo4j/quickstart/quickstart.md).