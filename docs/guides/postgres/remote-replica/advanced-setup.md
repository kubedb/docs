---
title: PostgreSQL Cross-Cluster DR Setup with Bidirectional Failover
menu:
  docs_{{ .version }}:
    identifier: pg-remote-replica-advanced-setup
    name: Advanced DR Setup
    parent: pg-remote-replica
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL Cross-Cluster Disaster Recovery with Bidirectional Failover

This guide walks through a production-grade Disaster Recovery (DR) setup for KubeDB-managed PostgreSQL across two Kubernetes clusters in different regions. You will:

- Deploy a 3-replica HA PostgreSQL cluster in the **primary region** (Singapore)
- Replicate it live to the **DR region** (London) as a remote replica
- Perform a **failover**: promote London to primary when Singapore goes down
- Bring Singapore back online as a remote replica of the new primary (London)
- Perform a **failback**: promote Singapore again and reconnect London as the DR

> Note: YAML files used in this tutorial are stored in [docs/guides/postgres/remote-replica/advanced-setup-yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

You need:

- **Two Kubernetes clusters** — one for each region. This guide uses `KUBECONFIG_PRIMARY` (Singapore) and `KUBECONFIG_DR` (London).
- **KubeDB operator** installed on both clusters — follow [setup instructions](/docs/setup/README.md).
- **cert-manager** installed on both clusters — follow [cert-manager.io/docs/installation](https://cert-manager.io/docs/installation/).
- **kubectl** and **kubectl-dba** plugin on your workstation.

Export your kubeconfig paths:

```bash
export KUBECONFIG_PRIMARY=/path/to/singapore-kubeconfig.yaml
export KUBECONFIG_DR=/path/to/london-kubeconfig.yaml
```

Create the `demo` namespace on both clusters:

```bash
$ kubectl create ns demo --kubeconfig $KUBECONFIG_PRIMARY
namespace/demo created

$ kubectl create ns demo --kubeconfig $KUBECONFIG_DR
namespace/demo created
```

## Architecture

```
 ┌──────────────────────────────────┐         ┌──────────────────────────────────┐
 │  Cluster: Singapore (Primary)    │         │  Cluster: London (DR)            │
 │                                  │         │                                  │
 │  pg-singapore-0  (primary)       │◄───────►│  pg-london-0  (remote replica)   │
 │  pg-singapore-1  (hot standby)   │  WAL    │  pg-london-1  (hot standby)      │
 │  pg-singapore-2  (hot standby)   │ stream  │  pg-london-2  (hot standby)      │
 │                                  │         │                                  │
 │  ingress-nginx → :5432           │         │  ingress-nginx → :5432           │
 │  ExternalIP: <PRIMARY_IP>        │         │  ExternalIP: <DR_IP>             │
 └──────────────────────────────────┘         └──────────────────────────────────┘
```

> **Key design principle:** Both clusters use the **same CA certificate** to issue TLS certificates. This allows mutual TLS verification across clusters — no separate CA bundle exchange needed.

## Step 1: Generate a Shared CA Certificate

Generate the CA once on your workstation:

```bash
$ openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout ca.key -out ca.crt \
    -subj "/CN=postgres/O=kubedb"
```

Create the `postgres-ca` secret on **both** clusters with the **same** CA files:

```bash
$ kubectl create secret tls postgres-ca \
    --cert=ca.crt --key=ca.key \
    --namespace=demo --kubeconfig $KUBECONFIG_PRIMARY
secret/postgres-ca created

$ kubectl create secret tls postgres-ca \
    --cert=ca.crt --key=ca.key \
    --namespace=demo --kubeconfig $KUBECONFIG_DR
secret/postgres-ca created
```

## Step 2: Create Issuers

Create a cert-manager `Issuer` backed by the shared CA on **both** clusters:

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-issuer.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
issuer.cert-manager.io/pg-issuer created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-issuer.yaml \
    --kubeconfig $KUBECONFIG_DR
issuer.cert-manager.io/pg-issuer created
```

Verify both are Ready:

```bash
$ kubectl get issuer pg-issuer -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME        READY   AGE
pg-issuer   True    0s

$ kubectl get issuer pg-issuer -n demo --kubeconfig $KUBECONFIG_DR
NAME        READY   AGE
pg-issuer   True    0s
```

## Step 3: Create Auth Secrets

> **Critical:** Both clusters must use **the same password**. When the remote replica performs `pg_basebackup` from the primary, it inherits the primary's password hashes. The local `authSecret` must match.

Create these secrets on **both** clusters:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pg-singapore-auth
  namespace: demo
stringData:
  username: postgres
  password: <your-strong-password>
type: kubernetes.io/basic-auth
---
apiVersion: v1
kind: Secret
metadata:
  name: pg-london-auth
  namespace: demo
stringData:
  username: postgres
  password: <your-strong-password>
type: kubernetes.io/basic-auth
```

```bash
$ kubectl apply \
    -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-singapore-auth.yaml \
    -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-london-auth.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
secret/pg-singapore-auth created
secret/pg-london-auth created

$ kubectl apply \
    -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-singapore-auth.yaml \
    -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-london-auth.yaml \
    --kubeconfig $KUBECONFIG_DR
secret/pg-singapore-auth created
secret/pg-london-auth created
```

## Step 4: Expose PostgreSQL via ingress-nginx

The remote replica connects to the primary over the public/external IP. We use ingress-nginx TCP passthrough on port 5432.

Install on the **primary cluster** (routes `:5432` → `pg-singapore`):

```bash
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx

$ helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx \
    --namespace demo \
    --set tcp.5432="demo/pg-singapore:5432" \
    --kubeconfig $KUBECONFIG_PRIMARY
```

Install on the **DR cluster** (routes `:5432` → `pg-london`):

```bash
$ helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx \
    --namespace demo \
    --set tcp.5432="demo/pg-london:5432" \
    --kubeconfig $KUBECONFIG_DR
```

Wait for LoadBalancer IPs:

```bash
$ kubectl get svc ingress-nginx-controller -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME                       TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                                     AGE
ingress-nginx-controller   LoadBalancer   10.43.238.105   10.2.0.188     80:32243/TCP,443:32688/TCP,5432:31152/TCP   21s

$ kubectl get svc ingress-nginx-controller -n demo --kubeconfig $KUBECONFIG_DR
NAME                       TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)                                     AGE
ingress-nginx-controller   LoadBalancer   10.43.3.197   10.2.0.30     80:30572/TCP,443:31732/TCP,5432:32539/TCP   8s
```

Note the `EXTERNAL-IP` values — you will need them for `kubectl-dba remote-config`.

## Step 5: Deploy the Primary PostgreSQL Cluster (Singapore)

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-singapore
  namespace: demo
spec:
  authSecret:
    name: pg-singapore-auth
  clientAuthMode: md5
  deletionPolicy: Delete
  replicas: 3
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-singapore.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
postgres.kubedb.com/pg-singapore created
```

Wait until Ready:

```bash
$ kubectl wait pg pg-singapore -n demo \
    --for=jsonpath='{.status.phase}'=Ready \
    --timeout=300s --kubeconfig $KUBECONFIG_PRIMARY
postgres.kubedb.com/pg-singapore condition met

$ kubectl get pg pg-singapore -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME           VERSION   STATUS   AGE
pg-singapore   17.4      Ready    3m11s

$ kubectl get pods -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME                                       READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-d4b9877f6-xp6n4   1/1     Running   0          54m
pg-singapore-0                             2/2     Running   0          3m1s
pg-singapore-1                             2/2     Running   0          2m22s
pg-singapore-2                             2/2     Running   0          104s
```

Each pod shows `2/2` — the `postgres` container and `pg-coordinator` (Raft HA manager).

### Seed test data

```bash
$ kubectl exec -i -n demo pg-singapore-0 -c postgres \
    --kubeconfig $KUBECONFIG_PRIMARY -- \
    psql -U postgres -d postgres -c \
    "CREATE TABLE hello (id SERIAL PRIMARY KEY, msg TEXT);
     INSERT INTO hello (msg) SELECT 'hello from singapore ' || i FROM generate_series(1,1000) i;
     SELECT count(*) FROM hello;"
CREATE TABLE
INSERT 0 1000
 count
-------
  1000
(1 row)
```

Verify replication within the Singapore cluster:

```bash
$ kubectl exec -i -n demo pg-singapore-2 -c postgres \
    --kubeconfig $KUBECONFIG_PRIMARY -- \
    psql -U postgres -d postgres -c "SELECT count(*) FROM hello;"
 count
-------
  1000
(1 row)
```

## Step 6: Generate Remote Replica Configuration

Use `kubectl-dba remote-config` to generate the AppBinding and TLS secrets the DR cluster needs to connect to the primary. Run this against the **primary cluster** from any working directory — the config file is written to the current directory:

```bash
$ kubectl-dba remote-config postgres -n demo pg-singapore \
    -upostgres -p'<your-password>' \
    -d <PRIMARY_EXTERNAL_IP> \
    --auth-secret pg-london-auth \
    -y pg-singapore-remote-config.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
```

This generates `pg-singapore-remote-config.yaml` containing:
- A `Secret` (`pg-london-auth`) — credentials the remote replica uses to authenticate
- A `Secret` (`pg-singapore-client-cert-postgres`) — TLS client certificate signed by the shared CA
- An `AppBinding` (`pg-singapore`) — points to `<PRIMARY_EXTERNAL_IP>:5432` with `sslmode=verify-ca`

Apply the generated file on the **DR cluster**:

```bash
$ kubectl apply -f pg-singapore-remote-config.yaml --kubeconfig $KUBECONFIG_DR
secret/pg-london-auth configured
secret/pg-singapore-client-cert-postgres created
appbinding.appcatalog.appscode.com/pg-singapore created
```

## Step 7: Deploy the Remote Replica (London)

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
    name: pg-london-auth
  clientAuthMode: md5
  standbyMode: Hot
  replicas: 3
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-london-remote-replica.yaml \
    --kubeconfig $KUBECONFIG_DR
postgres.kubedb.com/pg-london created
```

Wait until Ready:

```bash
$ kubectl wait pg pg-london -n demo \
    --for=jsonpath='{.status.phase}'=Ready \
    --timeout=300s --kubeconfig $KUBECONFIG_DR
postgres.kubedb.com/pg-london condition met

$ kubectl get pg pg-london -n demo --kubeconfig $KUBECONFIG_DR
NAME        VERSION   STATUS   AGE
pg-london   17.4      Ready    83s

$ kubectl get pods -n demo --kubeconfig $KUBECONFIG_DR
NAME                                       READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-d4b9877f6-664gs   1/1     Running   0          56m
pg-london-0                                1/1     Running   0          74s
pg-london-1                                1/1     Running   0          66s
pg-london-2                                1/1     Running   0          59s
```

Remote replica pods show `1/1` — no coordinator (remote replicas are managed by init scripts directly, without Raft leader election).

## Step 8: Verify Cross-Cluster Replication

Verify the seeded data arrived:

```bash
$ kubectl exec -i -n demo pg-london-0 -c postgres \
    --kubeconfig $KUBECONFIG_DR -- \
    psql -U postgres -d postgres -c "SELECT count(*) FROM hello;"
 count
-------
  1000
(1 row)
```

Insert a new row on the primary and confirm it streams to London within seconds:

```bash
$ kubectl exec -i -n demo pg-singapore-0 -c postgres \
    --kubeconfig $KUBECONFIG_PRIMARY -- \
    psql -U postgres -d postgres -c \
    "INSERT INTO hello (msg) VALUES ('live replication test'); SELECT count(*) FROM hello;"
INSERT 0 1
 count
-------
  1001
(1 row)

$ kubectl exec -i -n demo pg-london-0 -c postgres \
    --kubeconfig $KUBECONFIG_DR -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello; SELECT msg FROM hello ORDER BY id DESC LIMIT 1;"
 count
-------
  1001
(1 row)

          msg
-----------------------
 live replication test
(1 row)
```

Cross-cluster WAL streaming is confirmed.

## Step 9: Failover — Promote the DR Cluster (London)

This section simulates a primary region failure. Singapore goes down; London is promoted to primary.

### 9.1 Delete the Primary

```bash
$ kubectl delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-singapore.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
postgres.kubedb.com "pg-singapore" deleted
```

### 9.2 Promote London to Standalone Primary

Apply the standalone spec (removes `remoteReplica`) and restart the pods:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-london.yaml \
    --kubeconfig $KUBECONFIG_DR
postgres.kubedb.com/pg-london configured

$ kubectl delete pods -n demo \
    -l "app.kubernetes.io/instance=pg-london,app.kubernetes.io/component=database" \
    --kubeconfig $KUBECONFIG_DR
pod "pg-london-0" deleted
pod "pg-london-1" deleted
pod "pg-london-2" deleted
```

Wait for pg-london to come up as a standalone HA cluster:

```bash
$ kubectl wait pg pg-london -n demo \
    --for=jsonpath='{.status.phase}'=Ready \
    --timeout=300s --kubeconfig $KUBECONFIG_DR
postgres.kubedb.com/pg-london condition met

$ kubectl get pg pg-london -n demo --kubeconfig $KUBECONFIG_DR
NAME        VERSION   STATUS   AGE
pg-london   17.4      Ready    3m1s

$ kubectl get pods -n demo --kubeconfig $KUBECONFIG_DR
NAME                                       READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-d4b9877f6-664gs   1/1     Running   0          58m
pg-london-0                                2/2     Running   0          19s
pg-london-1                                2/2     Running   0          16s
pg-london-2                                2/2     Running   0          13s
```

Pods show `2/2` — the coordinator is running again, pg-london is a full HA cluster.

### 9.3 Verify Data Integrity After Failover

All prior data is intact and writes are accepted:

```bash
$ kubectl exec -i -n demo pg-london-0 -c postgres \
    --kubeconfig $KUBECONFIG_DR -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello;
     INSERT INTO hello (msg) VALUES ('inserted after london promoted to primary');
     SELECT count(*) FROM hello;"
 count
-------
  1001
(1 row)

INSERT 0 1

 count
-------
  1002
(1 row)
```

Verify HA replication within the London cluster:

```bash
$ kubectl exec -i -n demo pg-london-2 -c postgres \
    --kubeconfig $KUBECONFIG_DR -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello; SELECT msg FROM hello ORDER BY id DESC LIMIT 1;"
 count
-------
  1002
(1 row)

                    msg
-------------------------------------------
 inserted after london promoted to primary
(1 row)
```

## Step 10: Bring Singapore Back as a Remote Replica

### 10.1 Generate Remote Config from London (now primary)

```bash
$ kubectl-dba remote-config postgres -n demo pg-london \
    -upostgres -p'<your-password>' \
    -d <DR_EXTERNAL_IP> \
    --auth-secret pg-singapore-auth \
    -y pg-london-remote-config.yaml \
    --kubeconfig $KUBECONFIG_DR
```

Apply the generated config on the **primary (Singapore) cluster**:

```bash
$ kubectl apply -f pg-london-remote-config.yaml --kubeconfig $KUBECONFIG_PRIMARY
secret/pg-singapore-auth configured
secret/pg-london-client-cert-postgres created
appbinding.appcatalog.appscode.com/pg-london created
```

### 10.2 Deploy Singapore as Remote Replica of London

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-singapore
  namespace: demo
spec:
  remoteReplica:
    sourceRef:
      name: pg-london
      namespace: demo
  authSecret:
    name: pg-london-auth
  clientAuthMode: md5
  standbyMode: Hot
  replicas: 3
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-singapore-remote-replica.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
postgres.kubedb.com/pg-singapore created

$ kubectl wait pg pg-singapore -n demo \
    --for=jsonpath='{.status.phase}'=Ready \
    --timeout=600s --kubeconfig $KUBECONFIG_PRIMARY
postgres.kubedb.com/pg-singapore condition met

$ kubectl get pg pg-singapore -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME           VERSION   STATUS   AGE
pg-singapore   17.4      Ready    3m5s

$ kubectl get pods -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME                                       READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-d4b9877f6-xp6n4   1/1     Running   0          61m
pg-singapore-0                             1/1     Running   0          3m1s
pg-singapore-1                             1/1     Running   0          2m26s
pg-singapore-2                             1/1     Running   0          101s
```

### 10.3 Verify Recovery and Live Replication

All data is present including what was written while London was primary:

```bash
$ kubectl exec -i -n demo pg-singapore-0 -c postgres \
    --kubeconfig $KUBECONFIG_PRIMARY -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello; SELECT msg FROM hello ORDER BY id DESC LIMIT 1;"
 count
-------
  1002
(1 row)

                    msg
-------------------------------------------
 inserted after london promoted to primary
(1 row)
```

Insert a new row on London (current primary) and verify it streams to Singapore:

```bash
$ kubectl exec -i -n demo pg-london-0 -c postgres \
    --kubeconfig $KUBECONFIG_DR -- \
    psql -U postgres -d postgres -c \
    "INSERT INTO hello (msg) VALUES ('streaming to singapore after role reversal');
     SELECT count(*) FROM hello;"
INSERT 0 1
 count
-------
  1003
(1 row)

$ kubectl exec -i -n demo pg-singapore-0 -c postgres \
    --kubeconfig $KUBECONFIG_PRIMARY -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello; SELECT msg FROM hello ORDER BY id DESC LIMIT 1;"
 count
-------
  1003
(1 row)

                    msg
--------------------------------------------
 streaming to singapore after role reversal
(1 row)
```

## Step 11: Failback — Restore Singapore as Primary

To return to the original topology (Singapore primary, London DR), repeat the failover steps with the roles reversed.

### 11.1 Delete London, Promote Singapore

```bash
$ kubectl delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-london.yaml \
    --kubeconfig $KUBECONFIG_DR
postgres.kubedb.com "pg-london" deleted

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-singapore.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
postgres.kubedb.com/pg-singapore configured

$ kubectl delete pods -n demo \
    -l "app.kubernetes.io/instance=pg-singapore,app.kubernetes.io/component=database" \
    --kubeconfig $KUBECONFIG_PRIMARY
pod "pg-singapore-0" deleted
pod "pg-singapore-1" deleted
pod "pg-singapore-2" deleted
```

> **Note:** In some cases after pod restart, `pg-singapore` may stay `NotReady` because the coordinator's gRPC promote call returns exit code 1 when postgres already auto-promoted itself. If this happens, restart `pg-singapore-0` once:
> ```bash
> kubectl delete pod pg-singapore-0 -n demo --kubeconfig $KUBECONFIG_PRIMARY
> ```
> This is a known edge case when transitioning from remote-replica mode to HA mode.

```bash
$ kubectl wait pg pg-singapore -n demo \
    --for=jsonpath='{.status.phase}'=Ready \
    --timeout=300s --kubeconfig $KUBECONFIG_PRIMARY
postgres.kubedb.com/pg-singapore condition met

$ kubectl get pg pg-singapore -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME           VERSION   STATUS   AGE
pg-singapore   17.4      Ready    5m43s

$ kubectl get pods -n demo --kubeconfig $KUBECONFIG_PRIMARY
NAME                                       READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-d4b9877f6-xp6n4   1/1     Running   0          64m
pg-singapore-0                             2/2     Running   0          91s
pg-singapore-1                             2/2     Running   0          50s
pg-singapore-2                             2/2     Running   0          27s
```

Verify data and write to confirm primary is writable:

```bash
$ kubectl exec -i -n demo pg-singapore-0 -c postgres \
    --kubeconfig $KUBECONFIG_PRIMARY -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello;
     INSERT INTO hello (msg) VALUES ('inserted after singapore re-promoted to primary');
     SELECT count(*) FROM hello;"
 count
-------
  1003
(1 row)

INSERT 0 1

 count
-------
  1004
(1 row)
```

### 11.2 Reconnect London as Remote Replica of Singapore

Generate fresh remote config from Singapore (now primary again):

```bash
$ kubectl-dba remote-config postgres -n demo pg-singapore \
    -upostgres -p'<your-password>' \
    -d <PRIMARY_EXTERNAL_IP> \
    --auth-secret pg-london-auth \
    -y pg-singapore-remote-config.yaml \
    --kubeconfig $KUBECONFIG_PRIMARY
```

Apply on the DR cluster and deploy pg-london as remote replica:

```bash
$ kubectl apply -f pg-singapore-remote-config.yaml --kubeconfig $KUBECONFIG_DR
secret/pg-london-auth configured
secret/pg-singapore-client-cert-postgres configured
appbinding.appcatalog.appscode.com/pg-singapore configured

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/advanced-setup-yamls/pg-london-remote-replica.yaml \
    --kubeconfig $KUBECONFIG_DR
postgres.kubedb.com/pg-london created

$ kubectl wait pg pg-london -n demo \
    --for=jsonpath='{.status.phase}'=Ready \
    --timeout=300s --kubeconfig $KUBECONFIG_DR
postgres.kubedb.com/pg-london condition met

$ kubectl get pg pg-london -n demo --kubeconfig $KUBECONFIG_DR
NAME        VERSION   STATUS   AGE
pg-london   17.4      Ready    117s

$ kubectl get pods -n demo --kubeconfig $KUBECONFIG_DR
NAME                                       READY   STATUS    RESTARTS   AGE
ingress-nginx-controller-d4b9877f6-664gs   1/1     Running   0          66m
pg-london-0                                1/1     Running   0          114s
pg-london-1                                1/1     Running   0          107s
pg-london-2                                1/1     Running   0          100s
```

### 11.3 Verify Final State

All data (including writes during both London-as-primary periods) is present on London:

```bash
$ kubectl exec -i -n demo pg-london-0 -c postgres \
    --kubeconfig $KUBECONFIG_DR -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello; SELECT msg FROM hello ORDER BY id DESC LIMIT 1;"
 count
-------
  1004
(1 row)

                       msg
-------------------------------------------------
 inserted after singapore re-promoted to primary
(1 row)
```

Insert on Singapore and verify it streams to London:

```bash
$ kubectl exec -i -n demo pg-singapore-0 -c postgres \
    --kubeconfig $KUBECONFIG_PRIMARY -- \
    psql -U postgres -d postgres -c \
    "INSERT INTO hello (msg) VALUES ('final verification - back to original topology');
     SELECT count(*) FROM hello;"
INSERT 0 1
 count
-------
  1005
(1 row)

$ kubectl exec -i -n demo pg-london-0 -c postgres \
    --kubeconfig $KUBECONFIG_DR -- \
    psql -U postgres -d postgres -c \
    "SELECT count(*) FROM hello; SELECT msg FROM hello ORDER BY id DESC LIMIT 1;"
 count
-------
  1005
(1 row)

                      msg
------------------------------------------------
 final verification - back to original topology
(1 row)
```

Original topology fully restored: Singapore is primary, London is DR, live streaming confirmed.

## Failover Runbook Summary

| Scenario | Steps |
|---|---|
| **Primary down → promote DR** | 1. `kubectl delete -f <primary-standalone>.yaml` on PRIMARY<br>2. `kubectl apply -f <dr-standalone>.yaml` on DR<br>3. Delete DR pods<br>4. Wait for DR `Ready` |
| **Bring old primary back as remote replica** | 1. `kubectl-dba remote-config` from new primary<br>2. `kubectl apply -f <remote-config>.yaml` on old primary cluster<br>3. `kubectl apply -f <old-primary-remote-replica>.yaml` on old primary cluster<br>4. Wait for `Ready` |
| **Failback (reverse the above)** | Repeat the two rows above with clusters swapped |

## Cleaning Up

```bash
$ kubectl delete pg,secret,appbinding,issuer --all -n demo --kubeconfig $KUBECONFIG_PRIMARY
$ kubectl delete pg,secret,appbinding,issuer --all -n demo --kubeconfig $KUBECONFIG_DR
$ kubectl delete ns demo --kubeconfig $KUBECONFIG_PRIMARY
$ kubectl delete ns demo --kubeconfig $KUBECONFIG_DR
```

## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/stash/overview/index.md) PostgreSQL with Stash.
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Configure a [Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md).
- Detail concepts of the [Postgres object](/docs/guides/postgres/concepts/postgres.md).
