---
title: Deploy a TDE Encrypted Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-tde-guide
    name: TDE Guide
    parent: guides-postgres-tde
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Deploy a TDE Encrypted Postgres

This guide shows how to run a KubeDB Postgres with Transparent Data Encryption
(TDE) backed by HashiCorp Vault, verify that data is encrypted at rest, rotate
the principal key, and enable WAL encryption.

Read the [overview](/docs/guides/postgres/tde/overview/index.md) first.

## Before You Begin

- A Kubernetes cluster with [KubeDB installed](/docs/setup/README.md).
- A reachable HashiCorp Vault (or KMIP server). This guide uses Vault.
- A `PostgresVersion` with `spec.distribution: Percona` and
  `spec.tde.supported: true`. Confirm one is available:
- If the Percona TDE image is pulled from a private registry in your
  environment, you will need a docker-registry `Secret` referenced from
  `spec.podTemplate.spec.imagePullSecrets` on the `Postgres` object (see
  [backup & restore](/docs/guides/postgres/tde/backup/index.md#pulling-the-percona-tde-images)
  for details).

```bash
$ kubectl get postgresversion 17.9-percona -o jsonpath='{.spec.distribution} tde={.spec.tde.supported}{"\n"}'
Percona tde=true
```

All commands use the `demo` namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Provide the KMS credentials

Create the Secret holding the Vault token (key `token`). The operator projects it
into every pod at `/etc/pg-tde/vault.token`, outside the data directory.

```bash
$ kubectl create secret generic vault-token -n demo --from-literal=token='<your-vault-token>'
secret/vault-token created
```

## Deploy the encrypted Postgres

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: tde-postgres
  namespace: demo
spec:
  version: "17.9-percona"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
  tde:
    keyProvider:
      vault:
        address: https://vault.example.com:8200
        mountPath: secret
        tokenSecretRef:
          name: vault-token
    defaultEncryptedTables: true
    cipher: aes_128
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/tde/guide/yamls/tde-postgres.yaml
postgres.kubedb.com/tde-postgres created
```

Wait for it to be `Ready`:

```bash
$ kubectl get pg -n demo tde-postgres
NAME           VERSION        STATUS   AGE
tde-postgres   17.9-percona   Ready    4m
```

`defaultEncryptedTables: true` makes `tde_heap` the default access method, so
every table you create is encrypted without any extra syntax.

## Verify encryption

Exec into the primary and create a table, then confirm it is encrypted:

```bash
$ kubectl exec -it -n demo tde-postgres-0 -c postgres -- bash

# inside the pod
$ psql
postgres=# CREATE TABLE secrets (id int, data text);
postgres=# INSERT INTO secrets VALUES (1, 'sensitive');
postgres=# SELECT pg_tde_is_encrypted('secrets');
 pg_tde_is_encrypted
---------------------
 t
```

`pg_tde_is_encrypted` returning `t` confirms the table's data files are encrypted
on disk. You can inspect the active principal key with `SELECT pg_tde_key_info();`.

To convert a pre existing, unencrypted table:

```sql
ALTER TABLE legacy SET ACCESS METHOD tde_heap;
```

## Rotate the principal key

The principal key wraps the per relation internal keys, so rotating it does not
rewrite any data and needs no restart. Use a `RotatePrincipalKey`
`PostgresOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: tde-rotate-key
  namespace: demo
spec:
  type: RotatePrincipalKey
  databaseRef:
    name: tde-postgres
  # optional: pin the new key name; otherwise one is generated
  rotatePrincipalKey:
    keyName: tde-postgres-principal-2
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/tde/guide/yamls/rotate-principal-key.yaml
postgresopsrequest.ops.kubedb.com/tde-rotate-key created

$ kubectl get postgresopsrequest -n demo tde-rotate-key
NAME             TYPE                 STATUS       AGE
tde-rotate-key   RotatePrincipalKey   Successful   40s
```

## Enable WAL encryption

WAL encryption is cluster wide, requires a global provider (Vault or KMIP), and a
rolling restart. Enable it with an `EnableWALEncryption` `PostgresOpsRequest`,
which sets the server key, flips `spec.tde.encryptWAL`, and restarts every node:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: tde-enable-wal
  namespace: demo
spec:
  type: EnableWALEncryption
  databaseRef:
    name: tde-postgres
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/tde/guide/yamls/enable-wal-encryption.yaml
postgresopsrequest.ops.kubedb.com/tde-enable-wal created

$ kubectl get postgresopsrequest -n demo tde-enable-wal
NAME             TYPE                  STATUS       AGE
tde-enable-wal   EnableWALEncryption   Successful   3m

$ kubectl exec -n demo tde-postgres-0 -c postgres -- psql -Atqc "SHOW pg_tde.wal_encrypt;"
on
```

> WAL encryption is not compatible with every WAL tool. Validate your archiver
> (for example wal-g) against encrypted WAL before enabling it on a cluster with
> continuous archiving.

## Cleanup

```bash
$ kubectl delete postgresopsrequest -n demo tde-rotate-key tde-enable-wal
$ kubectl delete pg -n demo tde-postgres
$ kubectl delete secret -n demo vault-token
$ kubectl delete ns demo
```

## Next Steps

- Review the [TDE overview](/docs/guides/postgres/tde/overview/index.md) for the
  key hierarchy and limitations.
- Combine TDE (at rest) with [TLS/SSL](/docs/guides/postgres/tls/overview/index.md)
  (in transit) for defense in depth.
- Set up [backup, continuous archiving, and PITR](/docs/guides/postgres/tde/backup/index.md)
  for a TDE encrypted Postgres.
