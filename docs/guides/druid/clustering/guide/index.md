---
title: Druid Topology Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: guides-druid-clustering-guide
    name: Deploy Druid Cluster
    parent: guides-druid-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Druid Cluster

This tutorial will show you how to use KubeDB to provision a Druid Cluster.

## Before You Begin

Before proceeding:

- Read [druid topology cluster overview](/docs/guides/druid/clustering/overview/index.md) to get a basic idea about the design and architecture of Druid.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Druid=true` to ensure **Druid CRD** and `--set global.featureGates.ZooKeeper=true` to ensure **ZooKeeper CRD** as Druid depends on ZooKeeper for external dependency with helm command.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/druid/clustering/topology-cluster-guide/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/druid/clustering/topology-cluster-guide/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create External Dependency (Deep Storage)

Before proceeding further, we need to prepare deep storage, which is one of the external dependency of Druid and used for storing the segments. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/application-level/examples/deep-storage-config.yaml
secret/deep-storage-config created
```

## Deploy Druid Cluster

The following is an example `Druid` object which creates a Druid cluster of six nodes (coordinators, overlords, brokers, routers, historicals and middleManager). Each with one replica.

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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/clustering/guide/yamls/druid-with-monitoring.yaml
druid.kubedb.com/druid-cluster created
```

KubeDB operator watches for `Druid` objects using Kubernetes API. When a `Druid` object is created, KubeDB operator will create new PetSets and Services with the matching Druid object name. KubeDB operator will also create a governing service for the PetSet with the name `<druid-object-name>-pods`.

```bash
$ kubectl describe druid -n demo druid-cluster
Name:         druid-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Druid
Metadata:
  Creation Timestamp:  2024-10-21T06:01:32Z
  Finalizers:
    kubedb.com/druid
  Generation:  1
  Managed Fields:
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com/druid":
    Manager:      druid-operator
    Operation:    Update
    Time:         2024-10-21T06:01:32Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:deepStorage:
          .:
          f:configSecret:
          f:type:
        f:deletionPolicy:
        f:healthChecker:
          .:
          f:failureThreshold:
          f:periodSeconds:
          f:timeoutSeconds:
        f:topology:
          .:
          f:routers:
            .:
            f:replicas:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-21T06:01:32Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:phase:
    Manager:         druid-operator
    Operation:       Update
    Subresource:     status
    Time:            2024-10-21T06:04:29Z
  Resource Version:  52093
  UID:               a2e12db2-6694-419f-ad07-2c906df5b611
Spec:
  Auth Secret:
    Name:  druid-cluster-admin-cred
  Deep Storage:
    Config Secret:
      Name:         deep-storage-config
    Type:           s3
  Deletion Policy:  Delete
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     30
    Timeout Seconds:    10
  Metadata Storage:
    Create Tables:  true
    Linked DB:      druid
    Name:           druid-cluster-mysql-metadata
    Namespace:      demo
    Type:           MySQL
    Version:        8.0.35
  Topology:
    Brokers:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
    Coordinators:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
    Historicals:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:  1Gi
      Storage Type:   Durable
    Middle Managers:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
            Resources:
              Limits:
                Memory:  2560Mi
              Requests:
                Cpu:     500m
                Memory:  2560Mi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:  1Gi
      Storage Type:   Durable
    Routers:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
  Version:             28.0.1
  Zookeeper Ref:
    Name:       druid-cluster-zk
    Namespace:  demo
    Version:    3.7.2
Status:
  Conditions:
    Last Transition Time:  2024-10-21T06:01:32Z
    Message:               The KubeDB operator has started the provisioning of Druid: demo/druid-cluster
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
  Phase:                   Provisioning
Events:                    <none>

$ kubectl get petset -n demo
NAME                           AGE
druid-cluster-brokers          13m
druid-cluster-coordinators     13m
druid-cluster-historicals      13m
druid-cluster-middlemanagers   13m
druid-cluster-mysql-metadata   14m
druid-cluster-routers          13m
druid-cluster-zk               14m

$ kubectl get pvc -n demo -l app.kubernetes.io/name=druids.kubedb.com
NAME                                                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
druid-cluster-base-task-dir-druid-cluster-middlemanagers-0   Bound    pvc-d288b621-d281-4004-995d-7a25bb4149de   1Gi        RWO            standard       14m
druid-cluster-segment-cache-druid-cluster-historicals-0      Bound    pvc-ccca6be2-658a-46af-a270-de1c6a041af7   1Gi        RWO            standard       14m


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                              STORAGECLASS   REASON   AGE
pvc-4f8538f6-a6ce-4233-b533-8566852f5b98   1Gi        RWO            Delete           Bound    demo/druid-cluster-base-task-dir-druid-cluster-middlemanagers-0    standard                4m39s
pvc-8823d3ad-d614-4172-89ac-c2284a17f502   1Gi        RWO            Delete           Bound    demo/druid-cluster-segment-cache-druid-cluster-historicals-0       standard                4m35s

$ kubectl get service -n demo
NAME                                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                 AGE
druid-cluster-brokers                  ClusterIP   10.96.186.168   <none>        8082/TCP                                                17m
druid-cluster-coordinators             ClusterIP   10.96.122.235   <none>        8081/TCP                                                17m
druid-cluster-mysql-metadata           ClusterIP   10.96.109.2     <none>        3306/TCP                                                18m
druid-cluster-mysql-metadata-pods      ClusterIP   None            <none>        3306/TCP                                                18m
druid-cluster-mysql-metadata-standby   ClusterIP   10.96.97.152    <none>        3306/TCP                                                18m
druid-cluster-pods                     ClusterIP   None            <none>        8081/TCP,8090/TCP,8083/TCP,8091/TCP,8082/TCP,8888/TCP   17m
druid-cluster-routers                  ClusterIP   10.96.138.237   <none>        8888/TCP                                                17m
druid-cluster-zk                       ClusterIP   10.96.148.251   <none>        2181/TCP                                                18m
druid-cluster-zk-admin-server          ClusterIP   10.96.2.106     <none>        8080/TCP                                                18m
druid-cluster-zk-pods                  ClusterIP   None            <none>        2181/TCP,2888/TCP,3888/TCP                              18m
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified `Druid` object:

```bash
$ kubectl describe druid -n demo druid-cluster
Name:         druid-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Druid
Metadata:
  Creation Timestamp:  2024-10-21T06:01:32Z
  Finalizers:
    kubedb.com/druid
  Generation:  1
  Managed Fields:
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
          .:
          v:"kubedb.com/druid":
    Manager:      druid-operator
    Operation:    Update
    Time:         2024-10-21T06:01:32Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:deepStorage:
          .:
          f:configSecret:
          f:type:
        f:deletionPolicy:
        f:healthChecker:
          .:
          f:failureThreshold:
          f:periodSeconds:
          f:timeoutSeconds:
        f:topology:
          .:
          f:routers:
            .:
            f:replicas:
        f:version:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2024-10-21T06:01:32Z
    API Version:  kubedb.com/v1alpha2
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:phase:
    Manager:         druid-operator
    Operation:       Update
    Subresource:     status
    Time:            2024-10-21T06:04:29Z
  Resource Version:  52093
  UID:               a2e12db2-6694-419f-ad07-2c906df5b611
Spec:
  Auth Secret:
    Name:  druid-cluster-admin-cred
  Deep Storage:
    Config Secret:
      Name:         deep-storage-config
    Type:           s3
  Deletion Policy:  Delete
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     30
    Timeout Seconds:    10
  Metadata Storage:
    Create Tables:  true
    Linked DB:      druid
    Name:           druid-cluster-mysql-metadata
    Namespace:      demo
    Type:           MySQL
    Version:        8.0.35
  Topology:
    Brokers:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
    Coordinators:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
    Historicals:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:  1Gi
      Storage Type:   Durable
    Middle Managers:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
            Resources:
              Limits:
                Memory:  2560Mi
              Requests:
                Cpu:     500m
                Memory:  2560Mi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:  1Gi
      Storage Type:   Durable
    Routers:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  init-druid
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
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        1
  Version:             28.0.1
  Zookeeper Ref:
    Name:       druid-cluster-zk
    Namespace:  demo
    Version:    3.7.2
Status:
  Conditions:
    Last Transition Time:  2024-10-21T06:01:32Z
    Message:               The KubeDB operator has started the provisioning of Druid: demo/druid-cluster
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-10-21T06:03:03Z
    Message:               Database dependency is ready
    Observed Generation:   1
    Reason:                DatabaseDependencyReady
    Status:                True
    Type:                  DatabaseDependencyReady
    Last Transition Time:  2024-10-21T06:03:34Z
    Message:               All desired replicas are ready.
    Observed Generation:   1
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-10-21T06:04:04Z
    Message:               The Druid: demo/druid-cluster is accepting client requests and nodes formed a cluster
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-10-21T06:04:29Z
    Message:               The Druid: demo/druid-cluster is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-10-21T06:04:29Z
    Message:               The Druid: demo/druid-cluster is successfully provisioned.
    Observed Generation:   1
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```


## Connect with Druid Database
We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our routers of the Druid database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our Druid database is working well. It is also possible to use `External-IP` to access druid nodes if you make `service` type of that node as `LoadBalancer`.

### Check the Service Health

Let's port-forward the port `8888` to local machine:

```bash
$ kubectl port-forward -n demo svc/druid-cluster-routers 8888
Forwarding from 127.0.0.1:8888 -> 8888
Forwarding from [::1]:8888 -> 8888
```

Now, the Druid cluster is accessible at `localhost:8888`. Let's check the [Service Health](https://druid.apache.org/docs/latest/api-reference/service-status-api/#get-service-health) of Routers of the Druid database.

```bash
$ curl "http://localhost:8888/status/health"
true
```
From the retrieved health information above, we can see that our Druid cluster’s status is `true`,  indicating that the service can receive API calls and is healthy. In the same way it possible to check the health of other druid nodes by port-forwarding the appropriate services.

### Access the web console

We can also access the [web console](https://druid.apache.org/docs/latest/operations/web-console) of Druid database from any browser by port-forwarding the routers in the same way shown in the aforementioned step or directly using the `External-IP` if the router service type is `LoadBalancer`.

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
  <img alt="lifecycle"  src="/docs/images/druid/Druid-Web-Console.png">
</p>

You can use this web console for loading data, managing datasources and tasks, and viewing server status and segment information. You can also run SQL and native Druid queries in the console.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo druid druid-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kafka.kubedb.com/druid-cluster patched

$ kubectl delete dr druid-cluster  -n demo
druid.kubedb.com "druid-cluster" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Detail concepts of [DruidDBVersion object](/docs/guides/druid/concepts/druidversion.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
