---
title: Reconfigure MongoDB TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: mg-reconfigure-tls-rs
    name: Reconfigure MongoDB TLS/SSL Encryption
    parent: mg-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure MongoDB TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing MongoDB database via a MongoDBOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a MongoDB database

Here, We are going to create a MongoDB database without TLS and then reconfigure the database to use TLS.

### Deploy MongoDB without TLS

In this section, we are going to deploy a MongoDB Replicaset database without TLS. In the next few sections we will reconfigure TLS using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs
  namespace: demo
spec:
  version: "4.4.26"
  replicas: 3
  replicaSet:
    name: rs0
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure-tls/mg-replicaset.yaml
mongodb.kubedb.com/mg-rs created
```

Now, wait until `mg-replicaset` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME    VERSION    STATUS    AGE
mg-rs   4.4.26      Ready     10m

$ kubectl dba describe mongodb mg-rs -n demo
Name:               mg-rs
Namespace:          demo
CreationTimestamp:  Thu, 11 Mar 2021 13:25:05 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mg-rs","namespace":"demo"},"spec":{"replicaSet":{"name":"rs0"...
Replicas:           3  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  Delete

StatefulSet:          
  Name:               mg-rs
  CreationTimestamp:  Thu, 11 Mar 2021 13:25:05 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mg-rs
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:        <none>
  Replicas:           824639275080 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mg-rs
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mg-rs
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.70.27
  Port:         primary  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.63:27017

Service:        
  Name:         mg-rs-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mg-rs
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.63:27017,10.244.0.65:27017,10.244.0.67:27017

Auth Secret:
  Name:         mg-rs-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mg-rs
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         Opaque
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mg-rs","namespace":"demo"},"spec":{"replicaSet":{"name":"rs0"},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"4.4.26"}}

    Creation Timestamp:  2021-03-11T07:26:44Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mg-rs
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mongodbs.kubedb.com
    Name:                            mg-rs
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mg-rs
        Port:    27017
        Scheme:  mongodb
    Parameters:
      API Version:  config.kubedb.com/v1alpha1
      Kind:         MongoConfiguration
      Replica Sets:
        host-0:  rs0/mg-rs-0.mg-rs-pods.demo.svc,mg-rs-1.mg-rs-pods.demo.svc,mg-rs-2.mg-rs-pods.demo.svc
      Stash:
        Addon:
          Backup Task:
            Name:  mongodb-backup-4.4.6-v6
          Restore Task:
            Name:  mongodb-restore-4.4.6-v6
    Secret:
      Name:   mg-rs-auth
    Type:     kubedb.com/mongodb
    Version:  4.4.26

Events:
  Type    Reason      Age   From              Message
  ----    ------      ----  ----              -------
  Normal  Successful  14m   MongoDB operator  Successfully created stats service
  Normal  Successful  14m   MongoDB operator  Successfully created Service
  Normal  Successful  14m   MongoDB operator  Successfully  stats service
  Normal  Successful  14m   MongoDB operator  Successfully  stats service
  Normal  Successful  13m   MongoDB operator  Successfully  stats service
  Normal  Successful  13m   MongoDB operator  Successfully  stats service
  Normal  Successful  13m   MongoDB operator  Successfully  stats service
  Normal  Successful  13m   MongoDB operator  Successfully  stats service
  Normal  Successful  13m   MongoDB operator  Successfully  stats service
  Normal  Successful  12m   MongoDB operator  Successfully  stats service
  Normal  Successful  12m   MongoDB operator  Successfully patched StatefulSet demo/mg-rs
```

Now, we can connect to this database through [mongo-shell](https://docs.mongodb.com/v4.2/mongo/) and verify that the TLS is disabled.


```bash
$ kubectl get secrets -n demo mg-rs-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mg-rs-auth -o jsonpath='{.data.\password}' | base64 -d
U6(h_pYrekLZ2OOd

$ kubectl exec -it mg-rs-0 -n demo -- mongo admin -u root -p 'U6(h_pYrekLZ2OOd'
rs0:PRIMARY> db.adminCommand({ getParameter:1, sslMode:1 })
{
	"sslMode" : "disabled",
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1615468344, 1),
		"signature" : {
			"hash" : BinData(0,"Xdclj9Y67WKZ/oTDGT/E1XzOY28="),
			"keyId" : NumberLong("6938294279689207810")
		}
	},
	"operationTime" : Timestamp(1615468344, 1)
}
```

We can verify from the above output that TLS is disabled for this database.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in MongoDB. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls mongo-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/mongo-ca created
```

Now, Let's create an `Issuer` using the `mongo-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mg-issuer
  namespace: demo
spec:
  ca:
    secretName: mongo-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/mg-issuer created
```

### Create MongoDBOpsRequest

In order to add TLS to the database, we have to create a `MongoDBOpsRequest` CRO with our created issuer. Below is the YAML of the `MongoDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    issuerRef:
      name: mg-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - mongo
          organizationalUnits:
            - client
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mg-rs` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/mongodb/concepts/mongodb.md#spectls).

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure-tls/mops-add-tls.yaml
mongodbopsrequest.ops.kubedb.com/mops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CRO,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME           TYPE             STATUS        AGE
mops-add-tls   ReconfigureTLS   Successful    91s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-add-tls 
Name:         mops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-11T13:32:18Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:tls:
          .:
          f:certificates:
          f:issuerRef:
            .:
            f:apiGroup:
            f:kind:
            f:name:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-11T13:32:18Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-11T13:32:19Z
  Resource Version:  488264
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-add-tls
  UID:               0024ec16-0d43-4686-a2d7-1cdeb96e41a5
Spec:
  Database Ref:
    Name:  mg-rs
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          mongo
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mg-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2021-03-11T13:32:19Z
    Message:               MongoDB ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2021-03-11T13:32:25Z
    Message:               Successfully Updated StatefulSets
    Observed Generation:   1
    Reason:                TLSAdded
    Status:                True
    Type:                  TLSAdded
    Last Transition Time:  2021-03-11T13:34:25Z
    Message:               Successfully Restarted ReplicaSet nodes
    Observed Generation:   1
    Reason:                RestartReplicaSet
    Status:                True
    Type:                  RestartReplicaSet
    Last Transition Time:  2021-03-11T13:34:25Z
    Message:               Successfully Reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason             Age    From                        Message
  ----    ------             ----   ----                        -------
  Normal  PauseDatabase      2m10s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-rs
  Normal  PauseDatabase      2m10s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-rs
  Normal  TLSAdded           2m10s  KubeDB Ops-manager operator  Successfully Updated StatefulSets
  Normal  RestartReplicaSet  10s    KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  ResumeDatabase     10s    KubeDB Ops-manager operator  Resuming MongoDB demo/mg-rs
  Normal  ResumeDatabase     10s    KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-rs
  Normal  Successful         10s    KubeDB Ops-manager operator  Successfully Reconfigured TLS
```

Now, Let's exec into a database primary node and find out the username to connect in a mongo shell,

```bash
$ kubectl exec -it mg-rs-2 -n demo bash
root@mgo-rs-tls-2:/$ ls /var/run/mongodb/tls
ca.crt  client.pem  mongo.pem
root@mgo-rs-tls-2:/$ openssl x509 -in /var/run/mongodb/tls/client.pem -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=root,OU=client,O=mongo
```

Now, we can connect using `CN=root,OU=client,O=mongo` as root to connect to the mongo shell of the master pod,

```bash
root@mgo-rs-tls-2:/$ mongo --tls --tlsCAFile /var/run/mongodb/tls/ca.crt --tlsCertificateKeyFile /var/run/mongodb/tls/client.pem admin --host localhost --authenticationMechanism MONGODB-X509 --authenticationDatabase='$external' -u "CN=root,OU=client,O=mongo" --quiet
rs0:PRIMARY>
```

We are connected to the mongo shell. Let's run some command to verify the sslMode and the user,

```bash
rs0:PRIMARY> db.adminCommand({ getParameter:1, sslMode:1 })
{
	"sslMode" : "requireSSL",
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1615472249, 1),
		"signature" : {
			"hash" : BinData(0,"AAAAAAAAAAAAAAAAAAAAAAAAAAA="),
			"keyId" : NumberLong(0)
		}
	},
	"operationTime" : Timestamp(1615472249, 1)
}
```

We can see from the above output that, `sslMode` is set to `requireSSL`. So, database TLS is enabled successfully to this database.

## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it mg-rs-2 -n demo bash
root@mg-rs-2:/# openssl x509 -in /var/run/mongodb/tls/client.pem -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Jun  9 13:32:20 2021 GMT
```

So, the certificate will expire on this time `Jun  9 13:32:20 2021 GMT`. 

### Create MongoDBOpsRequest

Now we are going to increase it using a MongoDBOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mg-rs` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure-tls/mops-rotate.yaml
mongodbopsrequest.ops.kubedb.com/mops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CRO,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME           TYPE             STATUS        AGE
mops-rotate    ReconfigureTLS   Successful    112s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-rotate
Name:         mops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-11T16:17:55Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:tls:
          .:
          f:rotateCertificates:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-11T16:17:55Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-11T16:17:55Z
  Resource Version:  521643
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-rotate
  UID:               6d96ead2-a868-47d8-85fb-77eecc9a96b4
Spec:
  Database Ref:
    Name:  mg-rs
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2021-03-11T16:17:55Z
    Message:               MongoDB ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2021-03-11T16:17:55Z
    Message:               Successfully Added Issuing Condition in Certificates
    Observed Generation:   1
    Reason:                IssuingConditionUpdated
    Status:                True
    Type:                  IssuingConditionUpdated
    Last Transition Time:  2021-03-11T16:18:00Z
    Message:               Successfully Issued New Certificates
    Observed Generation:   1
    Reason:                CertificateIssuingSuccessful
    Status:                True
    Type:                  CertificateIssuingSuccessful
    Last Transition Time:  2021-03-11T16:19:45Z
    Message:               Successfully Restarted ReplicaSet nodes
    Observed Generation:   1
    Reason:                RestartReplicaSet
    Status:                True
    Type:                  RestartReplicaSet
    Last Transition Time:  2021-03-11T16:19:45Z
    Message:               Successfully Reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                        Age    From                        Message
  ----    ------                        ----   ----                        -------
  Normal  CertificateIssuingSuccessful  2m10s  KubeDB Ops-manager operator  Successfully Issued New Certificates
  Normal  RestartReplicaSet             25s    KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  Successful                    25s    KubeDB Ops-manager operator  Successfully Reconfigured TLS
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it mg-rs-2 -n demo bash
root@mg-rs-2:/# openssl x509 -in /var/run/mongodb/tls/client.pem -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Jun  9 16:17:55 2021 GMT
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
$ kubectl create secret tls mongo-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/mongo-new-ca created
```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mg-new-issuer
  namespace: demo
spec:
  ca:
    secretName: mongo-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/mg-new-issuer created
```

### Create MongoDBOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `MongoDBOpsRequest` CRO with the newly created issuer. Below is the YAML of the `MongoDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    issuerRef:
      name: mg-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mg-rs` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure-tls/mops-change-issuer.yaml
mongodbopsrequest.ops.kubedb.com/mops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CRO,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                  TYPE             STATUS        AGE
mops-change-issuer    ReconfigureTLS   Successful    105s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-change-issuer
Name:         mops-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-11T16:27:47Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:tls:
          .:
          f:issuerRef:
            .:
            f:apiGroup:
            f:kind:
            f:name:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-11T16:27:47Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-11T16:27:47Z
  Resource Version:  523903
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-change-issuer
  UID:               cdfe8a7d-52ef-466c-a5dd-97e74ad598ca
Spec:
  Database Ref:
    Name:  mg-rs
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mg-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2021-03-11T16:27:47Z
    Message:               MongoDB ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2021-03-11T16:27:52Z
    Message:               Successfully Issued New Certificates
    Observed Generation:   1
    Reason:                CertificateIssuingSuccessful
    Status:                True
    Type:                  CertificateIssuingSuccessful
    Last Transition Time:  2021-03-11T16:29:37Z
    Message:               Successfully Restarted ReplicaSet nodes
    Observed Generation:   1
    Reason:                RestartReplicaSet
    Status:                True
    Type:                  RestartReplicaSet
    Last Transition Time:  2021-03-11T16:29:37Z
    Message:               Successfully Reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                        Age    From                        Message
  ----    ------                        ----   ----                        -------
  Normal  CertificateIssuingSuccessful  2m27s  KubeDB Ops-manager operator  Successfully Issued New Certificates
  Normal  RestartReplicaSet             42s    KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  Successful                    42s    KubeDB Ops-manager operator  Successfully Reconfigured TLS
```

Now, Let's exec into a database node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it mg-rs-2 -n demo bash
root@mgo-rs-tls-2:/$ openssl x509 -in /var/run/mongodb/tls/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb-updated,CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a MongoDBOpsRequest.

### Create MongoDBOpsRequest

Below is the YAML of the `MongoDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mg-rs` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure-tls/mops-remove.yaml
mongodbopsrequest.ops.kubedb.com/mops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CRO,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME          TYPE             STATUS        AGE
mops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-remove
Name:         mops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-11T16:35:32Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:tls:
          .:
          f:remove:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-11T16:35:32Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-11T16:35:32Z
  Resource Version:  525550
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-remove
  UID:               99184cc4-1595-4f0f-b8eb-b65c5d0e86a6
Spec:
  Database Ref:
    Name:  mg-rs
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2021-03-11T16:35:32Z
    Message:               MongoDB ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2021-03-11T16:35:37Z
    Message:               Successfully Updated StatefulSets
    Observed Generation:   1
    Reason:                TLSRemoved
    Status:                True
    Type:                  TLSRemoved
    Last Transition Time:  2021-03-11T16:37:07Z
    Message:               Successfully Restarted ReplicaSet nodes
    Observed Generation:   1
    Reason:                RestartReplicaSet
    Status:                True
    Type:                  RestartReplicaSet
    Last Transition Time:  2021-03-11T16:37:07Z
    Message:               Successfully Reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason             Age   From                        Message
  ----    ------             ----  ----                        -------
  Normal  PauseDatabase      2m5s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-rs
  Normal  PauseDatabase      2m5s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-rs
  Normal  TLSRemoved         2m5s  KubeDB Ops-manager operator  Successfully Updated StatefulSets
  Normal  RestartReplicaSet  35s   KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  ResumeDatabase     35s   KubeDB Ops-manager operator  Resuming MongoDB demo/mg-rs
  Normal  ResumeDatabase     35s   KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-rs
  Normal  Successful         35s   KubeDB Ops-manager operator  Successfully Reconfigured TLS
```

Now, Let's exec into the database primary node and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo mg-rs-1 -- mongo admin -u root -p 'U6(h_pYrekLZ2OOd'
rs0:PRIMARY> db.adminCommand({ getParameter:1, sslMode:1 })
{
	"sslMode" : "disabled",
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1615480817, 1),
		"signature" : {
			"hash" : BinData(0,"CWJngDTQqDhKXyx7WMFJqqUfvhY="),
			"keyId" : NumberLong("6938294279689207810")
		}
	},
	"operationTime" : Timestamp(1615480817, 1)
}
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mongodb -n demo mg-rs
kubectl delete issuer -n demo mg-issuer mg-new-issuer
kubectl delete mongodbopsrequest mops-add-tls mops-remove mops-rotate mops-change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
