---
title: ClickHouse Topology TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: ch-tls-topology
    name: Cluster
    parent: ch-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run ClickHouse with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for ClickHouse. This tutorial will show you how to use KubeDB to run a ClickHouse cluster with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/clickhouse](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/clickhouse) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in ClickHouse.

- `spec:`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [clickhouse concept](/docs/guides/clickhouse/concepts/clickhouse.md),

`tls` is applicable for all types of ClickHouse (i.e., `standalone` and `clickhouse`).

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt`, `tls.key`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in ClickHouse. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=clickhouse/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls clickhouse-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: clickhouse-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: clickhouse-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/tls/clickhouse-issuer.yaml
issuer.cert-manager.io/clickhouse-ca-issuer created
```

## TLS/SSL encryption in ClickHouse Cluster

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-prod-tls
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      - name: appscode-cluster
        shards: 2
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: clickhouse
                resources:
                  limits:
                    memory: 4Gi
                  requests:
                    cpu: 500m
                    memory: 512Mi
            initContainers:
              - name: clickhouse-init
                resources:
                  limits:
                    memory: 1Gi
                  requests:
                    cpu: 500m
                    memory: 512Mi
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  sslVerificationMode: relaxed
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: clickhouse-ca-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  deletionPolicy: WipeOut

```

### Deploy ClickHouse Topology Cluster with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/tls/clickhouse-prod-tls.yaml
clickhouse.kubedb.com/clickhouse-prod-tls created
```

Now, wait until `clickhouse-prod-tls created` has status `Ready`. i.e,

```bash
➤ kubectl get clickhouse -n demo -w
NAME                  TYPE                  VERSION   STATUS         AGE
clickhouse-prod-tls   kubedb.com/v1alpha2   24.4.1    Provisioning   31s
clickhouse-prod-tls   kubedb.com/v1alpha2   24.4.1    Provisioning   51s
.
.
clickhouse-prod-tls   kubedb.com/v1alpha2   24.4.1    Ready          2m6s
```

### Verify TLS/SSL in ClickHouse Topology Cluster

```bash
➤ kubectl describe secret clickhouse-prod-tls-client-cert -n demo
Name:         clickhouse-prod-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=clickhouse-prod-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=clickhouses.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.clickhouse-prod-tls-pods.demo.svc.cluster.local,clickhouse-prod-tls,clickhouse-prod-tls-pods,clickhouse-prod-tls-pods.demo.svc,clickhous...
              cert-manager.io/certificate-name: clickhouse-prod-tls-client-cert
              cert-manager.io/common-name: clickhouse-prod-tls.demo.svc
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: clickhouse-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:            1164 bytes
tls-combined.pem:  3221 bytes
tls.crt:           1541 bytes
tls.key:           1679 bytes
```

Now, Let's exec into a clickhouse pod and verify the configuration that the TLS is enabled.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-tls-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-tls-appscode-cluster-shard-0-0:/$ openssl s_client -connect localhost:9440
CONNECTED(00000003)
Can't use SSL_get_servername
depth=1 CN = clickhouse, O = kubedb
verify error:num=19:self signed certificate in certificate chain
verify return:1
depth=1 CN = clickhouse, O = kubedb
verify return:1
depth=0 O = kubedb:server, CN = clickhouse-prod-tls
verify return:1
---
Certificate chain
 0 s:O = kubedb:server, CN = clickhouse-prod-tls
   i:CN = clickhouse, O = kubedb
 1 s:CN = clickhouse, O = kubedb
   i:CN = clickhouse, O = kubedb
```

We can see from the above output that, tls port is accessible by using openssl which means that TLS is enabled.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete clickhouse -n demo clickhouse-prod-tls
kubectl delete issuer -n demo clickhouse-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
