---
title: Rotate Authentication DocumentDB
menu:
  docs_{{ .version }}:
    identifier: dc-rotate-authentication-details
    name: Rotate Authentication
    parent: dc-rotate-authentication
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication Credentials of DocumentDB

KubeDB can rotate the credentials of a `DocumentDB` database on demand with a
`DocumentDBOpsRequest` of type `RotateAuth`. The operator generates a fresh password, applies it
to the database, rolls the pods so every replica picks up the change, and preserves the previous
password under a `password.prev` key so you can reconcile any clients that still hold the old
secret.

## Two credentials, one rotated

A DocumentDB database ships with **two** auth secrets (this is a key difference from KubeDB
Postgres, which has a single auth secret):

| Secret | User | Purpose |
| --- | --- | --- |
| `<db>-auth` | `default_user` | application / MongoDB-wire login (port `10260`) |
| `<db>-admin-auth` | `documentdb` | internal admin / backend-PostgreSQL superuser |

> [!IMPORTANT]
> **`RotateAuth` rotates only the `<db>-admin-auth` secret. The application `<db>-auth` secret is
> left untouched.** This is demonstrated explicitly below. After the OpsRequest completes,
> re-read `<db>-admin-auth` to obtain the new admin password.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo` (`kubectl create ns demo`).
- Deploy a `DocumentDB` cluster (`documentdb-cls-sample`) and wait for it to become `Ready`.

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Credentials before rotation

```bash
$ kubectl get secret -n demo documentdb-cls-sample-admin-auth -o jsonpath='{.data.password}' | base64 -d
EY1imAac)vqps)Ez
$ kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d
DQShSsn0Dqq7Uf*F
```

The `-admin-auth` secret has no `password.prev` key yet (nothing has been rotated):

```bash
$ kubectl get secret -n demo documentdb-cls-sample-admin-auth -o jsonpath='{.data.password\.prev}'
       # (empty)
```

## Create the RotateAuth OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: documentdb-cls-sample
```

```bash
$ kubectl apply -f cluster-rotate-auth.yaml
documentdbopsrequest.ops.kubedb.com/documentdb-cls-rotate-auth created

$ kubectl get dcops -n demo documentdb-cls-rotate-auth
NAME                         TYPE         STATUS       AGE
documentdb-cls-rotate-auth   RotateAuth   Successful   3m34s
```

The status conditions show the new credential being generated, applied to the primary, written
into the PetSet, and then a rolling restart so all replicas pick it up:

```bash
$ kubectl get dcops -n demo documentdb-cls-rotate-auth \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
RotateAuth=True :: DocumentDB ops request has started to rotate auth for documentdb
UpdateCredential=True :: Successfully generated new credentials
ApplyNewCredential=True :: Successfully applied rotated credential to the database primary
UpdatePetSets=True :: Successfully updated petsets for rotate auth type
EvictPod=True :: evict pod; ConditionStatus:True
CheckPodReady=True :: check pod ready; ConditionStatus:True
RestartNodes=True :: Successfully restarted all the nodes
RestartReadReplicas=True :: Successfully Restarted Read Replicas
Successful=True :: Successfully Rotated DocumentDB Auth Secret
UnsetRaftKeyOpsRequestProgressing=True :: Successfully Unset Raft Key OpsRequestProgressing
```

## Credentials after rotation

The admin password has changed, and the **old admin password is retained under
`password.prev`**:

```bash
$ kubectl get secret -n demo documentdb-cls-sample-admin-auth -o jsonpath='{.data.password}' | base64 -d
ELKnwAUT.I85QJ4g
$ kubectl get secret -n demo documentdb-cls-sample-admin-auth -o jsonpath='{.data.password\.prev}' | base64 -d
EY1imAac)vqps)Ez
```

The application `-auth` secret is **unchanged** — same password as before, and no
`password.prev` was written:

```bash
$ kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d
DQShSsn0Dqq7Uf*F          # identical to the "before" value
$ kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password\.prev}'
       # (empty)
```

Because the application credential did not change, existing MongoDB-wire clients keep working
with no reconfiguration:

```bash
$ PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo documentdb-cls-sample-1 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
{ ok: 1 }
```

## Summary

`RotateAuth` is a safe, targeted operation: it rotates the **admin** credential only
(`<db>-admin-auth`), keeps the prior value in `password.prev` for a grace window, and leaves the
application login (`<db>-auth`) and all client connections undisturbed.

## Standalone

The same `DocumentDBOpsRequest` applies to a standalone (`replicas: 1`) instance — point
`spec.databaseRef.name` at `documentdb-sa-sample`. On this build standalone instances did not
finish bootstrapping (see the [Restart](/docs/guides/documentdb/restart/) guide), so the
standalone `RotateAuth` could not be exercised live; the behavior above applies once a
standalone instance is healthy.

## Cleaning Up

```bash
kubectl delete documentdbopsrequest -n demo documentdb-cls-rotate-auth
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- [Restart](/docs/guides/documentdb/restart/) a DocumentDB database.
- Provision a database with [Custom Configuration](/docs/guides/documentdb/configuration/using-config-file/).
