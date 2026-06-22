---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-appbinding-concepts
    name: AppBinding
    parent: hanadb-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# AppBinding

## What is AppBinding?

An `AppBinding` is a Kubernetes custom resource that points to an application using either a Kubernetes service or an external URL, along with optional parameters and a credential secret.

KubeDB creates an `AppBinding` automatically for each `HanaDB` object. Applications and other operators can use this object to discover the service endpoint and credentials for the SAP HANA database.

## AppBinding Specification

An `AppBinding` object created by KubeDB for a HanaDB instance looks like this:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: hanadb-cluster
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: hanadb-cluster
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
spec:
  appRef:
    apiGroup: kubedb.com
    kind: HanaDB
    name: hanadb-cluster
    namespace: demo
  clientConfig:
    service:
      name: hanadb-cluster
      path: /
      port: 39017
      scheme: tcp
  secret:
    kind: Secret
    name: hanadb-cluster-auth
  type: kubedb.com/hanadb
  version: "2.0.82"
```

## Key fields

- `spec.type` identifies the application type as `kubedb.com/hanadb`.
- `spec.appRef` points back to the source `HanaDB` object.
- `spec.clientConfig.service` contains the in-cluster service endpoint for SAP HANA.
- `spec.secret` points to the Kubernetes Secret containing the database credentials.
- `spec.version` is the SAP HANA version resolved from the `HanaDBVersion` catalog.

## Next Steps

- Read the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
- Run the [HanaDB quickstart](/docs/guides/hanadb/quickstart/quickstart.md).
