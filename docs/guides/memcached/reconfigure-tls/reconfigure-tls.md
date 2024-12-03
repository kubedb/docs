---
title: Reconfigure Memcached TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: mc-reconfigure-tls
    name: Reconfigure TLS
    parent: reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Memcached TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. `add, remove, update and rotation of TLS/SSL certificates` for existing Memcached database via a `MemcachedOpsRequest`. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

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

In this section, we are going to deploy a `Memcached` database without TLS. In the next few sections we will add reconfigure TLS using `MemcachedOpsRequest` CRD. Below is the YAML of the `Memcached` CR that we are going to create,

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

Now, we can connect to this database through `telnet` to verify that the `TLS` is disabled.

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

### Create Issuer/ ClusterIssuer

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
Memcachedopsrequest.ops.kubedb.com/mc-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CRO,

```bash
$ kubectl get Memcachedopsrequest -n demo
Every 2.0s: kubectl get Memcachedopsrequest -n demo
NAME             TYPE             STATUS       AGE
mc-add-tls       ReconfigureTLS   Successful   79s
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded. If we describe the `MemcachedOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mcops -n demo mc-add-tls 
Name:         mc-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MemcachedOpsRequest
Metadata:
  Creation Timestamp:  2024-11-15T11:10:37Z
  Generation:          1
  Resource Version:    1782138
  UID:                 25123c6c-90e1-4a11-a060-42a1f75bc15d
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  memcd-quickstart
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          memcached
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       memcached-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-15T11:10:37Z
    Message:               Memcached ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-15T11:10:50Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-15T11:10:45Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-15T11:10:45Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-15T11:10:45Z
    Message:               check issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckIssuingCondition
    Last Transition Time:  2024-11-15T11:11:10Z
    Message:               Successfully restarted pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-15T11:10:55Z
    Message:               evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--memcd-quickstart-0
    Last Transition Time:  2024-11-15T11:10:55Z
    Message:               is pod ready; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsPodReady
    Last Transition Time:  2024-11-15T11:11:00Z
    Message:               is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--memcd-quickstart-0
    Last Transition Time:  2024-11-15T11:11:00Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>

```

Now, let's describe the client.crt of running Memcached database.

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

Now, we are going to rotate the certificate of this database. First let’s check the current expiration date of the certificate:
```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211

$ openssl x509 -in <(openssl s_client -connect 127.0.0.1:11211 -showcerts < /dev/null 2>/dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p') -noout -enddate
notAfter=Feb 16 04:58:37 2025 GMT
```
So, the certificate will expire on Feb 16 04:58:37 2025 GMT.

### Create MemcachedOpsRequest

Now we are going to rotate certificates using a MemcachedOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mc-ops-rotate
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

We can see from the above output that the `MemcachedOpsRequest` has succeeded. If we describe the `MemcachedOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mcops -n demo mc-ops-rotate
Name:         mc-ops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MemcachedOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T06:14:21Z
  Generation:          1
  Resource Version:    1802316
  UID:                 0c54644b-3006-4c3d-8c12-4566ad73a7eb
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  memcd-quickstart
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T06:14:21Z
    Message:               Memcached ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T06:14:24Z
    Message:               successfully add issuing condition to all the certificates
    Observed Generation:   1
    Reason:                IssueCertificatesSucceeded
    Status:                True
    Type:                  IssueCertificatesSucceeded
    Last Transition Time:  2024-11-18T06:14:35Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-18T06:14:29Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-18T06:14:29Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-18T06:14:29Z
    Message:               check issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckIssuingCondition
    Last Transition Time:  2024-11-18T06:14:55Z
    Message:               Successfully restarted pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-18T06:14:40Z
    Message:               evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--memcd-quickstart-0
    Last Transition Time:  2024-11-18T06:14:40Z
    Message:               is pod ready; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsPodReady
    Last Transition Time:  2024-11-18T06:14:45Z
    Message:               is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--memcd-quickstart-0
    Last Transition Time:  2024-11-18T06:14:45Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>

```

Now, let’s check the expiration date of the certificate:
```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211

$ openssl x509 -in <(openssl s_client -connect 127.0.0.1:11211 -showcerts < /dev/null 2>/dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p') -noout -enddate
notAfter=Feb 16 06:46:16 2025 GMT
```
As we can see from the above output, the certificate has been rotated successfully as the expire time got updated.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/mc-new-issuer.yaml
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

We can see from the above output that the `MemcachedlOpsRequest` has succeeded. If we describe the `MemcachedOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mcops -n demo mc-change-issuer
Name:         mc-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MemcachedOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T11:26:45Z
  Generation:          1
  Resource Version:    1830164
  UID:                 9d1e3477-7b22-4feb-8e32-97cd33c8b312
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  memcd-quickstart
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       my-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T11:26:45Z
    Message:               Memcached ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T11:26:58Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-18T11:26:53Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-18T11:26:53Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-18T11:26:53Z
    Message:               check issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckIssuingCondition
    Last Transition Time:  2024-11-18T11:27:18Z
    Message:               Successfully restarted pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-18T11:27:03Z
    Message:               evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--memcd-quickstart-0
    Last Transition Time:  2024-11-18T11:27:03Z
    Message:               is pod ready; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsPodReady
    Last Transition Time:  2024-11-18T11:27:08Z
    Message:               is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--memcd-quickstart-0
    Last Transition Time:  2024-11-18T11:27:08Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age   From                         Message
  ----     ------                                                          ----  ----                         -------
  Normal   PauseDatabase                                                   16m   KubeDB Ops-manager Operator  Pausing Memcached demo/memcd-quickstart
  Warning  get certificate; ConditionStatus:True                           16m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                     16m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True                   16m   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                           16m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                     16m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True                   16m   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                           16m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                     16m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True                   16m   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                               16m   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                           16m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                     16m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True                   16m   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                           16m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                     16m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True                   16m   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                           16m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                     16m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True                   16m   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                               16m   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0     16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
  Warning  is pod ready; ConditionStatus:False                             16m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False
  Warning  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0  16m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
  Normal   RestartPods                                                     16m   KubeDB Ops-manager Operator  Successfully restarted pods
  Normal   ResumeDatabase                                                  16m   KubeDB Ops-manager Operator  Resuming Memcached demo/memcd-quickstart
  Normal   ResumeDatabase                                                  16m   KubeDB Ops-manager Operator  Successfully resumed Memcached demo/memcd-quickstart
  Normal   Successful                                                      16m   KubeDB Ops-manager Operator  Successfully Reconfigured TLS
  Normal   PauseDatabase                                                   16m   KubeDB Ops-manager Operator  Pausing Memcached demo/memcd-quickstart
  Warning  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0     15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
  Warning  is pod ready; ConditionStatus:False                             15m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False
  Warning  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0  15m   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
  Normal   RestartPods                                                     15m   KubeDB Ops-manager Operator  Successfully restarted pods
```

Now, let’s port-forward the database pod and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211

$ openssl x509 -in <(openssl s_client -connect 127.0.0.1:11211 -showcerts < /dev/null 2>/dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p') -inform PEM -issuer -nameopt RFC2253 -noout
issuer=O=kubedb-updated,CN=memcached-updated
```
We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a MemcachedOpsRequest.

### Create MemcachedOpsRequest

Below is the YAML of the `MemcachedOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: mc-ops-tls-remove
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure-tls/mc-ops-tls-remove.yaml
Memcachedopsrequest.ops.kubedb.com/mc-ops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcachedOpsRequest` CRO,

```bash
$ kubectl get memcachedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME                TYPE             STATUS        AGE
mc-ops-tls-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded. If we describe the `MemcachedOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mcops -n demo mc-ops-tls-remove
Name:         mc-ops-tls-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MemcachedOpsRequest
Metadata:
  Creation Timestamp:  2024-11-12T12:49:09Z
  Generation:          1
  Resource Version:    1684823
  UID:                 c3260cc6-7862-4f22-9e12-93dcdb3edac8
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  memcd-quickstart
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-12T12:49:09Z
    Message:               Memcached ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-12T12:49:32Z
    Message:               Successfully restarted pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-11-12T12:49:17Z
    Message:               evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--memcd-quickstart-0
    Last Transition Time:  2024-11-12T12:49:17Z
    Message:               is pod ready; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsPodReady
    Last Transition Time:  2024-11-12T12:49:22Z
    Message:               is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--memcd-quickstart-0
    Last Transition Time:  2024-11-12T12:49:32Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```
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
