---
title: NATS Ops Requests
menu:
  docs_{{ .version }}:
    identifier: nats-ops-requests
    name: Ops Requests
    parent: nats-guides
    weight: 80
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# NATS Ops Requests

Day-2 operations use `NATSOpsRequest`. Fields that an ops request owns (version, replicas, resources, storage size, configuration, TLS, auth secret) are rejected when edited directly on a provisioned `NATS` object.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/opsrequests](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/opsrequests) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

Supported types: `UpdateVersion`, `HorizontalScaling`, `VerticalScaling`, `VolumeExpansion`, `Reconfigure`, `ReconfigureTLS`, `RotateAuth`, `Restart` and `StorageMigration`.

## How a roll works

Anything that restarts pods (`UpdateVersion`, `Restart`, memory changes, `Reconfigure` with a restart, enabling or disabling TLS, `StorageMigration`) uses one procedure:

1. Before the first pod, strict `/healthz` on every pod must be `ok` and the JetStream meta group must have a leader. Otherwise the request waits, then fails after its `timeout`.
2. Non-leader pods roll first, highest ordinal first. The meta leader rolls last, after KubeDB steps leadership down to a pod that is already current.
3. For each pod KubeDB deletes it; the pre-stop hook puts the server in lame-duck mode (listeners close, Raft leaderships transfer, clients are moved away over the lame-duck duration). It then waits for the pod to be Ready, for strict `/healthz` (all streams and consumers caught up) and for its routes to reappear on its peers.
4. The next pod is touched only after that. A pod that does not catch up within `timeout` fails the request and leaves the cluster paused for you to inspect.

Clients keep working only if they use a library that reconnects, and connect through the Service or a server list. A stream with `num_replicas: 1` is unavailable while its pod restarts; KubeDB emits a warning event naming such streams.

## UpdateVersion

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: NATSOpsRequest
metadata:
  name: nats-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: nats-cluster
  updateVersion:
    targetVersion: "2.15.1"
  timeout: 30m
  apply: IfReady
```

The target `NATSVersion` must exist and be allowed by the current version's update constraints. All three images (server, reloader, exporter) are updated. Mixed versions during the roll are supported by NATS. There is no downgrade and a failed pod is not rolled back, so take a [backup](/docs/guides/nats/backup.md) first; it is the supported recovery path.

## HorizontalScaling

```yaml
spec:
  type: HorizontalScaling
  databaseRef:
    name: nats-cluster
  horizontalScaling:
    replicas: 5
```

With JetStream the count must be 1 or an odd number `>= 3`.

- **Scale up** adds pods that join the mesh and the meta group on their own. No data moves: existing streams stay where they are, and rebalancing is a user action (`nats stream edit`).
- **Scale down** removes one ordinal at a time, highest first. It is refused if any stream or consumer has more replicas than the target size. KubeDB evacuates the server from the meta group and its streams, lowers the replica count, removes the peer once the pod is gone, and deletes the PVC of the removed ordinal.
- Going from 1 to 3 replicas with JetStream restarts the original pod with a cluster configuration first.

## VerticalScaling

```yaml
spec:
  type: VerticalScaling
  databaseRef:
    name: nats-cluster
  verticalScaling:
    nats:
      resources:
        requests: {cpu: "1", memory: 2Gi}
        limits: {cpu: "2", memory: 2Gi}
```

`nats`, `reloader` and `exporter` can be sized separately. `memoryStoreSize` must stay at or below 50% of the new memory limit. A memory change always restarts the pods, because `GOMEMLIMIT` is read at process start; a CPU-only change can be applied in place with `mode: InPlace`.

## VolumeExpansion

```yaml
spec:
  type: VolumeExpansion
  databaseRef:
    name: nats-cluster
  volumeExpansion:
    mode: Online
    nats: 20Gi
```

`mode` is `Online` or `Offline`. After the PVCs grow, KubeDB raises `max_file_store` to 90% of the new size and reloads the config. The server accepts increases without a restart, and rejects decreases. With a storage driver that resizes online, no pod restarts.

## Reconfigure

```yaml
spec:
  type: Reconfigure
  databaseRef:
    name: nats-cluster
  configuration:
    serverConfig:
      maxConnections: 20000
    restart: auto
```

You can also give `configSecret`, `applyConfig` (inline `nats.conf` or `accounts.conf` content) or `removeCustomConfig`, and set `lameDuck`. KubeDB validates the new config with `nats-server -t` first, then the reloader signals the servers.

| Reloads live | Needs a roll |
|---|---|
| users, accounts, authorization, resolver settings | `listen`, `http_port`, cluster `host`/`port` |
| `max_connections`, `max_payload`, `max_control_line`, `ping_interval`, `write_deadline`, `auth_timeout` | `max_subscriptions`, `max_pending` |
| `debug`, `trace`, logging, `server_tags` | `lame_duck_duration`, `lame_duck_grace_period`, `system_account` |
| client TLS, cluster routes and TLS | `no_auth_user`, `allow_non_tls` |
| `max_file_store` increases | `store_dir`, JetStream domain, decreases |
| leaf remotes added, removed or re-pointed | `gateway` (except TLS material), hub-side `leafnodes` block |

`restart` is `auto` (roll only if a right-hand key changed), `false` (fail with the server's own error if the change is not reloadable) or `true` (always roll). After a reload KubeDB checks `/varz` on every pod to confirm the new config was loaded, and fails the request, naming the pods that differ, if one rejected it.

## ReconfigureTLS

```yaml
spec:
  type: ReconfigureTLS
  databaseRef:
    name: nats-tls
  tls:
    rotateCertificates: true
```

`tls.issuerRef` and `tls.certificates` add or change TLS and `tls.remove: true` removes it. See the [TLS guide](/docs/guides/nats/tls.md#day-2-changes) for what each change does to running clients.

## RotateAuth

```yaml
spec:
  type: RotateAuth
  databaseRef:
    name: nats-cluster
```

Without `authentication.secretRef` KubeDB generates a new password. With it, KubeDB uses the credentials from your Secret. The old and new credentials are both valid during the change, so applications have a window to read the new Secret. The cluster-route and gateway credentials cannot hold two values at once. The route credential is replaced in one config write and reloaded. The gateway block does not reload, so rotating it is a rolling restart per DC. See [auth modes](/docs/guides/nats/auth-modes.md#rotating-credentials).

## Restart

```yaml
spec:
  type: Restart
  databaseRef:
    name: nats-cluster
```

A rolling restart of every pod using the procedure above.

## Multi-DC clusters

For a distributed `NATS`, a request runs inside one DC at a time, the Arbiter DC first and the DC with the meta leader last. After every request KubeDB steps the meta leader down to a data DC. `HorizontalScaling` takes per-DC targets and is refused if a stream placed in that DC would fall below its replica count:

```yaml
spec:
  type: HorizontalScaling
  databaseRef:
    name: nats-dc
  horizontalScaling:
    dataCenters:
      - clusterName: dc-a
        replicas: 5
      - clusterName: dc-b
        replicas: 5
```

`UpdateVersion` goes DC by DC. Adding a DC rolls only the new DC. Removing a DC is refused until its streams are moved or mirrored elsewhere.

## Autoscaling

A `NATSAutoscaler` object scales compute and, for JetStream, storage, using the same ops requests under the hood.

## Cleanup

```bash
$ kubectl delete natsopsrequest -n demo --all
```
