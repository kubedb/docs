---
title: PgBouncer TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: pb-tls-configure
    name: PgBouncer_SSL
    parent: pb-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run PgBouncer with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption (via, `sslMode` and `connectionPool.authType`) for PgBouncer. This tutorial will show you how to use KubeDB to run a PgBouncer database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in PgBouncer.

- `spec:`
  - `sslMode`
  - `tls:`
    - `issuerRef`
    - `certificate`
  - `connectionPool`
    - `authType`

Read about the fields in details in [pgbouncer concept](/docs/guides/pgbouncer/concepts/pgbouncer.md),

`sslMode` enables TLS/SSL or mixed TLS/SSL used for all network connections. The value of `sslMode` field can be one of the following:

|     Value     | Description                                                                                                                                 |
|:-------------:|:--------------------------------------------------------------------------------------------------------------------------------------------|
|  `disabled`   | The server does not use TLS/SSL.                                                                                                            |
|    `allow`    | If client requests TLS, it is used. If not, plain TCP is used. If the client presents a client certificate, it is not validated.            |
|    `prefer`   | Same as allow.                                                                                                                              |
|   `require`   | Client must use TLS. If not, the client connection is rejected. If the client presents a client certificate, it is not validated.           |
|  `verify-ca`  | Client must use TLS with valid client certificate.                                                                                          |
| `verify-full` | Same as verify-ca.                                                                                                                          | 

The specified ssl mode will be used by health checker and exporter of PgBouncer.

The value of `connectionPool.authType` field can be one of the following:

|      Value      | Description                                                                                                                                                                   |
|:---------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `scram-sha-256` | The server uses scram-sha-256 authentication method to authenticate the users.                                                                                                |
|      `md5`      | The server uses md5 authentication method to authenticate the users.                                                                                                          |

The  `userlist.txt` of PgBouncer will have the configuration based on the specified AuthType.

When, SSLMode is anything other than `disabled`, users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.pem`, `tls.crt` and `tls.key`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in PgBouncer. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=pgbouncer/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls pgbouncer-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pgbouncer-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: pgbouncer-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/tls/issuer.yaml
issuer.cert-manager.io/pgbouncer-ca-issuer created
```

## Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

## TLS/SSL encryption in PgBouncer

Below is the YAML for PgBouncer with TLS enabled:

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-tls
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "pg"
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
  sslMode: verify-ca
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pb-ca-issuer
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
```

### Deploy PgBouncer

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/tls/pgbouncer-ssl.yaml
pgbouncer.kubedb.com/pb-tls created
```

Now, wait until `pb-tls created` has status `Ready`. i.e,

```bash
$ watch kubectl get pb -n demo
Every 2.0s: kubectl get pgbouncer -n demo
NAME     VERSION   STATUS   AGE
pb-tls   1.18.0    Ready    108s
```

### Verify TLS/SSL in PgBouncer

Now, connect to this database through [psql](https://www.postgresql.org/docs/current/app-psql.html) and verify if `SSLMode` has been set up as intended (i.e, `require`).

```bash
$ kubectl describe secret -n demo pb-tls-client-cert
Name:         pb-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=connection-pooler
              app.kubernetes.io/instance=pb-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=pgbouncers.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: pb-tls-client-cert
              cert-manager.io/common-name: pgbouncer
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: pb-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1159 bytes
tls.crt:  1135 bytes
tls.key:  1679 bytes
```

Now, Lets save the client cert and key to two different files:

```bash
$ kubectl get secrets -n demo pb-tls-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > client.crt
$ cat client.crt
-----BEGIN CERTIFICATE-----
MIIDGTCCAgGgAwIBAgIQFzXjq6IExD5sjF7FW44NzTANBgkqhkiG9w0BAQsFADAl
MRIwEAYDVQQDDAlwZ2JvdW5jZXIxDzANBgNVBAoMBmt1YmVkYjAeFw0yNTAxMjMx
MDQ2MDBaFw0yNTA0MjMxMDQ2MDBaMBQxEjAQBgNVBAMTCXBnYm91bmNlcjCCASIw
DQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANtr22zMM8A0k7tsPvXICpNWUAfW
1xqDrEv5dsHP04Pd8YwioCP6lrDSahV8jkFhI4jrLCy4RYYhC8nzf3QLTkYIPTEd
PfYaS9jTfNPgGHMD8hSKFfO+gXSidg+PzUW2x8/hA8SFq9rJwn3/b39DVL71E4aU
D8aJYPc51LsIr2JoiGb0qPNSPpud/4bma1GcqCgsChkMLzsn88vOg0B9a74RUSKd
W78I37N2xNUwS5M7mgNmpzKVIhBfs0h01F6vfTVzOwOl/C9as1uQGDCIRBx6ONyl
7r1SJCENuEEr4Q33iTmBLRBwy5HKGy+UHc58DZ1lLwBaJsQdujUcbEoRrbMCAwEA
AaNWMFQwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMCMAwGA1Ud
EwEB/wQCMAAwHwYDVR0jBBgwFoAUfEEIxgLcuBXzCYzm48qnmkbxZvIwDQYJKoZI
hvcNAQELBQADggEBAAkFInhE2W8bVbuRM+PESMNDff3cfgH8bzi9A+iWDR0XmpBm
qLqq8zciebGmuqH8PLQr518U6dCI9g0iATfV/WQ6JlRFhxiO3h+7rAjwW77V49QM
06CkL2uSRk0GeO9a/VNXMmcNZGARgG+m7gYZJ/sOVnzlj5zEchfaH82FY5HnInRl
coSL5sY28QU1iS0bO3wHoFx6t8gzwluP/H040ImS60CE5t/b3njIgfWDHzhDOkKV
Rl66yC3j2YD8+Dvdl63Dp8r5KtWDvGAkiM8SVysASHnKAM/ipEqUoqyWBUT7gG/L
JbiZCRCTnewRU9/mzcn9FxxmAPt7yq9IEND1cMQ=
-----END CERTIFICATE-----
$ kubectl get secrets -n demo pb-tls-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > client.key
$ cat client.key
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA22vbbMwzwDSTu2w+9cgKk1ZQB9bXGoOsS/l2wc/Tg93xjCKg
I/qWsNJqFXyOQWEjiOssLLhFhiELyfN/dAtORgg9MR099hpL2NN80+AYcwPyFIoV
876BdKJ2D4/NRbbHz+EDxIWr2snCff9vf0NUvvUThpQPxolg9znUuwivYmiIZvSo
81I+m53/huZrUZyoKCwKGQwvOyfzy86DQH1rvhFRIp1bvwjfs3bE1TBLkzuaA2an
MpUiEF+zSHTUXq99NXM7A6X8L1qzW5AYMIhEHHo43KXuvVIkIQ24QSvhDfeJOYEt
EHDLkcobL5QdznwNnWUvAFomxB26NRxsShGtswIDAQABAoIBAQCXOLNmPSnhapry
TbzqiS54ssC/VlqzJFJnngsxsbjVpe2mJer2QOr//FQucMRd3MOvxlyQiYMo2LeW
PGH3qR8N9vmtUrj0VtU1HzRllYlkIzEA5NYSQZZYuurg+LuBM2JsK2j8VR/Gzsxj
J9tA+zd5z8/gLUTeEKoqWMn7CRZOm/OhorKM2PnduniazZjF9w9PZwSUlIjJnBNi
rx21RvVUw7UGCw/5jVvsDENSAkt/RHAQySu3Zzbk+gbpyhq2VIa/SADhKO9BgjxG
EQxWNQbi8anmVtSneGngfeY/OOnlyahsdzuQ9l53Iz/o511897TePDgvz6mmGxhS
4ht1QWk5AoGBAPThrun/G26f/GUxQN5QTywj926UBqpmfiCHaIoP/UM2M7Kck8Um
bgmZu2M9FvSErmmvi/KYHUJlY9yGRHX+8TqP8RVxHi0MxaPjCq0Jv/IRcSM15qVl
IoIbGPtAQrXNe1crLTeJboQ5mY20ekzkj1q9KYWWC/0Jc/Xj5kcIdO1vAoGBAOVi
PkTanAN7lMfBVEDk2dpcZTzf17WM//LGsZ1G/KTFjHq7hNMHakhla2CbPEkCOt9l
HgUOKqROsFf8lyWNIUnllEhyHfoBFRweplub2Zh3Y/JkQONA3MohKbkO28ZvDJDg
5AZB/eaTB36URqEr6hHdI037MwACZxOSKxjRp+n9AoGAZceTcrBUT4NxXQG+q2gH
sBn20l/18UcOLyj4m0GQCypxDFCl3nBdleHuj42ph9HJyCVtblQo/Rq1CchIlh5z
VtrS4g2U9DZ1wusv2cHOpKb5NiBGEAJb+GWY2XzY/UU9eXp5nbaiV5S1LL+RgXoR
1y3+HwbBTtdp+g5R/L4YE0MCgYBqxBeHpNkJJfRSJcI5kkt0P50/gFC+yCo5rhHt
yqS9bNW+KpngP4tQtyQLizW8JbWRVVdrsvRWFeouifswF0hvRNSIA9XAD9DrjbiQ
2zGkra1vnQo2vHIIAveQk0HoUrfel06LOxwavkS2vf1B91azieJs4YcTcgrYKSi2
HJ+zYQKBgQCKfewbwVLuexdW6yLrxwXuMAZljtHUQWe7Txx3k+bw+kAF46NEBlN2
bZc0zaz8cEn8d7GWVGGGulZA7XxZM+Tr3uD1t/8AkiS/GwRKcXBOjzQZS08bnTVJ
BwIhO4g2OiLojS6dQxrXtj/miB3pTZbVed7QhYOBUGEFs3lUV+KEVQ==
```

Now, if you see the common name of the client.crt you can see,
```bash
$ openssl x509 -in client.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=pgbouncer
```
Here common name of the client certificate is important if you want to connect with the client certificate, the `username must match the common name of the certificate`. Here, we can see the common name(CN) is, `pgbouncer`. So, we will use pgbouncer user to connect with PgBouncer.

Now, we can connect using `subject=CN=pgbouncer` to connect to the psql,

```bash
$ psql "sslmode=require port=9999 host=localhost dbname=pgbouncer user=pgbouncer sslrootcert=ca.crt sslcert=client.crt sslkey=client.key"
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, compression: off)
Type "help" for help.

pgbouncer=# 
```

We are connected to the pgbouncer database. Let's run some command to verify the sslMode and the user,

```bash
pgbouncer=# SHOW SERVERS;
type |   user   | database | state |     addr      | port | local_addr | local_port |      connect_time       |      request_time       | wait | wait_us | close_needed |      ptr       | link | remote_pid | tls | application_name 
------+----------+----------+-------+---------------+------+------------+------------+-------------------------+-------------------------+------+---------+--------------+----------------+------+------------+-----+------------------
 S    | postgres | postgres | idle  | 10.96.125.227 | 5432 | 10.244.0.6 |      57524 | 2025-01-24 05:17:47 UTC | 2025-01-24 06:06:07 UTC |    0 |       0 |            0 | 0x70872bea80a0 |      |        476 |     | 
(1 row)
~
~
(END)
pgbouncer=# exit
⏎  
```

## Changing the SSLMode & ClusterAuthMode

User can update `sslMode` & `connectionPool.authType` if needed. Some changes may be invalid from pgbouncer end, like using `sslMode: disabled` with `connectionPool.authType: cert`.

The good thing is, **KubeDB operator will throw error for invalid SSL specs while creating/updating the PgBouncer object.** i.e.,

```bash
$ kubectl patch -n demo pb/pb-tls -p '{"spec":{"sslMode": "disabled"}}' --type="merge"
The PgBouncer "pb-tls" is invalid: spec.sslMode: Unsupported value: "disabled": supported values: "disable", "allow", "prefer", "require", "verify-ca", "verify-full"
```

> Note: There is no official support from kubedb for PgBouncer to connect wit cert mode`.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pgbouncer -n demo pb-tls
kubectl delete issuer -n demo pb-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Detail concepts of [PgBouncerVersion object](/docs/guides/pgbouncer/concepts/catalog.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
