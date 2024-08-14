---
title: Run Pgpool with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: using-podtemplate-configuration-pp
    name: Customize PodTemplate
    parent: pp-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Pgpool with Custom PodTemplate

KubeDB supports providing custom configuration for Pgpool via [PodTemplate](/docs/guides/pgpool/concepts/pgpool.md#specpodtemplate). This tutorial will show you how to use KubeDB to run a Pgpool database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Pgpool database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
  - labels (pod's labels)
- controller:
  - annotations (statefulset's annotation)
  - labels (statefulset's labels)
- spec:
  - volumes
  - initContainers
  - containers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - livenessProbe
  - readinessProbe
  - lifecycle

Read about the fields in details in [PodTemplate concept](/docs/guides/pgpool/concepts/pgpool.md#specpodtemplate),

## Prepare Postgres
For a Pgpool surely we will need a Postgres server so, prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.


## CRD Configuration

Below is the YAML for the Pgpool created in this example. Here, `spec.podTemplate.spec.containers[].env` specifies additional environment variables by users.

In this tutorial, we will register additional two users at starting time of Pgpool. So, the fact is any environment variable with having `suffix: USERNAME` and `suffix: PASSWORD` will be key value pairs of username and password and will be registered in the `pool_passwd` file of Pgpool. So we can use these users after Pgpool initialize without even syncing them.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-misc-config
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      containers:
        - name: pgpool
          env:
            - name: "ALICE_USERNAME"
              value: alice
            - name: "ALICE_PASSWORD"
              value: '123'
            - name: "BOB_USERNAME"
              value: bob
            - name: "BOB_PASSWORD"
              value: '456'
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pp-misc-config.yaml
pgpool.kubedb.com/pp-misc-config created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pp-misc-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME               READY   STATUS    RESTARTS   AGE
pp-misc-config-0   1/1     Running   0          68s
```

Now, check if the pgpool has started with the custom configuration we have provided. We will exec in the pod and see the `pool_passwd` file if the user exists of not. We will also see if the environment variable is set or not.

```bash
$ kubectl exec -it -n demo pp-misc-config-0 -- bash
pp-misc-config-0:/$ echo $BOB_USERNAME
bob
pp-misc-config-0:/$ echo $BOB_PASSWORD
456
pp-misc-config-0:/$ echo $ALICE_USERNAME
alice
pp-misc-config-0:/$ echo $ALICE_PASSWORD
123
pp-misc-config-0:/$ cat opt/pgpool-II/etc/pool_passwd 
postgres:AESNz9O12b8N9Ngz1SpCYymv2K8wkHMWS+5TICOsbR5W1U=
bob:AESBw7fOtf4SCfFiI7vbAYpKg==
alice:AESgda2WBFwHQfKluCkXwo+MA==
pp-misc-config-0:/$ exit
exit
```
So, we can see that the additional two users Alice and Bob is successfully registered. Now we can use them. So, first let create the users through the root user postgres.

Now, you can connect to this pgpool through [psql](https://www.postgresql.org/docs/current/app-psql.html). Before that we need to port-forward to the primary service of pgpool.

```bash
$ kubectl port-forward -n demo svc/pp-misc-config 9999
Forwarding from 127.0.0.1:9999 -> 9999
```
Now, let's get the password for the root user.
```bash
$ kubectl get secrets -n demo ha-postgres-auth -o jsonpath='{.data.\password}' | base64 -d
qEeuU6cu5aH!O9CI⏎ 
```
We can use this password now,
```bash
$ psql --host=localhost --port=9999 --username=postgres postgres
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=# CREATE USER alice WITH PASSWORD '123';
CREATE ROLE
postgres=# CREATE USER bob WITH PASSWORD '456';
CREATE ROLE
postgres=# exit
```

Now, let's verify if we can to the database through pgpool with the new users,
```bash
$ export PGPASSWORD='123'
$ psql --host=localhost --port=9999 --username=alice postgres                                    
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=> exit
$ export PGPASSWORD='456'
$ psql --host=localhost --port=9999 --username=bob postgres
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=> exit
```

You can see we can use these new users to connect to the database.

## Custom Sidecar Containers

Here in this example we will add an extra sidecar container with our pgpool container. Suppose, you are running a KubeDB-managed Pgpool, and you need to monitor the general logs. We can configure pgpool to write those logs in any directory, in this example we will configure pgpool to write logs to `/tmp/pgpool_log` directory with file name format `pgpool-%Y-%m-%d_%H%M%S.log`. In order to export those logs to some remote monitoring solution (such as, Elasticsearch, Logstash, Kafka or Redis) will need a tool like [Filebeat](https://www.elastic.co/beats/filebeat). Filebeat is used to ship logs and files from devices, cloud, containers and hosts. So, it is required to run Filebeat as a sidecar container along with the KubeDB-managed Pgpool. Here’s a quick demonstration on how to accomplish it.

Firstly, we are going to make our custom filebeat image with our required configuration.
```yaml
filebeat.inputs:
  - type: log
    paths:
      - /tmp/pgpool_log/*.log
output.console:
  pretty: true
```
Save this yaml with name `filebeat.yml`. Now prepare the dockerfile,
```dockerfile
FROM elastic/filebeat:7.17.1
COPY filebeat.yml /usr/share/filebeat
USER root
RUN chmod go-w /usr/share/filebeat/filebeat.yml
USER filebeat
```
Now run these following commands to build and push the docker image to your docker repository.
```bash
$ docker build -t repository_name/custom_filebeat:latest .
$ docker push repository_name/custom_filebeat:latest
```
Now we will deploy our pgpool with custom sidecar container and will also use the `spec.initConfig` to configure the logs related settings. Here is the yaml of our pgpool:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-custom-sidecar
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      containers:
        - name: pgpool
          volumeMounts:
          - mountPath: /tmp/pgpool_log
            name: data
            readOnly: false
        - name: filebeat
          image: repository_name/custom_filebeat:latest
          volumeMounts:
          - mountPath: /tmp/pgpool_log
            name: data
            readOnly: true
      volumes:
      - name: data
        emptyDir: {}
  initConfig:
    pgpoolConfig:
      log_destination : 'stderr'
      logging_collector : on
      log_directory : '/tmp/pgpool_log'
      log_filename : 'pgpool-%Y-%m-%d_%H%M%S.log'
      log_file_mode : 0777
      log_truncate_on_rotation : off
      log_rotation_age : 1d
      log_rotation_size : 10MB
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pgpool-config-sidecar.yaml
pgpool.kubedb.com/pgpool-custom-sidecar created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pgpool-custom-sidecar-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                      READY   STATUS    RESTARTS      AGE
pgpool-custom-sidecar-0   2/2     Running   0             33s

```

Now, Let’s fetch the logs shipped to filebeat console output. The outputs will be generated in json format.

```bash
$ kubectl logs -f -n demo pgpool-custom-sidecar-0 -c filebeat
```
We will find the query logs in filebeat console output. Sample output:
```json
{
  "@timestamp": "2024-08-14T06:14:38.461Z",
  "@metadata": {
    "beat": "filebeat",
    "type": "_doc",
    "version": "7.17.1"
  },
  "host": {
    "name": "pgpool-custom-sidecar-0"
  },
  "agent": {
    "ephemeral_id": "17afa770-9fe2-450c-a4fd-eae1301fa3f5",
    "id": "3833c41c-e37c-49d7-9881-bf4a4796d31d",
    "name": "pgpool-custom-sidecar-0",
    "type": "filebeat",
    "version": "7.17.1",
    "hostname": "pgpool-custom-sidecar-0"
  },
  "log": {
    "offset": 2913,
    "file": {
      "path": "/tmp/pgpool_log/pgpool-2024-08-14_061421.log"
    }
  },
  "message": "2024-08-14 06:14:33.919: [unknown] pid 70: LOG:  pool_send_and_wait: Error or notice message from backend: : DB node id: 0 backend pid: 20986 statement: \"create table if not exists kubedb_write_check_pgpool (health_key varchar(50) NOT NULL, health_value varchar(50) NOT NULL, PRIMARY KEY (health_key));\" message: \"relation \"kubedb_write_check_pgpool\" already exists, skipping\"",
  "input": {
    "type": "log"
  },
  "ecs": {
    "version": "1.12.0"
  }
}
```
So, we have successfully extracted logs from pgpool to our sidecar filebeat container.

## Using Node Selector

Here in this example we will use [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to schedule our pgpool pod to a specific node. Applying nodeSelector to the Pod involves several steps. We first need to assign a label to some node that will be later used by the `nodeSelector` . Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:  

```bash
$ kubectl get nodes --show-labels
NAME                            STATUS   ROLES    AGE   VERSION   LABELS
lke212553-307295-339173d10000   Ready    <none>   36m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-339173d10000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=618158120a299c6fd37f00d01d355ca18794c467,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
lke212553-307295-5541798e0000   Ready    <none>   36m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-5541798e0000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=75cfe3dbbb0380f1727efc53f5192897485e95d5,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
lke212553-307295-5b53c5520000   Ready    <none>   36m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-5b53c5520000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=792bac078d7ce0e548163b9423416d7d8c88b08f,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
```
As you see, we have three nodes in the cluster: lke212553-307295-339173d10000, lke212553-307295-5541798e0000, and lke212553-307295-5b53c5520000.

Next, select a node to which you want to add a label. For example, let’s say we want to add a new label with the key `disktype` and value ssd to the `lke212553-307295-5541798e0000` node, which is a node with the SSD storage. To do so, run:
```bash
$ kubectl label nodes lke212553-307295-5541798e0000 disktype=ssd
node/lke212553-307295-5541798e0000 labeled
```
As you noticed, the command above follows the format `kubectl label nodes <node-name> <label-key>=<label-value>` .
Finally, let’s verify that the new label was added by running:
```bash
 $ kubectl get nodes --show-labels                                                                                                                                                                  
NAME                            STATUS   ROLES    AGE   VERSION   LABELS
lke212553-307295-339173d10000   Ready    <none>   41m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-339173d10000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=618158120a299c6fd37f00d01d355ca18794c467,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
lke212553-307295-5541798e0000   Ready    <none>   41m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,disktype=ssd,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-5541798e0000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=75cfe3dbbb0380f1727efc53f5192897485e95d5,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
lke212553-307295-5b53c5520000   Ready    <none>   41m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-5b53c5520000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=792bac078d7ce0e548163b9423416d7d8c88b08f,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
```
As you see, the lke212553-307295-5541798e0000 now has a new label disktype=ssd. To see all labels attached to the node, you can also run:
```bash
$ kubectl describe node "lke212553-307295-5541798e0000"
Name:               lke212553-307295-5541798e0000
Roles:              <none>
Labels:             beta.kubernetes.io/arch=amd64
                    beta.kubernetes.io/instance-type=g6-dedicated-4
                    beta.kubernetes.io/os=linux
                    disktype=ssd
                    failure-domain.beta.kubernetes.io/region=ap-south
                    kubernetes.io/arch=amd64
                    kubernetes.io/hostname=lke212553-307295-5541798e0000
                    kubernetes.io/os=linux
                    lke.linode.com/pool-id=307295
                    node.k8s.linode.com/host-uuid=75cfe3dbbb0380f1727efc53f5192897485e95d5
                    node.kubernetes.io/instance-type=g6-dedicated-4
                    topology.kubernetes.io/region=ap-south
                    topology.linode.com/region=ap-south
```
Along with the `disktype=ssd` label we’ve just added, you can see other labels such as `beta.kubernetes.io/arch` or `kubernetes.io/hostname`. These are all default labels attached to Kubernetes nodes.

Now let's create a pgpool with this new label as nodeSelector. Below is the yaml we are going to apply:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-node-selector
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      nodeSelector:
        disktype: ssd
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pgpool-node-selector.yaml
pgpool.kubedb.com/pgpool-node-selector created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pgpool-node-selector-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                     READY   STATUS    RESTARTS   AGE
pgpool-node-selector-0   1/1     Running   0          60s
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo pgpool-node-selector-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo pgpool-node-selector-0 -o wide
NAME                     READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
pgpool-node-selector-0   1/1     Running   0          3m19s   10.2.1.7   lke212553-307295-5541798e0000   <none>           <none>
```
We can successfully verify that our pod was scheduled to our desired node.

## Using Taints and Tolerations

Here in this example we will use [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to schedule our pgpool pod to a specific node and also prevent from scheduling to nodes. Applying taints and tolerations to the Pod involves several steps. Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

```bash
$ kubectl get nodes --show-labels
NAME                            STATUS   ROLES    AGE   VERSION   LABELS
lke212553-307295-339173d10000   Ready    <none>   36m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-339173d10000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=618158120a299c6fd37f00d01d355ca18794c467,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
lke212553-307295-5541798e0000   Ready    <none>   36m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-5541798e0000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=75cfe3dbbb0380f1727efc53f5192897485e95d5,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
lke212553-307295-5b53c5520000   Ready    <none>   36m   v1.30.3   beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=g6-dedicated-4,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=ap-south,kubernetes.io/arch=amd64,kubernetes.io/hostname=lke212553-307295-5b53c5520000,kubernetes.io/os=linux,lke.linode.com/pool-id=307295,node.k8s.linode.com/host-uuid=792bac078d7ce0e548163b9423416d7d8c88b08f,node.kubernetes.io/instance-type=g6-dedicated-4,topology.kubernetes.io/region=ap-south,topology.linode.com/region=ap-south
```
As you see, we have three nodes in the cluster: lke212553-307295-339173d10000, lke212553-307295-5541798e0000, and lke212553-307295-5b53c5520000.

Next, we are going to taint these nodes.
```bash
$ kubectl taint nodes lke212553-307295-339173d10000 key1=node1:NoSchedule
node/lke212553-307295-339173d10000 tainted

$ kubectl taint nodes lke212553-307295-5541798e0000 key1=node2:NoSchedule
node/lke212553-307295-5541798e0000 tainted

$ kubectl taint nodes lke212553-307295-5b53c5520000 key1=node3:NoSchedule
node/lke212553-307295-5b53c5520000 tainted
```
Let's see our tainted nodes here,
```bash
$ kubectl get nodes -o json | jq -r '.items[] | select(.spec.taints != null) | .metadata.name, .spec.taints'
lke212553-307295-339173d10000
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node1"
  }
]
lke212553-307295-5541798e0000
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node2"
  }
]
lke212553-307295-5b53c5520000
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node3"
  }
]
```
We can see that our taints were successfully assigned. Now let's try to create a pgpool without proper tolerations. Here is the yaml of pgpool we are going to createc
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-without-tolerations
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pgpool-without-tolerations.yaml
pgpool.kubedb.com/pgpool-without-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pgpool-without-tolerations-0` has been created and running.

Check that the petset's pod is running or not,
```bash
$ kubectl get pods -n demo
NAME                           READY   STATUS    RESTARTS   AGE
pgpool-without-tolerations-0   0/1     Pending   0          3m35s
```
Here we can see that the pod is not running. So let's describe the pod,
```bash
$ kubectl describe pods -n demo pgpool-without-tolerations-0 
Name:             pgpool-without-tolerations-0
Namespace:        demo
Priority:         0
Service Account:  default
Node:             <none>
Labels:           app.kubernetes.io/component=connection-pooler
                  app.kubernetes.io/instance=pgpool-without-tolerations
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=pgpools.kubedb.com
                  apps.kubernetes.io/pod-index=0
                  controller-revision-hash=pgpool-without-tolerations-5b85f9cd
                  statefulset.kubernetes.io/pod-name=pgpool-without-tolerations-0
Annotations:      <none>
Status:           Pending
IP:               
IPs:              <none>
Controlled By:    PetSet/pgpool-without-tolerations
Containers:
  pgpool:
    Image:           ghcr.io/appscode-images/pgpool2:4.4.5@sha256:7f2537e3dc69dae2cebea3500502e6a2b764b42911881e623195eeed32569217
    Ports:           9999/TCP, 9595/TCP
    Host Ports:      0/TCP, 0/TCP
    SeccompProfile:  RuntimeDefault
    Limits:
      memory:  1Gi
    Requests:
      cpu:     500m
      memory:  1Gi
    Environment:
      POSTGRES_USERNAME:                  postgres
      POSTGRES_PASSWORD:                  5ja8dHF79x4o6Ot6
      PGPOOL_PCP_USER:                    <set to the key 'username' in secret 'pgpool-without-tolerations-auth'>  Optional: false
      PGPOOL_PCP_PASSWORD:                <set to the key 'password' in secret 'pgpool-without-tolerations-auth'>  Optional: false
      PGPOOL_PASSWORD_ENCRYPTION_METHOD:  scram-sha-256
      PGPOOL_ENABLE_POOL_PASSWD:          true
      PGPOOL_SKIP_PASSWORD_ENCRYPTION:    false
    Mounts:
      /config from pgpool-config (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-69qx2 (ro)
Conditions:
  Type           Status
  PodScheduled   False 
Volumes:
  pgpool-config:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  pgpool-without-tolerations-config
    Optional:    false
  kube-api-access-69qx2:
    Type:                     Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:   3607
    ConfigMapName:            kube-root-ca.crt
    ConfigMapOptional:        <nil>
    DownwardAPI:              true
QoS Class:                    Burstable
Node-Selectors:               <none>
Tolerations:                  node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                              node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Topology Spread Constraints:  kubernetes.io/hostname:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=connection-pooler,app.kubernetes.io/instance=pgpool-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=pgpools.kubedb.com
                              topology.kubernetes.io/zone:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=connection-pooler,app.kubernetes.io/instance=pgpool-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=pgpools.kubedb.com
Events:
  Type     Reason             Age                   From                Message
  ----     ------             ----                  ----                -------
  Warning  FailedScheduling   5m20s                 default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Warning  FailedScheduling   11s                   default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Normal   NotTriggerScaleUp  13s (x31 over 5m15s)  cluster-autoscaler  pod didn't trigger scale-up:
```
Here we can see that the pod has no tolerations for the tainted nodes and because of that the pod is not able to scheduled.

So, let's add proper tolerations and create another pgpool. Here is the yaml we are going to apply,
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-with-tolerations
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      tolerations:
      - key: "key1"
        operator: "Equal"
        value: "node1"
        effect: "NoSchedule"
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pgpool-with-tolerations.yaml
pgpool.kubedb.com/pgpool-with-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pgpool-with-tolerations-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                           READY   STATUS    RESTARTS   AGE
pgpool-with-tolerations-0      1/1     Running   0          2m
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo pgpool-with-tolerations-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo pgpool-with-tolerations-0 -o wide
NAME                        READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
pgpool-with-tolerations-0   1/1     Running   0          3m49s   10.2.0.8   lke212553-307295-339173d10000   <none>           <none>
```
We can successfully verify that our pod was scheduled to the node which it has tolerations.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pp pp-misc-config pgpool-custom-sidecar pgpool-node-selector pgpool-with-tolerations pgpool-without-tolerations
kubectl delete -n demo pg/ha-postgres
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart Pgpool](/docs/guides/pgpool/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
