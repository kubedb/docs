---
title: ZooKeeper TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: zk-tls-configure
    name: ZooKeeper_SSL
    parent: zk-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run ZooKeeper Ensemble with TLS/SSL 

KubeDB supports providing TLS/SSL encryption for ZooKeeper Ensemble. This tutorial will show you how to use KubeDB to run a ZooKeeper Ensemble with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/zookeeper](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/zookeeper) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in ZooKeeper.

- `spec:`
    - `enableSSL`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [zookeeper Concept Guide](/docs/guides/zookeeper/concepts/zookeeper.md),

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt`, `tls.key`, `keystore.jks` and `truststore.jks`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in ZooKeeper. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=zookeeper/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls zookeeper-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: zookeeper-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: zookeeper-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/tls/zookeeper-issuer.yaml
issuer.cert-manager.io/zookeeper-ca-issuer created
```

## TLS/SSL encryption in ZooKeeper Ensemble

Below is the YAML for ZooKeeper with TLS enabled:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-tls
  namespace: demo
spec:
  version: "3.8.3"
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: zookeeper-ca-issuer
  adminServerPort: 8080
  replicas: 5
  storage:
    resources:
      requests:
        storage: "1Gi"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"

```

Here,
- `spec.enableSSL` is set to `true` to enable TLS/SSL encryption.
- `spec.tls.issuerRef` refers to the `Issuer` that we have created in the previous step.
- 
### Deploy ZOoKeeper Ensemble with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/tls/zookeeper-tls.yaml
zookeeper.kubedb.com/zk-tls created
```

Now, wait until `zookeeper-tls created` has status `Ready`. i.e,

```bash
$ watch kubectl get zookeeper -n demo
NAME        TYPE                    VERSION   STATUS    AGE
zk-tls      kubedb.com/v1alpha2     3.8.3     Ready     60s
```

### Verify TLS/SSL in ZooKeeper Ensemble

```bash
$ kubectl describe secret -n demo zk-quickstart-client-cert 
Name:         zk-quickstart-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=zk-quickstart
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=zookeepers.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.zk-quickstart-pods.demo.svc.cluster.local,localhost,zk-quickstart,zk-quickstart-pods,zk-quickstart-pods.demo.svc,zk-quickstart-pods.demo...
              cert-manager.io/certificate-name: zk-quickstart-client-cert
              cert-manager.io/common-name: zk-quickstart-pods.demo.svc
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: zookeeper-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:            1159 bytes
keystore.jks:      3258 bytes
tls-combined.pem:  3198 bytes
tls.crt:           1493 bytes
tls.key:           1704 bytes
truststore.jks:    873 bytes
```

Now, Let's exec into a ZooKeeper pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- bash
Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ cd ../var/private/ssl
zookeeper@zk-quickstart-0:/var/private/ssl$ openssl s_client -connect localhost:2182 -CAfile ca.crt -cert tls.crt -key tls.key
CONNECTED(00000003)
depth=1 CN = zookeeper, O = kubedb
verify return:1
depth=0 CN = zk-quickstart.demo.svc
verify return:1
---
Certificate chain
 0 s:CN = zk-quickstart.demo.svc
   i:CN = zookeeper, O = kubedb
   a:PKEY: rsaEncryption, 2048 (bit); sigalg: RSA-SHA256
   v:NotBefore: Nov  4 05:46:21 2024 GMT; NotAfter: Feb  2 05:46:21 2025 GMT
---
Server certificate
-----BEGIN CERTIFICATE-----
MIIEJTCCAw2gAwIBAgIQaWLGhg/TgVF8oXGcsLQkKjANBgkqhkiG9w0BAQsFADAl
MRIwEAYDVQQDDAl6b29rZWVwZXIxDzANBgNVBAoMBmt1YmVkYjAeFw0yNDExMDQw
NTQ2MjFaFw0yNTAyMDIwNTQ2MjFaMCExHzAdBgNVBAMTFnprLXF1aWNrc3RhcnQu
ZGVtby5zdmMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCeeiLZeNa7
wHOUwD76fmp45Ae9qlpHCW/lGz+lGO48FBDUBbG2Tm2BZVW2297HOzb/Lax6Molb
9qCDsV7ITCUYXLBGz0pCGqGYS/icZupShhKAvD33Gn8kH/QeANwFonpxBAtr36vi
WxwcRD+dfVAu7OCATwSakZh3zdbRPQXLiAVqj8qn4zNSYL5bzUXQ5dHFzvgwZve5
FR3QYLvVjUEu2tFjCKM+/HTzQ/IMUAjcU0lU4qnWqnhgcGp8ZE3hDyL9OOOsjrWx
CGNhB0Orf6Efztkqq4FMZ//w3DUQgnRglGKl1rGK015//W0MGSPlT4uve6Z7zaRU
aUqa7Y8P5wZxAgMBAAGjggFTMIIBTzAOBgNVHQ8BAf8EBAMCAqQwHQYDVR0lBBYw
FAYIKwYBBQUHAwEGCCsGAQUFBwMCMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYE
FC7Wrn4SOKhsT4TQFEMtSao72H5TMB8GA1UdIwQYMBaAFDe7/VhWOllB39U/xOht
MxmZu9wQMIHMBgNVHREEgcQwgcGCKyouemstcXVpY2tzdGFydC1wb2RzLmRlbW8u
c3ZjLmNsdXN0ZXIubG9jYWyCCWxvY2FsaG9zdIINemstcXVpY2tzdGFydIISemst
cXVpY2tzdGFydC1wb2Rzght6ay1xdWlja3N0YXJ0LXBvZHMuZGVtby5zdmOCKXpr
LXF1aWNrc3RhcnQtcG9kcy5kZW1vLnN2Yy5jbHVzdGVyLmxvY2FsghZ6ay1xdWlj
a3N0YXJ0LmRlbW8uc3ZjhwR/AAABMA0GCSqGSIb3DQEBCwUAA4IBAQCGGxgGzdjF
Vo9VALc6ddZD50M7bfh5L5z2KfSY4ZH7kuokM52LGzJYwREV3UpVAhjBqn0XEf9p
JX8ePo0Z9zjtWIIZg4ctjlCvKDy+HpKlqh2RJejnl+NoLPV628QJDiEksLzdVl4v
z36AwdGeUhADpvoGQiXUT6LgrD++Uv0akpDEzWOB2LUKsvCRKnxyBNyBqpsW8/Pu
DeC/RUGXT/JFtZtDBGp8d/FOIpJ0t/ZjrI9Hyu5DLFB08oTYmEVE3Lv2owZZV/o8
6YqlpTu2efKEzMFZudUWpnGUrb69sZeDR9hwxGcAdKobTB8SZOBU61nsRn95BH7O
S4dKhcrbzP70
-----END CERTIFICATE-----
subject=CN = zk-quickstart.demo.svc
issuer=CN = zookeeper, O = kubedb
---
Acceptable client certificate CA names
CN = zookeeper, O = kubedb
Client Certificate Types: ECDSA sign, RSA sign, DSA sign
Requested Signature Algorithms: ECDSA+SHA256:ECDSA+SHA384:ECDSA+SHA512:RSA-PSS+SHA256:RSA-PSS+SHA384:RSA-PSS+SHA512:RSA-PSS+SHA256:RSA-PSS+SHA384:RSA-PSS+SHA512:RSA+SHA256:RSA+SHA384:RSA+SHA512:DSA+SHA256:ECDSA+SHA224:RSA+SHA224:DSA+SHA224:ECDSA+SHA1:RSA+SHA1:DSA+SHA1
Shared Requested Signature Algorithms: ECDSA+SHA256:ECDSA+SHA384:ECDSA+SHA512:RSA-PSS+SHA256:RSA-PSS+SHA384:RSA-PSS+SHA512:RSA-PSS+SHA256:RSA-PSS+SHA384:RSA-PSS+SHA512:RSA+SHA256:RSA+SHA384:RSA+SHA512:DSA+SHA256:ECDSA+SHA224:RSA+SHA224:DSA+SHA224
Peer signing digest: SHA256
Peer signature type: RSA-PSS
Server Temp Key: X25519, 253 bits
---
SSL handshake has read 1611 bytes and written 2553 bytes
Verification: OK
---
New, TLSv1.2, Cipher is ECDHE-RSA-AES128-GCM-SHA256
Server public key is 2048 bit
Secure Renegotiation IS supported
Compression: NONE
Expansion: NONE
No ALPN negotiated
SSL-Session:
    Protocol  : TLSv1.2
    Cipher    : ECDHE-RSA-AES128-GCM-SHA256
    Session-ID: 057DF7D5B8BCE6DA3EAE6101136E644057BE67AF0A4931DC8FD15848D4E74D38
    Session-ID-ctx: 
    Master-Key: 807690ACC8782745D1C8AB6E4CF42FCAE7B13CAAC75A27FF4538FEA136DB9E6A332FDDB18703367593EBAD77629919C3
    PSK identity: None
    PSK identity hint: None
    SRP username: None
    Start Time: 1730703067
    Timeout   : 7200 (sec)
    Verify return code: 0 (ok)
    Extended master secret: yes
---
```

From the above output, we can see that we are able to connect to the ZooKeeper Ensemble using the TLS configuration.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete zookeeper -n demo zk-tls
kubectl delete issuer -n demo zookeeper-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ZooKeeper object](/docs/guides/zookeeper/concepts/zookeeper.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).