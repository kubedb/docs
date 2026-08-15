---
title: Run Etcd with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: etcd-using-config-configuration
    name: Customize Configuration
    parent: etcd-custom-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Run Etcd with Custom Configuration

KubeDB lets you tune an etcd cluster at creation time through the typed `spec.configuration.tuning` block. This guide covers what each knob does, how to set it, and how to verify it took effect.

This guide is about **day-1** configuration — the settings you bake into the `Etcd` object before the cluster exists. To change these on a cluster that is already running, use a `Reconfigure` [EtcdOpsRequest](/docs/guides/etcd/reconfigure/reconfigure.md) instead; the fields are the same.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure the operator was installed with `--set featureGates.Etcd=true`.

- To keep things isolated, this tutorial uses a separate namespace called `demo`.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Custom config files are not supported for etcd

Most KubeDB databases let you mount a configuration file — `spec.configuration.configSecret` pointing at a `Secret` holding `postgresql.conf`, `my.cnf`, `zoo.cfg` and so on — or merge in free-form key/value pairs with `spec.configuration.applyConfig`. **etcd supports neither.**

The reason is a property of etcd itself, not a KubeDB limitation: **etcd's `--config-file` flag is mutually exclusive with the individual command-line flags.** From etcd's own documentation, if you pass `--config-file`, etcd ignores every other flag on the command line and takes its entire configuration from that file.

That is fatal for an orchestrator. KubeDB *must* own a set of flags for the cluster to work at all:

- `--name`, `--initial-advertise-peer-urls`, `--advertise-client-urls` — per-member identity, resolved from the pod's own name through the downward API.
- `--listen-peer-urls`, `--listen-client-urls`, `--listen-metrics-urls` — the listeners, whose scheme flips between `http` and `https` when you enable TLS.
- `--cert-file`, `--key-file`, `--trusted-ca-file`, `--client-cert-auth`, and the `--peer-*` equivalents — the TLS material, whose paths come from the mounted certificate secrets.
- `--data-dir` — the mounted PersistentVolume.

If a user-supplied config file took over, all of those would silently stop being applied: the member would not know its own identity, would not join the right cluster, and would not use the certificates that were issued for it. Rather than merge a config file and a flag set — which etcd will not do — KubeDB exposes the safe subset as typed fields and renders them as additional flags.

So, concretely:

- `spec.configuration.tuning` — **supported**. This is the path described in the rest of this guide.
- `spec.configuration.applyConfig` — **not supported**. A `Reconfigure` `EtcdOpsRequest` carrying it is rejected by the admission webhook.
- `spec.configuration.configSecret` — **not supported**. Likewise rejected by the `Reconfigure` webhook.
- `spec.configSecret` (the deprecated top-level field) — present only for structural parity with the other KubeDB databases. It is not mounted into the etcd container and has no effect. Do not use it.

The webhook's rejection message spells it out:

```
spec.configuration.applyConfig and spec.configuration.configSecret are not supported for Etcd, use spec.configuration.tuning instead
```

If you need an etcd flag that `tuning` does not cover, use `spec.podTemplate` to set the corresponding `ETCD_*` environment variable, which etcd reads as an alternative to the flag — but be aware that any variable colliding with one KubeDB manages will be overwritten by the operator.

## The tuning knobs

`spec.configuration.tuning` has four fields, each mapping onto exactly one etcd flag:

| Field                     | etcd flag                     | Type     | Default (etcd)          |
|---------------------------|-------------------------------|----------|-------------------------|
| `quotaBackendBytes`       | `--quota-backend-bytes`       | integer  | 2 GiB                   |
| `autoCompactionMode`      | `--auto-compaction-mode`      | enum     | `periodic`              |
| `autoCompactionRetention` | `--auto-compaction-retention` | string   | `0` (disabled)          |
| `snapshotCount`           | `--snapshot-count`            | integer  | version dependent       |

Any field you leave unset simply produces no flag, so etcd's own default applies. KubeDB does not fill in defaults of its own here.

### `quotaBackendBytes`

Caps the size of etcd's backend database file (the bbolt file on disk that holds the keyspace and its MVCC history).

This is a **hard limit with a sharp edge**: when the backend grows past the quota, etcd raises a `NOSPACE` alarm and the entire cluster goes **read-only**. Writes are rejected until an operator compacts history, defragments the backend to actually release the space, and then disarms the alarm. It is not a soft warning — it is the mechanism etcd uses to stop a runaway keyspace from filling the disk and corrupting the cluster.

Practical guidance:

- etcd's own default is 2 GiB and upstream recommends staying below 8 GiB. Going much beyond that trades away recovery time: a larger backend means slower startup, slower snapshot transfer to a new member, and longer defragmentation pauses.
- Set the quota comfortably below the size of the volume in `spec.storage`. The volume also holds the write-ahead log and snapshots, so the backend is not the only thing consuming it.
- Monitor `etcd_mvcc_db_total_size_in_bytes` against `etcd_server_quota_backend_bytes` and alert well before the ratio reaches 1. See [monitoring](/docs/guides/etcd/monitoring/overview.md).

The value is in **bytes**, so `8589934592` is 8 GiB. Note this is a plain integer, not a Kubernetes `resource.Quantity` — `8Gi` is not accepted here.

### `autoCompactionMode` and `autoCompactionRetention`

etcd is an MVCC store: every write creates a new revision and the old ones are kept, so the keyspace grows even under a workload that only ever updates the same keys. **Compaction** discards revisions older than a chosen point. These two fields configure etcd to do it automatically, in the background, instead of requiring a periodic manual `Compact`.

The two fields work as a pair — the mode determines how the retention string is interpreted:

- `autoCompactionMode: periodic` — `autoCompactionRetention` is a **duration**, and etcd keeps at least that much history. `"1h"` means "discard revisions older than one hour". This is the mode you usually want: it bounds history in the terms you actually care about (how far back can a watch or a lagging client rewind?), regardless of write rate.

- `autoCompactionMode: revision` — `autoCompactionRetention` is a **revision count**. `"1000"` means "keep the last 1000 revisions". This bounds history in terms of writes rather than time, which is useful when your write rate is predictable and you want a hard ceiling on growth.

Leaving `autoCompactionRetention` unset (or `"0"`) disables auto-compaction entirely, which means the keyspace grows without bound until you hit `quotaBackendBytes`. For any long-lived cluster, set it.

> **Compaction alone does not shrink the file on disk.** It frees space *inside* the backend database for reuse, but the file stays the same size. To actually return the space to the filesystem you must defragment — use the `Defragment` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md), which walks the members one at a time (leader last) so the cluster never loses quorum while a member is blocked defragmenting.

### `snapshotCount`

The number of committed transactions after which etcd writes a snapshot of its state to disk and truncates the write-ahead log.

The trade-off is between memory/WAL size and recovery cost:

- A **lower** value snapshots more often. The WAL stays small and a restarting member replays less, but each snapshot costs I/O.
- A **higher** value snapshots less often. Fewer I/O spikes, but etcd holds more entries in memory, the WAL directory grows larger, and a member that restarts (or a slow follower catching up) has more to replay.

The main reason to raise it is a slow follower: if a follower falls behind by more than `snapshotCount` entries, the leader can no longer send it the missing log entries and has to ship an entire snapshot instead, which is far more expensive. Raising the count widens the window in which cheap log-based catch-up still works.

## Deploy Etcd with a tuning configuration

Below is an `Etcd` object with all four knobs set:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-custom-config
  namespace: demo
spec:
  version: 3.6.4
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  configuration:
    tuning:
      quotaBackendBytes: 8589934592
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
      snapshotCount: 10000
  deletionPolicy: WipeOut
```

Here,

- `quotaBackendBytes: 8589934592` caps the backend at 8 GiB — comfortably below the 10 GiB volume, which also has to hold the WAL and snapshots.
- `autoCompactionMode: periodic` with `autoCompactionRetention: "1h"` keeps one hour of MVCC history and discards the rest automatically.
- `snapshotCount: 10000` snapshots every 10,000 committed transactions.

Create it:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/custom-configuration/etcd-custom-config.yaml
etcd.kubedb.com/etcd-custom-config created
```

Wait until the cluster is `Ready`:

```bash
$ kubectl get etcd -n demo etcd-custom-config
NAME                 VERSION   STATUS   AGE
etcd-custom-config   3.6.4     Ready    3m
```

## Verify the configuration was applied

Because the knobs are rendered as command-line flags rather than into a config file, the pod spec itself is the source of truth:

```bash
$ kubectl get pod -n demo etcd-custom-config-0 -o jsonpath='{.spec.containers[0].args}' | tr ',' '\n'
["--name=$(POD_NAME)"
"--data-dir=/var/lib/etcd/data"
"--initial-advertise-peer-urls=$(ETCD_MEMBER_PEER_URL)"
"--listen-peer-urls=http://0.0.0.0:2380"
"--listen-client-urls=http://0.0.0.0:2379"
"--advertise-client-urls=$(ETCD_MEMBER_CLIENT_URL)"
"--listen-metrics-urls=http://0.0.0.0:2381"
"--quota-backend-bytes=8589934592"
"--auto-compaction-mode=periodic"
"--auto-compaction-retention=1h"
"--snapshot-count=10000"]
```

The tuning flags are appended after the flags KubeDB owns, in the order `quota-backend-bytes`, `auto-compaction-mode`, `auto-compaction-retention`, `snapshot-count`. Any knob you left unset simply does not appear.

You can also confirm the quota from etcd's own metrics endpoint, which is a good end-to-end check that etcd actually accepted the value rather than just that it was passed:

```bash
$ kubectl exec -it -n demo etcd-custom-config-0 -c etcd -- \
    curl -s http://127.0.0.1:2381/metrics | grep -E '^etcd_server_quota_backend_bytes|^etcd_mvcc_db_total_size_in_bytes'
etcd_server_quota_backend_bytes 8.589934592e+09
etcd_mvcc_db_total_size_in_bytes 2.0480e+04
```

And the settings are, of course, recorded on the `Etcd` object itself:

```bash
$ kubectl get etcd -n demo etcd-custom-config -o jsonpath='{.spec.configuration.tuning}'
{"autoCompactionMode":"periodic","autoCompactionRetention":"1h","quotaBackendBytes":8589934592,"snapshotCount":10000}
```

## Changing the configuration later

Editing `spec.configuration.tuning` on a live `Etcd` object is **not** the way to change these settings. etcd has no live-reload path for command-line flags — there is no equivalent of PostgreSQL's `pg_reload_conf()` — so a change only takes effect when a member restarts, and restarting an etcd cluster safely requires moving Raft leadership around.

Use a `Reconfigure` `EtcdOpsRequest` instead. It patches the `Etcd` object, re-renders the `PetSet`, and then performs a leader-aware rolling restart — followers first, one at a time with quorum checks in between, and the leader last after handing off its leadership. It also only touches the knobs your request actually names, so a partial reconfigure does not silently reset the others:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcdops-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: etcd-custom-config
  configuration:
    tuning:
      autoCompactionRetention: "2h"
```

See the [Reconfigure guide](/docs/guides/etcd/reconfigure/reconfigure.md) for the full walkthrough.

## Cleaning up

```bash
$ kubectl delete etcd -n demo etcd-custom-config
$ kubectl delete ns demo
```

## Next Steps

- Change the tuning knobs on a running cluster with a [Reconfigure EtcdOpsRequest](/docs/guides/etcd/reconfigure/reconfigure.md).
- Watch the backend size against the quota with [monitoring](/docs/guides/etcd/monitoring/overview.md).
- Run your etcd cluster with [TLS/SSL encryption](/docs/guides/etcd/tls/configure-ssl.md).
- Detail concepts of [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
