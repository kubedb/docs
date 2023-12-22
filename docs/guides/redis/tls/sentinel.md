---
title: Redis Sentinel TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: rd-tls-sentinel
    name: Sentinel
    parent: rd-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Redis with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for Redis. This tutorial will show you how to use KubeDB to run a Redis database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Redis and RedisSentinel.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificate`

There are two basic things to keep in mind when securing Redis using TLS in Sentinel Mode.

- Either Sentinel instance and Redis database both should have TLS enabled or both have TLS disabled.

- If TLS enabled, both Sentinel instance and Redis database should use the same `Issuer`. If they are in different namespace, in order to use same issuer, the certificates should be signed using `ClusterIssuer`

Read about the fields in details in [redis concept](/docs/guides/redis/concepts/redis.md) and [redissentinel concept](/docs/guides/redis/concepts/redissentinel.md)

## Create Issuer/ ClusterIssuer

We are going to create an example `ClusterIssuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Redis. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `ClusterIssuer`.

- Start off by generating you can certificate using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=redis/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated. The secret should be created in `cert-manager` namespace to create the `ClusterIssuer`.

```bash
$ kubectl create secret tls redis-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=cert-manager
```

Now, create an `ClusterIssuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: redis-ca-issuer
spec:
  ca:
    secretName: redis-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/tls/clusterissuer.yaml
clusterissuer.cert-manager.io/redis-ca-issuer created
```

## TLS/SSL encryption in Sentinel

Below is the YAML for Redis  in Sentinel Mode.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
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
      name: redis-ca-issuer
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

### Deploy Redis in Sentinel Mode

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/tls/sentinel-ssl.yaml
redissentinel.kubedb.com/sen-tls created
```

Now, wait until `sen-tls` has status `Ready`. i.e,

```bash
$ watch kubectl get redissentinel -n demo
Every 2.0s: kubectl get redis -n demo
NAME      VERSION   STATUS   AGE
sen-tls   6.2.14     Ready    111s
```

### Verify TLS/SSL in Redis in Sentinel Mode

Now, connect to this database by exec into a pod and verify if `tls` has been set up as intended.

```bash
$ kubectl describe secret -n demo sen-tls-client-cert
Name:         sen-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sen-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=redissentinels.kubedb.com
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: sen-tls-client-cert
              cert-manager.io/common-name: default
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: ClusterIssuer
              cert-manager.io/issuer-name: redis-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1147 bytes
tls.crt:  1127 bytes
tls.key:  1675 bytes
```

Now, Lets exec into a redis container and find out the username to connect in a redis shell,

```bash
$ kubectl exec -it -n demo sen-tls-0 -c redissentinel -- bash
 
root@sen-tls-0:/data# ls /certs
ca.crt	client.crt  client.key	server.crt  server.key

root@sen-tls-0:/data# apt-get update; apt-get install openssl;
...

root@sen-tls-0:/data# openssl x509 -in /certs/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb,CN=redis
```

## TLS/SSL encryption in Redis in Sentinel Mode

Below is the YAML for Redis  in Sentinel Mode.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-tls
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
      name: redis-ca-issuer
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

### Deploy Redis in Sentinel Mode

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/tls/rd-sentinel.yaml
redis.kubedb.com/rd-tls created
```

Now, wait until `rd-tls` has status `Ready`. i.e,

```bash
$ watch kubectl get rd -n demo
Every 2.0s: kubectl get redis -n demo
NAME      VERSION     STATUS     AGE
rd-tls    6.2.14       Ready      2m14s
```

### Verify TLS/SSL in Redis in Sentinel Mode

Now, connect to this database by exec into a pod and verify if `tls` has been set up as intended.

```bash
$ kubectl describe secret -n demo rd-tls-client-cert
Name:         rd-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=rd-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=redises.kubedb.com
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: rd-tls-client-cert
              cert-manager.io/common-name: default
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: ClusterIssuer
              cert-manager.io/issuer-name: redis-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
tls.key:  1679 bytes
ca.crt:   1147 bytes
tls.crt:  1127 bytes
```

Now, Lets exec into a redis container and find out the username to connect in a redis shell,

```bash
$ kubectl exec -it -n demo rd-tls-0 -c redis -- bash

root@rd-tls-0:/data# ls /certs
ca.crt	client.crt  client.key	server.crt  server.key

root@rd-tls-0:/data# apt-get update; apt-get install openssl;
...

root@rd-tls-0:/data# openssl x509 -in /certs/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb,CN=redis
```

Now, we can connect using `CN=redis,O=kubedb` as root to connect to the redis and write some data

```bash
$ kubectl exec -it -n demo rd-tls-0 -c redis -- bash

# Trying to connect without tls certificates
root@rd-tls-0:/data# redis-cli
127.0.0.1:6379> 
127.0.0.1:6379> set hello world
# Can not write data 
Error: Connection reset by peer 

# Trying to connect with tls certificates
root@rd-tls-0:/data# redis-cli --tls --cert "/certs/client.crt" --key "/certs/client.key" --cacert "/certs/ca.crt"
127.0.0.1:6379> 
127.0.0.1:6379> set hello world
OK
127.0.0.1:6379> exit
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo redis/rd-tls -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-tls patched

$ kubectl delete -n demo redis rd-tls
redis.kubedb.com "rd-tls" deleted

$ kubectl patch -n demo redissentinel/sen-tls -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-tls patched

$ kubectl delete -n demo redissentinel sen-tls
redissentinel.kubedb.com "sen-tls" deleted

$ kubectl delete clusterissuer redis-ca-issuer
clusterissuer.cert-manager.io "redis-ca-issuer" deleted
```

## Next Steps

- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- [Backup and Restore](/docs/guides/redis/backup/overview/index.md) Redis databases using Stash. .
- Monitor your Redis database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
