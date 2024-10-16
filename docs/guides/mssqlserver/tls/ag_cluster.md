---
title: MSSQLServer Sentinel TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: ms-tls-sentinel
    name: Sentinel
    parent: ms-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run SQL Server with TLS/SSL Encryption

KubeDB supports providing TLS/SSL encryption for MSSQLServer. This tutorial will show you how to use KubeDB to run a MSSQLServer database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in MSSQLServer and MSSQLServerSentinel.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificate`

There are two basic things to keep in mind when securing MSSQLServer using TLS in Sentinel Mode.

- Either Sentinel instance and MSSQLServer database both should have TLS enabled or both have TLS disabled.

- If TLS enabled, both Sentinel instance and MSSQLServer database should use the same `Issuer`. If they are in different namespace, in order to use same issuer, the certificates should be signed using `ClusterIssuer`

Read about the fields in details in [mssqlserver concept](/docs/guides/mssqlserver/concepts/mssqlserver.md) and [mssqlserversentinel concept](/docs/guides/mssqlserver/concepts/mssqlserversentinel.md)

## Create Issuer/ ClusterIssuer

We are going to create an example `ClusterIssuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in MSSQLServer. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `ClusterIssuer`.

- Start off by generating you can certificate using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mssqlserver/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated. The secret should be created in `cert-manager` namespace to create the `ClusterIssuer`.

```bash
$ kubectl create secret tls mssqlserver-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=cert-manager
```

Now, create an `ClusterIssuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: mssqlserver-ca-issuer
spec:
  ca:
    secretName: mssqlserver-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/tls/clusterissuer.yaml
clusterissuer.cert-manager.io/mssqlserver-ca-issuer created
```

## TLS/SSL encryption in Sentinel

Below is the YAML for MSSQLServer  in Sentinel Mode.
```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServerSentinel
metadata:
  name: sen-tls
  namespace: demo
spec:
  replicas: 3
  version: "6.2.14"
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: ClusterIssuer
      name: mssqlserver-ca-issuer
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

### Deploy MSSQLServer in Sentinel Mode

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/tls/sentinel-ssl.yaml
mssqlserversentinel.kubedb.com/sen-tls created
```

Now, wait until `sen-tls` has status `Ready`. i.e,

```bash
$ watch kubectl get mssqlserversentinel -n demo
Every 2.0s: kubectl get mssqlserver -n demo
NAME      VERSION   STATUS   AGE
sen-tls   6.2.14     Ready    111s
```

### Verify TLS/SSL in MSSQLServer in Sentinel Mode

Now, connect to this database by exec into a pod and verify if `tls` has been set up as intended.

```bash
$ kubectl describe secret -n demo sen-tls-client-cert
Name:         sen-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sen-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mssqlserversentinels.kubedb.com
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: sen-tls-client-cert
              cert-manager.io/common-name: default
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: ClusterIssuer
              cert-manager.io/issuer-name: mssqlserver-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1147 bytes
tls.crt:  1127 bytes
tls.key:  1675 bytes
```

## TLS/SSL encryption in MSSQLServer in Sentinel Mode

Below is the YAML for MSSQLServer  in Sentinel Mode.
```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: ms-tls
  namespace: demo
spec:
  version: "6.2.14"
  mode: Sentinel
  replicas: 3
  sentinelRef:
    name: sen-tls
    namespace: demo
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: ClusterIssuer
      name: mssqlserver-ca-issuer
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

### Deploy MSSQLServer in Sentinel Mode

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/tls/ms-sentinel.yaml
mssqlserver.kubedb.com/ms-tls created
```

Now, wait until `ms-tls` has status `Ready`. i.e,

```bash
$ watch kubectl get ms -n demo
Every 2.0s: kubectl get mssqlserver -n demo
NAME      VERSION     STATUS     AGE
ms-tls    6.2.14       Ready      2m14s
```

### Verify TLS/SSL in MSSQLServer in Sentinel Mode

Now, connect to this database by exec into a pod and verify if `tls` has been set up as intended.

```bash
$ kubectl describe secret -n demo ms-tls-client-cert
Name:         ms-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=ms-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mssqlserveres.kubedb.com
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: ms-tls-client-cert
              cert-manager.io/common-name: default
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: ClusterIssuer
              cert-manager.io/issuer-name: mssqlserver-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
tls.key:  1679 bytes
ca.crt:   1147 bytes
tls.crt:  1127 bytes
```


Now, we can connect using tls-certs connect to the mssqlserver and write some data

```bash
$ kubectl exec -it -n demo ms-tls-0 -c mssqlserver -- bash

# Trying to connect without tls certificates
root@ms-tls-0:/data# mssqlserver-cli
127.0.0.1:6379> 
127.0.0.1:6379> set hello world
# Can not write data 
Error: Connection reset by peer 

# Trying to connect with tls certificates
root@ms-tls-0:/data# mssqlserver-cli --tls --cert "/certs/client.crt" --key "/certs/client.key" --cacert "/certs/ca.crt"
127.0.0.1:6379> 
127.0.0.1:6379> set hello world
OK
127.0.0.1:6379> exit
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo mssqlserver/ms-tls -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
mssqlserver.kubedb.com/ms-tls patched

$ kubectl delete -n demo mssqlserver ms-tls
mssqlserver.kubedb.com "ms-tls" deleted

$ kubectl patch -n demo mssqlserversentinel/sen-tls -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
mssqlserversentinel.kubedb.com/sen-tls patched

$ kubectl delete -n demo mssqlserversentinel sen-tls
mssqlserversentinel.kubedb.com "sen-tls" deleted

$ kubectl delete clusterissuer mssqlserver-ca-issuer
clusterissuer.cert-manager.io "mssqlserver-ca-issuer" deleted
```

## Next Steps

- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- [Backup and Restore](/docs/guides/mssqlserver/backup/kubestash/overview/index.md) MSSQLServer databases using KubeStash. .
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mssqlserver/monitoring/using-builtin-prometheus.md).