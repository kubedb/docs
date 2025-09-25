---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-tls-configure
    name: SingleStore TLS/SSL Configuration
    parent: guides-sdb-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in SingleStore

`KubeDB` supports providing TLS/SSL encryption (via, `tls` mode) for `SingleStore`. This tutorial will show you how to use `KubeDB` to deploy a `SingleStore` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/singlestore/tls/configure/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/tls/configure/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Deploy SingleStore database with TLS/SSL configuration

As pre-requisite, at first, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a SingleStore standalone and cluster that will be configured with these certificates by `KubeDB` operator.

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=memsql/O=kubedb"
Generating a RSA private key
...........................................................................+++++
........................................................................................................+++++
writing new private key to './ca.key'
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls sdb-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/sdb-ca created
```

Now, we are going to create an `Issuer` using the `sdb-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: sdb-issuer
  namespace: demo
spec:
  ca:
    secretName: sdb-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/tls/configure/examples/issuer.yaml
issuer.cert-manager.io/sdb-issuer created
```

### Deploy SingleStore Cluster with TLS/SSL configuration

Here, our issuer `sdb-issuer`  is ready to deploy a `SingleStore` cluster with TLS/SSL configuration. Below is the YAML for SingleStore Cluster that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-tls
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "700m"
              requests:
                memory: "2Gi"
                cpu: "700m"
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "700m"
                requests:
                  memory: "2Gi"
                  cpu: "700m"                      
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    kind: Secret
    name: license-secret
  deletionPolicy: WipeOut
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: sdb-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
```

Here,

- `spec.tls.issuerRef` refers to the `sdb-issuer` issuer.

- `spec.tls.certificates` gives you a lot of options to configure so that the certificate will be renewed and kept up to date. 
You can found more details from [here](/docs/guides/singlestore/concepts/singlestore.md#spectls)

Let’s create the `SingleStore` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/tls/configure/examples/tls-cluster.yaml
singlestore.kubedb.com/sdb-tls created
```

**Wait for the database to be ready:**

Now, wait for `SingleStore` going on `Running` state and also wait for `PetSet` and its pod to be created and going to `Running` state,

```bash
$ kubectl get sdb,petset -n demo
NAME                             TYPE                  VERSION   STATUS   AGE
singlestore.kubedb.com/sdb-tls   kubedb.com/v1alpha2   8.7.10    Ready    3m57s

NAME                                              AGE
petset.apps.k8s.appscode.com/sdb-tls-aggregator   3m53s
petset.apps.k8s.appscode.com/sdb-tls-leaf         3m50s
```

**Verify tls-secrets created successfully:**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` Ops Manager. Default tls-secret name formed as _{singlestore-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secret -n demo | grep sdb-tls
sdb-tls-client-cert   kubernetes.io/tls          3      5m41s
sdb-tls-root-cred     kubernetes.io/basic-auth   2      5m41s
sdb-tls-server-cert   kubernetes.io/tls 
```

**Verify SingleStore configured with TLS/SSL:**

Now, we are going to connect to the database for verifying the `SingleStore` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo sdb-tls-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)

[memsql@sdb-tls-aggregator-0 /]$ ls etc/memsql/certs
ca.crt	client.crt  client.key	server.crt  server.key
 
[memsql@sdb-tls-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 237
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> show variables like '%ssl%';
+---------------------------------+------------------------------+
| Variable_name                   | Value                        |
+---------------------------------+------------------------------+
| default_user_require_ssl        | OFF                          |
| exporter_ssl_ca                 |                              |
| exporter_ssl_capath             |                              |
| exporter_ssl_cert               |                              |
| exporter_ssl_key                |                              |
| exporter_ssl_key_passphrase     | [redacted]                   |
| have_openssl                    | ON                           |
| have_ssl                        | ON                           |
| jwks_ssl_ca_certificate         |                              |
| node_replication_ssl_only       | OFF                          |
| openssl_version                 | 805306480                    |
| processlist_rpc_json_max_size   | 2048                         |
| ssl_ca                          | /etc/memsql/certs/ca.crt     |
| ssl_capath                      |                              |
| ssl_cert                        | /etc/memsql/certs/server.crt |
| ssl_cipher                      |                              |
| ssl_fips_mode                   | OFF                          |
| ssl_key                         | /etc/memsql/certs/server.key |
| ssl_key_passphrase              | [redacted]                   |
| ssl_last_reload_attempt_time    |                              |
| ssl_last_successful_reload_time |                              |
+---------------------------------+------------------------------+
21 rows in set (0.00 sec)
singlestore> exit
Bye

```

The above output shows that the `SingleStore` server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in `/etc/mysql/certs/` directory for client and server respectively.

**Verify secure connection for SSL required user:**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```bash
$ kubectl exec -it -n demo sdb-tls-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sdb-tls-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 316
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> CREATE USER 'new_user'@'localhost' IDENTIFIED BY '1234' REQUIRE SSL;
Query OK, 0 rows affected (0.05 sec)

singlestore> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.00 sec)

singlestore> exit
Bye

# accessing the database server newly created user with certificates
[memsql@sdb-tls-aggregator-0 /]$ memsql -unew_user -p1234 --ssl-ca=/etc/memsql/certs/ca.crt  --ssl-cert=/etc/memsql/certs/server.crt --ssl-key=/etc/memsql/certs/server.key
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 462
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> exit;
Bye

```

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/memsql/certs/` directory.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete  sdb demo sdb-tls
singlestore.kubedb.com "sdb-tls" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```