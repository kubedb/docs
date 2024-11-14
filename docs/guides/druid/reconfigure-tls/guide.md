---
title: Reconfigure Druid TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-druid-reconfigure-tls-guide
    name: Reconfigure Druid TLS/SSL Encryption
    parent: guides-druid-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Druid TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Druid database via a DruidOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/druid](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Druid database

Here, We are going to create a Druid without TLS and then reconfigure the database to use TLS.

### Deploy Druid without TLS

In this section, we are going to deploy a Druid topology cluster without TLS. In the next few sections we will reconfigure TLS using `DruidOpsRequest` CRD. Below is the YAML of the `Druid` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

Let's create the `Druid` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/reconfigure-tls/yamls/druid-cluster.yaml
druid.kubedb.com/druid-cluster created
```

Now, wait until `druid-cluster` has status `Ready`. i.e,

```bash
$ kubectl get dr -n demo -w
NAME            TYPE                  VERSION   STATUS         AGE
druid-cluster   kubedb.com/v1alpha2   28.0.1    Provisioning   15s
druid-cluster   kubedb.com/v1alpha2   28.0.1    Provisioning   37s
.
.
druid-cluster   kubedb.com/v1alpha2   28.0.1    Ready          2m27s
```

Now, we can exec one druid broker pod and verify configuration that the TLS is disabled.

```bash
$ kubectl exec -it -n demo druid-cluster-coordinators-0 -- bash
Defaulted container "druid" out of: druid, init-druid (init)
bash-5.1$ cat conf/druid/cluster/_common/common.runtime.properties                           
druid.auth.authenticator.basic.authorizerName=basic
druid.auth.authenticator.basic.credentialsValidator.type=metadata
druid.auth.authenticator.basic.initialAdminPassword={"type": "environment", "variable": "DRUID_ADMIN_PASSWORD"}
druid.auth.authenticator.basic.initialInternalClientPassword=*****
druid.auth.authenticator.basic.skipOnFailure=false
druid.auth.authenticator.basic.type=basic
druid.auth.authenticatorChain=["basic"]
druid.auth.authorizer.basic.type=basic
druid.auth.authorizers=["basic"]
druid.emitter.logging.logLevel=info
druid.emitter=noop
druid.escalator.authorizerName=basic
druid.escalator.internalClientPassword=******
druid.escalator.internalClientUsername=druid_system
druid.escalator.type=basic
druid.expressions.useStrictBooleans=true
druid.extensions.loadList=["druid-avro-extensions", "druid-kafka-indexing-service", "druid-kafka-indexing-service", "druid-datasketches", "druid-multi-stage-query", "druid-basic-security", "mysql-metadata-storage", "druid-s3-extensions"]
druid.global.http.eagerInitialization=false
druid.host=localhost
druid.indexer.logs.directory=var/druid/indexing-logs
druid.indexer.logs.type=file
druid.indexing.doubleStorage=double
druid.lookup.enableLookupSyncOnStartup=false
druid.metadata.storage.connector.connectURI=jdbc:mysql://druid-cluster-mysql-metadata.demo.svc:3306/druid
druid.metadata.storage.connector.createTables=true
druid.metadata.storage.connector.host=localhost
druid.metadata.storage.connector.password={"type": "environment", "variable": "DRUID_METADATA_STORAGE_PASSWORD"}
druid.metadata.storage.connector.port=1527
druid.metadata.storage.connector.user=root
druid.metadata.storage.type=mysql
druid.monitoring.monitors=["org.apache.druid.java.util.metrics.JvmMonitor", "org.apache.druid.server.metrics.ServiceStatusMonitor"]
druid.s3.accessKey=minio
druid.s3.enablePathStyleAccess=true
druid.s3.endpoint.signingRegion=us-east-1
druid.s3.endpoint.url=http://myminio-hl.demo.svc.cluster.local:9000/
druid.s3.protocol=http
druid.s3.secretKey=minio123
druid.selectors.coordinator.serviceName=druid/coordinator
druid.selectors.indexing.serviceName=druid/overlord
druid.server.hiddenProperties=["druid.s3.accessKey","druid.s3.secretKey","druid.metadata.storage.connector.password", "password", "key", "token", "pwd"]
druid.sql.enable=true
druid.sql.planner.useGroupingSetForExactDistinct=true
druid.startup.logging.logProperties=true
druid.storage.baseKey=druid/segments
druid.storage.bucket=druid
druid.storage.storageDirectory=var/druid/segments
druid.storage.type=s3
druid.zk.paths.base=/druid
druid.zk.service.host=druid-cluster-zk.demo.svc:2181
druid.zk.service.pwd={"type": "environment", "variable": "DRUID_ZK_SERVICE_PASSWORD"}
druid.zk.service.user=super
```

We can verify from the above output that TLS is disabled for this cluster as there is no TLS/SSL related configs provided for it.

#### Verify TLS/SSL is disabled using Druid UI

First port-forward the port `8888` to local machine:

```bash
$ kubectl port-forward -n demo svc/druid-cluster-routers 8888
Forwarding from 127.0.0.1:8888 -> 8888
Forwarding from [::1]:8888 -> 8888
```


Now hit the `http://localhost:8888` from any browser, and you will be prompted to provide the credential of the druid database. By following the steps discussed below, you can get the credential generated by the KubeDB operator for your Druid database.

**Connection information:**

- Username:

  ```bash
  $ kubectl get secret -n demo druid-cluster-admin-cred -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo druid-cluster-admin-cred -o jsonpath='{.data.password}' | base64 -d
  LzJtVRX5E8MorFaf
  ```

After providing the credentials correctly, you should be able to access the web console like shown below.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/druid/reconfigure-tls/images/druid-without-tls.png">
</p>

From the above screenshot, we can see that the connection is not secure now. In other words, TLS/SSL is disabled for this druid cluster.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Druid. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls druid-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/druid-ca created
```

Now, Let's create an `Issuer` using the `druid-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: druid-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: druid-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/reconfigure-tls/yamls/druid-issuer.yaml
issuer.cert-manager.io/druid-ca-issuer created
```

### Create DruidOpsRequest

In order to add TLS to the druid, we have to create a `DruidOpsRequest` CRO with our created issuer. Below is the YAML of the `DruidOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-cluster
  tls:
    issuerRef:
      name: druid-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - druid
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `druid-cluster` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on druid.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/druid/concepts/druid.md#spectls).

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/reconfigure-tls/yamls/drops-add-tls.yaml
druidopsrequest.ops.kubedb.com/drops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `DruidOpsRequest` to be `Successful`.  Run the following command to watch `DruidOpsRequest` CRO,

```bash
$  kubectl get drops -n demo -w
NAME            TYPE             STATUS        AGE
drops-add-tls   ReconfigureTLS   Progressing   39s
drops-add-tls   ReconfigureTLS   Progressing   44s
...
...
drops-add-tls   ReconfigureTLS   Successful    79s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe druidopsrequest -n demo drops-add-tls 
Name:         drops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-28T09:43:13Z
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
        f:apply:
        f:databaseRef:
        f:timeout:
        f:tls:
          .:
          f:certificates:
          f:issuerRef:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-28T09:43:13Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2024-10-28T09:44:32Z
  Resource Version:  409889
  UID:               b7f563c4-4773-49e9-aba2-17497e66f5f8
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   druid-cluster
  Timeout:  5m
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          druid
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       druid-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-28T09:43:13Z
    Message:               Druid ops-request has started to reconfigure tls for druid nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-28T09:43:26Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-10-28T09:43:21Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-10-28T09:43:21Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-10-28T09:43:21Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-10-28T09:43:31Z
    Message:               successfully reconciled the Druid with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-28T09:44:32Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-28T09:43:37Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T09:43:37Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T09:43:47Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T09:43:52Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T09:43:52Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T09:43:57Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T09:44:02Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T09:44:02Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T09:44:07Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T09:44:12Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T09:44:12Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T09:44:17Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T09:44:22Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T09:44:22Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T09:44:27Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T09:44:32Z
    Message:               Successfully completed reconfigureTLS for druid.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                           Age   From                         Message
  ----     ------                                                                           ----  ----                         -------
  Normal   Starting                                                                         103s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-add-tls
  Normal   Starting                                                                         103s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                       103s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-add-tls
  Warning  get certificate; ConditionStatus:True                                            95s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      95s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          95s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                            95s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      95s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          95s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                95s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                            90s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      90s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          90s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                            90s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      90s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          90s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                90s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                                    85s   KubeDB Ops-manager Operator  successfully reconciled the Druid with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0               79s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             79s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-0    74s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0     69s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0            64s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0          64s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0  59s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                   54s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                 54s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0         49s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0                   44s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0                 44s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0         39s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0              34s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0            34s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0    29s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Normal   RestartNodes                                                                     24s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                         24s   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                       24s   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-add-tls
```

Now, Lets exec into a druid coordinators pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo druid-cluster-coordinators-0 -- bash
Defaulted container "druid" out of: druid, init-druid (init)
bash-5.1$ cat conf/druid/cluster/_common/common.runtime.properties                                                  
druid.auth.authenticator.basic.authorizerName=basic
druid.auth.authenticator.basic.credentialsValidator.type=metadata
druid.auth.authenticator.basic.initialAdminPassword={"type": "environment", "variable": "DRUID_ADMIN_PASSWORD"}
druid.auth.authenticator.basic.initialInternalClientPassword=password2
druid.auth.authenticator.basic.skipOnFailure=false
druid.auth.authenticator.basic.type=basic
druid.auth.authenticatorChain=["basic"]
druid.auth.authorizer.basic.type=basic
druid.auth.authorizers=["basic"]
druid.client.https.trustStorePassword={"type": "environment", "variable": "DRUID_KEY_STORE_PASSWORD"}
druid.client.https.trustStorePath=/opt/druid/ssl/truststore.jks
druid.client.https.trustStoreType=jks
druid.client.https.validateHostnames=false
druid.emitter.logging.logLevel=info
druid.emitter=noop
druid.enablePlaintextPort=false
druid.enableTlsPort=true
druid.escalator.authorizerName=basic
druid.escalator.internalClientPassword=password2
druid.escalator.internalClientUsername=druid_system
druid.escalator.type=basic
druid.expressions.useStrictBooleans=true
druid.extensions.loadList=["druid-avro-extensions", "druid-kafka-indexing-service", "druid-kafka-indexing-service", "druid-datasketches", "druid-multi-stage-query", "druid-basic-security", "simple-client-sslcontext", "mysql-metadata-storage", "druid-s3-extensions"]
druid.global.http.eagerInitialization=false
druid.host=localhost
druid.indexer.logs.directory=var/druid/indexing-logs
druid.indexer.logs.type=file
druid.indexing.doubleStorage=double
druid.lookup.enableLookupSyncOnStartup=false
druid.metadata.storage.connector.connectURI=jdbc:mysql://druid-cluster-mysql-metadata.demo.svc:3306/druid
druid.metadata.storage.connector.createTables=true
druid.metadata.storage.connector.host=localhost
druid.metadata.storage.connector.password={"type": "environment", "variable": "DRUID_METADATA_STORAGE_PASSWORD"}
druid.metadata.storage.connector.port=1527
druid.metadata.storage.connector.user=root
druid.metadata.storage.type=mysql
druid.monitoring.monitors=["org.apache.druid.java.util.metrics.JvmMonitor", "org.apache.druid.server.metrics.ServiceStatusMonitor"]
druid.s3.accessKey=minio
druid.s3.enablePathStyleAccess=true
druid.s3.endpoint.signingRegion=us-east-1
druid.s3.endpoint.url=http://myminio-hl.demo.svc.cluster.local:9000/
druid.s3.protocol=http
druid.s3.secretKey=minio123
druid.selectors.coordinator.serviceName=druid/coordinator
druid.selectors.indexing.serviceName=druid/overlord
druid.server.hiddenProperties=["druid.s3.accessKey","druid.s3.secretKey","druid.metadata.storage.connector.password", "password", "key", "token", "pwd"]
druid.server.https.certAlias=druid
druid.server.https.keyStorePassword={"type": "environment", "variable": "DRUID_KEY_STORE_PASSWORD"}
druid.server.https.keyStorePath=/opt/druid/ssl/keystore.jks
druid.server.https.keyStoreType=jks
druid.sql.enable=true
druid.sql.planner.useGroupingSetForExactDistinct=true
druid.startup.logging.logProperties=true
druid.storage.baseKey=druid/segments
druid.storage.bucket=druid
druid.storage.storageDirectory=var/druid/segments
druid.storage.type=s3
druid.zk.paths.base=/druid
druid.zk.service.host=druid-cluster-zk.demo.svc:2181
druid.zk.service.pwd={"type": "environment", "variable": "DRUID_ZK_SERVICE_PASSWORD"}
druid.zk.service.user=super

```

We can see from the output above that all TLS related configs are added in the configuration file of the druid database.

#### Verify TLS/SSL using Druid UI

To check follow the following steps:

Druid uses separate ports for TLS/SSL. While the plaintext port for `routers` node is `8888`. For TLS, it is `9088`. Hence, we will use that port to access the UI.

First port-forward the port `9088` to local machine:

```bash
$ kubectl port-forward -n demo svc/druid-cluster-tls-routers 9088
Forwarding from 127.0.0.1:9088 -> 9088
Forwarding from [::1]:9088 -> 9088
```


Now hit the `https://localhost:9088/` from any browser. Here you may select `Advance` and then `Proceed to localhost (unsafe)` or you can add the `ca.crt` from the secret `druid-cluster-tls-client-cert` to your browser's Authorities.

After that you will be prompted to provide the credential of the druid database. By following the steps discussed below, you can get the credential generated by the KubeDB operator for your Druid database.

**Connection information:**

- Username:

  ```bash
  $ kubectl get secret -n demo druid-cluster-tls-admin-cred -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo druid-cluster-tls-admin-cred -o jsonpath='{.data.password}' | base64 -d
  LzJtVRX5E8MorFaf
  ```

After providing the credentials correctly, you should be able to access the web console like shown below.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/druid/reconfigure-tls/images/druid-with-tls.png">
</p>

From the above screenshot, we can see that the connection is secure.


## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
$ kubectl port-forward -n demo svc/druid-cluster-routers 9088
Forwarding from 127.0.0.1:9088 -> 9088
Forwarding from [::1]:9088 -> 9088
Handling connection for 9088
...

$ openssl s_client -connect localhost:9088 2>/dev/null | openssl x509 -noout -enddate
notAfter=Jan 26 09:43:16 2025 GMT
```

So, the certificate will expire on this time `Jan 26 09:43:16 2025 GMT`.

### Create DruidOpsRequest

Now we are going to increase it using a DruidOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druid-recon-tls-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-cluster
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `druid-cluster`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this druid cluster.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/reconfigure-tls/yamls/drops-rotate.yaml
druidopsrequest.ops.kubedb.com/drops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `DruidOpsRequest` to be `Successful`.  Run the following command to watch `DruidOpsRequest` CRO,

```bash
$ kubectl get druidopsrequests -n demo drops-rotate -w
NAME            TYPE             STATUS       AGE
drops-rotate    ReconfigureTLS   Successful   4m4s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe druidopsrequest -n demo drops-rotate
Name:         drops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-28T14:14:50Z
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
        f:apply:
        f:databaseRef:
        f:tls:
          .:
          f:rotateCertificates:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-28T14:14:50Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2024-10-28T14:16:04Z
  Resource Version:  440897
  UID:               ca3532fc-6e11-4962-bddb-f9cf946d3954
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-28T14:14:50Z
    Message:               Druid ops-request has started to reconfigure tls for druid nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-28T14:15:04Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-10-28T14:14:58Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-10-28T14:14:58Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-10-28T14:14:58Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-10-28T14:15:09Z
    Message:               successfully reconciled the Druid with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-28T14:16:04Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-28T14:15:14Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:15:14Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:15:19Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:15:24Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:15:24Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:15:29Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:15:34Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:15:34Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:15:39Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:15:44Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:15:44Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:15:49Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:15:54Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:15:54Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:15:59Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:16:04Z
    Message:               Successfully completed reconfigureTLS for druid.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                           Age   From                         Message
  ----     ------                                                                           ----  ----                         -------
  Normal   Starting                                                                         101s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-rotate
  Normal   Starting                                                                         101s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                       101s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-rotate
  Warning  get certificate; ConditionStatus:True                                            93s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      93s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          93s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                            93s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      93s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          93s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                93s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                            88s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      88s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          88s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                            88s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      88s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          88s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                87s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                                    82s   KubeDB Ops-manager Operator  successfully reconciled the Druid with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0               77s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             77s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0     72s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0            67s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0          67s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0  62s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                   57s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                 57s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0         52s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0                   47s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0                 47s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0         42s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0              37s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0            37s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0    32s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Normal   RestartNodes                                                                     27s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                         27s   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                       27s   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl port-forward -n demo svc/druid-cluster-routers 9088
Forwarding from 127.0.0.1:9088 -> 9088
Forwarding from [::1]:9088 -> 9088
Handling connection for 9088
...

$ openssl s_client -connect localhost:9088 2>/dev/null | openssl x509 -noout -enddate
notAfter=Jan 26 14:15:46 2025 GMT
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
$ kubectl create secret tls druid-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/druid-new-ca created
```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: dr-new-issuer
  namespace: demo
spec:
  ca:
    secretName: druid-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/reconfigure-tls/yamls/druid-new-issuer.yaml
issuer.cert-manager.io/dr-new-issuer created
```

### Create DruidOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `DruidOpsRequest` CRO with the newly created issuer. Below is the YAML of the `DruidOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-cluster
  tls:
    issuerRef:
      name: dr-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `druid-cluster` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our druid.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/reconfigure-tls/yamls/druid-update-tls-issuer.yaml
druidpsrequest.ops.kubedb.com/drops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `DruidOpsRequest` to be `Successful`.  Run the following command to watch `DruidOpsRequest` CRO,

```bash
$ kubectl get druidopsrequests -n demo drops-update-issuer -w
NAME                  TYPE             STATUS        AGE
drops-update-issuer   ReconfigureTLS   Progressing   14s
drops-update-issuer   ReconfigureTLS   Progressing   18s
...
...
drops-update-issuer   ReconfigureTLS   Successful    73s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe druidopsrequest -n demo drops-update-issuer
Name:         drops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-28T14:24:22Z
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
        f:apply:
        f:databaseRef:
        f:tls:
          .:
          f:issuerRef:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-28T14:24:22Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2024-10-28T14:25:35Z
  Resource Version:  442332
  UID:               5089e358-2dc2-4d62-8c13-92828de7c557
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       dr-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-28T14:24:22Z
    Message:               Druid ops-request has started to reconfigure tls for druid nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-28T14:24:35Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-10-28T14:24:30Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-10-28T14:24:30Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-10-28T14:24:30Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-10-28T14:24:40Z
    Message:               successfully reconciled the Druid with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-28T14:25:35Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-28T14:24:45Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:24:45Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:24:50Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:24:55Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:24:55Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:25:00Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:25:05Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:25:05Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:25:10Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:25:15Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:25:15Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:25:20Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:25:25Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:25:25Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:25:30Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:25:35Z
    Message:               Successfully completed reconfigureTLS for druid.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                           Age   From                         Message
  ----     ------                                                                           ----  ----                         -------
  Normal   Starting                                                                         92s   KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-update-issuer
  Normal   Starting                                                                         92s   KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                       92s   KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-update-issuer
  Warning  get certificate; ConditionStatus:True                                            84s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      84s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          84s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                            84s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      84s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          84s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                84s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                            79s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      79s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          79s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                            79s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                      79s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                          79s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                79s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                                    74s   KubeDB Ops-manager Operator  successfully reconciled the Druid with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0               69s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             69s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0     64s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0            59s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0          59s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0  54s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                   49s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                 49s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0         44s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0                   39s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0                 39s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0         34s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0              29s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0            29s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0    24s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Normal   RestartNodes                                                                     19s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                         19s   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                       19s   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-update-issuer
```

Now, Lets exec into a druid node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it druid-cluster-broker-0 -- bash
druid@druid-cluster-broker-0:~$ keytool -list -v -keystore /var/private/ssl/server.keystore.jks -storepass wt6f5pwxpg84 | grep 'Issuer'
Issuer: O=kubedb-updated, CN=ca-updated
Issuer: O=kubedb-updated, CN=ca-updated

$ kubectl port-forward -n demo svc/druid-cluster-routers 9088
Forwarding from 127.0.0.1:9088 -> 9088
Forwarding from [::1]:9088 -> 9088
Handling connection for 9088
...

$ openssl s_client -connect localhost:9088 2>/dev/null | openssl x509 -noout -issuer
issuer=CN = ca-updated, O = kubedb-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a DruidOpsRequest.

### Create DruidOpsRequest

Below is the YAML of the `DruidOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-cluster
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `druid-cluster` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Druid.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/reconfigure-tls/yamls/drops-remove.yaml
druidopsrequest.ops.kubedb.com/drops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `DruidOpsRequest` to be `Successful`.  Run the following command to watch `DruidOpsRequest` CRO,

```bash
$ kubectl get druidopsrequest -n demo drops-remove -w
NAME           TYPE             STATUS        AGE
drops-remove   ReconfigureTLS   Progressing   25s
drops-remove   ReconfigureTLS   Progressing   29s
...
...
drops-remove   ReconfigureTLS   Successful    114s

```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed.

```bash
$  kubectl describe druidopsrequest -n demo drops-remove
Name:         drops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-10-28T14:31:07Z
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
        f:apply:
        f:databaseRef:
        f:tls:
          .:
          f:remove:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-28T14:31:07Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2024-10-28T14:33:01Z
  Resource Version:  443725
  UID:               27234241-c72e-471c-8dd4-16fd485956cc
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-10-28T14:31:07Z
    Message:               Druid ops-request has started to reconfigure tls for druid nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-10-28T14:31:16Z
    Message:               successfully reconciled the Druid with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-28T14:33:01Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-10-28T14:31:21Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:31:21Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:31:26Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-10-28T14:31:31Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:31:31Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:31:36Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-middlemanagers-0
    Last Transition Time:  2024-10-28T14:31:41Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:31:41Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:31:46Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-brokers-0
    Last Transition Time:  2024-10-28T14:31:51Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:31:51Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:31:56Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-routers-0
    Last Transition Time:  2024-10-28T14:32:01Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:32:01Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:32:06Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-10-28T14:33:01Z
    Message:               Successfully completed reconfigureTLS for druid.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                           Age    From                         Message
  ----     ------                                                                           ----   ----                         -------
  Normal   Starting                                                                         2m12s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-remove
  Normal   Starting                                                                         2m12s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                       2m12s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: drops-remove
  Normal   UpdatePetSets                                                                    2m3s   KubeDB Ops-manager Operator  successfully reconciled the Druid with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0               118s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             118s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0     113s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0            108s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0          108s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0  103s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                   98s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                 98s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0         93s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0                   88s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0                 88s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0         83s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0              78s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0            78s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0    73s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0               68s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             68s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0     63s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0            58s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0          58s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0  53s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-middlemanagers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                   48s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0                 48s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0         43s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-brokers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0                   38s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0                 38s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0         33s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-routers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0              28s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0            28s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0    23s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Normal   RestartNodes                                                                     18s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                         18s    KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                       18s    KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: drops-remove
```

Now, Let's exec into one of the broker node and find out that TLS is disabled or not.

```bash
$$ kubectl exec -it -n demo druid-cluster-broker-0 -- druid-configs.sh --bootstrap-server localhost:9092 --command-config /opt/druid/config/clientauth.properties --describe --entity-type brokers --all | grep 'ssl.keystore'
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=null sensitive=false synonyms={}
  ssl.keystore.password=null sensitive=true synonyms={}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=null sensitive=false synonyms={}
  ssl.keystore.password=null sensitive=true synonyms={}
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete opsrequest drops-add-tls drops-remove drops-rotate drops-update-issuer
kubectl delete druid -n demo druid-cluster
kubectl delete issuer -n demo druid-ca-issuer dr-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

