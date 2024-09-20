---
title: Reconfigure RabbitMQ TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: rm-reconfigure-tls-ops
    name: Reconfigure TLS/SSL Encryption
    parent: rm-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure RabbitMQ TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing RabbitMQ database via a RabbitMQOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/rabbitmq](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/rabbitmq) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a RabbitMQ database

Here, We are going to create a RabbitMQ database without TLS and then reconfigure the database to use TLS.

### Deploy RabbitMQ without TLS

In this section, we are going to deploy a RabbitMQ Replicaset database without TLS. In the next few sections we will reconfigure TLS using `RabbitMQOpsRequest` CRD. Below is the YAML of the `RabbitMQ` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `RabbitMQ` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/rm.yaml
rabbitmq.kubedb.com/rm created
```

Now, wait until `rm` has status `Ready`. i.e,

```bash
$ kubectl get rm -n demo
NAME    VERSION    STATUS    AGE
rm      3.13.2     Ready     10m


```bash
$ kubectl get secrets -n demo rm-admin-cred -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo rm-admin-cred -o jsonpath='{.data.\password}' | base64 -d
U6(h_pYrekLZ2OOd

We can verify from the above output that TLS is disabled for this database.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in RabbitMQ. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls rabbitmq-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/rabbitmq-ca created
```

Now, Let's create an `Issuer` using the `mongo-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: rm-issuer
  namespace: demo
spec:
  ca:
    secretName: rabbitmq-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/issuer.yaml
issuer.cert-manager.io/rm-issuer created
```

### Create RabbitMQOpsRequest

In order to add TLS to the database, we have to create a `RabbitMQOpsRequest` CRO with our created issuer. Below is the YAML of the `RabbitMQOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rmops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: rm
  tls:
    issuerRef:
      name: rm-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - rabbitmq
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `rm` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/rabbitmq/concepts/rabbitmq.md#spectls).

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/rmops-add-tls.yaml
rabbitmqopsrequest.ops.kubedb.com/rmops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CRO,

```bash
$ kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME           TYPE             STATUS        AGE
rmops-add-tls   ReconfigureTLS   Successful    91s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rmops-add-tls 
Name:         rmops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/RabbitMQopsrequests/mops-add-tls
  UID:               0024ec16-0d43-4686-a2d7-1cdeb96e41a5
Spec:
  Database Ref:
    Name:  rm
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
    Message:               RabbitMQ ops request is reconfiguring TLS
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
  Normal  PauseDatabase      2m10s  KubeDB Ops-manager operator  Pausing RabbitMQ demo/rm
  Normal  PauseDatabase      2m10s  KubeDB Ops-manager operator  Successfully paused RabbitMQ demo/rm
  Normal  TLSAdded           2m10s  KubeDB Ops-manager operator  Successfully Updated StatefulSets
  Normal  RestartReplicaSet  10s    KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  ResumeDatabase     10s    KubeDB Ops-manager operator  Resuming RabbitMQ demo/rm
  Normal  ResumeDatabase     10s    KubeDB Ops-manager operator  Successfully resumed RabbitMQ demo/rm
  Normal  Successful         10s    KubeDB Ops-manager operator  Successfully Reconfigured TLS
```

## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it rm-2 -n demo bash
root@rm-2:/# openssl x509 -in /var/private/ssl/client.pem -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Jun  9 13:32:20 2021 GMT
```

So, the certificate will expire on this time `Jun  9 13:32:20 2021 GMT`. 

### Create RabbitMQOpsRequest

Now we are going to increase it using a RabbitMQOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rmops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: rm
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `rm` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/RabbitMQ/reconfigure-tls/rmops-rotate.yaml
RabbitMQopsrequest.ops.kubedb.com/rmops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CRO,

```bash
$ kubectl get RabbitMQopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME           TYPE             STATUS        AGE
rmops-rotate    ReconfigureTLS   Successful    112s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rmops-rotate
Name:         rmops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/rabbitmqopsrequests/rmops-rotate
  UID:               6d96ead2-a868-47d8-85fb-77eecc9a96b4
Spec:
  Database Ref:
    Name:  rm
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2021-03-11T16:17:55Z
    Message:               RabbitMQ ops request is reconfiguring TLS
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
$ kubectl exec -it rm-2 -n demo bash
root@rm-2:/# openssl x509 -in /var/run/rabbitmq/tls/client.pem -inform PEM -enddate -nameopt RFC2253 -noout
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
$ kubectl create secret tls rm-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/rm-new-ca created
```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: rm-new-issuer
  namespace: demo
spec:
  ca:
    secretName: rm-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/rm-new-issuer created
```

### Create RabbitMQOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `RabbitMQOpsRequest` CRO with the newly created issuer. Below is the YAML of the `RabbitMQOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rm-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: rm
  tls:
    issuerRef:
      name: rm-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `rm` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/reconfigure-tls/rm-change-issuer.yaml
rabbitmqopsrequest.ops.kubedb.com/rm-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CRO,

```bash
$ kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME                  TYPE             STATUS        AGE
rm-change-issuer      ReconfigureTLS   Successful    105s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rm-change-issuer
Name:         rm-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/rabbitmqopsrequests/rm-change-issuer
  UID:               cdfe8a7d-52ef-466c-a5dd-97e74ad598ca
Spec:
  Database Ref:
    Name:  rm
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       mg-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2021-03-11T16:27:47Z
    Message:               RabbitMQ ops request is reconfiguring TLS
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
$ kubectl exec -it rm-2 -n demo bash
root@mgo-rs-tls-2:/$ openssl x509 -in /var/run/rabbitmq/tls/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb-updated,CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a RabbitMQOpsRequest.

### Create RabbitMQOpsRequest

Below is the YAML of the `RabbitMQOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: mops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: rm
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `rm` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/reconfigure-tls/mops-remove.yaml
rabbitmqopsrequest.ops.kubedb.com/mops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CRO,

```bash
$ kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME          TYPE             STATUS        AGE
mops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe rabbitmqopsrequest -n demo mops-remove
Name:         mops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/RabbitMQopsrequests/mops-remove
  UID:               99184cc4-1595-4f0f-b8eb-b65c5d0e86a6
Spec:
  Database Ref:
    Name:  rm
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2021-03-11T16:35:32Z
    Message:               RabbitMQ ops request is reconfiguring TLS
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
  Normal  PauseDatabase      2m5s  KubeDB Ops-manager operator  Pausing RabbitMQ demo/rm
  Normal  PauseDatabase      2m5s  KubeDB Ops-manager operator  Successfully paused RabbitMQ demo/rm
  Normal  TLSRemoved         2m5s  KubeDB Ops-manager operator  Successfully Updated StatefulSets
  Normal  RestartReplicaSet  35s   KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  ResumeDatabase     35s   KubeDB Ops-manager operator  Resuming RabbitMQ demo/rm
  Normal  ResumeDatabase     35s   KubeDB Ops-manager operator  Successfully resumed RabbitMQ demo/rm
  Normal  Successful         35s   KubeDB Ops-manager operator  Successfully Reconfigured TLS
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rabbitmq -n demo rm
kubectl delete issuer -n demo rm-issuer rm-new-issuer
kubectl delete rabbitmqopsrequest rmops-add-tls rmops-remove rmops-rotate rm-change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Monitor your RabbitMQ database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md).
- Monitor your RabbitMQ database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md).
- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
