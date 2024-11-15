---
title: Reconfigure Memcached TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: mc-reconfigure-tls-standalone
    name: Memcahced Reconfigure TLS
    parent: mc-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Memcached TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Memcached database via a MemcachedOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/memcached](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/memcached) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Memcached database

In this tutorial, we are going to reconfigure TLS of Memcached.
Here, We are going to create a Memcached database without TLS and then reconfigure the database to use TLS.

### Deploy Memcached without TLS

In this section, we are going to deploy a Memcached database without TLS. In the next few sections we will add reconfigure TLS using `MemcachedOpsRequest` CRD. Below is the YAML of the `Memcached` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  deletionPolicy: WipeOut
```

Let's create the `Memcached` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/memcached.yaml
memcached.kubedb.com/memcd-quickstart created
```

Now, wait until `memcd-quickstart` has status `Ready`. i.e,

```bash
$ watch kubectl get mc -n demo
Every 2.0s: kubectl get mc -n demo
NAME               VERSION   STATUS   AGE
memcd-quickstart   1.6.22    Ready    26s
```

Now, we can connect to this database through memcached-cli verify that the TLS is disabled.

```bash
$ kc port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
Handling connection for 11211

$ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.

# Authentication
set key 0 0 21
user ukwcbtebbrwastqg
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
ssl_enabled no
ssl_chain_cert (null)
ssl_key (null)
ssl_ca_cert NULL
...
END

quit
```

We can verify from the above output that TLS is disabled for this database.

### Create Issuer/ StandaloneIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Memcached. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=memcached/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now, we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls memcached-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/memcached-ca created
```

Now, Let's create an `Issuer` using the `memcached-ca` secret that we have just created. The `YAML` file looks like this:

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

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/memcached-ca-issuer created
```

### Create MemcachedOpsRequest

In order to add TLS to the database, we have to create a `MemcachedOpsRequest` CRO with our created issuer. Below is the YAML of the `MemcachedOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mc-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: memcd-quickstart
  tls:
    issuerRef:
      name: memcached-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - memcached
          organizationalUnits:
            - client
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `memcd-quickstart` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and API group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/memcached/concepts/memcached.md#spectls).

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/mc-add-tls.yaml
Memcachedopsrequest.ops.kubedb.com/rd-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CRO,

```bash
$ kubectl get Memcachedopsrequest -n demo
Every 2.0s: kubectl get Memcachedopsrequest -n demo
NAME             TYPE             STATUS       AGE
mc-add-tls       ReconfigureTLS   Successful   79s
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded. 

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
              cert-manager.io/ip-sans: 
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: memcached-ca-issuer
              cert-manager.io/subject-organizationalunits: client
              cert-manager.io/subject-organizations: memcached
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:            1159 bytes
tls-combined.pem:  2868 bytes
tls.crt:           1188 bytes
tls.key:           1679 bytes
```

Now, we can connect using tls-certs to connect to the Memcached and write some data

```bash
```bash
$ kc port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
Handling connection for 11211

$ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.

# Authentication
set key 0 0 21
user ukwcbtebbrwastqg
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

## Rotate Certificate

Now, we are going to rotate the certificate of this database.

### Create MemcachedOpsRequest

Now we are going to rotate certificates using a MemcachedOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: myops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: memcd-quickstart
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `memcd-quickstart` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/mc-ops-rotate.yaml
memcachedopsrequest.ops.kubedb.com/mc-ops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CRO,

```bash
$ watch kubectl get memcachedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME             TYPE             STATUS        AGE
mc-ops-rotate    ReconfigureTLS   Successful    5m5s
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=memcached-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=memcached-updated/O=kubedb-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls memcached-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/memcached-new-ca created
```

Now, Let's create a new `Issuer` using the `memcached-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mc-new-issuer
  namespace: demo
spec:
  ca:
    secretName: memcached-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/mc-new-issuer created
```

### Create MemcachedOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `MemcachedOpsRequest` CRO with the newly created issuer. Below is the YAML of the `MemcachedOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mc-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: memcd-quickstart
  tls:
    issuerRef:
      name: mc-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `memcd-quickstart` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/mc-change-issuer.yaml
Memcachedopsrequest.ops.kubedb.com/mc-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CRO,

```bash
$ kubectl get memcachedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME                  TYPE             STATUS        AGE
mc-change-issuer      ReconfigureTLS   Successful    4m65s
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded. 

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a MemcachedOpsRequest.

### Create MemcachedOpsRequest

Below is the YAML of the `MemcachedOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mc-ops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: memcd-quickstart
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `memcd-quickstart` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/mc-ops-remove.yaml
Memcachedopsrequest.ops.kubedb.com/mc-ops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CRO,

```bash
$ kubectl get memcachedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME            TYPE             STATUS        AGE
mc-ops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded. 

Now, Lets check Memcached TLS is disabled or not.

```bash
$ kc port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
Handling connection for 11211

$ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.

# Authentication
set key 0 0 21
user ukwcbtebbrwastqg
STORED

# Current Stats Settings
stats settings
...
ssl_enabled no
ssl_chain_cert (null)
ssl_key (null)
ssl_ca_cert NULL
...
END

quit
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo memcached/memcd-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
memcached.kubedb.com/memcd-quickstart patched

$ kubectl delete memcached -n demo memcd-quickstart
memcached.kubedb.com/memcd-quickstart deleted

$ kubectl delete issuer -n demo memcached-ca-issuer mc-new-issuer
issuer.cert-manager.io "memcached-ca-issuer" deleted
issuer.cert-manager.io "mc-new-issuer" deleted

$ kubectl delete memcachedopsrequest -n demo mc-add-tls mc-ops-remove mc-ops-rotate mc-change-issuer
memcachedopsrequest.ops.kubedb.com "mc-add-tls" deleted
memcachedopsrequest.ops.kubedb.com "mc-ops-remove" deleted
memcachedopsrequest.ops.kubedb.com "mc-ops-rotate" deleted
memcachedopsrequest.ops.kubedb.com "mc-change-issuer" deleted
```

## Next Steps

- Detail concepts of [Memcached](/docs/guides/memcached/concepts/memcached.md).
- Monitor your Memcached database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/memcached/monitoring/using-prometheus-operator.md).
- Monitor your Memcached database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md).
