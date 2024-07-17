---
title: PostgreSQL Remote Replica
menu:
  docs_{{ .version }}:
    identifier: pg-remote-replica-details
    name: Overview
    parent: pg-remote-replica
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - PostgreSQL Remote Replica

This tutorial will show you how to use KubeDB to provision a PostgreSQL Remote Replica from a KubeDB managed PostgreSQL instance. Remote replica can used in in or across cluster

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/postgres/remote-replica/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Remote Replica

The remote replica allows you to replicate data from an KubeDB managed PostgreSQL server to a read-only PostgreSQL server. The whole process  uses PostgreSQL asynchronous replication to keep up-to-date the replica with  source server.
It's useful to use remote replica to scale of read-intensive workloads, can be a workaround for your  BI and analytical workloads and can be geo-replicated.

## Deploy PostgreSQL server

The following is an example `PostgreSQL` object which creates a PostgreSQL cluster instance.we will create a tls secure instance since were planing to replicated across cluster

Lets start with creating a secret first to access to database and we will deploy a tls secured instance since were replication across cluster

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=postgres/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls pg-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pg-ca created
```

Now, we are going to create an `Issuer` using the `pg-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pg-issuer
  namespace: demo
spec:
  ca:
    secretName: pg-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-issuer.yaml
issuer.cert-manager.io/pg-issuer created
```


### Create Auth Secret

```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cG9zdGdyZXM=
kind: Secret
metadata:
  name: pg-singapore-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

```bash 
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-singapore-auth.yaml
secret/pg-singapore-auth created
```

## Deploy PostgreSQL with TLS/SSL configuration
```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-singapore
  namespace: demo
spec:
  authSecret:
    name: pg-singapore-auth
  allowedSchemas:
    namespaces:
      from: Same
  autoOps: {}
  clientAuthMode: md5
  replicas: 3
  sslMode: verify-ca
  standbyMode: Hot
  streamingMode: Synchronous
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
        storage: 1Gi
    storageClassName: linode-block-storage
  storageType: Durable
  deletionPolicy: WipeOut
  version: "15.5"
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-singapore.yaml
postgres.kubedb.com/pg-singapore created
```
KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created

```bash
$ kubectl get pg -n demo
NAME              VERSION   STATUS   AGE
pg-singapore      15.3      Ready    22h
```

# Exposing to outside world
For Now we will expose our postgresql with ingress with to outside world
```bash
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
$ helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx  \
                                      --namespace demo --create-namespace \
                                      --set tcp.5432="demo/pg-singapore:5432"
```
Let's apply the ingress yaml thats refers to `pg-singpore` service

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: pg-singapore
  namespace: demo  
spec:
  ingressClassName: nginx
  rules:
  - host: pg-singapore.something.org
    http:
      paths:
      - backend:
          service:
            name: pg-singapore
            port:
              number: 5432
        path: /
        pathType: Prefix
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-ingress.yaml
ingress.networking.k8s.io/pg-singapore created
$ kubectl get ingress -n demo
NAME              CLASS   HOSTS                           ADDRESS          PORTS   AGE
pg-singapore      nginx   pg-singapore.something.org      172.104.37.147   80      22h
```

# Prepare for Remote Replica
We wil use the [kubedb_plugin](/docs/setup/README.md) for generating configuration for remote replica. It will create the appbinding and and necessary secrets to connect with source server
```bash
$ kubectl dba remote-config postgres -n demo pg-singapore -uremote -ppass -d 172.104.37.147 -y
home/mehedi/go/src/kubedb.dev/yamls/postgres/pg-singapore-remote-config.yaml
```

#  Create  Remote Replica
We have prepared another cluster in london region for replicating across cluster. follow the installation instruction [above](/docs/README.md).

### Create sourceRef

We will apply the generated config from kubeDB plugin to create the source refs and secrets for it
```bash
$ kubectl apply -f  /home/mehedi/go/src/kubedb.dev/yamls/pg-singapore-remote-config.yaml
secret/pg-singapore-remote-replica-auth created
secret/pg-singapore-client-cert-remote created
appbinding.appcatalog.appscode.com/pg-singapore created
```

### Create remote replica auth
We will need to use the same auth secrets for remote replicas as well since operations like clone also replicated the auth-secrets from source server

```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cG9zdGdyZXM=
kind: Secret
metadata:
  name: pg-london-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-london-auth.yaml
```

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
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
    disableWriteCheck: true
  authSecret:
    name: pg-london-auth
  clientAuthMode: md5
  standbyMode: Hot
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: linode-block-storage
  storageType: Durable
  deletionPolicy: WipeOut
  version: "15.5"
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-london.yaml
postgres.kubedb.com/pg-london created
```

Now we will be able to see kubedb will provision a Remote Replica from the source postgres instance. Lets checkout out the petSet , pvc , pv and services associated with it
.
KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified `PostgreSQL` object:
```bash
$ kubectl get pg -n demo 
NAME           VERSION   STATUS   AGE
pg-london      15.3      Ready    7m17s
```

##  Validate Remote Replica

At this point we want to validate the replication, we can see `pg-london-0` is connected as asynchronous replica

### Validate from source

```bash
$ kubectl exec -it -n demo pg-singapore-0 -c postgres -- psql -c "select * from pg_stat_replication";
  pid   | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |    flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
--------+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+-----------------+-----------------+---------------+------------+-------------------------------
    121 |       10 | postgres | pg-singapore-1   | 10.2.1.13   |                 |       37990 | 2023-10-12 06:53:50.402925+00 |              | streaming | 0/89758A8 | 0/89758A8 | 0/89758A8 | 0/89758A8  | 00:00:00.000745 | 00:00:00.00484  | 00:00:00.004848 |             1 | quorum     | 2023-10-13 05:43:53.817575+00
    209 |       10 | postgres | pg-singapore-2   | 10.2.0.11   |                 |       51270 | 2023-10-12 06:54:15.759067+00 |              | streaming | 0/89758A8 | 0/89758A8 | 0/89758A8 | 0/89758A8  | 00:00:00.000581 | 00:00:00.009797 | 00:00:00.009955 |             1 | quorum     | 2023-10-13 05:43:53.823562+00
 205338 |    16394 | remote   | pg-london-0      | 10.2.1.10   |                 |       34850 | 2023-10-12 20:15:07.751715+00 |              | streaming | 0/89758A8 | 0/89758A8 | 0/89758A8 | 0/89758A8  | 00:00:00.158877 | 00:00:00.163418 | 00:00:00.163425 |             0 | async      | 2023-10-13 05:43:53.900061+00
(3 rows)

### Validate from remote replica

$ kubectl exec -it -n demo pg-london-0 -c postgres -- psql -c "select * from pg_stat_wal_receiver";
 pid  |  status   | receive_start_lsn | receive_start_tli | written_lsn | flushed_lsn | received_tli |      last_msg_send_time       |     last_msg_receipt_time     | latest_end_lsn |        latest_end_time        | slot_name |  sender_host   | sender_port |                                                                                                                                                                                                               conninfo                                                                                                                                                                                                               
------+-----------+-------------------+-------------------+-------------+-------------+--------------+-------------------------------+-------------------------------+----------------+-------------------------------+-----------+----------------+-------------+--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 4813 | streaming | 0/8000000         |                 1 | 0/8DC01E0   | 0/8DC01E0   |            1 | 2023-10-13 05:54:33.812544+00 | 2023-10-13 05:54:33.893159+00 | 0/8DC01E0      | 2023-10-13 05:54:33.812544+pplication_name=walreceiver sslmode=verify-full sslcompression=0 sslcert=/tls/certs/remote/client.crt sslkey=/tls/certs/remote/client.key sslrootcert=/tls/certs/remote/ca.crt sslsni=1 ssl_min_protocol_version=TLSv1.2 gssencmode=prefer krbsrvname=postgres target_session_attrs=any
(1 row)  
## Validation data replication
lets create a a database and insert some data

$ kubectl exec -it -n demo pg-singapore-0 -c postgres -- psql -c "create database hi";
CREATE DATABASE

$ kubectl exec -it -n demo pg-singapore-0 -c postgres -- psql -c "create table tab_1 ( a int); insert into tab_1 values(generate_series(1,5))";
CREATE TABLE
INSERT 0 5

### Validate data on primary
kubectl exec -it -n demo pg-singapore-0 -c postgres -- psql -c "select * from tab_1";
 a 
---
 1
 2
 3
 4
 5
(5 rows)

### Validate data on remote replica

$ kubectl exec -it -n demo pg-london-0 -c postgres -- psql -c "select * from tab_1";
 a 
---
 1
 2
 3
 4
 5
(5 rows)

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pg/pg-singapore
kubectl delete -n demo pg/pg-london
kubectl delete secret -n demo pg-singapore-auth
kubectl delete secret -n demo pg-london-auth
kubectl delete ingres -n demo pg-singapore
kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/overview/index.md) PostgreSQL database using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Learn about [custom PostgresVersions](/docs/guides/postgres/custom-versions/setup.md).
- Want to setup PostgreSQL cluster? Check how to [configure Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy PostgreSQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).