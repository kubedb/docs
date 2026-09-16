---
title: Run PostgreSQL with Custom pg_hba.conf Rules
menu:
  docs_{{ .version }}:
    identifier: pg-custom-pg-hba-configuration
    name: Custom pg_hba.conf
    parent: pg-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom pg_hba.conf Rules

KubeDB generates `pg_hba.conf` for every PostgreSQL pod from the database's TLS and `clientAuthMode` settings. This tutorial shows how to add your own host-based authentication rules on top of the generated ones — for example, to reject superuser connections from outside the pod network.

## How it works

Add a `user_hba.conf` key to the same `configSecret` you already use for `user.conf`. Its content is standard [pg_hba.conf syntax](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html), and it is spliced into the generated `pg_hba.conf` at a fixed position:

```
local      all             all                                     trust          ┐ generated:
host(ssl)  all             all             127.0.0.1/32            <auth>         │ operator-essential
local      replication     all                                     trust          │ rules stay ABOVE
host(ssl)  replication     all             127.0.0.1/32            <auth>         ┘ your rules

<your user_hba.conf rules>                                                        ← inserted here

host(ssl)  all             all             0.0.0.0/0               <auth>         ┐ generated
host(ssl)  replication     postgres        0.0.0.0/0               <auth>         │ catch-alls sit
...                                                                               ┘ BELOW your rules
```

`pg_hba.conf` is **first-match-wins** — the opposite of `postgresql.conf`, where the last setting wins. Placing your rules above the catch-alls means they take precedence over the defaults; placing them below the local/loopback essentials means a mistaken rule cannot lock KubeDB's own scripts, health checks, or sidecars out of the database.

- **PostgreSQL ≥ 16**: the generated file references your rules with `include_if_exists`, so after editing the secret a `SELECT pg_reload_conf();` (or `pg ctl reload`) applies them live — no restart.
- **PostgreSQL ≤ 15**: `pg_hba.conf` cannot include files, so your rules are copied in when the pod starts. Changes to the secret take effect on the next pod restart (a `PostgresOpsRequest` reconfigure restart, or delete the pods).

## Before You Begin

Install KubeDB following the steps [here](/docs/setup/README.md), and create the `demo` namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Example: restrict the postgres superuser to the pod network

Create the config secret. `user.conf` must exist (it may be empty); `user_hba.conf` carries the rules:

```bash
$ kubectl create secret generic -n demo pg-configuration \
    --from-literal=user.conf="" \
    --from-file=user_hba.conf=./user_hba.conf
```

with `user_hba.conf`:

```
# allow the postgres role from the pod network (replication, coordinator, probes)
host    all    postgres    10.42.0.0/16    scram-sha-256
# reject the postgres role from everywhere else
host    all    postgres    0.0.0.0/0       reject
host    all    postgres    ::/0            reject
```

Reference the secret from the Postgres object:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: custom-postgres
  namespace: demo
spec:
  version: "18.6"
  replicas: 3
  configSecret:
    name: pg-configuration
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 2Gi
    accessModes:
      - ReadWriteOnce
```

After the database is ready, verify the rules landed in position:

```bash
$ kubectl exec -n demo custom-postgres-0 -c postgres -- \
    psql -c "SELECT rule_number, type, database, user_name, address, auth_method \
             FROM pg_hba_file_rules ORDER BY rule_number;"
```

The rows from `user_hba.conf` appear after the loopback/replication rules and before the `0.0.0.0/0` catch-alls. A connection as `postgres` from outside the pod CIDR now fails with:

```
FATAL:  pg_hba.conf rejects connection for host "...", user "postgres"
```

while replication, health checks, and in-cluster clients are untouched.

## Rules of the road

- **Order within your file matters** — it is first-match-wins there too. Put narrower `allow` rules before wider `reject` rules, as in the example above.
- **Don't reject `postgres` from the pod network.** The pg-coordinator sidecar connects to *peer pods* as `postgres` with `replication=database`, which matches ordinary `host` rules (not the `replication` keyword). A blanket `host all postgres 0.0.0.0/0 reject` without a preceding pod-CIDR allow breaks failover and standby sync. Always pair a reject with a pod-network allow, as shown.
- **Physical replication** (`replication` keyword in the database column) between pods is protected by generated rules only on loopback; the pod-network replication catch-alls sit *below* your rules. If you write rules with `replication` in the database column, make sure streaming between pods still has a matching allow.
- **Validate before you rely on a restart.** A syntactically invalid rule is refused at reload (PostgreSQL keeps the old rules and logs the error), but at *pod start* it is fatal and the pod will crash-loop. After editing the secret, reload and check the server log or `SELECT * FROM pg_hba_file_rules WHERE error IS NOT NULL;` before the next restart.
- **`local` and loopback access cannot be overridden.** Rules above your file's position guarantee KubeDB's own scripts and probes keep working. Use `spec.allowedSchemas`/network policies for tighter isolation goals.

## Cleaning up

```bash
kubectl delete pg -n demo custom-postgres
kubectl delete secret -n demo pg-configuration
kubectl delete ns demo
```

## Next Steps

- [Custom configuration](/docs/guides/postgres/configuration/using-config-file.md) for `postgresql.conf` settings.
- [Reconfigure](/docs/guides/postgres/reconfigure/overview.md) a running database with a `PostgresOpsRequest`.
