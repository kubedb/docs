---
title: FerretDB TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: fr-tls-configure
    name: FerretDB TLS/SSL Configuration
    parent: fr-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run FerretDB with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption (via, `sslMode`) for FerretDB. This tutorial will show you how to use KubeDB to run a FerretDB database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/ferretdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ferretdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Mongodb.

- `spec:`
    - `sslMode`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [ferretdb concept](/docs/guides/ferretdb/concepts/ferretdb.md),

`sslMode` enables TLS/SSL or mixed TLS/SSL used for all network connections. The value of `sslMode` field can be one of the following:

|    Value     | Description                                                                                                                    |
| :----------: | :----------------------------------------------------------------------------------------------------------------------------- |
|  `disabled`  | The server does not use TLS/SSL.                                                                                               |
| `requireSSL` | The server uses and accepts only TLS/SSL encrypted connections.                                                                |

The specified ssl mode will be used by health checker and exporter of FerretDB.

When, SSLMode is anything other than `disabled`, users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.pem`, `tls.crt` and `tls.key`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in FerretDB. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ferretdb/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls ferretdb-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: ferretdb-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: ferretdb-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/tls/issuer.yaml
issuer.cert-manager.io/ferretdb-ca-issuer created
```

## TLS/SSL encryption in FerretDB

Below is the YAML for FerretDB with TLS enabled. Backend Postgres will automatically managed by KubeDB:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: fr-tls
  namespace: demo
spec:
  version: "2.0.0"
  authSecret:
    kind: Secret
    name: ferret-auth
    externallyManaged: false
  backend:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 500Mi
  deletionPolicy: WipeOut
  sslMode: requireSSL
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: ferretdb-ca-issuer
```

### Deploy FerretDB

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/tls/ferretdb-tls.yaml
ferretdb.kubedb.com/fr-tls created
```

Now, wait until `fr-tls created` has status `Ready`. i.e,

```bash
$ watch kubectl get fr -n demo
Every 2.0s: kubectl get ferretdb -n demo
NAME     TYPE                  VERSION   STATUS   AGE
fr-tls   kubedb.com/v1alpha2   2.0.0     Ready    60s
```

### Verify TLS/SSL in FerretDB

Now, connect to this database through [mongosh](https://www.mongodb.com/docs/mongodb-shell/) and verify if `SSLMode` has been set up as intended (i.e, `require`).

```bash
$ kubectl describe secret -n demo fr-tls-client-cert
Name:         fr-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=fr-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=ferretdbs.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: fr-tls-client-cert
              cert-manager.io/common-name: fr-tls
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: ferretdb-ca-issuer
              cert-manager.io/subject-organizationalunits: client
              cert-manager.io/subject-organizations: kubedb
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1155 bytes
tls.crt:  1176 bytes
tls.key:  1679 bytes
```

Now we need save the client cert and key to two different files and make a pem file.
Additionally, to verify server, we need to store ca.crt.

```bash
$ kubectl get secrets -n demo fr-tls-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > client.crt
$ kubectl get secrets -n demo fr-tls-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > client.key
$ kubectl get secrets -n demo fr-tls-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
$ cat client.crt client.key > client.pem
```

Now, we can connect to our FerretDB with these files with mongosh client.

```bash
$ kubectl get secrets -n demo fr-tls-auth -o jsonpath='{.data.\username}' | base64 -d
postgres
$ kubectl get secrets -n demo fr-tls-auth -o jsonpath='{.data.\\password}' | base64 -d
l*jGp8u*El8WRSDJ

$ kubectl port-forward svc/fr-tls -n demo 27017
Forwarding from 127.0.0.1:27017 -> 27018
Forwarding from [::1]:27017 -> 27018
Handling connection for 27017
Handling connection for 27017
```

Now in another terminal

```bash
$ mongosh 'mongodb://postgres:l*jGp8u*El8WRSDJ@localhost:27017/ferretdb?authMechanism=PLAIN&tls=true&tlsCertificateKeyFile=./client.pem&tlsCaFile=./ca.crt'
Current Mongosh Log ID:	67ee22bbd9c3422c286b140a
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.4.2
Using MongoDB:		7.0.77
Using Mongosh:		2.4.2

For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/

------
   The server generated these startup warnings when booting
   2025-04-03T05:55:07.528Z: Powered by FerretDB v2.0.0-1-g7fb2c9a8 and DocumentDB 0.102.0 (PostgreSQL 17.4).
   2025-04-03T05:55:07.528Z: Please star 🌟 us on GitHub: https://github.com/FerretDB/FerretDB and https://github.com/microsoft/documentdb.
   2025-04-03T05:55:07.528Z: The telemetry state is undecided. Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.
------

ferretdb> 
```

So our connection is now tls encrypted.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ferretdb -n demo fr-tls
kubectl delete issuer -n demo ferretdb-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Detail concepts of [FerretDBVersion object](/docs/guides/ferretdb/concepts/catalog.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ferretdb/monitoring/using-prometheus-operator.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ferretdb/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
