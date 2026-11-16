---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-appbinding-concepts
    name: AppBinding
    parent: yugabytedb-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition` (CRD) that points to an application endpoint and its access credentials. KubeDB creates one automatically for every `YugabyteDB`, with the same name, and tools such as KubeStash use it to discover how to connect.

## AppBinding for YugabyteDB

This is the AppBinding KubeDB created for the `yb-demo` universe of the quickstart (metadata trimmed):

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: yb-demo
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: yb-demo
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: yugabytedbs.kubedb.com
spec:
  appRef:
    apiGroup: kubedb.com
    kind: YugabyteDB
    name: yb-demo
    namespace: demo
  clientConfig:
    service:
      name: yb-demo
      port: 5433
      scheme: postgresql
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: YugabyteDBConfiguration
    ysql:
      database: yugabyte
      host: yb-demo.demo.svc
      port: 5433
  secret:
    apiGroup: ""
    kind: Secret
    name: yb-demo-auth
  type: kubedb.com/YugabyteDB
  version: 2025.2.8
```

### spec.type

`kubedb.com/YugabyteDB`.

### spec.appRef

Points back to the `YugabyteDB` object that owns the binding.

### spec.secret

The Secret with the credentials, in the same namespace. For YugabyteDB it is `<name>-auth` with the keys `username` (always `yugabyte`) and `password`. After a [RotateAuth](/docs/guides/yugabytedb/rotate-auth/rotateauth.md) with a user provided Secret, this reference is updated to the new Secret.

### spec.clientConfig

`clientConfig.service` is the client Service (`<name>`, port 5433, scheme `postgresql`). The Service only routes to tservers that finished the auth bootstrap.

### spec.parameters

A `YugabyteDBConfiguration` with the YSQL endpoint (`host`, `port`, `database`). When TLS is enabled, KubeDB also publishes the CA and the client certificate Secret in the AppBinding (not verified in a cluster).

## Verify

```bash
$ kubectl get appbinding -n demo
NAME      TYPE                    VERSION    AGE
yb-demo   kubedb.com/YugabyteDB   2025.2.8   2m6s
```

## Next Steps

- Read the [YugabyteDB CRD concept](/docs/guides/yugabytedb/concepts/yugabytedb.md).
- Run the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md).
