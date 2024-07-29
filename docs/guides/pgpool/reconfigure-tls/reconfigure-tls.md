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
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME            TYPE             STATUS       AGE
ppops-add-tls   ReconfigureTLS   Successful   107s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo ppops-add-tls 
Name:         ppops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T06:47:24Z
  Generation:          1
  Resource Version:    8910
  UID:                 679969d1-4a1b-460e-a64a-d0255db4f1c8
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   pgpool
  Timeout:  5m
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          pgpool
    Client Auth Mode:  cert
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       pgpool-issuer
    Ssl Mode:     require
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T06:47:24Z
    Message:               Pgpool ops-request has started to reconfigure tls for RabbitMQ nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-29T06:47:27Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-29T06:47:38Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-07-29T06:47:33Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-07-29T06:47:33Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-07-29T06:47:33Z
    Message:               check issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckIssuingCondition
    Last Transition Time:  2024-07-29T06:47:45Z
    Message:               successfully reconciled the Pgpool with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T06:48:30Z
    Message:               Successfully Restarted Pgpool pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T06:47:50Z
    Message:               get pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgpool-0
    Last Transition Time:  2024-07-29T06:47:50Z
    Message:               evict pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgpool-0
    Last Transition Time:  2024-07-29T06:48:25Z
    Message:               check pod running; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgpool-0
    Last Transition Time:  2024-07-29T06:48:37Z
    Message:               Successfully updated Pgpool
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T06:48:39Z
    Message:               Successfully updated Pgpool TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age    From                         Message
  ----     ------                                                      ----   ----                         -------
  Normal   Starting                                                    2m12s  KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/ppops-add-tls
  Normal   Starting                                                    2m12s  KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pgpool
  Normal   Successful                                                  2m12s  KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-add-tls
  Warning  get certificate; ConditionStatus:True                       2m3s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m3s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m3s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       2m3s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m3s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m3s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       2m3s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m3s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m3s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                           2m3s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                       118s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 118s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               118s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       118s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 118s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               118s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       118s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 118s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               118s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                           118s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                               111s   KubeDB Ops-manager Operator  successfully reconciled the Pgpool with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgpool-0             106s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgpool-0
  Warning  evict pod; ConditionStatus:True; PodName:pgpool-0           106s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:False; PodName:pgpool-0  101s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:True; PodName:pgpool-0   71s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgpool-0
  Normal   RestartPods                                                 66s    KubeDB Ops-manager Operator  Successfully Restarted Pgpool pods
  Normal   Starting                                                    57s    KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pgpool
  Normal   Successful                                                  57s    KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-add-tls
```

Now, we let exec into a pgpool pod and verify that the TLS is enabled.

```bash
$ kubectl exec -it -n demo pgpool-0 -- bash
pgpool-0:/$ cat opt/pgpool-II/etc/pgpool.conf
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
ssl_ca_cert = '/opt/pgpool-II/tls/ca.pem'
ssl = on
failover_on_backend_error = 'off'
log_min_messages = 'warning'
statement_level_load_balance = 'off'
memory_cache_enabled = 'off'
memqcache_oiddir = '/tmp/oiddir/'
allow_clear_text_frontend_auth = 'false'
ssl = 'on'
ssl_key = '/opt/pgpool-II/tls/tls.key'
ssl_cert = '/opt/pgpool-II/tls/tls.crt'
failover_on_backend_error = 'off'
pgpool-0:/$ exit
exit
```
We can see from the above output that `ssl='on'` so we can verify that TLS is enabled for this pgpool.

Now, let's connect with just client certificate using psql. For that first save the `tls.crt` and `tls.key` from the secret named `pgpool-client-cert`.
```bash
$ kubectl get secrets -n demo pgpool-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > client.crt                                    master ⬆ ⬇ ✱ ◼
$ kubectl get secrets -n demo pgpool-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > client.key
```
Now let's port forward to the main service of the pgpool:
```bash
$ kubectl port-forward -n demo svc/pgpool 9999                                                                                                                                         pgpool ✱ ◼
Forwarding from 127.0.0.1:9999 -> 9999
```
Now connect with `psql`:
```bash
psql "sslmode=require port=9999 host=localhost dbname=postgres user=postgres sslrootcert=ca.crt sslcert=client.crt sslkey=client.key"     master ⬆ ⬇ ✱ ◼
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, compression: off)
Type "help" for help.

postgres=# 
```
So, here we have connected using the client certificate and now password was needed and the connection is tls secured. So, we can safely assume that tls enabling was successful.
## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it -n demo pgpool-0 -- bash                                                                                                 master ⬆ ⬇ ✱ ◼
pgpool-0:/$ openssl x509 -in /opt/pgpool-II/tls/ca.pem -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Oct 27 06:47:28 2024 GMT
```

So, the certificate will expire on this time `27 06:47:28 2024 GMT`. 

### Create PgpoolOpsRequest

Now we are going to increase it using a PgpoolOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: ppops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgpool
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pgpool`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our pgpool.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this pgpool.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/ppops-rotate.yaml
pgpoolopsrequest.ops.kubedb.com/ppops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CRO,

```bash
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME           TYPE             STATUS       AGE
ppops-rotate   ReconfigureTLS   Successful   113s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo ppops-rotate
Name:         ppops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T07:10:15Z
  Generation:          1
  Resource Version:    10505
  UID:                 6399fdad-bf2a-43de-b542-9ad09f032844
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgpool
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T07:10:15Z
    Message:               Pgpool ops-request has started to reconfigure tls for RabbitMQ nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-29T07:10:18Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-29T07:10:19Z
    Message:               successfully add issuing condition to all the certificates
    Observed Generation:   1
    Reason:                IssueCertificatesSucceeded
    Status:                True
    Type:                  IssueCertificatesSucceeded
    Last Transition Time:  2024-07-29T07:10:31Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-07-29T07:10:25Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-07-29T07:10:25Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-07-29T07:10:25Z
    Message:               check issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckIssuingCondition
    Last Transition Time:  2024-07-29T07:10:39Z
    Message:               successfully reconciled the Pgpool with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T07:11:25Z
    Message:               Successfully Restarted Pgpool pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T07:10:45Z
    Message:               get pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgpool-0
    Last Transition Time:  2024-07-29T07:10:45Z
    Message:               evict pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgpool-0
    Last Transition Time:  2024-07-29T07:11:20Z
    Message:               check pod running; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgpool-0
    Last Transition Time:  2024-07-29T07:11:25Z
    Message:               Successfully updated Pgpool
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T07:11:25Z
    Message:               Successfully updated Pgpool TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age    From                         Message
  ----     ------                                                      ----   ----                         -------
  Normal   Starting                                                    2m16s  KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/ppops-rotate
  Normal   Starting                                                    2m16s  KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pgpool
  Normal   Successful                                                  2m16s  KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-rotate
  Warning  get certificate; ConditionStatus:True                       2m6s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m6s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m6s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       2m6s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m6s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m6s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       2m6s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m6s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m6s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                           2m6s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                       2m1s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m1s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m1s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       2m1s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m1s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m1s   KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       2m     KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 2m     KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               2m     KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                           2m     KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                               112s   KubeDB Ops-manager Operator  successfully reconciled the Pgpool with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgpool-0             106s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgpool-0
  Warning  evict pod; ConditionStatus:True; PodName:pgpool-0           106s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:False; PodName:pgpool-0  101s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:True; PodName:pgpool-0   71s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgpool-0
  Normal   RestartPods                                                 66s    KubeDB Ops-manager Operator  Successfully Restarted Pgpool pods
  Normal   Starting                                                    66s    KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pgpool
  Normal   Successful                                                  66s    KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo pgpool-0 -- bash                                                                                                 master ⬆ ⬇ ✱ ◼
pgpool-0:/$ openssl x509 -in /opt/pgpool-II/tls/ca.pem -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Oct 27 07:10:20 2024 GMT
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
$ kubectl create secret tls pgpool-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pgpool-new-ca created
```

Now, Let's create a new `Issuer` using the `pgpool-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pp-new-issuer
  namespace: demo
spec:
  ca:
    secretName: pgpool-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/pp-new-issuer created
```

### Create PgpoolOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `PgpoolOpsRequest` CRO with the newly created issuer. Below is the YAML of the `PgpoolOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: ppops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgpool
  tls:
    issuerRef:
      name: pp-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pgpool`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our pgpool.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/ppops-change-issuer.yaml
pgpoolopsrequest.ops.kubedb.com/ppops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CRO,

```bash
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                  TYPE             STATUS       AGE
ppops-change-issuer   ReconfigureTLS   Successful   87s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo ppops-change-issuer
Name:         ppops-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T07:37:09Z
  Generation:          1
  Resource Version:    12367
  UID:                 f48452ed-7264-4e99-80f1-58d7e826d9a9
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgpool
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       pp-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T07:37:09Z
    Message:               Pgpool ops-request has started to reconfigure tls for RabbitMQ nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-29T07:37:12Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-29T07:37:24Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-07-29T07:37:18Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-07-29T07:37:18Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-07-29T07:37:18Z
    Message:               check issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckIssuingCondition
    Last Transition Time:  2024-07-29T07:37:30Z
    Message:               successfully reconciled the Pgpool with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T07:38:15Z
    Message:               Successfully Restarted Pgpool pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T07:37:35Z
    Message:               get pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgpool-0
    Last Transition Time:  2024-07-29T07:37:35Z
    Message:               evict pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgpool-0
    Last Transition Time:  2024-07-29T07:38:10Z
    Message:               check pod running; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgpool-0
    Last Transition Time:  2024-07-29T07:38:15Z
    Message:               Successfully updated Pgpool
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T07:38:16Z
    Message:               Successfully updated Pgpool TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age    From                         Message
  ----     ------                                                      ----   ----                         -------
  Normal   Starting                                                    3m39s  KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/ppops-change-issuer
  Normal   Starting                                                    3m39s  KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pgpool
  Normal   Successful                                                  3m39s  KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-change-issuer
  Warning  get certificate; ConditionStatus:True                       3m30s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 3m30s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               3m30s  KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       3m30s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 3m30s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               3m30s  KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       3m30s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 3m30s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               3m30s  KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                           3m30s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                       3m25s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 3m25s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               3m24s  KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       3m24s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 3m24s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               3m24s  KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                       3m24s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                 3m24s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  check issuing condition; ConditionStatus:True               3m24s  KubeDB Ops-manager Operator  check issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                           3m24s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                               3m18s  KubeDB Ops-manager Operator  successfully reconciled the Pgpool with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgpool-0             3m13s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgpool-0
  Warning  evict pod; ConditionStatus:True; PodName:pgpool-0           3m13s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:False; PodName:pgpool-0  3m8s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:True; PodName:pgpool-0   2m38s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgpool-0
  Normal   RestartPods                                                 2m33s  KubeDB Ops-manager Operator  Successfully Restarted Pgpool pods
  Normal   Starting                                                    2m32s  KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pgpool
  Normal   Successful                                                  2m32s  KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-change-issuer
```

Now, Let's exec pgpool and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo pgpool-0 -- bash
pgpool-0:/$ openssl x509 -in /opt/pgpool-II/tls/ca.pem -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb-updated,CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the pgpool

Now, we are going to remove TLS from this pgpool using a PgpoolOpsRequest.

### Create PgpoolOpsRequest

Below is the YAML of the `PgpoolOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: ppops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgpool
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pgpool`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our pgpool.
- `spec.tls.remove` specifies that we want to remove tls from this pgpool.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/reconfigure-tls/ppops-remove.yaml
pgpoolopsrequest.ops.kubedb.com/ppops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CRO,

```bash
$ wacth kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME           TYPE             STATUS       AGE
ppops-remove   ReconfigureTLS   Successful   65s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgpoolopsrequest -n demo ppops-remove
Name:         ppops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T08:38:35Z
  Generation:          1
  Resource Version:    16378
  UID:                 f848e04f-0fd1-48ce-813d-67dbdc3e4a55
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgpool
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T08:38:37Z
    Message:               Pgpool ops-request has started to reconfigure tls for RabbitMQ nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-29T08:38:41Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-29T08:38:47Z
    Message:               successfully reconciled the Pgpool with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T08:39:32Z
    Message:               Successfully Restarted Pgpool pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T08:38:52Z
    Message:               get pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgpool-0
    Last Transition Time:  2024-07-29T08:38:52Z
    Message:               evict pod; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgpool-0
    Last Transition Time:  2024-07-29T08:39:27Z
    Message:               check pod running; ConditionStatus:True; PodName:pgpool-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgpool-0
    Last Transition Time:  2024-07-29T08:39:32Z
    Message:               Successfully updated Pgpool
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T08:39:33Z
    Message:               Successfully updated Pgpool TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age   From                         Message
  ----     ------                                                      ----  ----                         -------
  Normal   Starting                                                    84s   KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/ppops-remove
  Normal   Starting                                                    84s   KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pgpool
  Normal   Successful                                                  83s   KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-remove
  Normal   UpdatePetSets                                               74s   KubeDB Ops-manager Operator  successfully reconciled the Pgpool with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgpool-0             69s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgpool-0
  Warning  evict pod; ConditionStatus:True; PodName:pgpool-0           69s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:False; PodName:pgpool-0  64s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgpool-0
  Warning  check pod running; ConditionStatus:True; PodName:pgpool-0   34s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgpool-0
  Normal   RestartPods                                                 29s   KubeDB Ops-manager Operator  Successfully Restarted Pgpool pods
  Normal   Starting                                                    29s   KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pgpool
  Normal   Successful                                                  28s   KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pgpool for PgpoolOpsRequest: ppops-remove
```

Now, Let's exec into pgpool and find out that TLS is disabled or not.

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
```

We can see from the above output that `ssl='off'` so we can verify that TLS is disabled successfully for this pgpool.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pgpool -n demo pgpool
kubectl delete issuer -n demo pgpool-issuer pp-new-issuer
kubectl delete pgpoolopsrequest -n demo ppops-add-tls ppops-remove ppops-rotate ppops-change-issuer
kubectl delete pg -n demo ha-postgres
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
