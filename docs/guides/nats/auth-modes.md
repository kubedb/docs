---
title: NATS Authentication Modes
menu:
  docs_{{ .version }}:
    identifier: nats-auth
    name: Auth Modes
    parent: nats-guides
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# NATS Authentication Modes

NATS has no built-in admin user, so KubeDB defines two accounts and generates the credentials.

| Account | User | Purpose |
|---|---|---|
| `SYS` (system account) | `kubedb-sys` | operator only: evacuate, peer-remove, meta-leader step-down, `$SYS` events. Stored in `<name>-sys-auth`. You never need it. |
| `APP` | `admin` | applications and the KubeDB health checker. Stored in `<name>-auth`. |

`spec.auth.mode` selects how these are expressed. It is **immutable**: a running cluster cannot move between modes.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/auth) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not have one, create it with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install KubeDB following the steps [here](/docs/setup/README.md), with `--set global.featureGates.NATS=true` so the NATS CRDs are installed.
- This tutorial uses the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Config mode (default)

Accounts and users are rendered into the server configuration. Passwords are stored as bcrypt hashes in the config; the plaintext stays in the Secrets.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-config-auth
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 3
  auth:
    mode: Config
  deletionPolicy: WipeOut
```

- Use your own admin password by creating a Secret with `username: admin` and a `password`, and setting `spec.authSecret.name`. The default name is `<name>-auth`.
- Add accounts and users with `accounts.conf` in [`spec.configuration`](/docs/guides/nats/clustering.md#custom-configuration).
- `spec.disableSecurity: true` lets clients connect without credentials. They land in the `APP` account through `no_auth_user`. The system account keeps its password. This is immutable and is not allowed in Operator mode.

## Operator mode (decentralized JWT)

The trust chain is *operator, accounts, users*, all signed JWTs. KubeDB generates the whole chain when `spec.auth.operator` has no `externalSecret`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-operator-auth
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 3
  auth:
    mode: Operator
    operator:
      resolverStorage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
      allowDelete: false
  deletionPolicy: WipeOut
```

KubeDB creates these Secrets:

| Secret | Contents |
|---|---|
| `<name>-operator` | operator seed, signing-key seed, `operator.jwt`. Never mounted into pods. |
| `<name>-sys-creds` | `sys.creds` for the `SYS` user; also what you need to `nsc push` new accounts. |
| `<name>-auth` | `user.creds` for the `admin` user of `APP`, plus `username` and `password` keys so the AppBinding keeps its usual shape. |
| `<name>-account-<acct>` | one per managed account: account seed, signing key, `account.jwt`. |

- `resolverStorage` is the PVC of the full account resolver, mounted at `/resolver` (default `1Gi`).
- `allowDelete` is the resolver's `allow_delete` (default `false`).
- Users and accounts live outside the config. Manage them with `nsc` or `nats auth` using `<name>-sys-creds`. The `accounts` and `authorization` keys are rejected in `spec.configuration`.
- Connect with the credentials file:

  ```bash
  $ kubectl get secret -n demo nats-operator-auth -o jsonpath='{.data.user\.creds}' | base64 -d > user.creds
  $ nats --server nats://127.0.0.1:4222 --creds user.creds server check connection
  ```

### Bring your own operator

If you created the operator with `nsc`, put `operator.jwt`, `sys-account.jwt` and `sys.creds` in a Secret and reference it. KubeDB then generates nothing and only renders the server configuration.

```yaml
spec:
  auth:
    mode: Operator
    operator:
      externalSecret:
        name: my-nats-operator
```

Rotating the operator key is your responsibility in this case. KubeDB does not rotate the operator key in either variant.

> **Keep the key Secrets.** In Operator mode the seeds exist nowhere else. Losing them makes the account JWTs in the resolver unusable for further changes. They are included in [backups](/docs/guides/nats/backup.md).

## Rotating credentials

A `RotateAuth` ops request rotates credentials. In Config mode the old and new `admin` hash are both valid until the new one is active, then the old one is removed. Only the removed user's connections are closed. In Operator mode it issues a new user key and JWT, revokes the old key and pushes the updated account JWT, with no config reload and no restart. See [ops requests](/docs/guides/nats/ops-requests.md#rotateauth).

## Cleanup

```bash
$ kubectl delete -n demo nats/nats-config-auth nats/nats-operator-auth
```
