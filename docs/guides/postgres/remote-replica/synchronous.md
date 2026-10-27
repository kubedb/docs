---
title: PostgreSQL Zero-Data-Loss Cross-Cluster Replication
menu:
  docs_{{ .version }}:
    identifier: pg-remote-replica-synchronous
    name: Synchronous Cross-Cluster
    parent: pg-remote-replica
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL Zero-Data-Loss Cross-Cluster Replication

An asynchronous remote replica can lose the tail of the WAL if the primary region disappears:
the primary acknowledges a commit before the DR site has it. This guide configures
**synchronous** replication that spans two Kubernetes clusters, so that **every acknowledged
commit is already durable in the DR cluster before the client sees success**.

You will:

- Deploy a 2-replica PostgreSQL cluster in the primary region (Singapore)
- Attach a 1-replica remote replica in the DR region (London)
- Configure `synchronous_standby_names` so **at least one standby in each cluster** must
  confirm every commit
- Prove zero data loss by destroying the entire primary cluster and comparing a content
  fingerprint — with **no promotion** of the DR site

> Note: YAML files used in this tutorial are stored in [docs/guides/postgres/remote-replica/synchronous-yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/synchronous-yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- **Two Kubernetes clusters**, one per region. This guide uses `KUBECONFIG_PRIMARY` (Singapore)
  and `KUBECONFIG_DR` (London).
- **KubeDB operator** on both — see [setup](/docs/setup/README.md).
- **cert-manager** on both — see [cert-manager.io/docs/installation](https://cert-manager.io/docs/installation/).
- **kubectl** and the **kubectl-dba** plugin.

```bash
export KUBECONFIG_PRIMARY=/path/to/singapore-kubeconfig.yaml
export KUBECONFIG_DR=/path/to/london-kubeconfig.yaml

kubectl create ns demo --kubeconfig $KUBECONFIG_PRIMARY
kubectl create ns demo --kubeconfig $KUBECONFIG_DR
```

## How the quorum is chosen

With Singapore at `replicas: 2`, exactly one Singapore pod is primary at any time, so the
primary always has **one local standby plus `pg-london-0`**. Listing all three pods in
priority order and requiring two confirmations therefore always selects one standby from each
cluster, whichever Singapore pod happens to hold the primary role:

```
synchronous_standby_names = FIRST 2 ("pg-london-0","pg-singapore-0","pg-singapore-1")
```

| Primary | Priority 1 | Priority 2 | Priority 3 | Chosen (FIRST 2) |
|---|---|---|---|---|
| `pg-singapore-0` | `pg-london-0` | *(is primary, skipped)* | `pg-singapore-1` | `pg-london-0` + `pg-singapore-1` |
| `pg-singapore-1` | `pg-london-0` | `pg-singapore-0` | *(is primary, skipped)* | `pg-london-0` + `pg-singapore-0` |

`pg-london-0` is first in the list, so the cross-cluster standby is always one of the two.
Combined with `commitLevel: "On"` — which waits for the standby to **write and flush WAL to
disk** — a commit cannot be acknowledged unless London has it on disk.

## Step 1: Shared CA and Issuers

Both clusters issue certificates from the **same CA**, so each side can verify the other
without exchanging separate bundles.

```bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout ca.key -out ca.crt \
    -subj "/CN=postgres/O=kubedb"

for KC in $KUBECONFIG_PRIMARY $KUBECONFIG_DR; do
  kubectl create secret tls postgres-ca \
    --cert=ca.crt --key=ca.key -n demo --kubeconfig $KC
done
```

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pg-issuer
  namespace: demo
spec:
  ca:
    secretName: postgres-ca
```

```bash
for KC in $KUBECONFIG_PRIMARY $KUBECONFIG_DR; do
  kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/synchronous-yamls/pg-issuer.yaml --kubeconfig $KC
done

kubectl get issuer pg-issuer -n demo --kubeconfig $KUBECONFIG_PRIMARY
```

```
NAME        READY   AGE
pg-issuer   True    3s
```

> **Re-running this guide in a namespace that previously used a different CA?** Delete the old
> leaf certificate Secrets first. cert-manager does **not** reissue a certificate just because
> the CA Issuer's backing secret changed — as long as the existing leaf is still valid and
> matches its `Certificate` spec, the stale Secret is reused. Deleting the `Certificate`
> objects is not enough; the `Secret`s must go too. The symptom is the DR replica never
> reaching `Ready`, looping on `Attempting pg_isready on primary`, while the primary logs:
>
> ```
> LOG:  could not accept SSL connection: tlsv1 alert unknown ca
> ```
>
> ```bash
> for KC in $KUBECONFIG_PRIMARY $KUBECONFIG_DR; do
>   kubectl delete certificate --all -n demo --kubeconfig $KC
>   kubectl get secret -n demo --kubeconfig $KC -o name | grep cert | \
>     xargs -r kubectl delete -n demo --kubeconfig $KC
> done
> ```
>
> You can confirm which CA signed a given certificate with:
>
> ```bash
> kubectl get secret pg-singapore-server-cert -n demo \
>   -o jsonpath='{.data.ca\.crt}' --kubeconfig $KUBECONFIG_PRIMARY \
>   | base64 -d | openssl x509 -noout -serial
> ```
>
> The serial must match your `ca.crt` on **both** clusters. A fresh namespace is unaffected.

## Step 2: Auth secrets

> **Critical:** both clusters must use the **same password**. The remote replica seeds itself
> with `pg_basebackup` from the primary and inherits its password hashes, so a mismatched local
> `authSecret` breaks authentication after the seed.

Create the secrets from a shell variable so no password is ever written to a file:

```bash
export PG_PASSWORD='<choose-a-strong-password>'

for KC in $KUBECONFIG_PRIMARY $KUBECONFIG_DR; do
  for NAME in pg-singapore-auth pg-london-auth; do
    kubectl create secret generic $NAME -n demo \
      --type=kubernetes.io/basic-auth \
      --from-literal=username=postgres \
      --from-literal=password="$PG_PASSWORD" \
      --kubeconfig $KC
  done
done
```

## Step 3: Expose PostgreSQL to the other cluster

The DR replica connects to the primary over an external address. This guide uses
ingress-nginx TCP passthrough on 5432.

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx

helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx --namespace demo \
    --set tcp.5432="demo/pg-singapore:5432" --kubeconfig $KUBECONFIG_PRIMARY

helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx --namespace demo \
    --set tcp.5432="demo/pg-london:5432" --kubeconfig $KUBECONFIG_DR

kubectl get svc ingress-nginx-controller -n demo --kubeconfig $KUBECONFIG_PRIMARY
```

```
NAME                       TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)                                     AGE
ingress-nginx-controller   LoadBalancer   10.43.84.237   10.2.0.30     80:32521/TCP,443:30734/TCP,5432:30621/TCP   24s
```

Record the primary's `EXTERNAL-IP` as `PRIMARY_IP`.

## Step 4: Deploy the primary cluster — asynchronous at first

Deploy Singapore **without** synchronous replication:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-singapore
  namespace: demo
spec:
  authSecret:
    kind: Secret
    name: pg-singapore-auth
  clientAuthMode: md5
  deletionPolicy: Delete
  replicas: 2
  sslMode: verify-ca
  standbyMode: Hot
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pg-issuer
      kind: Issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 3Gi
  storageType: Durable
  version: "17.4"
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/synchronous-yamls/pg-singapore.yaml --kubeconfig $KUBECONFIG_PRIMARY

kubectl wait pg pg-singapore -n demo --for=jsonpath='{.status.phase}'=Ready \
    --timeout=600s --kubeconfig $KUBECONFIG_PRIMARY
```

> **Why asynchronous first?** `numSyncReplicas: 2` requires **two** connected standbys. Until
> London exists, Singapore has only one, and *every commit would block indefinitely* — including
> the writes KubeDB performs during bootstrap, so the database would never become `Ready`.
> Enable synchronous replication only after the DR replica is streaming (Step 7).

`replicas: 2` is an even number, so KubeDB also creates a coordinator-only **arbiter** pod to
break Raft leader-election ties. It stores no data and never appears in `pg_stat_replication`.

```bash
kubectl get pods -n demo -L kubedb.com/role --kubeconfig $KUBECONFIG_PRIMARY
```

```
NAME                     READY   STATUS    RESTARTS   AGE   ROLE
pg-singapore-0           2/2     Running   0          72s   primary
pg-singapore-1           2/2     Running   0          65s   standby
pg-singapore-arbiter-0   1/1     Running   0          54s   arbiter
```

### Seed some data

The primary is not always `pg-singapore-0` — Raft elects it — so resolve it by label:

```bash
PRIMARY=$(kubectl get pods -n demo \
  -l "app.kubernetes.io/instance=pg-singapore,kubedb.com/role=primary" \
  -o jsonpath='{.items[0].metadata.name}' --kubeconfig $KUBECONFIG_PRIMARY)

kubectl exec -i -n demo $PRIMARY -c postgres --kubeconfig $KUBECONFIG_PRIMARY -- \
  psql -U postgres -d postgres -c \
  "CREATE TABLE ledger (id SERIAL PRIMARY KEY, amount BIGINT, note TEXT);
   INSERT INTO ledger (amount, note) SELECT i, 'seed-'||i FROM generate_series(1,500) i;
   SELECT count(*) FROM ledger;"
```

```
 count
-------
   500
(1 row)
```

## Step 5: Generate the remote replica configuration

```bash
kubectl-dba remote-config postgres -n demo pg-singapore \
    -upostgres -p"$PG_PASSWORD" \
    -d <PRIMARY_IP> \
    --auth-secret pg-london-auth \
    -y \
    --kubeconfig $KUBECONFIG_PRIMARY
```

The file is written to the current directory as `pg-singapore-remote-config.yaml`. Apply it on
the DR cluster:

```bash
kubectl apply -f pg-singapore-remote-config.yaml --kubeconfig $KUBECONFIG_DR
```

```
secret/pg-london-auth configured
secret/pg-singapore-client-cert-postgres created
appbinding.appcatalog.appscode.com/pg-singapore created
```

## Step 6: Deploy the remote replica

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-london
  namespace: demo
spec:
  remoteReplica:
    sourceRef:
      name: pg-singapore
      namespace: demo
  authSecret:
    kind: Secret
    name: pg-london-auth
  clientAuthMode: md5
  standbyMode: Hot
  replicas: 1
  sslMode: verify-ca
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pg-issuer
      kind: Issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 3Gi
  storageType: Durable
  deletionPolicy: Delete
  version: "17.4"
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/synchronous-yamls/pg-london.yaml --kubeconfig $KUBECONFIG_DR

kubectl wait pg pg-london -n demo --for=jsonpath='{.status.phase}'=Ready \
    --timeout=600s --kubeconfig $KUBECONFIG_DR
```

Confirm both standbys are attached, still asynchronous:

```bash
kubectl exec -i -n demo $PRIMARY -c postgres --kubeconfig $KUBECONFIG_PRIMARY -- \
  psql -U postgres -d postgres -c \
  "SELECT application_name,state,sync_state FROM pg_stat_replication ORDER BY application_name;"
```

```
 application_name |   state   | sync_state
------------------+-----------+------------
 pg-london-0      | streaming | async
 pg-singapore-1   | streaming | async
(2 rows)
```

## Step 7: Enable synchronous replication

Now that both standbys are connected, switch Singapore to synchronous:

```yaml
  streamingMode: Synchronous
  synchronousReplicationConfig:
    mode: First
    numSyncReplicas: 2
    commitLevel: "On"
    standbyNames:
    - pg-london-0
    - pg-singapore-0
    - pg-singapore-1
```

> **Quote `"On"`.** In YAML 1.1 the bare word `On` is a boolean, and the admission webhook
> rejects it with `json: cannot unmarshal bool into Go struct field
> PostgresSynchronousReplicationSpec.spec.synchronousReplicationConfig.commitLevel`. The same
> applies to `Off`.

| Field | Meaning |
|---|---|
| `mode: First` | Priority-based selection: the highest-priority connected standbys in list order |
| `numSyncReplicas: 2` | Two standbys must confirm before a commit is acknowledged |
| `commitLevel: "On"` | Wait until the standby has **written and flushed** WAL to disk |
| `standbyNames` | Explicit priority order; `pg-london-0` first pins the cross-cluster standby |

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/synchronous-yamls/pg-singapore-sync.yaml --kubeconfig $KUBECONFIG_PRIMARY

# the setting is rendered into postgresql.conf at pod start, so restart the data pods
kubectl delete pods -n demo \
  -l "app.kubernetes.io/instance=pg-singapore,app.kubernetes.io/component=database" \
  --kubeconfig $KUBECONFIG_PRIMARY

kubectl wait pg pg-singapore -n demo --for=jsonpath='{.status.phase}'=Ready \
    --timeout=600s --kubeconfig $KUBECONFIG_PRIMARY
```

## Step 8: Verify the replication is synchronous

Re-resolve the primary (leadership may have moved during the restart) and inspect:

```bash
PRIMARY=$(kubectl get pods -n demo \
  -l "app.kubernetes.io/instance=pg-singapore,kubedb.com/role=primary" \
  -o jsonpath='{.items[0].metadata.name}' --kubeconfig $KUBECONFIG_PRIMARY)

kubectl exec -i -n demo $PRIMARY -c postgres --kubeconfig $KUBECONFIG_PRIMARY -- \
  psql -U postgres -d postgres -c \
  "SHOW synchronous_standby_names; SHOW synchronous_commit;"
```

```
        synchronous_standby_names
-------------------------------------------------------
 FIRST 2 ("pg-london-0","pg-singapore-0","pg-singapore-1")
(1 row)

 synchronous_commit
--------------------
 on
(1 row)
```

```bash
kubectl exec -i -n demo $PRIMARY -c postgres --kubeconfig $KUBECONFIG_PRIMARY -- \
  psql -U postgres -d postgres -c \
  "SELECT application_name,state,sync_state,sync_priority FROM pg_stat_replication ORDER BY sync_priority;"
```

```
 application_name |   state   | sync_state | sync_priority
------------------+-----------+------------+---------------
 pg-london-0      | streaming | sync       |             1
 pg-singapore-1   | streaming | sync       |             3
(2 rows)
```

Both standbys report `sync_state = sync`, and one of them is in the DR cluster. Priority 2
(`pg-singapore-0`) is absent because it is the primary.

## Step 9: Prove zero data loss

Write a batch of rows. Every one of these commits returns only after London has flushed the
corresponding WAL to disk:

```bash
kubectl exec -i -n demo $PRIMARY -c postgres --kubeconfig $KUBECONFIG_PRIMARY -- \
  psql -U postgres -d postgres -c \
  "INSERT INTO ledger (amount, note) SELECT i, 'sync-'||i FROM generate_series(501,5500) i;
   SELECT count(*) FROM ledger;"
```

Take a content fingerprint on both sides — row count, checksum of the values, and an MD5 over
the ordered contents:

```bash
FINGERPRINT="SELECT count(*)||' | '||coalesce(sum(amount),0)||' | '||md5(string_agg(id||':'||amount||':'||note, ',' ORDER BY id)) FROM ledger;"

kubectl exec -i -n demo $PRIMARY -c postgres --kubeconfig $KUBECONFIG_PRIMARY -- \
  psql -U postgres -d postgres -tAc "$FINGERPRINT"

kubectl exec -i -n demo pg-london-0 -c postgres --kubeconfig $KUBECONFIG_DR -- \
  psql -U postgres -d postgres -tAc "$FINGERPRINT"
```

```
5500 | 15127750 | a9707c95783d7728772819064282790e
5500 | 15127750 | a9707c95783d7728772819064282790e
```

### Destroy the primary cluster

`deletionPolicy: Delete` removes the PVCs along with the database — the primary region is gone
for good, not merely stopped:

```bash
kubectl delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/synchronous-yamls/pg-singapore-sync.yaml --kubeconfig $KUBECONFIG_PRIMARY

kubectl get pods,pvc -n demo --kubeconfig $KUBECONFIG_PRIMARY
```

### Verify the DR cluster still holds everything

```bash
kubectl exec -i -n demo pg-london-0 -c postgres --kubeconfig $KUBECONFIG_DR -- \
  psql -U postgres -d postgres -tAc "$FINGERPRINT"

kubectl exec -i -n demo pg-london-0 -c postgres --kubeconfig $KUBECONFIG_DR -- \
  psql -U postgres -d postgres -tAc "SELECT pg_is_in_recovery();"
```

```
5500 | 15127750 | a9707c95783d7728772819064282790e
t
```

The fingerprint is byte-for-byte identical to the primary's last known state, and
`pg_is_in_recovery()` is still `t` — **London was never promoted**. Nothing was recovered or
replayed after the fact; the data was already durable in the DR cluster at commit time.

To turn London into a writable primary, see
[Cross-Cluster DR with Bidirectional Failover](/docs/guides/postgres/remote-replica/advanced-setup.md).

## Recommended: run the primary with three replicas

The 2-replica layout above has no slack. The primary has exactly two standbys and both must
confirm every commit, so **losing either one blocks writes** until it comes back.

Going to `replicas: 3` fixes that, and it is close to free:

```yaml
  replicas: 3
  synchronousReplicationConfig:
    mode: First
    numSyncReplicas: 2
    commitLevel: "On"
    standbyNames:
    - pg-london-0
    - pg-singapore-0
    - pg-singapore-1
    - pg-singapore-2
```

**It costs no extra pods.** `replicas: 2` is an even number, so KubeDB adds a coordinator-only
**arbiter** pod to break Raft ties. `replicas: 3` is odd and needs none — so both layouts run
three pods in the primary cluster, except that with three replicas the third pod holds data and
can serve reads instead of only voting.

| Layout | Data pods | Arbiter | Total | Tolerates a standby loss |
|---|---|---|---|---|
| `replicas: 2` | 2 | 1 | 3 | no — writes block |
| `replicas: 3` | 3 | 0 | 3 | **yes** |

With four names and `FIRST 2`, the primary keeps a **spare** standby in reserve:

```
 application_name |   state   | sync_state | sync_priority
------------------+-----------+------------+---------------
 pg-london-0      | streaming | sync       |             1
 pg-singapore-0   | streaming | sync       |             2
 pg-singapore-2   | streaming | potential  |             4
(3 rows)
```

`pg-singapore-2` is `potential` — connected and caught up, but not currently counted. Lose the
active local standby and PostgreSQL promotes the spare with no operator action:

```
 application_name |   state   | sync_state | sync_priority
------------------+-----------+------------+---------------
 pg-london-0      | streaming | sync       |             1
 pg-singapore-2   | streaming | sync       |             4
(2 rows)
```

Writes are unaffected — measured at 0.09 s with a standby down versus 0.12 s with the full
quorum — and the row is still on the DR site before the commit returns, so the zero-RPO
guarantee is intact throughout.

> **What three replicas does *not* buy you.** If **London** is the standby that fails, `FIRST 2`
> falls back to the two remaining *local* standbys. Writes keep flowing, but they are no longer
> being confirmed by the DR cluster — the zero-RPO guarantee is silently downgraded to
> single-cluster durability. This is inherent to `FIRST N` whenever the list is longer than `N`:
> you cannot both keep accepting writes and guarantee cross-cluster durability while the DR site
> is unreachable, because those are contradictory requirements. Decide which one you want:
>
> - **Durability first** — list only `pg-london-0` plus one local standby, so a London outage
>   blocks writes rather than weakening the guarantee.
> - **Availability first** — the four-name list above, plus monitoring on
>   `pg_stat_replication` so you notice when `pg-london-0` stops being `sync`.

## Operational notes

- **Reads on the DR site** are served normally (`standbyMode: Hot`), so London can carry
  read-only traffic.
- **Adding Singapore replicas** means extending `standbyNames`; names not present in the list
  never become synchronous.
- **Alert on the sync set, not just on replication.** `sync_state` dropping from `sync` to
  `potential` for `pg-london-0` is the signal that cross-cluster durability has been lost, and
  nothing else in the system will complain about it.

## Cleaning up

```bash
kubectl delete pg pg-london -n demo --kubeconfig $KUBECONFIG_DR
kubectl delete pg pg-singapore -n demo --kubeconfig $KUBECONFIG_PRIMARY
kubectl delete ns demo --kubeconfig $KUBECONFIG_PRIMARY
kubectl delete ns demo --kubeconfig $KUBECONFIG_DR
```

## Next Steps

- [Cross-Cluster DR with Bidirectional Failover](/docs/guides/postgres/remote-replica/advanced-setup.md)
- [Remote Replica overview](/docs/guides/postgres/remote-replica/remotereplica.md)
- Configure a [Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor with the [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md)
