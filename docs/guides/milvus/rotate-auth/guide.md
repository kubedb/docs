---
title: Rotate Authentication of Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-rotate-auth-guide
    name: Guide
    parent: milvus-rotate-auth
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Milvus

This guide will show you how to use the `KubeDB` Ops-manager operator to rotate the authentication credentials of a Milvus database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
  - [Rotate Authentication Overview](/docs/guides/milvus/rotate-auth/overview.md)

- An object-storage secret named `my-release-minio` must exist in the `demo` namespace.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/rotate-auth/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/rotate-auth/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Auth Secret Basics

Milvus authentication is enabled by default. When `spec.authSecret` is omitted, KubeDB creates a `kubernetes.io/basic-auth` secret named `<db>-auth` with a `root` user and a random password:

```bash
$ kubectl get secret milvus-standalone-auth -n demo
NAME                     TYPE                       DATA   AGE
milvus-standalone-auth   kubernetes.io/basic-auth   2      92s

$ kubectl get secret milvus-standalone-auth -n demo -o jsonpath='{.data.username}' | base64 -d
root
```

There are two ways to rotate this credential.

## Rotate Standalone Milvus (User-Supplied Secret)

This sample provides a new secret with a known password and asks the operator to switch to it.

`rotate-auth-standalone.yaml`

```yaml
---
apiVersion: v1
kind: Secret
metadata:
 name: milvus-new-auth1
 namespace: demo
type: Opaque
stringData:
 username: root
 password: NewPassword1
---
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-rotate-auth-user-secret
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: milvus-standalone
  authentication:
    secretRef:
      kind: Secret
      name: milvus-new-auth1
```

Here, `spec.authentication.secretRef.name` points at the user-created secret. To let the operator generate a random password instead, simply omit `spec.authentication`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/rotate-auth/yamls/rotate-auth-standalone.yaml
secret/milvus-new-auth1 created
milvusopsrequest.ops.kubedb.com/milvus-rotate-auth-user-secret created
```

### Watch Progress

```bash
$ kubectl get milvusopsrequest milvus-rotate-auth-user-secret -n demo
NAME                             TYPE         STATUS       AGE
milvus-rotate-auth-user-secret   RotateAuth   Successful   83s
```

```bash
$ kubectl describe milvusopsrequest milvus-rotate-auth-user-secret -n demo
...
Status:
  Conditions:
    Message:  Milvus ops-request has started to rotate auth for milvus nodes
    Reason:   RotateAuth
    Type:     RotateAuth
    Message:  Successfully referenced the user provided authSecret
    Reason:   UpdateCredential
    Type:     UpdateCredential
    Message:  Successfully updated milvus credential dynamically
    Reason:   UpdateCredentialDynamically
    Type:     UpdateCredentialDynamically
    Message:  successfully reconciled the Milvus with new auth secret
    Reason:   UpdatePetSets
    Type:     UpdatePetSets
    Message:  Successfully restarted all milvus nodes
    Reason:   RestartNodes
    ...
  Phase:      Successful
```

### Verify the Rotation

The database now references the new secret:

```bash
$ kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.spec.authSecret.name}'
milvus-new-auth1

$ kubectl get secret milvus-standalone-auth -n demo -o jsonpath='{.metadata.annotations.kubedb\.com/auth-active-from}'
2026-06-30T17:19:33Z
```

## Rotate Distributed Milvus

For a distributed Milvus, point `spec.databaseRef.name` at the distributed database (`milvus-cluster`). The credential is updated dynamically inside every distributed role.

`rotate-auth-distributed.yaml`

```yaml
---
apiVersion: v1
kind: Secret
metadata:
 name: milvus-new-auth1
 namespace: demo
type: Opaque
stringData:
 username: root
 password: NewPassword1
---
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-rotate-auth-user-secret
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: milvus-cluster
  authentication:
    secretRef:
      kind: Secret
      name: milvus-new-auth1
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/rotate-auth/yamls/rotate-auth-distributed.yaml
secret/milvus-new-auth1 created
milvusopsrequest.ops.kubedb.com/milvus-rotate-auth-user-secret created

$ kubectl get milvusopsrequest milvus-rotate-auth-user-secret -n demo
NAME                             TYPE         STATUS       AGE
milvus-rotate-auth-user-secret   RotateAuth   Successful   3m56s
```

The credential is updated dynamically and then every role is reconciled and restarted:

```bash
$ kubectl describe milvusopsrequest milvus-rotate-auth-user-secret -n demo
...
Status:
  Conditions:
    Message:  Successfully referenced the user provided authSecret
    Reason:   UpdateCredential
    Type:     UpdateCredential
    Message:  Successfully updated milvus credential dynamically
    Reason:   UpdateCredentialDynamically
    Type:     UpdateCredentialDynamically
    Message:  successfully reconciled the Milvus with new auth secret
    Reason:   UpdatePetSets
    Type:     UpdatePetSets
  Phase:      Successful

$ kubectl get milvuses.kubedb.com milvus-cluster -n demo -o jsonpath='{.spec.authSecret.name}'
milvus-new-auth1
```

## Automatic Rotation Recommendations

The Recommendation Engine can generate `RotateAuth` recommendations automatically based on two fields on `spec.authSecret`:

- **`rotateAfter`** — once the credential is older than this duration, a `RotateAuth` recommendation is generated.
- **`activeFrom`** — the timestamp the credential became active (also reflected in the `kubedb.com/auth-active-from` annotation on the secret); `rotateAfter` is measured from this point.

See the [Recommendation Engine guide](/docs/guides/milvus/recommendation/guide.md) for an end-to-end walkthrough.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo milvus-rotate-auth-user-secret
$ kubectl delete secret -n demo milvus-new-auth1
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Learn more about the [Recommendation Engine](/docs/guides/milvus/recommendation/guide.md).
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
