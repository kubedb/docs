---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: etcd-appbinding-concepts
    name: AppBinding
    parent: etcd-concepts-etcd
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://appscode.com/blog/post/the-case-for-appbinding).

When you deploy an [Etcd](/docs/guides/etcd/concepts/etcd.md) object with KubeDB, an `AppBinding` object is created automatically for it. Backup and restore tooling — [KubeStash](https://kubestash.com/) and the etcd backup addon — uses it to find out how to reach the cluster and which credentials to use.

## The AppBinding created for an Etcd object

An `AppBinding` object created by KubeDB for an etcd cluster is shown below:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2026-01-14T10:02:45Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: etcd-quickstart
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: etcds.kubedb.com
  name: etcd-quickstart
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Etcd
      name: etcd-quickstart
      uid: 20e00408-abf1-470b-a049-bdf272b3e994
  resourceVersion: "12345"
  uid: 8fd15549-ab9c-4523-b85d-77275f620bb5
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Etcd
    name: etcd-quickstart
    namespace: demo
  clientConfig:
    service:
      name: etcd-quickstart
      port: 2379
      scheme: etcd
      path: /
    insecureSkipTLSVerify: false
  secret:
    name: etcd-quickstart-auth
  type: kubedb.com/etcd
  version: 3.6.4
```

What is etcd-specific about it:

| Field                             | Value for etcd                                                                                                             |
|-----------------------------------|------------------------------------------------------------------------------------------------------------------------------|
| `spec.type`                       | `kubedb.com/etcd`                                                                                                            |
| `spec.clientConfig.service.scheme` | `etcd`                                                                                                                      |
| `spec.clientConfig.service.name`  | The load balanced client Service, which has the same name as the `Etcd` object.                                              |
| `spec.clientConfig.service.port`  | `2379`, etcd's client port. (The peer port `2380` and the metrics port `2381` are not part of the AppBinding — clients never talk to them.) |
| `spec.secret`                     | The auth Secret holding the `username`/`password` of the etcd `root` user.                                                    |
| `spec.version`                    | The etcd version, taken from the referenced [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md).                          |

There is a single client Service rather than a primary/standby pair, because every etcd member serves the client API and forwards writes to the current Raft leader — see [Etcd CRD](/docs/guides/etcd/concepts/etcd.md#specservicetemplates).

### When TLS is enabled

If [`spec.tls`](/docs/guides/etcd/concepts/etcd.md#spectls) is configured on the `Etcd` object, KubeDB additionally fills in:

- `spec.clientConfig.caBundle` — the CA certificate read from the `client` certificate Secret, so consumers can verify the etcd server certificate. `insecureSkipTLSVerify` stays `false`.
- `spec.tlsSecret` — a reference to the `client` certificate Secret (`<db-name>-client-cert` by default), which holds the client certificate and key used to authenticate to etcd.

### spec.secret

`spec.secret` references the Secret holding the credentials required to access the cluster. It must be in the same namespace as the `AppBinding`, and contains the following keys:

| Key        | Usage                                                                     |
|------------|-----------------------------------------------------------------------------|
| `username` | Username of the etcd superuser. KubeDB uses etcd's built-in `root` user.     |
| `password` | Password for the user specified by `username`.                              |

## Next Steps

- Learn about the [Etcd](/docs/guides/etcd/concepts/etcd.md) crd.
- Learn how to use KubeDB to manage various databases [here](/docs/guides/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
