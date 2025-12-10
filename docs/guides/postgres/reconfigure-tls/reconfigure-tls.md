---
title: Reconfigure Postgres TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: pg-reconfigure-tls-cluster
    name: Reconfigure Postgres TLS/SSL Encryption
    parent: pg-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Postgres TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates, changing issuer for existing Postgres database via a PostgresOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `cert-manger` v1.0.0 or later to your cluster to manage your SSL/TLS certificates from [here](https://cert-manager.io/docs/installation/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Postgres database

Here, We are going to create a Postgres database without TLS and then reconfigure the database to use TLS.

### Deploy Postgres without TLS

In this section, we are going to deploy a Postgres Replicaset database without TLS. In the next few sections we will reconfigure TLS using `PostgresOpsRequest` CRD. Below is the YAML of the `Postgres` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 3
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  version: "13.13"
```

Let's create the `Postgres` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/ha-postgres.yaml
postgres.kubedb.com/ha-postgres created
```

Now, wait until `ha-postgres` has status `Ready`. i.e,

```bash
$ kubectl get pg -n demo
NAME          VERSION   STATUS   AGE
ha-postgres   13.13     Ready    87s

$ kubectl dba describe postgres ha-postgres -n demo
Name:               ha-postgres
Namespace:          demo
CreationTimestamp:  Mon, 19 Aug 2024 13:38:28 +0600
Labels:             <none>
Replicas:           3  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  WipeOut

Service:        
  Name:         ha-postgres
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=ha-postgres
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=postgreses.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.193.243
  Port:         primary  5432/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.76:5432
  Port:         coordinatclient  2379/TCP
  TargetPort:   coordinatclient/TCP
  Endpoints:    10.244.0.76:2379

Service:        
  Name:         ha-postgres-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=ha-postgres
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=postgreses.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  5432/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.76:5432,10.244.0.78:5432,10.244.0.80:5432
  Port:         coordinator  2380/TCP
  TargetPort:   coordinator/TCP
  Endpoints:    10.244.0.76:2380,10.244.0.78:2380,10.244.0.80:2380
  Port:         coordinatclient  2379/TCP
  TargetPort:   coordinatclient/TCP
  Endpoints:    10.244.0.76:2379,10.244.0.78:2379,10.244.0.80:2379

Service:        
  Name:         ha-postgres-standby
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=ha-postgres
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=postgreses.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.65.45
  Port:         standby  5432/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.78:5432,10.244.0.80:5432

Auth Secret:
  Name:         ha-postgres-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=ha-postgres
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=postgreses.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  8 bytes

Topology:
  Type     Pod            StartTime                      Phase
  ----     ---            ---------                      -----
  primary  ha-postgres-0  2024-08-19 13:38:34 +0600 +06  Running
           ha-postgres-1  2024-08-19 13:38:41 +0600 +06  Running
           ha-postgres-2  2024-08-19 13:38:48 +0600 +06  Running

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"Postgres","metadata":{"annotations":{},"name":"ha-postgres","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","version":"13.13"}}

    Creation Timestamp:  2024-08-19T07:38:31Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    ha-postgres
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        postgreses.kubedb.com
    Name:                            ha-postgres
    Namespace:                       demo
  Spec:
    App Ref:
      API Group:  kubedb.com
      Kind:       Postgres
      Name:       ha-postgres
      Namespace:  demo
    Client Config:
      Service:
        Name:    ha-postgres
        Path:    /
        Port:    5432
        Query:   sslmode=disable
        Scheme:  postgresql
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  postgres-backup-13.1
          Restore Task:
            Name:  postgres-restore-13.1
    Secret:
      Name:   ha-postgres-auth
    Type:     kubedb.com/postgres
    Version:  13.13

Events:
  Type    Reason      Age   From             Message
  ----    ------      ----  ----             -------
  Normal  Successful  2m    KubeDB Operator  Successfully created governing service
  Normal  Successful  2m    KubeDB Operator  Successfully created Service
  Normal  Successful  2m    KubeDB Operator  Successfully created Service
  Normal  Successful  2m    KubeDB Operator  Successfully created Postgres
  Normal  Successful  49s   KubeDB Operator  Successfully patched Postgres
```

Now, we can connect to this database through `psql` and verify that the TLS is disabled.


```bash
$ kubectl get secrets -n demo ha-postgres-auth -o jsonpath='{.data.\username}' | base64 -d
postgres

$ kubectl get secrets -n demo ha-postgres-auth -o jsonpath='{.data.\password}' | base64 -d
U6(h_pYrekLZ2OOd

$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql -h ha-postgres.demo.svc -U postgres
Password for user postgres: 
psql (13.13)
Type "help" for help.

postgres=# 
postgres=# SELECT name, setting  FROM pg_settings  WHERE name IN ('ssl');
 name | setting 
------+---------
 ssl  | off
(1 row)


```

We can verify from the above output that TLS is disabled for this database.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Postgres. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls postgres-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/postgres-ca created
```

Now, Let's create an `Issuer` using the `postgres-ca` secret that we have just created. The `YAML` file looks like this:

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

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/pg-issuer created
```

```bash
$ kubectl get issuer -n demo
NAME        READY   AGE
pg-issuer   True    11s
```
Issuer is ready(true).

### Create PostgresOpsRequest

In order to add TLS to the database, we have to create a `PostgresOpsRequest` CRO with our created issuer. Below is the YAML of the `PostgresOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ha-postgres
  tls:
    sslMode: verify-full
    clientAuthMode: cert
    issuerRef:
      name: pg-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - postgres
          organizationalUnits:
            - client
  apply: Always
```
Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ha-postgres` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates.

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/add-tls.yaml
postgresopsrequest.ops.kubedb.com/add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `PostgresOpsRequest` to be `Successful`.  Run the following command to watch `PostgresOpsRequest` CRO,

```bash
$ kubectl get pgops -n demo add-tls 
NAME      TYPE             STATUS       AGE
add-tls   ReconfigureTLS   Successful   5m23s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. 

Now, Let's exec into a database primary pods to see if certificates are added there.
```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ ls -R /tls
tls:
certs

tls/certs:
client    exporter  server

tls/certs/client:
ca.crt      client.crt  client.key

tls/certs/exporter:
ca.crt   tls.crt  tls.key

tls/certs/server:
ca.crt      server.crt  server.key

```
All the certs are added. Now lets connect with the postgres using client certs
```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql -h ha-postgres.demo.svc -U postgres -d "sslmode=verify-full sslrootcert=/tls/certs/client/ca.crt sslcert=/tls/certs/client/client.crt sslkey=/tls/certs/client/client.key"
psql (13.13)
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

postgres=# 
```
We can see our connection is now `SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)`

Lets check whether ssl is on.
```bash
postgres=# SELECT name, setting 
postgres-# FROM pg_settings 
postgres-# WHERE name IN ('ssl', 'ssl_cert_file', 'ssl_key_file');
     name      |           setting            
---------------+------------------------------
 ssl           | on
 ssl_cert_file | /tls/certs/server/server.crt
 ssl_key_file  | /tls/certs/server/server.key
(3 rows)

```

> Note: We by default set local connection to trust. So you can connect to postgres without password or certificate from inside of the pods.
> ```bash
> $ kubectl exec -it -n demo ha-postgres-0 -- bash
> Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
> ha-postgres-0:/$ psql
> psql (13.13)
> Type "help" for help.
> postgres=#


## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
kubectl get secrets -n demo ha-postgres-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -dates
notBefore=Aug 21 05:25:05 2024 GMT
notAfter=Nov 19 05:25:05 2024 GMT
```

So, the certificate will expire on this time `Nov 19 05:25:05 2024 GMT`. 

### Create PostgresOpsRequest

Now we are going to increase it using a PostgresOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ha-postgres
  tls:
    rotateCertificates: true

```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ha-postgres` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/rotate-tls.yaml
postgresopsrequest.ops.kubedb.com/rotate-tls created
```

#### Verify Certificate Rotated Successfully

Let's wait for `PostgresOpsRequest` to be `Successful`.  Run the following command to watch `PostgresOpsRequest` CRO,

```bash
$ kubectl get pgops -n demo 
NAME         TYPE             STATUS       AGE
rotate-tls   ReconfigureTLS   Successful   3m10s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. And we can check that the tls.crt has been updated.
```bash
$  kubectl get secrets -n demo ha-postgres-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates

notBefore=Aug 21 05:40:49 2024 GMT
notAfter=Nov 19 05:40:49 2024 GMT

$ kubectl get secrets -n demo ha-postgres-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates

notBefore=Aug 21 05:40:49 2024 GMT
notAfter=Nov 19 05:40:49 2024 GMT
```


As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubedb-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls postgres-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/postgres-new-ca created
```

Now, Let's create a new `Issuer` using the `postgres-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pg-new-issuer
  namespace: demo
spec:
  ca:
    secretName: postgres-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/pg-new-issuer created
```

### Create PostgresOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `PostgresOpsRequest` CRO with the newly created issuer. Below is the YAML of the `PostgresOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ha-postgres
  tls:
    issuerRef:
      name: pg-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ha-postgres` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/change-issuer.yaml
postgresopsrequest.ops.kubedb.com/change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `PostgresOpsRequest` to be `Successful`.  Run the following command to watch `PostgresOpsRequest` CRO,

```bash
$ kubectl get pgops -n demo change-issuer
NAME            TYPE             STATUS       AGE
change-issuer   ReconfigureTLS   Successful   3m54s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded.

Now, Let's exec into a database node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl get secrets -n demo ha-postgres-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -subject

subject=CN = ca-updated, O = kubedb-updated

$ kubectl get secrets -n demo ha-postgres-server-cert -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -subject

subject=CN = ca-updated, O = kubedb-updated

# other way to check this is
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ cat /tls/certs/server/ca.crt 
-----BEGIN CERTIFICATE-----
MIIDPTCCAiWgAwIBAgIUGBW8oXbOFPLOZ6p7iAqVnx7tdsgwDQYJKoZIhvcNAQEL
BQAwLjETMBEGA1UEAwwKY2EtdXBkYXRlZDEXMBUGA1UECgwOa3ViZWRiLXVwZGF0
ZWQwHhcNMjQwODIxMDYwNTIxWhcNMjUwODIxMDYwNTIxWjAuMRMwEQYDVQQDDApj
YS11cGRhdGVkMRcwFQYDVQQKDA5rdWJlZGItdXBkYXRlZDCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAMafBxDa1r35z4yK/6bcJC22j9/JsV7EujxtN8pP
DbuLLUiAKhSZtqSjflS1EshPzVbesie/zgBY0BZRRZNTw7YEsGn/0fCLy/gtSSeD
c6tOilB7a31gH7EHUTm46tbiUcSduUXF+9KFbg54d34RVy/ozB7GULIPI5XqA/FE
E8FSRUZpYYnUaLBqqW+kJZCOS5K9wqT4mgicFWVc5kgcrkNouxwd1bdNdhaKURdL
oNsWpRT71LI+fwR4TV+Xzh2o4BR71YrW7ojbUu8+x1GIMcZmE43iGzhVELHMA+bw
KUwWOfSsDQ4eXvnOLvrXkTrdFOSxFlmKoKJfbo163dxoZPsCAwEAAaNTMFEwHQYD
VR0OBBYEFM/1iTxEvn2JTgfeHCpzZ+5/Oy4/MB8GA1UdIwQYMBaAFM/1iTxEvn2J
TgfeHCpzZ+5/Oy4/MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
AHZEfGBJU9ctE+tP70hWqUJF4WVDyeO9VFXxfVlDJq2/w5id7tq8G75vtkh3wOMx
StbJa8z8ys3LPuPiCOcVP3i30x4sKN6xMgdY7xAbCD65UJ53XXqqTfSlaz/RW9UN
Swb3YKUHZvlMKrienp8qMjGWQE0thk4zJzH/MbvE/RV5W7fWTCuUop6zRDcL14e4
sOhjQoxh3hMrHh1IDDsa5S+r1jyWSr6lkCkf5dAeIx/CVZgJUnnou68sVkNL5P3g
5sXwCzQQnRA+lw6nQFC3mbbNWP+klOqf27eFz6ve1VmPAKyMAGazQhKMqQS8gIzA
aLcixLL6zhgM40K56RE7b14=
-----END CERTIFICATE-----
```
Now you can check any certificate decoding website.

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.



## Remove TLS from the Database

Now, we are going to remove TLS from this database using a PostgresOpsRequest.

### Create PostgresOpsRequest

Below is the YAML of the `PostgresOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ha-postgres
  tls:
    clientAuthMode: md5
    remove: true
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ha-postgres` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.
- `spec.tls.clientAuthMode` defines clientAuthentication mode after removing tls. It can't be `cert`. Possible values are `md5` `scram`.
  

Let's create the `PostgresOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/remove-tls.yaml
postgresopsrequest.ops.kubedb.com/remove-tls created
```

#### Verify TLS Removed Successfully

Let's wait for `PostgresOpsRequest` to be `Successful`.  Run the following command to watch `PostgresOpsRequest` CRO,

```bash
$ kubectl get pgops -n demo remove-tls 
NAME         TYPE             STATUS       AGE
remove-tls   ReconfigureTLS   Successful   4m

```

Now first verify if we can connect without using certs.

```bash
$ kubectl get secrets -n demo ha-postgres-auth -o jsonpath='{.data.\username}' | base64 -d
postgres

$ kubectl get secrets -n demo ha-postgres-auth -o jsonpath='{.data.\password}' | base64 -d
U6(h_pYrekLZ2OOd
```

```bash
kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql -h ha-postgres.demo.svc -U postgres
Password for user postgres: 
psql (13.13)
Type "help" for help.

postgres=# SELECT name, setting 
postgres-# FROM pg_settings 
postgres-# WHERE name IN ('ssl', 'ssl_cert_file', 'ssl_key_file');
     name      |  setting   
---------------+------------
 ssl           | off
 ssl_cert_file | server.crt
 ssl_key_file  | server.key

```

SSL is off now.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete postgres -n demo ha-postgres
kubectl delete issuer -n demo pg-issuer pg-new-issuer
kubectl delete postgresopsrequest add-tls remove-tls rotate-tls change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Monitor your Postgres database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Monitor your Postgres database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy Postgres with KubeDB.
- Use [kubedb cli](/docs/guides/postgres/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
