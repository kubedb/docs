---
title: Reconfigure Standalone MSSQLServer TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: ms-reconfigure-tls-standalone
    name: Standalone
    parent: ms-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Standalone MSSQLServer TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing MSSQLServer database via a MSSQLServerOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a MSSQLServer database

Here, We are going to create a MSSQLServer database without TLS and then reconfigure the database to use TLS.

At first, we need to create an Issuer/ClusterIssuer which will be used to generate the certificate used for TLS configurations.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
-
- Create a secret using the certificate files we have just generated,
```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```
Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

### Deploy MSSQLServer without TLS

In this section, we are going to deploy a MSSQLServer Standalone database without TLS. In the next few sections we will reconfigure TLS using `MSSQLServerOpsRequest` CRD. Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ms-standalone
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MSSQLServer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/ms-standalone.yaml
mssqlserver.kubedb.com/ms-standalone created
```

Now, wait until `ms-standalone` has status `Ready`. i.e,

```bash
$ kubectl get ms  -n demo
NAME            VERSION     STATUS   AGE
ms-standalone   2022-cu12   Ready    4m3s

$ kubectl describe ms -n demo ms-standalone
Name:         ms-standalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         MSSQLServer
Metadata:
  Creation Timestamp:  2024-11-14T12:45:36Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  438804
  UID:               83ebe191-3754-41af-8d86-ed211bf9c31c
Spec:
  Auth Secret:
    Name:           ms-standalone-auth
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
        Env:
          Name:   ACCEPT_EULA
          Value:  Y
          Name:   MSSQL_PID
          Value:  Evaluation
        Name:     mssql
        Resources:
          Limits:
            Memory:  4Gi
          Requests:
            Cpu:     500m
            Memory:  4Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Add:
              NET_BIND_SERVICE
            Drop:
              ALL
          Run As Group:     10001
          Run As Non Root:  true
          Run As User:      10001
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  mssql-init
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  512Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     10001
          Run As Non Root:  true
          Run As User:      10001
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:  10001
  Replicas:        1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Tls:
    Certificates:
      Alias:        server
      Secret Name:  ms-standalone-server-cert
      Subject:
        Organizational Units:
          server
        Organizations:
          kubedb
      Alias:        client
      Secret Name:  ms-standalone-client-cert
      Subject:
        Organizational Units:
          client
        Organizations:
          kubedb
    Client TLS:  false
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mssqlserver-ca-issuer
  Version:        2022-cu12
Status:
  Conditions:
    Last Transition Time:  2024-11-14T12:45:36Z
    Message:               The KubeDB operator has started the provisioning of MSSQLServer: demo/ms-standalone
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-11-14T12:46:36Z
    Message:               All replicas are ready for MSSQLServer demo/ms-standalone
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-11-14T12:46:46Z
    Message:               database demo/ms-standalone is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-11-14T12:46:46Z
    Message:               database demo/ms-standalone is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-11-14T12:47:06Z
    Message:               The MSSQLServer: demo/ms-standalone is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:
  Type    Reason      Age    From                         Message
  ----    ------      ----   ----                         -------
  Normal  Successful  4m20s  KubeDB Ops-manager Operator  Successfully created MSSQLServer server certificates
  Normal  Successful  4m20s  KubeDB Ops-manager Operator  Successfully created MSSQLServer client certificates

```

Now, connect to this database by exec into a pod and verify the TLS is disabled. 

> when we connect using the sqlcmd tool, the -N option is available with [s|m|o] parameters, where 's' stands for strict, 'm' for mandatory, and 'o' for optional. The default setting is mandatory.


```bash
$ kubectl get secrets -n demo ms-standalone-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secrets -n demo ms-standalone-auth -o jsonpath='{.data.\password}' | base64 -d
b1HLv9EV4CaSalX6

$ kubectl exec -it -n demo ms-standalone-0 -c mssql -- bash
mssql@ms-standalone-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "b1HLv9EV4CaSalX6" -N
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : SSL Provider: [error:0A000086:SSL routines::certificate verify failed:self-signed certificate].
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Client unable to establish connection.


So Now, we have to connect with -C [Trust Server Certificate]
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "b1HLv9EV4CaSalX6" -N -C
1> 
```

We can verify from the above output that TLS is disabled for this database, `mssql.conf` file has no tls configuration.

Now we will enable tls configuration using MSSQLServerOpsRequest
### Create MSSQLServerOpsRequest

In order to add TLS to the database, we have to create a `MSSQLServerOpsRequest` CRO with our issuer. Below is the YAML of the `MSSQLServerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ms-standalone
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - mssqlserver
          organizationalUnits:
            - client
    clientTLS: true
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ms-standalone` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/mssqlserver/concepts/mssqlserver.md#spectls).

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-add-tls.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ watch kubectl get msops -n demo
Every 2.0s: kubectl get msops -n demo

NAME            TYPE             STATUS       AGE
msops-add-tls   ReconfigureTLS   Successful   115s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-add-tls 
Name:         msops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-15T11:48:03Z
  Generation:          1
  Resource Version:    491162
  UID:                 007ad725-0a3f-4290-8814-d85592cfc247
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   ms-standalone
  Timeout:  5m
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          mssqlserver
    Client TLS:  true
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mssqlserver-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-15T11:48:03Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-15T11:48:06Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-15T11:48:16Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-15T11:48:11Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-15T11:48:11Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-15T11:48:11Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-15T11:48:21Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-15T11:49:06Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-15T11:48:26Z
    Message:               get pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ms-standalone-0
    Last Transition Time:  2024-11-15T11:48:26Z
    Message:               evict pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ms-standalone-0
    Last Transition Time:  2024-11-15T11:49:01Z
    Message:               check pod running; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ms-standalone-0
    Last Transition Time:  2024-11-15T11:49:07Z
    Message:               Successfully completed reconfigureTLS for mssqlserver.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```

Now, Let's exec into a database node



```bash
$ kubectl exec -it ms-standalone-0 -n demo -c mssql -- bash
mssql@ms-standalone-0:/$ ls /var/opt/mssql/tls
ca.crt	client.crt  client.key	server.crt  server.key
mssql@ms-standalone-0:/$ openssl x509 -in /var/opt/mssql/tls/client.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=mssql,OU=client,O=mssqlserver
mssql@ms-standalone-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
[network]
forceencryption = 1
tlscert = /var/opt/mssql/tls/server.crt
tlskey = /var/opt/mssql/tls/server.key
tlsprotocols = 1.2,1.1,1.0
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P l2mGQRMETAS96QRb -N
1> 
```

We can verify from the above output that TLS is enabled for this database, `mssql.conf` file has tls configurations.  So, TLS is enabled successfully to this database.


## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it ms-standalone-0 -n demo -c mssql -- bash
mssql@ms-standalone-0:/$ openssl x509 -in /var/opt/mssql/tls/client.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb 16 13:11:02 2025 GMT
mssql@ms-standalone-0:/$ 
```

So, the certificate will expire on this time `Feb 16 13:11:02 2025 GMT`.


### Create MSSQLServerOpsRequest

Now we are going to increase it using a MSSQLServerOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ms-standalone
  tls:
    rotateCertificates: true
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ms-standalone` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-rotate.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ kubectl get mssqlserveropsrequest -n demo
Every 2.0s: kubectl get mssqlserveropsrequest -n demo
NAME            TYPE             STATUS       AGE
msops-rotate    ReconfigureTLS   Successful   2m47s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-rotate
Name:         msops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T13:17:50Z
  Generation:          1
  Resource Version:    549743
  UID:                 af51934d-1fb4-4fa6-b254-46b1de199fae
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   ms-standalone
  Timeout:  5m
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T13:17:50Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T13:17:50Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-18T13:17:50Z
    Message:               successfully add issuing condition to all the certificates
    Observed Generation:   1
    Reason:                IssueCertificatesSucceeded
    Status:                True
    Type:                  IssueCertificatesSucceeded
    Last Transition Time:  2024-11-18T13:18:00Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-18T13:17:55Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-18T13:17:55Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-18T13:17:55Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-18T13:18:05Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-18T13:18:51Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-18T13:18:11Z
    Message:               get pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ms-standalone-0
    Last Transition Time:  2024-11-18T13:18:11Z
    Message:               evict pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ms-standalone-0
    Last Transition Time:  2024-11-18T13:18:46Z
    Message:               check pod running; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ms-standalone-0
    Last Transition Time:  2024-11-18T13:18:51Z
    Message:               Successfully completed reconfigureTLS for mssqlserver.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it ms-standalone-0 -n demo -c mssql -- bash
mssql@ms-standalone-0:/$ openssl x509 -in /var/opt/mssql/tls/client.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb 16 13:17:50 2025 GMT
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
$ kubectl create secret tls mssqlserver-new-ca \
           --cert=ca.crt \
           --key=ca.key \
           --namespace=demo
secret/mssqlserver-new-ca created
```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssqlserver-new-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: mssqlserver-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/mssqlserver-new-ca-issuer created
```

### Create MSSQLServerOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `MSSQLServerOpsRequest` CRO with the newly created issuer. Below is the YAML of the `MSSQLServerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ms-standalone
  tls:
    issuerRef:
      name: mssqlserver-new-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ms-standalone` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-change-issuer.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ kubectl get mssqlserveropsrequest -n demo
Every 2.0s: kubectl get mssqlserveropsrequest -n demo
NAME                  TYPE             STATUS       AGE
msops-change-issuer   ReconfigureTLS   Successful   3m28s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-change-issuer
Name:         msops-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T13:38:48Z
  Generation:          1
  Resource Version:    551920
  UID:                 551ce6a4-742a-43ed-a994-be4ba4809bca
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ms-standalone
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mssqlserver-new-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T13:38:48Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T13:38:51Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-18T13:39:01Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-18T13:38:56Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-18T13:38:56Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-18T13:38:56Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-18T13:39:06Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-18T13:42:11Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-18T13:39:11Z
    Message:               get pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ms-standalone-0
    Last Transition Time:  2024-11-18T13:39:11Z
    Message:               evict pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ms-standalone-0
    Last Transition Time:  2024-11-18T13:42:06Z
    Message:               check pod running; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ms-standalone-0
    Last Transition Time:  2024-11-18T13:42:11Z
    Message:               Successfully completed reconfigureTLS for mssqlserver.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```

Now, Lets exec into a database node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it ms-standalone-0 -n demo -c mssql -- bash
mssql@ms-standalone-0:/$  openssl x509 -in /var/opt/mssql/tls/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb-updated,CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a MSSQLServerOpsRequest.

### Create MSSQLServerOpsRequest

Below is the YAML of the `MSSQLServerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ms-standalone
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `ms-standalone` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-remove.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo
Every 2.0s: kubectl get mssqlserveropsrequest -n demo
NAME                  TYPE             STATUS       AGE
msops-remove          ReconfigureTLS   Successful   2m36s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-remove
Name:         msops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T13:49:15Z
  Generation:          1
  Resource Version:    552812
  UID:                 7e4b9c39-7fd2-44f6-9972-367c95198105
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ms-standalone
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T13:49:15Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T13:49:42Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-18T13:49:47Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-18T13:51:32Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-18T13:49:52Z
    Message:               get pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--ms-standalone-0
    Last Transition Time:  2024-11-18T13:49:52Z
    Message:               evict pod; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--ms-standalone-0
    Last Transition Time:  2024-11-18T13:51:27Z
    Message:               check pod running; ConditionStatus:True; PodName:ms-standalone-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--ms-standalone-0
    Last Transition Time:  2024-11-18T13:51:32Z
    Message:               Successfully completed reconfigureTLS for mssqlserver.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```


Now, Lets exec into the pod find out that TLS is disabled or not.
```bash
$ kubectl exec -it -n demo ms-standalone-0 -c mssql -- bash
mssql@ms-standalone-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P b1HLv9EV4CaSalX6 -N
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : SSL Provider: [error:0A000086:SSL routines::certificate verify failed:self-signed certificate].
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Client unable to establish connection.

So Now, we have to connect with -C [Trust Server Certificate]
mssql@ms-standalone-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P b1HLv9EV4CaSalX6 -N -C
1> 
```

We can verify from the above output that TLS is disabled for this database, `mssql.conf` file has no tls configuration.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mssqlserver -n demo ms-standalone
kubectl delete issuer -n demo mssqlserver-ca-issuer mssqlserver-new-ca-issuer
kubectl delete mssqlserveropsrequest msops-add-tls msops-remove msops-rotate msops-change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
