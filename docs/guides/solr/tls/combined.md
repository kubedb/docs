---
title: Solr TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: sl-tls-combined
    name: Combined Cluster
    parent: sl-tls-solr
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Solr with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for `Solr`. This tutorial will show you how to use KubeDB to run a Solr combined with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes combined, and the kubectl command-line tool must be configured to communicate with your combined. If you do not already have a combined, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your combined to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your combined following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/Solr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Solr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Solr.

- `spec:`
    - `enableSSL`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [Solr Concept](/docs/guides/solr/concepts/solr.md),

`tls` is applicable for all types of Solr (i.e., `combined` and `combined`).

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `combinedIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt`, `tls.key`, `keystore.jks` and `truststore.jks`.

## Create Issuer/ combinedIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Solr. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=Solr/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls solr-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: solr-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: solr-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/tls/sl-issuer.yaml
issuer.cert-manager.io/solr-ca-issuer created
```

## TLS/SSL encryption in Solr Combined

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-combined
  namespace: demo
spec:
  version: 9.4.1
  replicas: 2
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: solr-ca-issuer
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
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard

```

### Deploy Solr Combined with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/tls/solr-combined.yaml
solr.kubedb.com/solr-combined created
```

Now, wait until `solr-combined created` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-combined   kubedb.com/v1alpha2   9.4.1     Ready    2m31s
```

### Verify TLS/SSL in Solr Combined

```bash
$ kubectl describe secret solr-combined-client-cert -n demo
Name:         solr-combined-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=solr-combined
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=solrs.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.solr-combined-pods.demo,*.solr-combined-pods.demo.svc.combined.local,localhost,solr-combined,solr-combined-pods,solr-combined-pods.demo.svc,so...
              cert-manager.io/certificate-name: solr-combined-client-cert
              cert-manager.io/common-name: solr-combined
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: combinedIssuer
              cert-manager.io/issuer-name: self-signed-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
truststore.p12:  1090 bytes
ca.crt:          1147 bytes
keystore.p12:    3511 bytes
tls.crt:         1497 bytes
tls.key:         1679 bytes
```

Now, Let's exec into a solr data pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo solr-combined-data-0 -- bash
Defaulted container "solr" out of: solr, init-solr (init)
solr@solr-combined-data-0:/opt/solr-9.4.1$ env | grep -i SSL
JAVA_OPTS= -Djavax.net.ssl.trustStore=/var/solr/etc/truststore.p12 -Djavax.net.ssl.trustStorePassword=QyHKB(dYoT1MQYMu -Djavax.net.ssl.keyStore=/var/solr/etc/keystore.p12 -Djavax.net.ssl.keyStorePassword=QyHKB(dYoT1MQYMu -Djavax.net.ssl.keyStoreType=PKCS12 -Djavax.net.ssl.trustStoreType=PKCS12
SOLR_SSL_TRUST_STORE_PASSWORD=QyHKB(dYoT1MQYMu
SOLR_SSL_ENABLED=true
SOLR_SSL_WANT_CLIENT_AUTH=false
SOLR_SSL_KEY_STORE_PASSWORD=QyHKB(dYoT1MQYMu
SOLR_SSL_TRUST_STORE=/var/solr/etc/truststore.p12
SOLR_SSL_KEY_STORE=/var/solr/etc/keystore.p12
SOLR_SSL_NEED_CLIENT_AUTH=false

```

We can see from the above output that, keystore location is `/var/solr/etc` which means that TLS is enabled.

```bash
solr@solr-combined-data-0:/var/solr/etc$ ls
ca.crt	keystore.p12  tls.crt  tls.key	truststore.p12

```

From the above output, we can see that we are able to connect to the Solr combined using the TLS configuration.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete solr -n demo solr-combined
kubectl delete issuer -n demo solr-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Monitor your Solr combined with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).
- Monitor your Solr combined with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md).
- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
