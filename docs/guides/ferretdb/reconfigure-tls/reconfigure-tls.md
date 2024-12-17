---
title: Reconfigure FerretDB TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: fr-reconfigure-tls-rs
    name: Reconfigure FerretDB TLS/SSL Encryption
    parent: fr-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure FerretDB TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing FerretDB database via a FerretDBOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

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

## Add TLS to a FerretDB

Here, We are going to create a FerretDB database without TLS and then reconfigure the ferretdb to use TLS.

### Deploy FerretDB without TLS

In this section, we are going to deploy a FerretDB without TLS. In the next few sections we will reconfigure TLS using `FerretDBOpsRequest` CRD. Below is the YAML of the `FerretDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb-x
  namespace: demo
spec:
  version: "1.18.0"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  backend:
    externallyManaged: false
  replicas: 2
```

Let's create the `FerretDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/reconfigure-tls/ferretdb.yaml
ferretdb.kubedb.com/ferretdb created
```

Now, wait until `ferretdb` has status `Ready`. i.e,

```bash
$ kubectl get fr -n demo
NAME       NAMESPACE   VERSION   STATUS   AGE
ferretdb   demo        1.18.0    Ready    75s

$ kubectl dba describe ferretdb ferretdb -n demo
Name:         ferretdb
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         FerretDB
Metadata:
  Creation Timestamp:  2024-10-17T11:04:08Z
  Finalizers:
    kubedb.com
  Generation:        4
  Resource Version:  158199
  UID:               7da85335-bac0-4247-ad69-85a7c44831df
Spec:
  Auth Secret:
    Name:  ferretdb-auth
  Backend:
    Externally Managed:  false
    Linked DB:           ferretdb
    Postgres Ref:
      Name:         ferretdb-pg-backend
      Namespace:    demo
    Version:        13.13
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  ferretdb
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1000
          Run As Non Root:  true
          Run As User:      1000
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:  1000
  Replicas:        2
  Ssl Mode:        disabled
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:  500Mi
  Storage Type:   Durable
  Version:        1.18.0
Status:
  Conditions:
    Last Transition Time:  2024-10-17T11:04:08Z
    Message:               The KubeDB operator has started the provisioning of FerretDB: demo/ferretdb
    Observed Generation:   2
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-10-17T11:05:04Z
    Message:               All replicas are ready for FerretDB demo/ferretdb
    Observed Generation:   4
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-10-17T11:05:14Z
    Message:               The FerretDB: demo/ferretdb is accepting client requests.
    Observed Generation:   4
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-10-17T11:05:14Z
    Message:               The FerretDB: demo/ferretdb is ready.
    Observed Generation:   4
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-10-17T11:05:14Z
    Message:               The FerretDB: demo/ferretdb is successfully provisioned.
    Observed Generation:   4
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in FerretDB. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls ferretdb-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/ferretdb-ca created
```

Now, Let's create an `Issuer` using the `ferretdb-ca` secret that we have just created. The `YAML` file looks like this:

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

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/ferretdb-issuer created
```

### Create FerretDBOpsRequest

In order to add TLS to the ferretdb, we have to create a `FerretDBOpsRequest` CRO with our created issuer. Below is the YAML of the `FerretDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: frops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ferretdb
  tls:
    issuerRef:
      name: ferretdb-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mg-rs` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/ferretdb/concepts/ferretdb.md#spectls).
- `spec.tls.sslMode` is the ssl mode of the server. You can see the details [here](/docs/guides/ferretdb/concepts/ferretdb.md#specsslmode).
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/ferretdb/concepts/opsrequest.md#spectimeout)

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/reconfigure-tls/frops-add-tls.yaml
ferretdbopsrequest.ops.kubedb.com/frops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CRO,

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME            TYPE             STATUS       AGE
frops-add-tls   ReconfigureTLS   Successful   13m
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe ferretdbopsrequest -n demo frops-add-tls
Name:         frops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2024-10-17T11:15:12Z
  Generation:          1
  Resource Version:    159329
  UID:                 071189ab-275f-4a25-99b9-72da3fa2fb6a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   ferretdb
  Timeout:  5m
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       ferretdb-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-17T11:15:12Z
    Message:               FerretDB ops-request has started to reconfigure tls for FerretDB nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-17T11:15:15Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-17T11:15:20Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-10-17T11:15:20Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2024-10-17T11:15:20Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-10-17T11:15:20Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-10-17T11:15:25Z
    Message:               successfully reconciled the FerretDB with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-17T11:15:30Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-0
    Last Transition Time:  2024-10-17T11:15:31Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-0
    Last Transition Time:  2024-10-17T11:15:35Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-0
    Last Transition Time:  2024-10-17T11:15:40Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-1
    Last Transition Time:  2024-10-17T11:15:41Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-1
    Last Transition Time:  2024-10-17T11:15:45Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-1
    Last Transition Time:  2024-10-17T11:15:50Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-17T11:15:51Z
    Message:               Successfully completed the ReconfigureTLS for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     13m   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/frops-add-tls
  Normal   Starting                                                     13m   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/ferretdb
  Normal   Successful                                                   13m   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-add-tls
  Warning  get certificate; ConditionStatus:True                        13m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                        13m   KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                      13m   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                        13m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                        13m   KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                      13m   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                            13m   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                13m   KubeDB Ops-manager Operator  successfully reconciled the FerretDB with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-0            13m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-0          13m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-0  13m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-0
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-1            13m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-1          13m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-1  13m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-1
  Normal   RestartNodes                                                 13m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                     13m   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/ferretdb
  Normal   Successful                                                   13m   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-add-tls
```

Now let's connect with this ferretdb with certs. We need save the client cert and key to two different files and make a pem file.
Additionally, to verify server, we need to store ca.crt.

```bash
$ kubectl get secrets -n demo ferretdb-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > client.crt
$ kubectl get secrets -n demo ferretdb-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > client.key
$ kubectl get secrets -n demo ferretdb-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
$ cat client.crt client.key > client.pem
```

Now, we can connect to our FerretDB with these files with mongosh client.

```bash
$ kubectl get secrets -n demo ferretdb-auth -o jsonpath='{.data.\username}' | base64 -d
postgres
$ kubectl get secrets -n demo ferretdb-auth -o jsonpath='{.data.\\password}' | base64 -d
l*jGp8u*El8WRSDJ

$ kubectl port-forward svc/ferretdb -n demo 27017
Forwarding from 127.0.0.1:27017 -> 27018
Forwarding from [::1]:27017 -> 27018
Handling connection for 27017
Handling connection for 27017
```

Now in another terminal

```bash
$ mongosh 'mongodb://postgres:l*jGp8u*El8WRSDJ@localhost:27017/ferretdb?authMechanism=PLAIN&tls=true&tlsCertificateKeyFile=./client.pem&tlsCaFile=./ca.crt'
Current Mongosh Log ID:	65efeea2a3347fff66d04c70
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.1.5
Using MongoDB:		7.0.42
Using Mongosh:		2.1.5

For mongosh info see: https://docs.mongodb.com/mongodb-shell/

------
   The server generated these startup warnings when booting
   2024-03-12T05:56:50.979Z: Powered by FerretDB v1.23.0 and PostgreSQL 13.13 on x86_64-pc-linux-musl, compiled by gcc.
   2024-03-12T05:56:50.979Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2024-03-12T05:56:50.979Z: The telemetry state is undecided.
   2024-03-12T05:56:50.979Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb>
```
So, here we have connected using the client certificate and the connection is tls secured. So, we can safely assume that tls enabling was successful.

## Rotate Certificate

Now we are going to rotate the certificate of this database. First we can store the current expiration date of the certificate by exec into `ferretdb-0` pod. Certs are located in `/etc/certs/server/` path.

### Create FerretDBOpsRequest

Now we are going to increase it using a FerretDBOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: frops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ferretdb
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ferretdb`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our ferretdb.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this ferretdb.

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/reconfigure-tls/frops-rotate.yaml
ferretdbopsrequest.ops.kubedb.com/frops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CRO,

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME           TYPE             STATUS       AGE
frops-rotate   ReconfigureTLS   Successful   113s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe ferretdbopsrequest -n demo frops-rotate
Name:         frops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2024-10-17T11:37:29Z
  Generation:          1
  Resource Version:    161772
  UID:                 6d9acf23-2701-40f9-9187-da221f3e4158
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ferretdb
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-17T11:37:29Z
    Message:               FerretDB ops-request has started to reconfigure tls for FerretDB nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-17T11:37:32Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-17T11:37:38Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-10-17T11:37:38Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2024-10-17T11:37:38Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-10-17T11:37:38Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-10-17T11:37:43Z
    Message:               successfully reconciled the FerretDB with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-17T11:37:48Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-0
    Last Transition Time:  2024-10-17T11:37:48Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-0
    Last Transition Time:  2024-10-17T11:37:53Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-0
    Last Transition Time:  2024-10-17T11:37:58Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-1
    Last Transition Time:  2024-10-17T11:37:58Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-1
    Last Transition Time:  2024-10-17T11:38:03Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-1
    Last Transition Time:  2024-10-17T11:38:08Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-17T11:38:08Z
    Message:               Successfully completed the ReconfigureTLS for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     55s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/frops-rotate
  Normal   Starting                                                     55s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/ferretdb
  Normal   Successful                                                   55s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-rotate
  Warning  get certificate; ConditionStatus:True                        46s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                        46s   KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                      46s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                        46s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                        46s   KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                      46s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                            46s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                41s   KubeDB Ops-manager Operator  successfully reconciled the FerretDB with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-0            36s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-0          36s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-0  31s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-0
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-1            26s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-1          26s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-1  21s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-1
  Normal   RestartNodes                                                 16s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                     16s   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/ferretdb
  Normal   Successful                                                   16s   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-rotate
```

As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubedb-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls ferretdb-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/ferretdb-new-ca created
```

Now, Let's create a new `Issuer` using the `ferretdb-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: fr-new-issuer
  namespace: demo
spec:
  ca:
    secretName: ferretdb-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/fr-new-issuer created
```

### Create FerretDBOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `FerretDBOpsRequest` CRO with the newly created issuer. Below is the YAML of the `FerretDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: frops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ferretdb
  tls:
    issuerRef:
      name: fr-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ferretdb`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our ferretdb.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/reconfigure-tls/frops-change-issuer.yaml
ferretdbopsrequest.ops.kubedb.com/frops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CRO,

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME                  TYPE             STATUS       AGE
frops-change-issuer   ReconfigureTLS   Successful   87s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe ferretdbopsrequest -n demo frops-change-issuer
Name:         frops-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2024-10-18T10:14:38Z
  Generation:          1
  Resource Version:    423126
  UID:                 1bf730e8-603e-4f30-b9ab-5a4e75d3a4d4
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ferretdb
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       fr-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-18T10:14:38Z
    Message:               FerretDB ops-request has started to reconfigure tls for FerretDB nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-18T10:14:41Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-18T10:14:46Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-10-18T10:14:46Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2024-10-18T10:14:46Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-10-18T10:14:46Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-10-18T10:14:51Z
    Message:               successfully reconciled the FerretDB with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-18T10:14:56Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-0
    Last Transition Time:  2024-10-18T10:14:56Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-0
    Last Transition Time:  2024-10-18T10:15:01Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-0
    Last Transition Time:  2024-10-18T10:15:06Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-1
    Last Transition Time:  2024-10-18T10:15:06Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-1
    Last Transition Time:  2024-10-18T10:15:11Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-1
    Last Transition Time:  2024-10-18T10:15:16Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-18T10:15:16Z
    Message:               Successfully completed the ReconfigureTLS for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     88s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/frops-change-issuer
  Normal   Starting                                                     88s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/ferretdb
  Normal   Successful                                                   88s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-change-issuer
  Warning  get certificate; ConditionStatus:True                        80s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                        80s   KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                      80s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                        80s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                        80s   KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                      80s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                            80s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                75s   KubeDB Ops-manager Operator  successfully reconciled the FerretDB with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-0            70s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-0          70s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-0  65s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-0
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-1            60s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-1          60s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-1  55s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-1
  Normal   RestartNodes                                                 50s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                     50s   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/ferretdb
  Normal   Successful                                                   50s   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-change-issuer
```

Now, If exec ferretdb and find out the ca subject in `/etc/certs/server` location, we can see that the CN and O is updated according to out new ca.crt.

We can see that subject name of this ca.crt matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the ferretdb

Now, we are going to remove TLS from this ferretdb using a FerretDBOpsRequest.

### Create FerretDBOpsRequest

Below is the YAML of the `FerretDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: frops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ferretdb
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ferretdb`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our ferretdb.
- `spec.tls.remove` specifies that we want to remove tls from this ferretdb.

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/reconfigure-tls/frops-remove.yaml
ferretdbopsrequest.ops.kubedb.com/frops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CRO,

```bash
$ wacth kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME           TYPE             STATUS       AGE
frops-remove   ReconfigureTLS   Successful   65s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe ferretdbopsrequest -n demo frops-remove
Name:         frops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2024-10-18T11:11:55Z
  Generation:          1
  Resource Version:    428244
  UID:                 28a6ba72-0a2d-47f1-97b0-1e9609845acc
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ferretdb
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-18T11:11:55Z
    Message:               FerretDB ops-request has started to reconfigure tls for FerretDB nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-18T11:11:58Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-18T11:12:04Z
    Message:               successfully reconciled the FerretDB with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-18T11:12:09Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-0
    Last Transition Time:  2024-10-18T11:12:09Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-0
    Last Transition Time:  2024-10-18T11:12:14Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-0
    Last Transition Time:  2024-10-18T11:12:19Z
    Message:               get pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ferretdb-1
    Last Transition Time:  2024-10-18T11:12:19Z
    Message:               evict pod; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ferretdb-1
    Last Transition Time:  2024-10-18T11:12:24Z
    Message:               check pod running; ConditionStatus:True; PodName:ferretdb-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ferretdb-1
    Last Transition Time:  2024-10-18T11:12:29Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-18T11:12:29Z
    Message:               Successfully completed the ReconfigureTLS for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     87s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/frops-remove
  Normal   Starting                                                     87s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/ferretdb
  Normal   Successful                                                   87s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-remove
  Normal   UpdatePetSets                                                78s   KubeDB Ops-manager Operator  successfully reconciled the FerretDB with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-0            73s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-0          73s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-0
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-0  68s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-0
  Warning  get pod; ConditionStatus:True; PodName:ferretdb-1            63s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  evict pod; ConditionStatus:True; PodName:ferretdb-1          63s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:ferretdb-1
  Warning  check pod running; ConditionStatus:True; PodName:ferretdb-1  58s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:ferretdb-1
  Normal   RestartNodes                                                 53s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                     53s   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/ferretdb
  Normal   Successful                                                   53s   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/ferretdb for FerretDBOpsRequest: frops-remove
```

Now, Let's try to connect with ferretdb without TLS certs.

```bash
$ kubectl get secrets -n demo ferret-auth -o jsonpath='{.data.\username}' | base64 -d
postgres
$ kubectl get secrets -n demo ferret-auth -o jsonpath='{.data.\\password}' | base64 -d
l*jGp8u*El8WRSDJ

$ kubectl port-forward svc/ferret -n demo 27017
Forwarding from 127.0.0.1:27017 -> 27017
Forwarding from [::1]:27017 -> 27017
Handling connection for 27017
Handling connection for 27017
```

Now in another terminal

```bash
$ mongosh 'mongodb://postgres:l*jGp8u*El8WRSDJ@localhost:27017/ferretdb?authMechanism=PLAIN'
Current Mongosh Log ID:	65efeea2a3347fff66d04c70
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?authMechanism=PLAIN&directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.1.5
Using MongoDB:		7.0.42
Using Mongosh:		2.1.5

For mongosh info see: https://docs.mongodb.com/mongodb-shell/

------
   The server generated these startup warnings when booting
   2024-03-12T05:56:50.979Z: Powered by FerretDB v1.18.0 and PostgreSQL 13.13 on x86_64-pc-linux-musl, compiled by gcc.
   2024-03-12T05:56:50.979Z: Please star us on GitHub: https://github.com/FerretDB/FerretDB.
   2024-03-12T05:56:50.979Z: The telemetry state is undecided.
   2024-03-12T05:56:50.979Z: Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.
------

ferretdb>
```

We can see that we can now connect without providing TLS certs. So TLS connection is successfully disabled

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ferretdb -n demo ferretdb
kubectl delete issuer -n demo ferretdb-issuer fr-new-issuer
kubectl delete ferretdbopsrequest -n demo frops-add-tls frops-remove frops-rotate frops-change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ferretdb/monitoring/using-prometheus-operator.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ferretdb/monitoring/using-builtin-prometheus.md).
- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
