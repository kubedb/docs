---
title: Reconfigure PgBouncer TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: pb-reconfigure-tls-cluster
    name: Reconfigure PgBouncer TLS/SSL Encryption
    parent: pb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure PgBouncer TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates, changing issuer for existing PgBouncer database via a PgBouncerOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `cert-manger` v1.0.0 or later to your cluster to manage your SSL/TLS certificates from [here](https://cert-manager.io/docs/installation/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Prepare Postgres
For a PgBouncer surely we will need a Postgres server so, prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

Now, we are going to deploy a  `PgBouncer` using a supported version by `KubeDB` operator. Then we are going to apply `PgBouncerOpsRequest` to reconfigure its configuration.

### Prepare PgBouncer

Now, we are going to deploy a `PgBouncer` with version `1.18.0`.

## Add TLS to a PgBouncer database

Here, We are going to create a PgBouncer database without TLS and then reconfigure the database to use TLS.

### Deploy PgBouncer without TLS

In this section, we are going to deploy a PgBouncer Replicaset database without TLS. In the next few sections we will reconfigure TLS using `PgBouncerOpsRequest` CRD. Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    poolMode: session
    port: 5432
    reservePoolSize: 5
    maxClientConnections: 87
    defaultPoolSize: 2
    minPoolSize: 1
    authType: md5
  deletionPolicy: WipeOut
```

Let's create the `PgBouncer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/pb.yaml
pgbouncer.kubedb.com/pb created
```

Now, wait until `pb` has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
NAME   VERSION   STATUS   AGE
pb     1.18.0    Ready    131m

$ kubectl dba describe pgbouncer pb -n demo
Name:         pb
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1
Kind:         PgBouncer
Metadata:
  Creation Timestamp:  2025-01-25T09:21:57Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  157918
  UID:               d3635680-b216-47db-8e4f-aac7e42f196c
Spec:
  Auth Secret:
    Name:  pb-auth
  Auto Ops:
  Connection Pool:
    Auth Type:               md5
    Default Pool Size:       2
    Max Client Connections:  87
    Min Pool Size:           1
    Pool Mode:               session
    Port:                    5432
    Reserve Pool Size:       5
  Database:
    Database Name:  postgres
    Database Ref:
      Name:         ha-postgres
      Namespace:    demo
    Sync Users:     true
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  pgbouncer
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Privileged:    false
          Run As Group:  70
          Run As User:   70
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:            70
        Run As Group:        70
        Run As User:         70
      Service Account Name:  pb
  Replicas:                  1
  Ssl Mode:                  disable
  Version:                   1.18.0
Status:
  Conditions:
    Last Transition Time:  2025-01-25T09:22:17Z
    Message:               The KubeDB operator has started the provisioning of PgBouncer: demo/pb
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-01-25T09:22:29Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-01-25T09:22:49Z
    Message:               pgBouncer demo/pb is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-01-25T09:22:49Z
    Message:               pgBouncer demo/pb is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-01-25T09:23:01Z
    Message:               The PgBouncer: demo/pb is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:                    <none>
```

Now, we can verify that the TLS is disabled.

$ kubectl exec -it -n demo pb-0 -- /bin/sh
cat /etc/config/pgbouncer.ini
[databases]
postgres= host=ha-postgres.demo.svc port=5432 dbname=postgres

[pgbouncer]
max_client_conn = 87
min_pool_size = 1
reserve_pool_size = 5
max_user_connections = 2
listen_addr = *
admin_users =  pgbouncer
pool_mode = session
reserve_pool_timeout = 5
max_db_connections = 1
logfile = /tmp/pgbouncer.log
auth_file =  /var/run/pgbouncer/secret/userlist
listen_port = 5432
default_pool_size = 2
stats_period = 60
pidfile = /tmp/pgbouncer.pid
auth_type = md5
ignore_startup_parameters = extra_float_digits
```
Here we can see `client_tls_sslmode` is not present. That means it is by default in `disable` mode.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in PgBouncer. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=pgbouncer/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls pgbouncer-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pgbouncer-ca created
```

Now, Let's create an `Issuer` using the `pgbouncer-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pb-issuer
  namespace: demo
spec:
  ca:
    secretName: pgbouncer-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/pb-issuer created
```

```bash
$ kubectl get issuer -n demo
NAME        READY   AGE
pb-issuer   True    30s
```
Issuer is ready(true).

### Create PgBouncerOpsRequest

In order to add TLS to the database, we have to create a `PgBouncerOpsRequest` CRO with our created issuer. Below is the YAML of the `PgBouncerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pb
  tls:
    sslMode: verify-full
    clientAuthMode: md5
    issuerRef:
      name: pb-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - pgbouncer
          organizationalUnits:
            - client
  apply: Always
```
Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/add-tls.yaml
pgbounceropsrequest.ops.kubedb.com/add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ kubectl get pbops -n demo add-tls 
NAME      TYPE             STATUS       AGE
add-tls   ReconfigureTLS   Successful   2m27s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. 

Now, Let's exec into a database primary pods to see if certificates are added there.
```bash
$ kubectl exec -it -n demo pb-0 -- /bin/sh
/ $ cat /etc/config/pgbouncer.ini
[databases]
postgres= host=ha-postgres.demo.svc port=5432 dbname=postgres

[pgbouncer]
max_db_connections = 1
max_user_connections = 2
auth_type = md5
ignore_startup_parameters = extra_float_digits
pidfile = /tmp/pgbouncer.pid
auth_file =  /var/run/pgbouncer/secret/userlist
min_pool_size = 1
stats_period = 60
client_tls_cert_file = /var/run/pgbouncer/tls/serving/server/tls.crt
reserve_pool_timeout = 5
pool_mode = session
max_client_conn = 87
logfile = /tmp/pgbouncer.log
listen_addr = *
client_tls_sslmode = verify-full
admin_users =  pgbouncer
listen_port = 5432
reserve_pool_size = 5
client_tls_ca_file = /var/run/pgbouncer/tls/serving/server/ca.crt
client_tls_key_file = /var/run/pgbouncer/tls/serving/server/tls.key
default_pool_size = 2
```
Here we can see the presence of `client_tls_sslmode`, `client_tls_cert_file`, `client_tls_ca_file` and `client_tls_key_file`.

## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
kubectl get secrets -n demo pb-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -dates
notBefore=Jan 25 11:39:53 2025 GMT
notAfter=Jan 25 11:39:53 2026 GMT
```

So, the certificate will expire on this time `Jan 25 11:39:53 2026 GMT`. 

### Create PgBouncerOpsRequest

Now we are going to increase it using a PgBouncerOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pb
  tls:
    rotateCertificates: true

```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/rotate-tls.yaml
pgbounceropsrequest.ops.kubedb.com/rotate-tls created
```

#### Verify Certificate Rotated Successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ kubectl get pbops -n demo rotate-tls
NAME         TYPE             STATUS       AGE
rotate-tls   ReconfigureTLS   Successful   109s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. And we can check that the tls.crt has been updated.
```bash
$  kubectl get secrets -n demo pb-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
notBefore=Jan 25 11:53:14 2025 GMT
notAfter=Apr 25 11:53:14 2025 GMT

$ kubectl get secrets -n demo pb-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
notBefore=Jan 25 11:53:14 2025 GMT
notAfter=Apr 25 11:53:14 2025 GMT
```


As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=pgbouncer/O=kubedb-updated"
........+....+.....+...+....+...+..+.+...+..+...+............+....+.........+..+.......+.....+..........+..+.+...+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+..+.............+......+...........+....+.....+....+........+....+...+...+..+.......+......+..+.+.....+.+...+..+......+....+...........+.......+...+.........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.....+...........+.......+...+.....+....+..+.+..+............+.........+.+......+.........+......+.....+............+............+...+.+..+....+...+........+...+.+......+.........+........+.+..+...+.......+........+.+...........+...+....+..+...............+..........+...........+...+.+.....+.+...+............+...+...........+......+.......+...+...+..+............+..........+............+.........+.....+.+.....+....+...........+.+..+.+............+........+.......+........+......+..................+.......+........+...+...+....+..................+..+.......+...+........+....+.....+....+.........+...+...+......+...+..+..............................+...+......+......+.............+...+..+......+....+..+.........+............+....+...+........+...+.+........+.......+.....+...+......+..........+..+.......+.....+..........+...+........+....+..+.+..............+.............+...+..+..........+..+...................+..+...+.+...+...........+.+...+...........+................+..............+.........+......+....+..+..........+.....+.+..+...+....+.....+......+....+.........+..+...+....+......+..............+...+...+.+...........+...+.......+..+.+...........+...+.+.....+.+...+...+..............................+...+......+..............+.+........+.+....................+......+.........+.+...........+....+.....+......+.......+...+..+.+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
...........+.......+.....+...+....+..+...................+..+....+...+...+...+..+...+....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*....+......+..+...+....+..+.+.........+...+......+..+.......+...+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.............+.......+...............+...+...........+...+.+............+........+.........+.............+..+...+....+.....+................+...+..+...+.......+..+..........+.....+...+.............+..+...+.+..............+.+......+...+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls pgbouncer-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pgbouncer-new-ca created
```

Now, Let's create a new `Issuer` using the `pgbouncer-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pb-new-issuer
  namespace: demo
spec:
  ca:
    secretName: pgbouncer-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/pb-new-issuer created
```

### Create PgBouncerOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `PgBouncerOpsRequest` CRO with the newly created issuer. Below is the YAML of the `PgBouncerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pb
  tls:
    issuerRef:
      name: pb-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/change-issuer.yaml
pgbounceropsrequest.ops.kubedb.com/change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ kubectl get pbops -n demo change-issuer
NAME            TYPE             STATUS       AGE
change-issuer   ReconfigureTLS   Successful   104s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded.

Now, Let's exec into a database node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl get secrets -n demo pb-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -subject
subject=CN = pgbouncer, O = kubedb-updated

$ kubectl get secrets -n demo pb-server-cert -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -subject
subject=CN = pgbouncer, O = kubedb-updated
```
Now you can check [here](https://certlogik.com/decoder/).

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.


## Remove TLS from the Database

Now, we are going to remove TLS from this database using a PgBouncerOpsRequest.

### Create PgBouncerOpsRequest

Below is the YAML of the `PgBouncerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pb
  tls:
    clientAuthMode: md5
    remove: true
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.
- `spec.tls.clientAuthMode` defines clientAuthentication mode after removing tls. Possible values are `md5` `scram`.
  

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/remove-tls.yaml
pgbounceropsrequest.ops.kubedb.com/remove-tls created
```

#### Verify TLS Removed Successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ kubectl get pbops -n demo remove-tls
NAME         TYPE             STATUS       AGE
remove-tls   ReconfigureTLS   Successful   104s
```

Now first verify if this works in config.

```bash
kubectl exec -it -n demo pb-0 -- /bin/sh
/ $ cat etc/config/pgbouncer.ini
[databases]
postgres= host=ha-postgres.demo.svc port=5432 dbname=postgres

[pgbouncer]
max_db_connections = 1
max_user_connections = 2
pidfile = /tmp/pgbouncer.pid
logfile = /tmp/pgbouncer.log
listen_port = 5432
pool_mode = session
max_client_conn = 87
min_pool_size = 1
default_pool_size = 2
reserve_pool_size = 5
admin_users =  pgbouncer
listen_addr = *
auth_file =  /var/run/pgbouncer/secret/userlist
reserve_pool_timeout = 5
stats_period = 60
auth_type = md5
ignore_startup_parameters = extra_float_digits
```

SSL is off now.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pgbouncer -n demo pb
kubectl delete issuer -n demo pb-issuer pb-new-issuer
kubectl delete pgbounceropsrequest add-tls remove-tls rotate-tls change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/pgbouncer/private-registry/using-private-registry.md) to deploy PgBouncer with KubeDB.
- Use [kubedb cli](/docs/guides/pgbouncer/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
