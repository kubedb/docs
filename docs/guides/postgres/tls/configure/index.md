---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-tls-configure
    name: Postgres TLS/SSL Configuration
    parent: guides-postgres-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in Postgres

`KubeDB` provides support for TLS/SSL encryption with SSLMode (`allow`, `prefer`, `require`, `verify-ca`, `verify-full`) for `Postgres`. This tutorial will show you how to use `KubeDB` to deploy a `Postgres` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.4.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/postgres/tls/configure/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/tls/configure/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Deploy Postgres database with TLS/SSL configuration

As pre-requisite, at first, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a Postgres with TLS/SSL configuration.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=postgres/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
$ kubectl create secret tls postgres-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/postgres-ca created
```

Now, we are going to create an `Issuer` using the `postgres-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: postgres-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: postgres-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/tls/configure/yamls/issuer.yaml
issuer.cert-manager.io/postgres-ca-issuer created
```

### Deploy Postgres cluster with TLS/SSL configuration

Here, our issuer `postgres-ca-issuer`  is ready to deploy a `Postgres` Cluster with TLS/SSL configuration. Below is the YAML for Postgres Cluster that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: demo-pg
  namespace: demo
spec:
  version: "13.2"
  replicas: 3
  standbyMode: Hot
  sslMode: verify-full
  storageType: Durable
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: postgres-ca-issuer
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
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Here,

- `spec.sslMode` specifies the SSL/TLS client connection to the server is required.  

- `spec.tls.issuerRef` refers to the `postgres-ca-issuer` issuer.

- `spec.tls.certificates` gives you a lot of options to configure so that the certificate will be renewed and kept up to date. 
You can found more details from [here](/docs/guides/postgres/concepts/postgres.md#tls)

**Deploy Postgres Cluster:**

Let’s create the `Postgres` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/tls/configure/yamls/tls-postgres.yaml
postgres.kubedb.com/pg created
```

**Wait for the database to be ready:**

Now, watch `Postgres` is going to `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
$  watch kubectl get postgres -n demo pg

Every 2.0s: kubectl get postgres --all-namespaces                       ac-emon: Fri Dec  3 15:14:11 2021

NAMESPACE   NAME   VERSION   STATUS   AGE
demo        pg     13.2      Ready    62s


$ watch -n 3 kubectl get sts -n demo pg
Every 2.0s: kubectl get sts -n demo pg                                  ac-emon: Fri Dec  3 15:15:41 2021

NAME   READY   AGE
pg     3/3     2m30s

$  watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/name=postgreses.kubedb.com,app.kubernetes.io/instance=pg
Every 3.0s: kubectl get pod -n demo -l app.kubernetes.io/name=postg...  ac-emon: Fri Dec  3 15:17:10 2021

NAME   READY   STATUS    RESTARTS   AGE
pg-0   2/2     Running   0          3m59s
pg-1   2/2     Running   0          3m54s
pg-2   2/2     Running   0          3m49s
```

**Verify tls-secrets created successfully:**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` enterprise operator. Default tls-secret name formed as _{postgres-object-name}-{cert-alias}-cert_.

Let's check if the tls-secrets have been created properly,

```bash
$ kubectl get secrets -n demo | grep pg
pg-auth                    kubernetes.io/basic-auth              2      4m41s
pg-client-cert             kubernetes.io/tls                     3      4m40s
pg-metrics-exporter-cert   kubernetes.io/tls                     3      4m40s
pg-server-cert             kubernetes.io/tls                     3      4m41s
pg-token-xvk9p             kubernetes.io/service-account-token   3      4m41s
```

**Verify Postgres Cluster configured with TLS/SSL:**

Now, we are going to connect to the database to verify that `Postgres` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n  demo  pg-0 -- bash
bash-5.1$ ls /tls/certs
client    exporter  server

bash-5.1$ ls /tls/certs/server
ca.crt      server.crt  server.key

bash-5.1$ psql
psql (13.2)
Type "help" for help.

postgres=# SELECT * FROM pg_stat_ssl;
 pid  | ssl | version |         cipher         | bits | compression | client_dn | client_serial | issuer_dn 
------+-----+---------+------------------------+------+-------------+-----------+---------------+-----------
  129 | t   | TLSv1.3 | TLS_AES_256_GCM_SHA384 |  256 | f           |           |               | 
  130 | t   | TLSv1.3 | TLS_AES_256_GCM_SHA384 |  256 | f           |           |               | 
 2175 | f   |         |                        |      |             |           |               | 
(3 rows)

postgres=# exit

bash-5.1$ cat /var/pv/data/postgresql.conf  | grep ssl
ssl =on
ssl_cert_file ='/tls/certs/server/server.crt'
ssl_key_file ='/tls/certs/server/server.key'
ssl_ca_file ='/tls/certs/server/ca.crt'
primary_conninfo = 'application_name=pg-0 host=pg user=postgres password=0WpDlAbHsrNs-7hp sslmode=verify-full sslrootcert=/tls/certs/client/ca.crt'
#ssl = off
#ssl_ca_file = ''
#ssl_cert_file = 'server.crt'
#ssl_crl_file = ''
#ssl_key_file = 'server.key'
#ssl_ciphers = 'HIGH:MEDIUM:+3DES:!aNULL' # allowed SSL ciphers
#ssl_prefer_server_ciphers = on
#ssl_ecdh_curve = 'prime256v1'
#ssl_min_protocol_version = 'TLSv1.2'
#ssl_max_protocol_version = ''
#ssl_dh_params_file = ''
#ssl_passphrase_command = ''
#ssl_passphrase_command_supports_reload = off

```

The above output shows that the `Postgres` server is configured with TLS/SSL configuration and in `/var/pv/data/postgresql.conf ` you can see that `ssl= on`. You can also see that the `.crt` and `.key` files are stored in the `/tls/certs/` directory for client and server.

**Verify secure connection for SSL required user:**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```bash
# creating SSL required user
$ kubectl exec -it -n  demo  pg-0 -- bash

bash-5.1$ psql -d "user=postgres password=$POSTGRES_PASSWORD host=pg port=5432 connect_timeout=15 dbname=postgres sslmode=verify-full sslrootcert=/tls/certs/client/ca.crt"
psql (13.2)
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

postgres=# exit

bash-5.1$ psql -d "user=postgres password=$POSTGRES_PASSWORD host=pg port=5432 connect_timeout=15 dbname=postgres sslmode=verify-full"
psql: error: root certificate file "/var/lib/postgresql/.postgresql/root.crt" does not exist
Either provide the file or change sslmode to disable server certificate verification.
```

From the above output, you can see that only using ca certificate we can access the database securely, otherwise, it ask for the ca verification. Our client certificate is stored in `ls /tls/certs/client` directory.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pg -n demo  pg
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).