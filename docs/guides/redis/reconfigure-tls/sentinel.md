---
title: Reconfigure Redis Sentinel TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: rd-reconfigure-tls-sentinel
    name: Sentinel
    parent: rd-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Redis TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Redis database via a RedisOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

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

## Add TLS to a Redis database

Here we are going to reconfigure TLS of Redis in Sentinel Mode. First we are going to deploy a RedisSentinel instance and a Redis instance. Then wer are going to 
add TLS to them. 

### Deploy RedisSentinel without TLS :

In this section, we are going to deploy a `RedisSentinel` instance. Below is the YAML of the `RedisSentinel` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
metadata:
  name: sen-sample
  namespace: demo
spec:
  version: 6.2.14
  replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  terminationPolicy: DoNotTerminate
```

Let's create the `RedisSentinel` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure-tls/sentinel.yaml
redissentinel.kubedb.com/sen-sample created
```

Now, wait until `sen-sample` created has status `Ready`. i.e,

```bash
$ kubectl get redissentinel -n demo
NAME         VERSION   STATUS   AGE
sen-sample   6.2.14     Ready    5m20s
```

### Deploy Redis without TLS

In this section, we are going to deploy a Redis Standalone database without TLS. In the next few sections we will reconfigure TLS using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-sample
  namespace: demo
spec:
  version: 6.2.14
  replicas: 3
  sentinelRef:
    name: sen-sample
    namespace: demo
  mode: Sentinel
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  terminationPolicy: DoNotTerminate
```

Let's create the `Redis` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure-tls/rd-sentinel.yaml
redis.kubedb.com/rd-sample created
```

Now, wait until `redis-standalone` has status `Ready`. i.e,

```bash
$ watch kubectl get rd -n demo
Every 2.0s: kubectl get rd -n demo
NAME        VERSION   STATUS   AGE
rd-sample   6.2.14     Ready    88s
```

Now, we can connect to this database through redis-cli verify that the TLS is disabled.

```bash
$ kubectl exec -it -n demo rd-sample-0 -c redis -- bash

root@rd-sample-0:/data# redis-cli

127.0.0.1:6379> config get tls-cert-file
1) "tls-cert-file"
2) ""
127.0.0.1:6379> exit
root@rd-sample-0:/data# 
```

We can verify from the above output that TLS is disabled for this database.

## Create Issuer/ ClusterIssuer

Now, We are going to create an example `ClusterIssuer` that will be used to enable SSL/TLS in Redis. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `ClusterIssuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure-tls/clusterissuer.yaml
clusterissuer.cert-manager.io/redis-ca-issuer created
```

### Create RedisOpsRequest
There are two basic things to keep in mind when securing Redis using TLS in Sentinel Mode.

- Either Sentinel instance and Redis database both should have TLS enabled or both have TLS disabled.
- If TLS enabled, both Sentinel instance and Redis database should use the same `Issuer`. If they are in different namespace, in order to use same issuer, the certificates should be signed using `ClusterIssuer`

Currently, both Sentinel and Redis is tls disabled. If we want to add TLS to Redis database, we need to give reference to name/namespace of a Sentinel which 
is tls enabled. If a Sentinel is not found in given name/namespace KubeDB operator will create one.

In order to add TLS to the database, we have to create a `RedisOpsRequest` CRO with our created issuer. Below is the YAML of the `RedisOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: rd-sample
  tls:
    sentinel:
      ref:
        name: sen-demo-tls
        namespace: demo
      removeUnusedSentinel: true
    issuerRef:
      apiGroup: cert-manager.io
      name: redis-ca-issuer
      kind: ClusterIssuer
    certificates:
      - alias: client
        subject:
          organizations:
            - redis
          organizationalUnits:
            - client
```

Here,
- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `rd-sample` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.sentinel.ref` specifies the new sentinel which will monitor the redis after adding tls. If it does not exist, KubeDB will create one with given issuer.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/redis/concepts/redis.md#spectls).

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure-tls/rd-add-tls.yaml
redisopsrequest.ops.kubedb.com/rd-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CRO,

```bash
$ kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME           TYPE             STATUS        AGE
rd-add-tls     ReconfigureTLS   Successful    9m
```
We can see from the above output that the `RedisOpsRequest` has succeeded.

Let's check if new sentinel named `sen-demo-tls` is created 
```bash
$ kubectl get redissentinel -n demo
NAME           VERSION   STATUS   AGE
sen-demo-tls   6.2.14     Ready    17m
```

Now, connect to this database by exec into a pod and verify if `tls` has been set up as intended.

```bash
$ kubectl describe secret -n demo rd-sample-client-cert
Name:         rd-sample-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=rd-sample
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=redises.kubedb.com
Annotations:  cert-manager.io/alt-names: 
              cert-manager.io/certificate-name: rd-sample-client-cert
              cert-manager.io/common-name: default
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: ClusterIssuer
              cert-manager.io/issuer-name: redis-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1139 bytes
tls.crt:  1168 bytes
tls.key:  1675 bytes
```

Now, Lets exec into a redis container and find out the username to connect in a redis shell,

```bash
$ kubectl exec -it -n demo rd-sample-0 -c redis -- bash

root@rd-sample-0:/data# ls /certs
ca.crt	client.crt  client.key	server.crt  server.key

root@rd-sample-0:/data# redis-cli --tls --cert "/certs/client.crt" --key "/certs/client.key" --cacert "/certs/ca.crt" config get tls-cert-file
1) "tls-cert-file"
2) "/certs/server.crt

root@rd-sample-0:/data# apt-get update; apt-get install openssl;
...

root@rd-sample-0:/data# openssl x509 -in /certs/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb,CN=redis
```

Now, we can connect using `CN=redis,O=kubedb` as root to connect to the redis and write some data

```bash
$ kubectl exec -it -n demo rd-sample-0 -c redis -- bash
# Trying to connect without tls certificates
root@rd-sample-0:/data# redis-cli
127.0.0.1:6379> 
127.0.0.1:6379> set hello world
# Can not write data 
Error: Connection reset by peer 

# Trying to connect with tls certificates
root@rd-sample-0:/data# redis-cli --tls --cert "/certs/client.crt" --key "/certs/client.key" --cacert "/certs/ca.crt"
127.0.0.1:6379> 
127.0.0.1:6379> set hello world
OK
127.0.0.1:6379> exit
```

## Rotate Certificate

Now we are going to rotate the certificate of sentinel and database. First let's check the current expiration date of the certificate.

```bash
# Check Redis Certificate
$ kubectl exec -it -n demo rd-sample-0 -c redis -- bash

root@rd-sample-0:/data# apt-get update; apt-get install openssl;
...

root@rd-sample-0:/data# openssl x509 -in /certs/server.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=May 10 05:42:14 2023 GMT

# Check Sentinel Certificate
$ kubectl exec -it -n demo sen-demo-tls-0 -c redissentinel -- bash

root@sen-demo-tls-0:/data# apt-get update; apt-get install openssl;
...

root@sen-demo-tls-0:/data# openssl x509 -in /certs/server.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=May 10 05:41:19 2023 GMT
```

So, the redis certificate will expire on `May 10 05:42:14 2023 GMT` and sentinel certificate will expire on `notAfter=May 10 05:41:19 2023 GMT`. 

### Create RedisOpsRequest

Now we are going to increase it using a RedisOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-ops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: rd-sample
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `rd-sample` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure-tls/rd-ops-rotate.yaml
redisopsrequest.ops.kubedb.com/rd-ops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CRO,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME             TYPE             STATUS        AGE
rd-ops-rotate    ReconfigureTLS   Successful    5m5s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo rd-sample-0 -c redis -- bash

root@rd-sample-0:/data# apt-get update; apt-get install openssl;
...

root@rd-sample-0:/data# openssl x509 -in /certs/server.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=May 10 06:04:12 2023 GMT
```

As we can see from the above output, the certificate has been rotated successfully.

### Create RedisSentinelOpsRequest

Now we are going to increase it using a RedisOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: sen-ops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sen-demo-tls
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sen-demo-tls` sentinel.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure-tls/sen-ops-rotate.yaml
redisopsrequest.ops.kubedb.com/rd-ops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CRO,

```bash
$ watch kubectl get redissentinelopsrequest -n demo
Every 2.0s: kubectl get redissentinelopsrequest -n demo
NAME             TYPE             STATUS       AGE
sen-ops-rotate   ReconfigureTLS   Successful   78s
```

We can see from the above output that the `RedisSentinelOpsRequest` has succeeded.

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo sen-demo-tls-0 -c redissentinel -- bash

root@rd-sample-0:/data# apt-get update; apt-get install openssl;
...

root@rd-sample-0:/data# openssl x509 -in /certs/server.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=May 10 06:10:43 2023 GMT
```

As we can see from the above output, the certificate has been rotated successfully.


## Remove TLS from the Database

Now, we are going to remove TLS from this database using a RedisOpsRequest.

Currently, both Sentinel and Redis is tls enabled. If we want to remove TLS from Redis database, we need to give reference to name/namespace of a Sentinel which
is tls disabled. If a Sentinel is not found in given name/namespace KubeDB operator will create one.

### Create RedisOpsRequest

Below is the YAML of the `RedisOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-ops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: rd-sample
  tls:
    sentinel:
      ref:
        name: sen-sample
        namespace: demo
      removeUnusedSentinel: true
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `rd-sample` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.sentinel.ref` specifies the new sentinel which will monitor the redis after removing tls. If it does not exist, KubeDB will create a sentinel with given name/namespace.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure-tls/sen-ops-remove.yaml
redisopsrequest.ops.kubedb.com/rd-ops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CRO,

```bash
$ kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME            TYPE             STATUS        AGE
rd-ops-remove   ReconfigureTLS   Successful    2m5s
```
We can see from the above output that the `RedisOpsRequest` has succeeded.

Let's check if new sentinel named `sen-sample` is created
```bash
$ kubectl get redissentinel -n demo
NAME         VERSION   STATUS   AGE
sen-sample   6.2.14     Ready    7m56s
```

Now, Lets exec into the database primary node and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo rd-sample-0 -c redis -- bash
#
root@rd-sample-0:/data# redis-cli

127.0.0.1:6379> config get tls-cert-file
1) "tls-cert-file"
2) ""
127.0.0.1:6379> exit
root@rd-sample-0:/data# 
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
# Delete Redis and RedisOpsRequest
$ kubectl patch -n demo rd/rd-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-sample patched

$ kubectl delete -n demo redis rd-sample
redis.kubedb.com "rd-sample" deleted

$ kubectl delete -n demo redisopsrequest rd-add-tls rd-ops-remove rd-ops-rotate
redisopsrequest.ops.kubedb.com "rd-add-tls" deleted
redisopsrequest.ops.kubedb.com "rd-ops-remove" deleted
redisopsrequest.ops.kubedb.com "rd-ops-rotate" deleted

# Delete RedisSentinel and RedisSentinelOpsRequest
$ kubectl patch -n demo redissentinel/sen-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-sample patched

$ kubectl delete -n demo redissentinel sen-sample
redissentinel.kubedb.com "sen-sample" deleted

$ kubectl delete -n demo redissentinelopsrequests sen-ops-rotate
redissentinelopsrequest.ops.kubedb.com "sen-ops-rotate" deleted
```

## Next Steps

- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- [Backup and Restore](/docs/guides/redis/backup/overview/index.md) Redis databases using Stash. .
- Monitor your Redis database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
