---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-rotate-auth-overview
    name: Overview
    parent: etcd-rotate-authentication
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Etcd

This guide gives an overview of how the KubeDB Ops-manager operator rotates the credentials of an
`Etcd` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## How etcd authentication is stored

KubeDB creates a `kubernetes.io/basic-auth` shaped Secret named `<db-name>-auth` for every `Etcd`
object (unless you point `spec.authSecret.name` somewhere else). It carries two keys:

| Key        | Description                                                                        |
|------------|------------------------------------------------------------------------------------|
| `username` | The etcd RBAC user. It is always `root` for a KubeDB provisioned cluster.           |
| `password` | That user's password.                                                              |

These are the standard `core.BasicAuthUsernameKey` / `core.BasicAuthPasswordKey` keys, so the same
Secret can be consumed directly by an application, by `etcdctl --user=<username>:<password>`, and by
KubeDB's own health checker through the generated AppBinding.

## Why rotating etcd credentials does not restart anything

This is the one thing that makes etcd's `RotateAuth` different from most other KubeDB databases.

An etcd RBAC user does not live in a config file or in an environment variable — it lives *inside
the keyspace*, managed by etcd's own `UserChangePassword` / `UserAdd` / `UserGrantRole` RPCs. A
password change therefore takes effect on the next authentication, with no process restart and no
pod eviction involved. Compare that with PostgreSQL, where the credential is threaded through the
pod's environment and the ops request has to roll every pod for the change to be picked up.

So an `EtcdOpsRequest` of type `RotateAuth`:

- writes the new password into etcd's RBAC store,
- updates the auth Secret,
- and reaches `Successful` **without evicting a single pod**.

You will not see any `EvictPod--<pod>`, `CheckPodReady--<pod>` or `RestartEtcdPods` condition on an
etcd `RotateAuth` request, and pod `AGE` / `RESTARTS` are unchanged when it finishes.

## How the rotation is staged

The rotation writes to two independent places — etcd's own RBAC store and a Kubernetes Secret — so
it is deliberately staged, in order to stay crash safe:

1. The new password is parked in a `password.next` key on the Secret, while `password` remains the
   live one. Nothing that reads the Secret is disturbed yet.
2. The operator applies the staged password inside etcd. If it is re-run after a crash, it first
   probes whether the new credential already authenticates, so the step is idempotent.
3. Only once etcd has accepted the change is the staged value promoted: it becomes `password`, the
   superseded value is kept under `password.prev`, and `password.next`/`username.next` are removed.
   The `Etcd` object's `spec.authSecret.activeFrom` timestamp is stamped at the same time.

The `.prev` keys are kept for rollback purposes; nothing in KubeDB reads them afterwards.

## Two ways to rotate

1. **Operator generated.** Apply an `EtcdOpsRequest` of type `RotateAuth` with no
   `spec.authentication`. KubeDB generates a random password, stages it and promotes it in the
   existing auth Secret. The Secret's name does not change.

2. **User provided.** Create your own `kubernetes.io/basic-auth` Secret containing the desired
   `username` and `password`, and reference it from `spec.authentication.secretRef`. The operator
   validates that both keys are present and non-empty, applies the password to etcd, records the
   superseded credential in your Secret under `username.prev` / `password.prev`, and repoints
   `spec.authSecret` of the `Etcd` object at your Secret with `externallyManaged: true`.

## How the Rotate Authentication Process Works

1. A user first creates an `Etcd` Custom Resource Object (CRO).

2. The `KubeDB Provisioner` operator continuously watches for `Etcd` CROs.

3. When the operator detects an `Etcd` CR, it provisions the required `PetSet` along with related
   resources such as Secrets, Services and RBAC objects — including the `<db-name>-auth` Secret.

4. To initiate a credential rotation, the user creates an `EtcdOpsRequest` CR of type `RotateAuth`,
   optionally referring to their own Secret.

5. The `KubeDB Ops-manager` operator watches for `EtcdOpsRequest` CRs.

6. Upon detecting one, it pauses the referenced `Etcd` object so that the Provisioner operator does
   not act on it during the rotation.

7. The operator stages the new credential on the Secret (`UpdateCredential` condition).

8. It then applies the credential to the cluster through etcd's RBAC API
   (`EtcdCredentialApplied` condition). If the cluster has never had etcd authentication enabled,
   this step bootstraps it: it creates the `root` user, grants it the `root` role and calls
   `AuthEnable`.

9. The staged credential is promoted to the live one, and the `Etcd` object's `spec.authSecret`
   reference and `activeFrom` timestamp are updated (`UpdateDatabase` condition).

10. Finally the operator resumes the `Etcd` object and marks the request `Successful`. No pod is
    restarted at any point.

In the [next](/docs/guides/etcd/rotate-authentication/rotateauth.md) doc, we walk through both
rotation flows step by step.
