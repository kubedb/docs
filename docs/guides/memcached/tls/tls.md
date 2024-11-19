---
title: Memcached TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: mc-tls
    name: TLS
    parent: tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Memcached with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for Memcached. This tutorial will show you how to use KubeDB to run a Memcached database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/Memcached](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Memcached) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Memcached.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificate`

Read about the fields in details in [Memcached concept](/docs/guides/memcached/concepts/memcached.md),

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Memcached. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=Memcached/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
$ kubectl create secret tls memcached-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: memcached-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: memcached-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/tls/issuer.yaml
issuer.cert-manager.io/memcached-ca-issuer created
```

## TLS/SSL encryption in Memcached Standalone

Below is the YAML for Memcached Standalone.

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: memcached-ca-issuer
    certificates:
      - alias: client
        ipAddresses:
          - 127.0.0.1
          - 192.168.0.252
  deletionPolicy: WipeOut
```

### Deploy Memcached

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/tls/mc-tls.yaml
memcached.kubedb.com/memcd-quickstart created
```

Now, wait until `memcd-quickstart` has status `Ready`. i.e,

```bash
$ watch kubectl get rd -n demo
Every 2.0s: kubectl get memcached -n demo
NAME               VERSION   STATUS   AGE
memcd-quickstart   1.6.22    Ready    19m
```

### Verify TLS/SSL in Memcached

Now, connect to this database by exec into a pod and verify if `tls` has been set up as intended.

```bash
$ kubectl describe secret -n demo memcd-quickstart-client-cert
Name:         memcd-quickstart-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=memcd-quickstart
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=memcacheds.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: memcd-quickstart-client-cert
              cert-manager.io/common-name: memcached
              cert-manager.io/ip-sans: 127.0.0.1,192.168.0.252
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: memcached-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
tls.crt:           1168 bytes
tls.key:           1675 bytes
ca.crt:            1159 bytes
tls-combined.pem:  2844 bytes
```

Now, we can connect to the Memcached and read/write some data

```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
orwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
```

Telnet doesn't support TLS. To overcome this, we will use socat:
```bash
$ socat -d -d \
            TCP-LISTEN:12345,reuseaddr,fork \
            OPENSSL:localhost:11211,cert=/path/client.crt,key=/path/client.key,cafile=/path/ca.crt,verify=1
2024/11/15 12:02:41 socat[46145] N listening on AF=10 [0000:0000:0000:0000:0000:0000:0000:0000]:12345
```
Now connect to the memcached via socat using telnet:
```bash
$ telnet 127.0.0.1 12345
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.

# Set username and password for authentication.
set key 0 0 21
user **znjl**ketkdj**
STORED

# Set/Write a value
set mc-key 0 9999 8
mc-value
STORED

# Get/Read a value
get mc-key
VALUE mc-key 0 8
mc-value
END

# Current Stats Settings
stats settings
...
ssl_enabled yes
ssl_chain_cert /usr/certs/server.crt
ssl_key /usr/certs/server.key
ssl_ca_cert /usr/certs/ca.crt
...
END

quit
```
## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo memcached/memcd-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
memcached.kubedb.com/memcd-quickstart patched

$ kubectl delete -n demo memcached memcd-quickstart
memcached.kubedb.com "memcd-quickstart" deleted

$ kubectl delete issuer -n demo memcached-ca-issuer
issuer.cert-manager.io "memcached-ca-issuer" deleted
```

## Next Steps

- Detail concepts of [Memcached object](/docs/guides/memcached/concepts/memcached.md).
- Monitor your Memcached database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/memcached/monitoring/using-prometheus-operator.md).
- Monitor your Memcached database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md).
