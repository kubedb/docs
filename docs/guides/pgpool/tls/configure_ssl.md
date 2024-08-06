---
title: Pgpool TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: pp-tls-configure
    name: Pgpool_SSL
    parent: pp-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Pgpool with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption (via, `sslMode` and `clientAuthMode`) for Pgpool. This tutorial will show you how to use KubeDB to run a Pgpool database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Mongodb.

- `spec:`
  - `sslMode`
  - `tls:`
    - `issuerRef`
    - `certificate`
  - `clientAuthMode`

Read about the fields in details in [pgpool concept](/docs/guides/pgpool/concepts/pgpool.md),

`sslMode` enables TLS/SSL or mixed TLS/SSL used for all network connections. The value of `sslMode` field can be one of the following:

|     Value     | Description                                                                                                                                                                   |
|:-------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|  `disabled`   | The server does not use TLS/SSL.                                                                                                                                              |
|   `require`   | The server uses and accepts only TLS/SSL encrypted connections.                                                                                                               |
|  `verify-ca`  | The server uses and accepts only TLS/SSL encrypted connections and client want to be sure that client connect to a server that client trust.                                  |
| `verify-full` | The server uses and accepts only TLS/SSL encrypted connections and client want to be sure that client connect to a server client trust, and that it's the one client specify. |

The specified ssl mode will be used by health checker and exporter of Pgpool.

The value of `clientAuthMode` field can be one of the following:

|     Value     | Description                                                                                                                                                                   |
|:-------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|    `scram`    | The server uses scram-sha-256 authentication method to authenticate the users.                                                                                                |
|     `md5`     | The server uses md5 authentication method to authenticate the users.                                                                                                          |
|    `cert`     | The server uses tls certificates to authenticate the users and for this `sslMode` must not be disabled                                                                        |

The  `pool_hba.conf` of Pgpool will have the configuration based on the specified clientAuthMode.

When, SSLMode is anything other than `disabled`, users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.pem`, `tls.crt` and `tls.key`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Pgpool. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=pgpool/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls pgpool-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pgpool-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: pgpool-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/tls/issuer.yaml
issuer.cert-manager.io/pgpool-ca-issuer created
```

## Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

## TLS/SSL encryption in Pgpool

Below is the YAML for Pgpool with TLS enabled:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-tls
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  sslMode: require
  clientAuthMode: cert
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pgpool-ca-issuer
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
  syncUsers: true
  deletionPolicy: WipeOut
```

### Deploy Pgpool

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/tls/pgpool-ssl.yaml
pgpool.kubedb.com/pp-tls created
```

Now, wait until `pp-tls created` has status `Ready`. i.e,

```bash
$ watch kubectl get pp -n demo
Every 2.0s: kubectl get pgpool -n demo
NAME     TYPE                  VERSION   STATUS   AGE
pp-tls   kubedb.com/v1alpha2   4.5.0     Ready    60s
```

### Verify TLS/SSL in Pgpool

Now, connect to this database through [psql](https://www.postgresql.org/docs/current/app-psql.html) and verify if `SSLMode` has been set up as intended (i.e, `require`).

```bash
$ kubectl describe secret -n demo pp-tls-client-cert
Name:         pp-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=connection-pooler
              app.kubernetes.io/instance=pp-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=pgpools.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: pp-tls-client-cert
              cert-manager.io/common-name: postgres
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: pgpool-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
tls.key:  1675 bytes
ca.crt:   1151 bytes
tls.crt:  1131 bytes
```

Now, Lets save the client cert and key to two different files:

```bash
$ kubectl get secrets -n demo pp-tls-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > client.crt
$ cat client.crt
-----BEGIN CERTIFICATE-----
MIIDFjCCAf6gAwIBAgIRAO9tAQn/9lqHN4Pfi+UCe2IwDQYJKoZIhvcNAQELBQAw
IjEPMA0GA1UEAwwGcGdwb29sMQ8wDQYDVQQKDAZrdWJlZGIwHhcNMjQwNzE2MTAz
NzEzWhcNMjQxMDE0MTAzNzEzWjATMREwDwYDVQQDEwhwb3N0Z3JlczCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBAMo7Wikoc8XMAYorLO3lRJbcebiO+8ij
cI96UZ0SMzf7edyBbO1vVrlrBean9toCcW8Wf43o+Q+jRnvZJzzVnfV2C7gLeabC
o1I/g0JUmHdTxnIOLl6C4pvyoYZt4qB/cDorj89u6NIWnzs/fjFhYwCePQfPo7vM
eIeb8Mjngf77Cj5XENFpxR7+2Uy7SrpkeBoxuoPxrnStcwOIE06MbEHsLzGS59Kg
8qVLSyZAdguo8hBVV96gaRzW41uuf6MtBrrvwbv5IwtY9iHhlS9I3Uk+abnEnO6X
5aPUIQMcZixV2NXyEWZpbkoalvsywNcwNO/ESfN1DNA1oXknTknKeLkCAwEAAaNW
MFQwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMCMAwGA1UdEwEB
/wQCMAAwHwYDVR0jBBgwFoAUXmZGbZwUM7ojagZqJW24+Z+c2PMwDQYJKoZIhvcN
AQELBQADggEBAG0+bEOLTAXeYUIeyA5+aqCs4WTxT4KJo0jY8vWjNH2que4fydDI
860ZP/RWogO+dpVKWYmvmYVkCcNalYs5HnGJDhs+4IyiKcARcKWimuYXy3FtrLWu
UzKACXptOpnLKzD9v1vv1HfyfeB1hXyaEJLUuBPDGb05YSNFehcipbGFBHSWFv13
rMQCOKGt8R0JJUXR0fcuDEGKv+jpz5P+n5dBtPQ40CrE34mhpa3m00Y64X4PVDI6
RusaLKyNGkaU+15WErg44/zM3LayvMImRnnoIttO7NkOe/9ige8C3hgEjZoivZKM
0Jc7koXlrnszBH2K/MOst9kHRTPk0VVmxBo=
-----END CERTIFICATE-----
$ kubectl get secrets -n demo pp-tls-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > client.key
$ cat client.key
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAyjtaKShzxcwBiiss7eVEltx5uI77yKNwj3pRnRIzN/t53IFs
7W9WuWsF5qf22gJxbxZ/jej5D6NGe9knPNWd9XYLuAt5psKjUj+DQlSYd1PGcg4u
XoLim/Khhm3ioH9wOiuPz27o0hafOz9+MWFjAJ49B8+ju8x4h5vwyOeB/vsKPlcQ
0WnFHv7ZTLtKumR4GjG6g/GudK1zA4gTToxsQewvMZLn0qDypUtLJkB2C6jyEFVX
3qBpHNbjW65/oy0Guu/Bu/kjC1j2IeGVL0jdST5pucSc7pflo9QhAxxmLFXY1fIR
ZmluShqW+zLA1zA078RJ83UM0DWheSdOScp4uQIDAQABAoIBAQC7PPSfQsreCaIr
UQpKQImevAcer5PDEj/3N6M0sFMns/gCBvrZYsqC5eoSwtS0yKpJ1iTHOTrQFbX+
mPHRS17ykxcKkeVSVsdsMU3QLg7z/Gax1xtrefdht+WBV2AKhbNcyFRgFCoPyc4n
xwOJqMdHHTsYblEEYa3+sIzhFiev8Qv/eORQGcA/RFWier4sIZMGlhz/2aRmWXqr
9g+3g1VdMqor/0QX/MzAmuffoeGK/lSDZZLZ0fjZHxTXT56hMknzHkOoz3XK/BXj
AnOAosdg/MXCZEiP13aLMYcZ13Lx9txdcgBIU2tZmssNiN9XiGq+tcaYwnbqF7YS
mwyyS4NZAoGBANO3qMoRP3OOR5butEy03K0nLv8oMEgFy4nkFUBsCnr9+ABgZrUn
c/i8pt1kuwTElkV2MMNgCrNIZK6EAWoeuVqzpEwkI/F5FDvyOcXq76okpd21IoqE
PWtn7CR0VQqRfn6XWhU3lle8VrujE+w4OJqWE7a9x1AfxqLqqY2aAgfLAoGBAPSH
ynOi8+aM7LxCTGrROXVE8E4YWaMegLNkYj6WygQYxsww7eQ3Jk6cm+vVgwqIdZyW
hcsDzNutlMBuLreQwCP+asciHBhw8F58Z+6TZO/cBShTq+gXCWMG2zXP5pmxqrKD
UcyQ5F7WcNl+sv9zUYrPry+Jt2zdDzbPsfpyigkLAoGAIaH8a1VAGjBRCRYUiFb2
837VBW2x9c8N3XLhOWGwbIdp3U1zI3YIA0ycyXDWENTV9mTnLDJWoNJwRBTuUJhe
45zEDeBz4UlVwIwjR2CiAApgWw8KVKzbQPO6XLQqSkqAqMWMZvB0rq1ZrecjJBRu
UYhjy1Tsk7roiDr1Amyjw+8CgYABt2JIZYBowdx3hc+bgFRy6kT1h145suEcYTv/
THemh7X9gOpqi6iNLLQ7d4gv7r1EmBngTuqFMDa3Ew7o4u82UXbWZvrjgQdu4lio
aAhxVo4CtnOicWbzdvza59aqhYC5OAq+8NVphP/NxwHioSCVZNfJ8aGD9hlBPTv2
kg89+QKBgAQDWZkq2mPZMmb+ltW1TZO2HqmEXBP9plgYGfrSjpofTjsBzykoaHnA
J/ocHs2cNkW8arrhiZQzDyokZRc1j5+PIYLfXZ1gSK7WfOe6HO/667eCNuoEcfDv
w8MtuCJgbYP8J0BXun982+EnLkuyDAoyX9GvEqyGQagme1ENiwFm
-----END RSA PRIVATE KEY-----
```

Now, if you see the common name of the client.crt you can see,
```bash
$ openssl x509 -in client.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=postgres
```
Here common name of the client certificate is important if you want to connect with the client certificate, the `username must match the common name of the certificate`. Here, we can see the common name(CN) is, `postgres`. So, we will use postgres user to connect with Pgpool.

Now, we can connect using `subject=CN=postgres` to connect to the psql,

```bash
$ psql "sslmode=require port=9999 host=localhost dbname=postgres user=postgres sslrootcert=ca.crt sslcert=client.crt sslkey=client.key"
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, compression: off)
Type "help" for help.

postgres=# 
```

We are connected to the postgres database. Let's run some command to verify the sslMode and the user,

```bash
postgres=# SELECT
    usename,
    ssl
FROM
    pg_stat_ssl
JOIN
    pg_stat_activity
ON
    pg_stat_ssl.pid = pg_stat_activity.pid;
 usename  | ssl 
----------+-----
 postgres | t
 postgres | t
 postgres | t
(3 rows)

postgres=# \q
⏎  
```

You can see here that, `postgres` user with `ssl` status as t or true.

## Changing the SSLMode & ClusterAuthMode

User can update `sslMode` & `clientAuthMode` if needed. Some changes may be invalid from pgpool end, like using `sslMode: disabled` with `clientAuthMode: cert`.

The good thing is, **KubeDB operator will throw error for invalid SSL specs while creating/updating the Pgpool object.** i.e.,

```bash
$ kubectl patch -n demo pp/pp-tls -p '{"spec":{"sslMode": "disabled","clientAuthMode": "cert"}}' --type="merge"
The Pgpool "pp-tls" is invalid: spec.sslMode: Unsupported value: "disabled": supported values: "disable", "allow", "prefer", "require", "verify-ca", "verify-full"
```

> Note: There is no official support for Pgpool with the Postgres cluster having `clientAuthMode` as `cert`. Check [here](https://www.pgpool.net/docs/42/en/html/auth-methods.html#:~:text=Note%3A%20The%20certificate%20authentication%20works%20between%20only%20client%20and%20Pgpool%2DII.%20The%20certificate%20authentication%20does%20not%20work%20between%20Pgpool%2DII%20and%20PostgreSQL.%20For%20backend%20authentication%20you%20can%20use%20any%20other%20authentication%20method.).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pgpool -n demo pp-tls
kubectl delete issuer -n demo pp-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Detail concepts of [PgpoolVersion object](/docs/guides/pgpool/concepts/catalog.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
