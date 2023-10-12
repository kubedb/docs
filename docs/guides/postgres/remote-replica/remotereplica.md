---
title: PostgreSQL Remote Replica
menu:
  docs_{{ .version }}:
    identifier: pg-remote-replica-details
    name: Overview
    parent: pg-remote-replica
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running PostgreSQL

This tutorial will show you how to use KubeDB to run a PostgreSQL database.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/postgres/lifecycle.png">
</p>

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/postgres/remote-replica/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).
## Deploy PostgreSQL server

The following is an example `PostgreSQL` object which creates a PostgreSQL cluster instance.we will create a tls secure instance since were planing to replicated across cluster

Lets start with creating a secret first to access to database and we will deploy a tls secured instance since were replication across cluster

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=postgres/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls pg-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pg-ca created
```

Now, we are going to create an `Issuer` using the `pg-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pg-issuer
  namespace: demo
spec:
  ca:
    secretName: pg-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-issuer.yaml
issuer.cert-manager.io/pg-issuer created
```


### Create Auth Secret

```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cG9zdGdyZXM=
kind: Secret
metadata:
  name: pg-singapore-auth
  namespace: demo
type: kubernetes.io/basic-auth
```
```bash 
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-singapore-auth.yaml
secret/pg-singapore-auth created
```

## Deploy PostgreSQL server
## Deploy PostgreSQL with TLS/SSL configuration
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: pg-singapore
  namespace: demo
spec:
  authSecret:
    name: pg-singapore-auth
  allowedSchemas:
    namespaces:
      from: Same
  autoOps: {}
  clientAuthMode: md5
  replicas: 3
  sslMode: verify-ca
  standbyMode: Hot
  streamingMode: Synchronous
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pg-issuer
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
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: linode-block-storage
  storageType: Durable
  terminationPolicy: WipeOut
  version: "15.3"
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-singapore.yaml
postgres.kubedb.com/pg-singapore created
```
KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created

```bash
$ kubectl get pg -n demo
NAME              VERSION   STATUS   AGE
pg-singapore      15.3      Ready    22h
```

# Exposing to outside world
For Now we will expose our postgresql with ingress with to outside world
```bash
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
$ helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx  \
                                      --namespace demo --create-namespace \
                                      --set tcp.5432="demo/pg-singapore:5432"
```
Let's apply the ingress yaml thats refers to `pg-singpore` service

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: pg-singapore
  namespace: demo  
spec:
  ingressClassName: nginx
  rules:
  - host: pg-singapore.something.org
    http:
      paths:
      - backend:
          service:
            name: pg-singapore
            port:
              number: 5432
        path: /
        pathType: Prefix
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-ingress.yaml
ingress.networking.k8s.io/pg-singapore created
$ kubectl get ingress -n demo
NAME              CLASS   HOSTS                           ADDRESS          PORTS   AGE
pg-singapore      nginx   pg-singapore.something.org      172.104.37.147   80      22h
```

# Prepare for Remote Replica
We wil use the [kubedb_plugin](somelink) for generating configuration for remote replica. It will create the appbinding and and necessary secrets to connect with source server
```bash
$ kubectl dba remote-config postgres -n demo pg-singapore -uremote -ppass -d 172.104.37.147 -y
home/mehedi/go/src/kubedb.dev/yamls/mysql/pg-singapore-remote-config.yaml
```

#  Create  Remote Replica
We have prepared another cluster in london region for replicating across cluster. follow the installation instruction [above](/docs/README.md).

### create sourceRef 

We will apply the generated config from kubeDB plugin to create the source refs and secrets for it
```bash
$ kubectl apply -f  /home/mehedi/go/src/kubedb.dev/yamls/pg-singapore-remote-config.yaml
secret/pg-singapore-remote-replica-auth created
secret/pg-singapore-client-cert-remote created
appbinding.appcatalog.appscode.com/pg-singapore created
```

### create remote replica auth
we will need to use the same auth secrets for remote replicas as well since operations like clone also replicated the auth-secrets from source server

```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cG9zdGdyZXM=
kind: Secret
metadata:
  name: pg-london-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls/pg-london-auth.yaml
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: pg-london
  namespace: demo
spec:
  remoteReplica:
    sourceRef:
      name: pg-singapore
      namespace: demo
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
    disableWriteCheck: true
  authSecret:
    name: pg-london-auth
  clientAuthMode: md5
  standbyMode: Hot
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: linode-block-storage
  storageType: Durable
  terminationPolicy: WipeOut
  version: "15.3"
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/pg-london.yaml
mysql.kubedb.com/pg-london created
```

Now we will be able to see kubedb will provision a Remote Replica from the source mysql instance. Lets checkout out the statefulSet , pvc , pv and services associated with it
.
KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified `MySQL` object:
```bash
$ kubectl get pg -n demo 
NAME           VERSION   STATUS   AGE
pg-london      15.3      Ready    7m17s
```

##  Validate Remote Replica

