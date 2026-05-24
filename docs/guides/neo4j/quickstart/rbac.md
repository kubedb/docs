---
title: RBAC for Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-rbac-quickstart
    name: RBAC
    parent: neo4j-quickstart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for Neo4j

When you deploy a `Neo4j` object, KubeDB **automatically** creates a `Role`, `ServiceAccount`, and `RoleBinding` for it — you do not need to create these manually. This page documents what gets created and why.

Neo4j pods need to watch `Services` and `Endpoints` in their namespace so each cluster member can discover its peers and build the routing table during startup and after pod restarts.

| Kubernetes Resource | Permission required | Why Neo4j needs it |
|---------------------|---------------------|--------------------|
| services            | get, list, watch    | Discover peer pod addresses for cluster formation |
| endpoints           | get, list, watch    | Resolve headless Service endpoints for Raft and discovery ports |

> If you want to provide your own ServiceAccount and Role instead of the auto-generated ones, see [Custom RBAC for Neo4j](/docs/guides/neo4j/custom-rbac/using-custom-rbac.md).

> Prerequisites: A running `neo4j-test` database in the `demo` namespace. If you haven't deployed one yet, follow the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md) first.

## What KubeDB Creates

When the `Neo4j` object is created, the operator provisions three RBAC resources with names matching the database name. Let's inspect each one.

### Role

KubeDB operator creates a Role object `neo4j-test-role` in the same namespace as the Neo4j object.

```yaml
$ kubectl get role -n demo neo4j-test-role -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: "2026-05-14T06:54:08Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: neo4j-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
  name: neo4j-test-role
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Neo4j
      name: neo4j-test
      uid: 0034a30c-d33d-4596-a6d8-7cf47aa3d9e6
  resourceVersion: "1461745"
  uid: 1f3850bc-4d28-4780-88ad-b31e9c7fa21e
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
    verbs:
      - get
      - list
      - watch
```

### ServiceAccount

KubeDB operator creates a ServiceAccount object `neo4j-test` in the same namespace as the Neo4j object.

```yaml
$ kubectl get serviceaccount -n demo neo4j-test -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2026-05-14T06:54:08Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: neo4j-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
  name: neo4j-test
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Neo4j
      name: neo4j-test
      uid: 0034a30c-d33d-4596-a6d8-7cf47aa3d9e6
  resourceVersion: "1461744"
  uid: 8bb16bdc-2a76-454c-8a58-284f0cc33da3
```

This ServiceAccount is used by Neo4j pods created for the `neo4j-test` database.

### RoleBinding

KubeDB operator creates a RoleBinding object `neo4j-test-rolebinding` in the same namespace as the Neo4j object.

```yaml
$ kubectl get rolebinding -n demo neo4j-test-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2026-05-14T06:54:09Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: neo4j-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
  name: neo4j-test-rolebinding
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Neo4j
      name: neo4j-test
      uid: 0034a30c-d33d-4596-a6d8-7cf47aa3d9e6
  resourceVersion: "1461748"
  uid: f5e8ae7f-62a9-4390-bae0-918f4d5b54d1
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: neo4j-test-role
subjects:
  - kind: ServiceAccount
    name: neo4j-test
```

This object binds Role `neo4j-test-role` with ServiceAccount `neo4j-test`.

## Next Steps

- To use your own ServiceAccount and Role rather than the auto-generated ones, see [Custom RBAC for Neo4j](/docs/guides/neo4j/custom-rbac/using-custom-rbac.md).
- To clean up, delete the database and namespace from the [quickstart cleanup step](/docs/guides/neo4j/quickstart/quickstart.md#cleaning-up).

