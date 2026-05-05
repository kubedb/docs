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

> New to KubeDB? Please start [here](/docs/README.md).

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
  name: hana-cluster
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: hana-cluster
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
spec:
  appRef:
    apiGroup: kubedb.com
    kind: HanaDB
    name: hana-cluster
    namespace: demo
  clientConfig:
    service:
      name: hana-cluster
      path: /
      port: 39017
      scheme: tcp
  secret:
    name: hana-cluster-auth
  type: kubedb.com/hanadb
  version: "2.0.82"
```

## Key fields

- `spec.type` identifies the application type as `kubedb.com/hanadb`.
- `spec.appRef` points back to the source `HanaDB` object.
- `spec.clientConfig.service` contains the in-cluster service endpoint for SAP HANA.
- `spec.secret.name` points to the secret containing the database credentials.
- `spec.version` is the SAP HANA version resolved from the `HanaDBVersion` catalog.

## Next Steps

- Read the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
- Run the [HanaDB quickstart](/docs/guides/hanadb/quickstart/quickstart.md).
