---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-concepts-appbinding
    name: AppBinding
    parent: guides-hanadb-concepts
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `Custom Resource Definition` (CRD) that points to an application or
database. It lets other KubeDB components (and tools such as KubeStash) discover how to connect to a
database without hard-coding connection details. When KubeDB provisions a `HanaDB`, it automatically
creates an `AppBinding` with the same name in the same namespace.

## HanaDB AppBinding

Below is the `AppBinding` created by KubeDB for the `hanadb-quickstart` database from the
[Quickstart](/docs/guides/hanadb/quickstart/quickstart.md):

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: hanadb-quickstart
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: hanadb-quickstart
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
spec:
  type: kubedb.com/hanadb
  version: 2.0.82
  appRef:
    apiGroup: kubedb.com
    kind: HanaDB
    name: hanadb-quickstart
    namespace: demo
  clientConfig:
    service:
      name: hanadb-quickstart
      path: /
      port: 39017
      scheme: tcp
  secret:
    name: hanadb-quickstart-auth
```

### spec.type

`spec.type` identifies the kind of application. For HanaDB it is `kubedb.com/hanadb`.

### spec.clientConfig

`spec.clientConfig` describes how to reach the database. For HanaDB it points at the primary `Service`
on the SQL port `39017` (`scheme: tcp`).

### spec.secret

`spec.secret.name` references the authentication `Secret` (`<name>-auth`) that holds the `SYSTEM`
username and password. Consumers read the credentials from this secret.

### spec.appRef

`spec.appRef` is a back-reference to the owning `HanaDB` object.

## Retrieve the AppBinding

```bash
kubectl get appbinding -n demo hanadb-quickstart -o yaml
```

## Next Steps

- Learn about the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
- Read about [HanaDBOpsRequest](/docs/guides/hanadb/concepts/opsrequest.md).
