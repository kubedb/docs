---
title: Reconfigure MSSQLServer Availability Group TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: ms-reconfigure-tls-ag-cluster
    name: Availability Group (HA Cluster)
    parent: ms-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure MSSQLServer Availability Group TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing MSSQLServer Availability Group Cluster database via a MSSQLServerOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

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

## Add TLS to a MSSQLServer Availability Group Cluster

Here, We are going to create a MSSQLServer Availability Group Cluster without TLS and then reconfigure the database to use TLS.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/standalone/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer created
```

### Deploy MSSQLServer without TLS

In this section, we are going to deploy a MSSQLServer Availability Group Cluster without TLS. In the next few sections we will reconfigure to add TLS using `MSSQLServerOpsRequest` CRD. Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-ag-cluster
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
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
          resources:
            requests:
              cpu: "500m"
              memory: "1.5Gi"
            limits:
              memory: "2Gi"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/mssql-ag-cluster.yaml
mssqlserver.kubedb.com/mssql-ag-cluster created
```

Now, wait until `mssql-ag-cluster` has status `Ready`. i.e,

```bash
$ kubectl get ms  -n demo
NAME               VERSION     STATUS   AGE
mssql-ag-cluster   2022-cu12   Ready    4m38s
```

Now, connect to this database by exec into a pod and verify the TLS is disabled. 

> when we connect using the sqlcmd tool, the -N option is available with [s|m|o] parameters, where 's' stands for strict, 'm' for mandatory, and 'o' for optional. The default setting is mandatory.


```bash
$ kubectl get secrets -n demo mssql-ag-cluster-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secrets -n demo mssql-ag-cluster-auth -o jsonpath='{.data.\password}' | base64 -d
Q9kDWVQMnawLcnZq

$ kubectl exec -it -n demo mssql-ag-cluster-0 -c mssql -- bash
mssql@mssql-ag-cluster-0:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
mssql@mssql-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "Q9kDWVQMnawLcnZq" -N
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : SSL Provider: [error:0A000086:SSL routines::certificate verify failed:self-signed certificate].
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Client unable to establish connection.

So Now, we have to connect with -C [Trust Server Certificate]
mssql@mssql-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "Q9kDWVQMnawLcnZq" -N -C
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
  name: msops-ag-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mssql-ag-cluster
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

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mssql-ag-cluster` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/mssqlserver/concepts/mssqlserver.md#spectls).

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-ag-add-tls.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-ag-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ watch kubectl get msops -n demo
Every 2.0s: kubectl get msops -n demo
NAME               TYPE             STATUS       AGE
msops-ag-add-tls   ReconfigureTLS   Successful   3m32s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-ag-add-tls 
Name:         msops-ag-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T14:13:46Z
  Generation:          1
  Resource Version:    555629
  UID:                 329b1815-6002-4d20-8df8-662ec6bedb2a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mssql-ag-cluster
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
    Last Transition Time:  2024-11-18T14:13:46Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T14:13:49Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-18T14:13:59Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-18T14:13:54Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-18T14:13:54Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-18T14:13:54Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-18T14:14:04Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-18T14:16:10Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-18T14:14:10Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:14:10Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:14:45Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:14:50Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:14:50Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:15:25Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:15:30Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:15:30Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:16:05Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:16:10Z
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
$ kubectl exec -it mssql-ag-cluster-0 -n demo -c mssql -- bash
mssql@mssql-ag-cluster-0:/$ ls /var/opt/mssql/tls
mssql@mssql-ag-cluster-0:/$ openssl x509 -in /var/opt/mssql/tls/client.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=mssql,OU=client,O=mssqlserver
mssql@mssql-ag-cluster-0:/$  cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
[network]
forceencryption = 1
tlscert = /var/opt/mssql/tls/server.crt
tlskey = /var/opt/mssql/tls/server.key
tlsprotocols = 1.2,1.1,1.0
mssql@mssql-ag-cluster-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P Q9kDWVQMnawLcnZq -N
1> 
```

We can verify from the above output that TLS is enabled for this database, `mssql.conf` file has tls configurations.  So, TLS is enabled successfully to this database.


## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it mssql-ag-cluster-0 -n demo -c mssql -- bash
mssql@mssql-ag-cluster-0:/$ openssl x509 -in /var/opt/mssql/tls/client.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb 16 14:13:49 2025 GMT
mssql@mssql-ag-cluster-0:/$ 
```

So, the certificate will expire on this time `Feb 16 13:11:02 2025 GMT`.


### Create MSSQLServerOpsRequest

Now we are going to increase it using a MSSQLServerOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-ag-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mssql-ag-cluster
  tls:
    rotateCertificates: true
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mssql-ag-cluster` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-ag-rotate.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-ag-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ kubectl get mssqlserveropsrequest -n demo
Every 2.0s: kubectl get mssqlserveropsrequest -n demo
NAME               TYPE             STATUS       AGE
msops-ag-rotate    ReconfigureTLS   Successful   5m14s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-ag-rotate
Name:         msops-ag-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T14:36:38Z
  Generation:          1
  Resource Version:    558973
  UID:                 54eca6e2-5e08-4730-a18a-1a754d2d8ea3
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   mssql-ag-cluster
  Timeout:  5m
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T14:36:38Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T14:36:54Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-18T14:36:54Z
    Message:               successfully add issuing condition to all the certificates
    Observed Generation:   1
    Reason:                IssueCertificatesSucceeded
    Status:                True
    Type:                  IssueCertificatesSucceeded
    Last Transition Time:  2024-11-18T14:37:04Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-18T14:36:59Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-18T14:36:59Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-18T14:36:59Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-18T14:37:09Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-18T14:40:14Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-18T14:37:14Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:37:14Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:38:19Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:38:24Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:38:24Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:39:09Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:39:14Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:39:14Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:40:09Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:40:14Z
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
$ kubectl exec -it mssql-ag-cluster-0 -n demo -c mssql -- bash
mssql@mssql-ag-cluster-0:/$ openssl x509 -in /var/opt/mssql/tls/client.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb 16 14:36:54 2025 GMT
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

Now, Let's create a new `Issuer` using the `mssqlserver-new-ca` secret that we have just created. The `YAML` file looks like this:

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
  name: msops-ag-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mssql-ag-cluster
  tls:
    issuerRef:
      name: mssqlserver-new-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mssql-ag-cluster` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-ag-change-issuer.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-ag-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ kubectl get mssqlserveropsrequest -n demo
Every 2.0s: kubectl get mssqlserveropsrequest -n demo
NAME                     TYPE             STATUS       AGE
msops-ag-change-issuer   ReconfigureTLS   Successful   3m56s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-ag-change-issuer
Name:         msops-ag-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T14:46:57Z
  Generation:          1
  Resource Version:    560150
  UID:                 5bf3e378-01b4-4dc9-aeeb-8cf5765aed10
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mssql-ag-cluster
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mssqlserver-new-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T14:46:57Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T14:47:00Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-18T14:47:10Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-18T14:47:05Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-18T14:47:05Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-18T14:47:05Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-18T14:47:15Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-18T14:49:40Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-18T14:47:20Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:47:20Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:48:05Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:48:10Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:48:10Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:48:50Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:48:55Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:48:55Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:49:35Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T14:49:40Z
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
$ kubectl exec -it mssql-ag-cluster-2 -n demo -c mssql -- bash
mssql@mssql-ag-cluster-2:/$ openssl x509 -in /var/opt/mssql/tls/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
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
  name: msops-ag-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mssql-ag-cluster
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mssql-ag-cluster` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure-tls/msops-ag-remove.yaml
mssqlserveropsrequest.ops.kubedb.com/msops-ag-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CRO,

```bash
$ watch kubectl get mssqlserveropsrequest -n demo
Every 2.0s: kubectl get mssqlserveropsrequest -n demo
NAME                     TYPE             STATUS       AGE
msops-ag-remove          ReconfigureTLS   Successful   5m17s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mssqlserveropsrequest -n demo msops-ag-remove
Name:         msops-ag-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2024-11-18T14:56:16Z
  Generation:          1
  Resource Version:    561471
  UID:                 7f43e8b9-4ae9-4f5d-9355-c58a4dbf4504
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mssql-ag-cluster
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-18T14:56:16Z
    Message:               MSSQLServer ops-request has started to reconfigure tls for mssqlserver nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-18T14:56:19Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-11-18T14:56:24Z
    Message:               successfully reconciled the MSSQLServer with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-18T15:01:29Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-18T14:56:29Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:56:29Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:59:19Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-0
    Last Transition Time:  2024-11-18T14:59:24Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T14:59:25Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T15:00:09Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-1
    Last Transition Time:  2024-11-18T15:00:14Z
    Message:               get pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T15:00:14Z
    Message:               evict pod; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T15:01:24Z
    Message:               check pod running; ConditionStatus:True; PodName:mssql-ag-cluster-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--mssql-ag-cluster-2
    Last Transition Time:  2024-11-18T15:01:29Z
    Message:               Successfully completed reconfigureTLS for mssqlserver.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
```

Now, Lets exec into the database node and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo mssql-ag-cluster-1 -c mssql -- bash
mssql@mssql-ag-cluster-1:/$ cat /var/opt/mssql/mssql.conf
[language]
lcid = 1033
mssql@mssql-ag-cluster-1:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P Q9kDWVQMnawLcnZq -N
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : SSL Provider: [error:0A000086:SSL routines::certificate verify failed:self-signed certificate].
Sqlcmd: Error: Microsoft ODBC Driver 17 for SQL Server : Client unable to establish connection.


So Now, we have to connect with -C [Trust Server Certificate]
mssql@mssql-ag-cluster-1:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P Q9kDWVQMnawLcnZq -N -C
1> 
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mssqlserver -n demo mssql-ag-cluster
kubectl delete issuer -n demo mssqlserver-ca-issuer mssqlserver-new-ca-issuer
kubectl delete mssqlserveropsrequest msops-ag-add-tls msops-ag-remove msops-ag-rotate msops-ag-change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
