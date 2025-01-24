---
title: Reconfigure PgBouncer TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: pb-reconfigure-tls-rs
    name: Reconfigure PgBouncer TLS/SSL Encryption
    parent: pb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure PgBouncer TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing PgBouncer database via a PgBouncerOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a PgBouncer

Here, We are going to create a PgBouncer database without TLS and then reconfigure the pgbouncer to use TLS.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Deploy PgBouncer without TLS

In this section, we are going to deploy a PgBouncer without TLS. In the next few sections we will reconfigure TLS using `PgBouncerOpsRequest` CRD. Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    poolMode: session
    port: 5432
    reservePoolSize: 5
    maxClientConnections: 87
    defaultPoolSize: 2
    minPoolSize: 1
    authType: md5
  deletionPolicy: WipeOut
```

Let's create the `PgBouncer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/pgbouncer.yaml
pgbouncer.kubedb.com/pgbouncer created
```

Now, wait until `pgbouncer` has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
NAME   VERSION   STATUS   AGE
pb     1.18.0    Ready    65s

$ kubectl describe pgbouncer pb -n demo
Name:         pb
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1
Kind:         PgBouncer
Metadata:
  Creation Timestamp:  2025-01-24T13:15:24Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  130502
  UID:               bc199683-2564-4aff-aa2b-832bd78be875
Spec:
  Auth Secret:
    Name:  pb-auth
  Auto Ops:
  Connection Pool:
    Auth Type:               md5
    Default Pool Size:       2
    Max Client Connections:  87
    Min Pool Size:           1
    Pool Mode:               session
    Port:                    5432
    Reserve Pool Size:       5
  Database:
    Database Name:  postgres
    Database Ref:
      Name:         ha-postgres
      Namespace:    demo
    Sync Users:     true
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
        Name:  pgbouncer
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Privileged:    false
          Run As Group:  70
          Run As User:   70
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:            70
        Run As Group:        70
        Run As User:         70
      Service Account Name:  pb
  Replicas:                  1
  Ssl Mode:                  disable
  Version:                   1.18.0
Status:
  Conditions:
    Last Transition Time:  2025-01-24T13:15:44Z
    Message:               The KubeDB operator has started the provisioning of PgBouncer: demo/pb
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-01-24T13:15:56Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-01-24T13:16:15Z
    Message:               pgBouncer demo/pb is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-01-24T13:16:15Z
    Message:               pgBouncer demo/pb is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-01-24T13:16:28Z
    Message:               The PgBouncer: demo/pb is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:
  Type     Reason      Age   From               Message
  ----     ------      ----  ----               -------
  Normal   Successful  95s   Postgres operator  Successfully created governing service
  Normal   Successful  95s   Postgres operator  Successfully created Service
  Normal   Successful  84s   Postgres operator  Successfully created PgBouncer
  Normal   Successful  84s   Postgres operator  Successfully created appbinding
  Warning  Failed      84s   Postgres operator  Fail to be ready PgBouncer: "pb". Reason: Operation cannot be fulfilled on pgbouncers.kubedb.com "pb": the object has been modified; please apply your changes to the latest version and try again
  Warning  Failed      74s   Postgres operator  Fail to be ready PgBouncer: "pb". Reason: Operation cannot be fulfilled on pgbouncers.kubedb.com "pb": the object has been modified; please apply your changes to the latest version and try again
  Normal   Successful  63s   Postgres operator  Successfully patched PgBouncer
```

Now, we let exec into a pgbouncer pod and verify that the TLS is disabled.


```bash
$ kubectl exec -it -n demo pgbouncer-0 -- bash
pgbouncer-0:/$ cat opt/pgbouncer-II/etc/pgbouncer.conf
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
socket_dir = '/var/run/pgbouncer'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgbouncer'
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
pgbouncer-0:/$ exit
exit
```
We can see from the above output that `ssl='off'` so we can verify that TLS is disabled for this pgbouncer.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in PgBouncer. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls pgbouncer-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pgbouncer-ca created
```

Now, Let's create an `Issuer` using the `pgbouncer-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pgbouncer-issuer
  namespace: demo
spec:
  ca:
    secretName: pgbouncer-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/pgbouncer-issuer created
```

### Create PgBouncerOpsRequest

In order to add TLS to the pgbouncer, we have to create a `PgBouncerOpsRequest` CRO with our created issuer. Below is the YAML of the `PgBouncerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgbouncer
  tls:
    sslMode: require
    clientAuthMode: cert
    issuerRef:
      name: pgbouncer-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - pgbouncer
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mg-rs` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/pgbouncer/concepts/pgbouncer.md#spectls).
- `spec.tls.sslMode` is the ssl mode of the server. You can see the details [here](/docs/guides/pgbouncer/concepts/pgbouncer.md#specsslmode).
- `spec.tls.clientAuthMode` is the authentication mode of the server. You can see the details [here](/docs/guides/pgbouncer/concepts/pgbouncer.md#specclientauthmode).
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/pgbouncer/concepts/opsrequest.md#spectimeout)

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/pbops-add-tls.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME            TYPE             STATUS       AGE
pbops-add-tls   ReconfigureTLS   Successful   107s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-add-tls 
Name:         pbops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T06:47:24Z
  Generation:          1
  Resource Version:    8910
  UID:                 679969d1-4a1b-460e-a64a-d0255db4f1c8
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   pgbouncer
  Timeout:  5m
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          pgbouncer
    Client Auth Mode:  cert
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       pgbouncer-issuer
    Ssl Mode:     require
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T06:47:24Z
    Message:               PgBouncer ops-request has started to reconfigure tls for RabbitMQ nodes
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
    Message:               successfully reconciled the PgBouncer with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T06:48:30Z
    Message:               Successfully Restarted PgBouncer pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T06:47:50Z
    Message:               get pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgbouncer-0
    Last Transition Time:  2024-07-29T06:47:50Z
    Message:               evict pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgbouncer-0
    Last Transition Time:  2024-07-29T06:48:25Z
    Message:               check pod running; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgbouncer-0
    Last Transition Time:  2024-07-29T06:48:37Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T06:48:39Z
    Message:               Successfully updated PgBouncer TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age    From                         Message
  ----     ------                                                      ----   ----                         -------
  Normal   Starting                                                    2m12s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-add-tls
  Normal   Starting                                                    2m12s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pgbouncer
  Normal   Successful                                                  2m12s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pgbouncer for PgBouncerOpsRequest: ppops-add-tls
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
  Normal   UpdatePetSets                                               111s   KubeDB Ops-manager Operator  successfully reconciled the PgBouncer with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgbouncer-0             106s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  evict pod; ConditionStatus:True; PodName:pgbouncer-0           106s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:False; PodName:pgbouncer-0  101s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:True; PodName:pgbouncer-0   71s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgbouncer-0
  Normal   RestartPods                                                 66s    KubeDB Ops-manager Operator  Successfully Restarted PgBouncer pods
  Normal   Starting                                                    57s    KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pgbouncer
  Normal   Successful                                                  57s    KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pgbouncer for PgBouncerOpsRequest: ppops-add-tls
```

Now, we let exec into a pgbouncer pod and verify that the TLS is enabled.

```bash
$ kubectl exec -it -n demo pgbouncer-0 -- bash
pgbouncer-0:/$ cat opt/pgbouncer-II/etc/pgbouncer.conf
pgbouncer-0:/$ cat opt/pgbouncer-II/etc/pgbouncer.conf
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
socket_dir = '/var/run/pgbouncer'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgbouncer'
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
ssl_ca_cert = '/opt/pgbouncer-II/tls/ca.pem'
ssl = on
failover_on_backend_error = 'off'
log_min_messages = 'warning'
statement_level_load_balance = 'off'
memory_cache_enabled = 'off'
memqcache_oiddir = '/tmp/oiddir/'
allow_clear_text_frontend_auth = 'false'
ssl = 'on'
ssl_key = '/opt/pgbouncer-II/tls/tls.key'
ssl_cert = '/opt/pgbouncer-II/tls/tls.crt'
failover_on_backend_error = 'off'
pgbouncer-0:/$ exit
exit
```
We can see from the above output that `ssl='on'` so we can verify that TLS is enabled for this pgbouncer.

Now, let's connect with just client certificate using psql. For that first save the `tls.crt` and `tls.key` from the secret named `pgbouncer-client-cert`.
```bash
$ kubectl get secrets -n demo pgbouncer-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > client.crt                                    master ⬆ ⬇ ✱ ◼
$ kubectl get secrets -n demo pgbouncer-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > client.key
```
Now let's port forward to the main service of the pgbouncer:
```bash
$ kubectl port-forward -n demo svc/pgbouncer 9999                                                                                                                                         pgbouncer ✱ ◼
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
$ kubectl exec -it -n demo pgbouncer-0 -- bash                                                                                                 master ⬆ ⬇ ✱ ◼
pgbouncer-0:/$ openssl x509 -in /opt/pgbouncer-II/tls/ca.pem -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Oct 27 06:47:28 2024 GMT
```

So, the certificate will expire on this time `27 06:47:28 2024 GMT`. 

### Create PgBouncerOpsRequest

Now we are going to increase it using a PgBouncerOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgbouncer
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pgbouncer`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our pgbouncer.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this pgbouncer.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/pbops-rotate.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME           TYPE             STATUS       AGE
pbops-rotate   ReconfigureTLS   Successful   113s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-rotate
Name:         pbops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T07:10:15Z
  Generation:          1
  Resource Version:    10505
  UID:                 6399fdad-bf2a-43de-b542-9ad09f032844
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgbouncer
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T07:10:15Z
    Message:               PgBouncer ops-request has started to reconfigure tls for RabbitMQ nodes
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
    Message:               successfully reconciled the PgBouncer with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T07:11:25Z
    Message:               Successfully Restarted PgBouncer pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T07:10:45Z
    Message:               get pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgbouncer-0
    Last Transition Time:  2024-07-29T07:10:45Z
    Message:               evict pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgbouncer-0
    Last Transition Time:  2024-07-29T07:11:20Z
    Message:               check pod running; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgbouncer-0
    Last Transition Time:  2024-07-29T07:11:25Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T07:11:25Z
    Message:               Successfully updated PgBouncer TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age    From                         Message
  ----     ------                                                      ----   ----                         -------
  Normal   Starting                                                    2m16s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-rotate
  Normal   Starting                                                    2m16s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pgbouncer
  Normal   Successful                                                  2m16s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pgbouncer for PgBouncerOpsRequest: pbops-rotate
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
  Normal   UpdatePetSets                                               112s   KubeDB Ops-manager Operator  successfully reconciled the PgBouncer with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgbouncer-0             106s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  evict pod; ConditionStatus:True; PodName:pgbouncer-0           106s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:False; PodName:pgbouncer-0  101s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:True; PodName:pgbouncer-0   71s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgbouncer-0
  Normal   RestartPods                                                 66s    KubeDB Ops-manager Operator  Successfully Restarted PgBouncer pods
  Normal   Starting                                                    66s    KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pgbouncer
  Normal   Successful                                                  66s    KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pgbouncer for PgBouncerOpsRequest: pbops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo pgbouncer-0 -- bash                                                                                                 master ⬆ ⬇ ✱ ◼
pgbouncer-0:/$ openssl x509 -in /opt/pgbouncer-II/tls/ca.pem -inform PEM -enddate -nameopt RFC2253 -noout
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
$ kubectl create secret tls pgbouncer-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/pgbouncer-new-ca created
```

Now, Let's create a new `Issuer` using the `pgbouncer-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pb-new-issuer
  namespace: demo
spec:
  ca:
    secretName: pgbouncer-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/new-issuer.yaml
issuer.cert-manager.io/pb-new-issuer created
```

### Create PgBouncerOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `PgBouncerOpsRequest` CRO with the newly created issuer. Below is the YAML of the `PgBouncerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgbouncer
  tls:
    issuerRef:
      name: pb-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pgbouncer`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our pgbouncer.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/pbops-change-issuer.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                  TYPE             STATUS       AGE
pbops-change-issuer   ReconfigureTLS   Successful   87s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-change-issuer
Name:         pbops-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T07:37:09Z
  Generation:          1
  Resource Version:    12367
  UID:                 f48452ed-7264-4e99-80f1-58d7e826d9a9
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgbouncer
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       pb-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T07:37:09Z
    Message:               PgBouncer ops-request has started to reconfigure tls for RabbitMQ nodes
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
    Message:               successfully reconciled the PgBouncer with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T07:38:15Z
    Message:               Successfully Restarted PgBouncer pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T07:37:35Z
    Message:               get pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgbouncer-0
    Last Transition Time:  2024-07-29T07:37:35Z
    Message:               evict pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgbouncer-0
    Last Transition Time:  2024-07-29T07:38:10Z
    Message:               check pod running; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgbouncer-0
    Last Transition Time:  2024-07-29T07:38:15Z
    Message:               Successfully updated pgbouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T07:38:16Z
    Message:               Successfully updated PgBouncer TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age    From                         Message
  ----     ------                                                      ----   ----                         -------
  Normal   Starting                                                    3m39s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-change-issuer
  Normal   Starting                                                    3m39s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pgbouncer
  Normal   Successful                                                  3m39s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pgbouncer for PgBouncerOpsRequest: pbops-change-issuer
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
  Normal   UpdatePetSets                                               3m18s  KubeDB Ops-manager Operator  successfully reconciled the PgBouncer with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgbouncer-0             3m13s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  evict pod; ConditionStatus:True; PodName:pgbouncer-0           3m13s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:False; PodName:pgbouncer-0  3m8s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:True; PodName:pgbouncer-0   2m38s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgbouncer-0
  Normal   RestartPods                                                 2m33s  KubeDB Ops-manager Operator  Successfully Restarted PgBouncer pods
  Normal   Starting                                                    2m32s  KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pgbouncer
  Normal   Successful                                                  2m32s  KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/PgBouncer for PgBouncerOpsRequest: pbops-change-issuer
```

Now, Let's exec pgbouncer and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo pgbouncer-0 -- bash
pgbouncer-0:/$ openssl x509 -in /opt/pgbouncer-II/tls/ca.pem -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb-updated,CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the pgbouncer

Now, we are going to remove TLS from this pgbouncer using a PgBouncerOpsRequest.

### Create PgBouncerOpsRequest

Below is the YAML of the `PgBouncerOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pbops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgbouncer
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `pgbouncer`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our pgbouncer.
- `spec.tls.remove` specifies that we want to remove tls from this pgbouncer.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/reconfigure-tls/pbops-remove.yaml
pgbounceropsrequest.ops.kubedb.com/pbops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CRO,

```bash
$ wacth kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME           TYPE             STATUS       AGE
pbops-remove   ReconfigureTLS   Successful   65s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-remove
Name:         pbops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-29T08:38:35Z
  Generation:          1
  Resource Version:    16378
  UID:                 f848e04f-0fd1-48ce-813d-67dbdc3e4a55
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgbouncer
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-29T08:38:37Z
    Message:               PgBouncer ops-request has started to reconfigure tls for RabbitMQ nodes
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
    Message:               successfully reconciled the PgBouncer with TLS
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-29T08:39:32Z
    Message:               Successfully Restarted PgBouncer pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-29T08:38:52Z
    Message:               get pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgbouncer-0
    Last Transition Time:  2024-07-29T08:38:52Z
    Message:               evict pod; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgbouncer-0
    Last Transition Time:  2024-07-29T08:39:27Z
    Message:               check pod running; ConditionStatus:True; PodName:pgbouncer-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgbouncer-0
    Last Transition Time:  2024-07-29T08:39:32Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-29T08:39:33Z
    Message:               Successfully updated PgBouncer TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                      Age   From                         Message
  ----     ------                                                      ----  ----                         -------
  Normal   Starting                                                    84s   KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-remove
  Normal   Starting                                                    84s   KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pgbouncer
  Normal   Successful                                                  83s   KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pgbouncer for PgBouncerOpsRequest: pbops-remove
  Normal   UpdatePetSets                                               74s   KubeDB Ops-manager Operator  successfully reconciled the PgBouncer with TLS
  Warning  get pod; ConditionStatus:True; PodName:pgbouncer-0             69s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  evict pod; ConditionStatus:True; PodName:pgbouncer-0           69s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:False; PodName:pgbouncer-0  64s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgbouncer-0
  Warning  check pod running; ConditionStatus:True; PodName:pgbouncer-0   34s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgbouncer-0
  Normal   RestartPods                                                 29s   KubeDB Ops-manager Operator  Successfully Restarted PgBouncer pods
  Normal   Starting                                                    29s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pgbouncer
  Normal   Successful                                                  28s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pgbouncer for PgBouncerOpsRequest: pbops-remove
```

Now, Let's exec into pgbouncer and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo pgbouncer-0 -- bash
pgbouncer-0:/$ cat opt/pgbouncer-II/etc/pgbouncer.conf
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
socket_dir = '/var/run/pgbouncer'
pcp_listen_addresses = *
pcp_port = 9595
pcp_socket_dir = '/var/run/pgbouncer'
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

We can see from the above output that `ssl='off'` so we can verify that TLS is disabled successfully for this pgbouncer.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pgbouncer -n demo pgbouncer
kubectl delete issuer -n demo pgbouncer-issuer pb-new-issuer
kubectl delete pgbounceropsrequest -n demo pbops-add-tls pbops-remove pbops-rotate pbops-change-issuer
kubectl delete pg -n demo ha-postgres
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Monitor your PgBouncer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
