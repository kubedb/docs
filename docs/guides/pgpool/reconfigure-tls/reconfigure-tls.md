---
title: Reconfigure Pgpool TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: pp-reconfigure-tls-rs
    name: Reconfigure Pgpool TLS/SSL Encryption
    parent: pp-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Pgpool TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Pgpool database via a PgpoolOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Pgpool

Here, We are going to create a Pgpool database without TLS and then reconfigure the pgpool to use TLS.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Deploy Pgpool without TLS

In this section, we are going to deploy a Pgpool without TLS. In the next few sections we will reconfigure TLS using `PgpoolOpsRequest` CRD. Below is the YAML of the `Pgpool` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  deletionPolicy: WipeOut
```

Let's create the `Pgpool` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/pgpool.yaml
pgpool.kubedb.com/pgpool created
```

Now, wait until `pgpool` has status `Ready`. i.e,

```bash
$ kubectl get pp -n demo
NAME     TYPE                  VERSION   STATUS   AGE
pgpool   kubedb.com/v1alpha2   4.5.0     Ready    21s

$ kubectl dba describe pgpool pgpool -n demo
Name:         pgpool
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Pgpool
Metadata:
  Creation Timestamp:  2024-07-18T07:38:54Z
  Finalizers:
    kubedb.com
  Generation:  2
  Managed Fields:
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:clientAuthMode:
        f:deletionPolicy:
        f:healthChecker:
          .:
          f:failureThreshold:
          f:periodSeconds:
          f:timeoutSeconds:
        f:postgresRef:
          .:
          f:name:
          f:namespace:
        f:replicas:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-07-18T07:38:54Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com":
      f:spec:
        f:authSecret:
    Manager:      kubedb-provisioner
    Operation:    Update
    Time:         2024-07-18T07:38:54Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:phase:
    Manager:         kubedb-provisioner
    Operation:       Update
    Subresource:     status
    Time:            2024-07-18T07:39:07Z
  Resource Version:  98658
  UID:               4c7a00d5-9c52-4e6b-aa74-98e82285d9e1
Spec:
  Auth Secret:
    Name:            pgpool-auth
  Client Auth Mode:  md5
  Deletion Policy:   WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  pgpool
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
          Run As Group:     70
          Run As Non Root:  true
          Run As User:      70
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:  70
  Postgres Ref:
    Name:       ha-postgres
    Namespace:  demo
  Replicas:     1
  Ssl Mode:     disable
  Version:      4.5.0
Status:
  Conditions:
    Last Transition Time:  2024-07-18T07:38:54Z
    Message:               The KubeDB operator has started the provisioning of Pgpool: demo/pgpool
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-07-18T07:38:56Z
    Message:               All replicas are ready for Pgpool demo/pgpool
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-07-18T07:39:06Z
    Message:               pgpool demo/pgpool is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-07-18T07:39:06Z
    Message:               pgpool demo/pgpool is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-07-18T07:39:06Z
    Message:               The Pgpool: demo/pgpool is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

Now, we let exec into a pgpool pod and verify that the TLS is disabled.


```bash
$ kubectl exec -it -n demo pgpool-0 -- bash
pgpool-0:/$ cat opt/pgpool-II/etc/pgpool.conf
backend_hostname0 = 'ha-postgres.demo.svc'
backend_port0 = 5432
backend_weight0 = 1
backend_flag0 = 'ALWAYS_PRIMARY|DISALLOW_TO_FAILOVER'
backend_hostname1 = 'ha-postgres-standby.demo.svc'
backend_port1 = 5432
backend_weight1 = 1
backend_flag1 = 'DISALLOW_TO_FAILOVER'
enable_pool_hba = on
listen_addresses = *
port = 9999
socket_dir = '/var/run/pgpool'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgpool'
log_per_node_statement = on
sr_check_period = 0
health_check_period = 0
backend_clustering_mode = 'streaming_replication'
num_init_children = 5
max_pool = 15
child_life_time = 300
child_max_connections = 0
connection_life_time = 0
client_idle_limit = 0
connection_cache = on
load_balance_mode = on
ssl = 'off'
failover_on_backend_error = 'off'
log_min_messages = 'warning'
statement_level_load_balance = 'off'
memory_cache_enabled = 'off'
memqcache_oiddir = '/tmp/oiddir/'
allow_clear_text_frontend_auth = 'false'


failover_on_backend_error = 'off'
pgpool-0:/$ exit
exit
```
We can see from the above output that `ssl='off'` so we can verify that TLS is disabled for this pgpool.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Pgpool. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls pgpool-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pgpool-ca created
```

Now, Let's create an `Issuer` using the `pgpool-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pgpool-issuer
  namespace: demo
spec:
  ca:
    secretName: pgpool-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/pgpool-issuer created
```

### Create PgpoolOpsRequest

In order to add TLS to the pgpool, we have to create a `PgpoolOpsRequest` CRO with our created issuer. Below is the YAML of the `PgpoolOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: ppops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgpool
  tls:
    sslMode: require
    clientAuthMode: cert
    issuerRef:
      name: pgpool-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - pgpool
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mg-rs` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/pgpool/concepts/pgpool.md#spectls).
- `spec.tls.sslMode` is the ssl mode of the server. You can see the details [here](/docs/guides/pgpool/concepts/pgpool.md#specsslmode).
- `spec.tls.clientAuthMode` is the authentication mode of the server. You can see the details [here](/docs/guides/pgpool/concepts/pgpool.md#specclientauthmode).
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/pgpool/concepts/opsrequest.md#spectimeout)

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/ppops-add-tls.yaml
pgpoolopsrequest.ops.kubedb.com/ppops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CRO,

```bash
$ kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME           TYPE             STATUS        AGE
mops-add-tls   ReconfigureTLS   Successful    91s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo mops-add-tls 
Name:         mops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/pgpoolopsrequests/mops-add-tls
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
    Message:               Pgpool ops request is reconfiguring TLS
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
  Normal  PauseDatabase      2m10s  KubeDB Ops-manager operator  Pausing Pgpool demo/mg-rs
  Normal  PauseDatabase      2m10s  KubeDB Ops-manager operator  Successfully paused Pgpool demo/mg-rs
  Normal  TLSAdded           2m10s  KubeDB Ops-manager operator  Successfully Updated StatefulSets
  Normal  RestartReplicaSet  10s    KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  ResumeDatabase     10s    KubeDB Ops-manager operator  Resuming Pgpool demo/mg-rs
  Normal  ResumeDatabase     10s    KubeDB Ops-manager operator  Successfully resumed Pgpool demo/mg-rs
  Normal  Successful         10s    KubeDB Ops-manager operator  Successfully Reconfigured TLS
```

Now, Let's exec into a database primary node and find out the username to connect in a mongo shell,

```bash
$ kubectl exec -it mg-rs-2 -n demo bash
root@mgo-rs-tls-2:/$ ls /var/run/pgpool/tls
ca.crt  client.pem  mongo.pem
root@mgo-rs-tls-2:/$ openssl x509 -in /var/run/pgpool/tls/client.pem -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=root,OU=client,O=mongo
```

Now, we can connect using `CN=root,OU=client,O=mongo` as root to connect to the mongo shell of the master pod,

```bash
root@mgo-rs-tls-2:/$ mongo --tls --tlsCAFile /var/run/pgpool/tls/ca.crt --tlsCertificateKeyFile /var/run/pgpool/tls/client.pem admin --host localhost --authenticationMechanism MONGODB-X509 --authenticationDatabase='$external' -u "CN=root,OU=client,O=mongo" --quiet
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
root@mg-rs-2:/# openssl x509 -in /var/run/pgpool/tls/client.pem -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Jun  9 13:32:20 2021 GMT
```

So, the certificate will expire on this time `Jun  9 13:32:20 2021 GMT`. 

### Create PgpoolOpsRequest

Now we are going to increase it using a PgpoolOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
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

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/mops-rotate.yaml
pgpoolopsrequest.ops.kubedb.com/mops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CRO,

```bash
$ kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME           TYPE             STATUS        AGE
mops-rotate    ReconfigureTLS   Successful    112s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo mops-rotate
Name:         mops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/pgpoolopsrequests/mops-rotate
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
    Message:               Pgpool ops request is reconfiguring TLS
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
root@mg-rs-2:/# openssl x509 -in /var/run/pgpool/tls/client.pem -inform PEM -enddate -nameopt RFC2253 -noout
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/mg-new-issuer created
```

### Create PgpoolOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `PgpoolOpsRequest` CRO with the newly created issuer. Below is the YAML of the `PgpoolOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
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

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/mops-change-issuer.yaml
pgpoolopsrequest.ops.kubedb.com/mops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CRO,

```bash
$ kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                  TYPE             STATUS        AGE
mops-change-issuer    ReconfigureTLS   Successful    105s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo mops-change-issuer
Name:         mops-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/pgpoolopsrequests/mops-change-issuer
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
    Message:               Pgpool ops request is reconfiguring TLS
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
root@mgo-rs-tls-2:/$ openssl x509 -in /var/run/pgpool/tls/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb-updated,CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a PgpoolOpsRequest.

### Create PgpoolOpsRequest

Below is the YAML of the `PgpoolOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
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

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/mops-remove.yaml
pgpoolopsrequest.ops.kubedb.com/mops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CRO,

```bash
$ kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME          TYPE             STATUS        AGE
mops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo mops-remove
Name:         mops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
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
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/pgpoolopsrequests/mops-remove
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
    Message:               Pgpool ops request is reconfiguring TLS
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
  Normal  PauseDatabase      2m5s  KubeDB Ops-manager operator  Pausing Pgpool demo/mg-rs
  Normal  PauseDatabase      2m5s  KubeDB Ops-manager operator  Successfully paused Pgpool demo/mg-rs
  Normal  TLSRemoved         2m5s  KubeDB Ops-manager operator  Successfully Updated StatefulSets
  Normal  RestartReplicaSet  35s   KubeDB Ops-manager operator  Successfully Restarted ReplicaSet nodes
  Normal  ResumeDatabase     35s   KubeDB Ops-manager operator  Resuming Pgpool demo/mg-rs
  Normal  ResumeDatabase     35s   KubeDB Ops-manager operator  Successfully resumed Pgpool demo/mg-rs
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
kubectl delete pgpool -n demo mg-rs
kubectl delete issuer -n demo mg-issuer mg-new-issuer
kubectl delete pgpoolopsrequest mops-add-tls mops-remove mops-rotate mops-change-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
