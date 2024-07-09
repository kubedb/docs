---
title: Run ProxySQL with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-custom-rbac
    name: Custom RBAC
    parent: guides-proxysql
    weight: 31
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a ProxySQL instance. This tutorial will show you how to use KubeDB to run ProxySQL instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/guides/proxysql/custom-rbac/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/proxysql/custom-rbac/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for ProxySQL. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in ProxySQL crd.   If this field is left empty, the KubeDB operator will create a service account name matching ProxySQL crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a ProxySQL instance named `proxy-server` to provide the bare minimum access permissions.

## Custom RBAC for ProxySQL

At first, let's create a `Service Acoount` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo prx-custom-sa
serviceaccount/prx-custom-sa created
```

It should create a service account.

```yaml
$ kubectl get serviceaccount -n demo prx-custom-sa -oyaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2022-12-07T04:31:17Z"
  name: prx-custom-sa
  namespace: demo
  resourceVersion: "494665"
  uid: 4a8d9571-4bae-4af8-976e-061c5dd70a22
secrets:
  - name: prx-custom-sa-token-57whl

```

Now, we need to create a role that has necessary access permissions for the ProxySQL instance named `proxy-server`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/custom-rbac/yamls/prx-custom-role.yaml
role.rbac.authorization.k8s.io/prx-custom-role created
```

Below is the YAML for the Role we just created.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: prx-custom-role
  namespace: demo
rules:
  - apiGroups:
      - policy
    resourceNames:
      - proxy-server
    resources:
      - podsecuritypolicies
    verbs:
      - use
```

This permission is required for ProxySQL pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```bash
$ kubectl create rolebinding prx-custom-rb --role=prx-custom-role --serviceaccount=demo:prx-custom-sa --namespace=demo
rolebinding.rbac.authorization.k8s.io/prx-custom-rb created

```

It should bind `prx-custom-role` and `prx-custom-sa` successfully.

```yaml
$ kubectl get rolebinding -n demo prx-custom-rb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2022-12-07T04:35:58Z"
  name: prx-custom-rb
  namespace: demo
  resourceVersion: "495245"
  uid: d0286421-a0a2-46c8-b3aa-8e7cac9c5cf8
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: prx-custom-role
subjects:
  - kind: ServiceAccount
    name: prx-custom-sa
    namespace: demo

```

Now, create a ProxySQL crd specifying `spec.podTemplate.spec.serviceAccountName` field to `prx-custom-sa`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/custom-rbac/yamls/my-custom-db.yaml
proxysql.kubedb.com/proxy-server created
```

Below is the YAML for the ProxySQL crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.4.4-debian"
  replicas: 1
  backend:
    name: xtradb-galera-appbinding
  syncUsers: true
  podTemplate:
    spec:
      serviceAccountName: prx-custom-sa
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 3

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, PetSet, services, secret etc. If everything goes well, we should see that a pod with the name `proxy-server-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo proxy-server-0
NAME            READY   STATUS    RESTARTS   AGE
proxy-server-0   1/1     Running   0          2m44s
```

Check the pod's log to see if the proxy server is ready

```bash
$ kubectl logs -f -n demo proxy-server-0
...
2022-12-07 04:42:04 [INFO] Cluster: detected a new checksum for mysql_users from peer proxy-server-0.proxy-server-pods.demo:6032, version 2, epoch 1670388124, checksum 0xE6BB9970689336DB . Not syncing yet ...
2022-12-07 04:42:04 [INFO] Cluster: checksum for mysql_users from peer proxy-server-0.proxy-server-pods.demo:6032 matches with local checksum 0xE6BB9970689336DB , we won't sync.

```

Once we see the local checksum matched in the log, the proxysql server is ready.
