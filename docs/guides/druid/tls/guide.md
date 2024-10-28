---
title: Druid Combined TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-druid-tls-guide
    name: Guide
    parent: guides-druid-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Druid with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for Druid. This tutorial will show you how to use KubeDB to run a Druid cluster with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/druid](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Druid.

- `spec:`
    - `enableSSL`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [druid concept](/docs/guides/druid/concepts/druid.md),

`tls` is applicable for all types of Druid (i.e., `combined` and `topology`).

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt`, `tls.key`, `keystore.jks` and `truststore.jks`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Druid. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=druid/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls druid-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: druid-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: druid-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/tls/yamls/druid-ca-issuer.yaml
issuer.cert-manager.io/druid-ca-issuer created
```

## TLS/SSL encryption in Druid Cluster

### Create External Dependency (Deep Storage)

Before proceeding further, we need to prepare deep storage, which is one of the external dependency of Druid and used for storing the segments. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/tls/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

Now, lets go ahead and create a druid database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster-tls
  namespace: demo
spec:
  version: 28.0.1
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: druid-ca-issuer
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

### Deploy Druid Topology Cluster with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/tls/yamls/druid-cluster-tls.yaml
druid.kubedb.com/druid-cluster-tls created
```

Now, wait until `druid-cluster-tls created` has status `Ready`. i.e,

```bash
$ kubectl get druid -n demo -w

Every 2.0s: kubectl get druid -n demo                                                                                                                          aadee: Fri Sep  6 12:34:51 2024
NAME                TYPE                  VERSION   STATUS          AGE
druid-cluster-tls   kubedb.com/v1alpha2   28.0.1    Ready           20s
druid-cluster-tls   kubedb.com/v1alpha2   28.0.1    Provisioning    1m
...
...
druid-cluster-tls   kubedb.com/v1alpha2   28.0.1    Ready           38m
```

### Verify TLS/SSL in Druid Cluster

```bash
$ kubectl describe secret druid-cluster-tls-client-cert -n demo
Name:         druid-cluster-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=druid-cluster-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=druids.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.druid-cluster-tls-pods.demo.svc.cluster.local,druid-cluster-tls-brokers-0.druid-cluster-tls-pods.demo.svc.cluster.local:8282,druid-clust...
              cert-manager.io/certificate-name: druid-cluster-tls-client-cert
              cert-manager.io/common-name: druid-cluster-tls-pods.demo.svc
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: druid-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:            1147 bytes
keystore.jks:      3720 bytes
tls-combined.pem:  3835 bytes
tls.crt:           2126 bytes
tls.key:           1708 bytes
truststore.jks:    865 bytes
```

Now, Lets exec into a druid coordinators pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo druid-cluster-tls-coordinators-0 -- bash
Defaulted container "druid" out of: druid, init-druid (init)
bash-5.1$ cat conf/druid/cluster/_common/common.runtime.properties 
druid.client.https.trustStorePassword={"type": "environment", "variable": "DRUID_KEY_STORE_PASSWORD"}
druid.client.https.trustStorePath=/opt/druid/ssl/truststore.jks
druid.client.https.trustStoreType=jks
druid.emitter=noop
druid.enablePlaintextPort=false
druid.enableTlsPort=true
druid.metadata.mysql.ssl.clientCertificateKeyStorePassword=password
druid.metadata.mysql.ssl.clientCertificateKeyStoreType=JKS
druid.metadata.mysql.ssl.clientCertificateKeyStoreUrl=/opt/druid/ssl/metadata/keystore.jks
druid.metadata.mysql.ssl.useSSL=true
druid.server.https.certAlias=druid
druid.server.https.keyStorePassword={"type": "environment", "variable": "DRUID_KEY_STORE_PASSWORD"}
druid.server.https.keyStorePath=/opt/druid/ssl/keystore.jks
druid.server.https.keyStoreType=jks
```

We can see from the above output that, all the TLS related configuration is added. Here the `MySQL` and `ZooKeeper` deployed with Druid is also TLS secure and their connection configs are added as well.

#### Verify TLS/SSL using Druid UI

To check follow the following steps:

Druid uses separate ports for TLS/SSL. While the plaintext port for `routers` node is `8888`. For TLS, it is `9088`. Hence, we will use that port to access the UI. 

First port-forward the port `9088` to local machine:

```bash
$ kubectl port-forward -n demo svc/druid-cluster-tls-routers 9088
Forwarding from 127.0.0.1:9088 -> 9088
Forwarding from [::1]:9088 -> 9088
```


Now hit the `https://localhost:9088/` from any browser. Here you may select `Advance` and then `Proceed to localhost (unsafe)` or you can add the `ca.crt` from the secret `druid-cluster-tls-client-cert` to your browser's Authorities.

After that you will be prompted to provide the credential of the druid database. By following the steps discussed below, you can get the credential generated by the KubeDB operator for your Druid database.

**Connection information:**

- Username:

  ```bash
  $ kubectl get secret -n demo druid-cluster-tls-admin-cred -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo druid-cluster-tls-admin-cred -o jsonpath='{.data.password}' | base64 -d
  LzJtVRX5E8MorFaf
  ```

After providing the credentials correctly, you should be able to access the web console like shown below.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/druid/tls/images/druid-ui.png">
</p>

From the above output, we can see that the connection is secure.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete druid -n demo druid-cluster-tls
kubectl delete issuer -n demo druid-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Monitor your Druid cluster with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- Monitor your Druid cluster with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/druid/monitoring/using-builtin-prometheus.md).
- Use [kubedb cli](/docs/guides/druid/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
