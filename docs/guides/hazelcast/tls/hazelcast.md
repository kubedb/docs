---
title: Hazelcast TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: hz-tls-hazelcast
    name: Combined Cluster
    parent: hz-tls-hazelcast
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Hazelcast with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for `Hazelcast`. This tutorial will show you how to use KubeDB to run a Hazelcast with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes combined, and the kubectl command-line tool must be configured to communicate with your combined. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your combined to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your combined following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/Hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Hazelcast.

- `spec:`
    - `enableSSL`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [Hazelcast Concept](/docs/guides/hazelcast/concepts/hazelcast.md),

`tls` is applicable for  Hazelcast cluster.

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt`, `tls.key`, `keystore.jks` and `truststore.jks`.

## Create Issuer/ clusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Hazelcast. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=hazelcast /O=kubedb"
......................+...........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+............+.....+...+...+.....................+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*....+..+.+..+...............+.+.....+....+...+.....+.............+.........+..+...+.+......+........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
.+......+.+............+..+.........+....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*......+...+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*..............+..........+...+.....+.+.....+....+......+.....+.+......+..+.+..+.........+...+..........+..+.................................+.......+.....................+..+......................+......+.....+...+....+..+....+.........+...+..............+....+..+...+.+........+.+..+.........+...+................+..+...+.......+............+...+........+..........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
-----

```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls hz-ca --cert=ca.crt  --key=ca.key --namespace=cert-manager 
secret/hz-ca created
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: self-signed-issuer
spec:
  ca:
    secretName: hz-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/tls/hz-issuer.yaml
issuer.cert-manager.io/self-signed-issuer created
```

## TLS/SSL encryption in Hazelcast

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hazelcast-sample
  namespace: demo
spec:
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: self-signed-issuer
      kind: ClusterIssuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
      - alias: client
        subject:
          organizations:
            - kubedb
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  enableSSL: true
  deletionPolicy: WipeOut
  licenseSecret:
    name: hz-license-key
  replicas: 3
  version: 5.5.2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
    storageClassName: standard

```

### Deploy Hazelcast Combined with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/tls/hazelcast-sample.yaml
hazelcast.kubedb.com/hazelcast-sample created
```

Now, wait until `hazelcast-sample` created has status `Ready`. i.e,

```bash
$ kubectl get hz -n demo
NAME               TYPE                  VERSION   STATUS   AGE
hazelcast-sample   kubedb.com/v1alpha2   5.5.2     Ready    165m
```

### Verify TLS/SSL in Hazelcast Combined

```bash
$ kubectl describe secret hazelcast-sample-client-cert -n demo
Name:         hazelcast-sample-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=hazelcast-sample
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=hazelcasts.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.hazelcast-sample-pods.demo,*.hazelcast-sample-pods.demo.svc,*.hazelcast-sample-pods.demo.svc.cluster.local,hazelcast-sample,hazelcast-sa...
              cert-manager.io/certificate-name: hazelcast-sample-client-cert
              cert-manager.io/common-name: hazelcast-sample
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: ClusterIssuer
              cert-manager.io/issuer-name: self-signed-issuer
              cert-manager.io/subject-organizations: kubedb
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:          1164 bytes
keystore.p12:    3615 bytes
tls.crt:         1627 bytes
tls.key:         1675 bytes
truststore.p12:  1114 bytes
```

We can see from the above output that, keystore location is `/var/Hazelcast/etc` which means that TLS is enabled.

```bash
bash-5.1$ cd /data/etc/server
bash-5.1$ ls
ca.crt	keystore.p12  tls.crt  tls.key	truststore.p12
```

From the above output, we can see that we are able to connect to the Hazelcast combined using the TLS configuration.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcast -n demo hazelcast-sample
kubectl delete clusterissuer -n demo self-signed-issuer
kubectl delete ns demo
```

## Next Steps

- Monitor your Hazelcast combined with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).
- Monitor your Hazelcast combined with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/hazelcast/monitoring/prometheus-builtin.md).
- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
