---
title: Run RabbitMQ with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: using-podtemplate-configuration-rm
    name: Customize PodTemplate
    parent: rm-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run RabbitMQ with Custom PodTemplate

KubeDB supports providing custom configuration for RabbitMQ via [PodTemplate](/docs/guides/rabbitmq/concepts/rabbitmq.md#specpodtemplate). This tutorial will show you how to use KubeDB to run a RabbitMQ database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/rabbitmq](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/rabbitmq) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for RabbitMQ database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
  - labels (pod's labels)
- controller:
  - annotations (petset's annotation)
  - labels (petset's labels)
- spec:
  - volumes
  - initContainers
  - containers
  - imagePullSecrets
  - nodeSelector
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

Read about the fields in details in [PodTemplate concept](/docs/guides/rabbitmq/concepts/rabbitmq.md#specpodtemplate),


## CRD Configuration

Below is the YAML for the RabbitMQ created in this example. Here, `spec.podTemplate.spec.containers[].env` specifies additional environment variables by users.

In this tutorial, we will register additional two additional environment variable on rabbitmq bootstrap. These variables will update rabbitmq base log directory and configure the console logs to only view new logs once accessed. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm-misc-config
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  podTemplate:
    spec:
      containers:
        - name: rabbitmq
          env:
            - name: "RABBITMQ_LOG_BASE"
              value: '/var/log/cluster'
            - name: "RABBITMQ_CONSOLE_LOG"
              value: 'new'
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/configuration/rm-misc-config.yaml
rabbitmq.kubedb.com/rm-misc-config created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `rm-misc-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME               READY   STATUS    RESTARTS   AGE
rm-misc-config-0   1/1     Running   0          68s
```

Now, check if the rabbitmq has started with the custom configuration we have provided. We will fetch log in the pod and see the `RABBITMQ_LOG_BASE`, the new log directory exists of not.

```bash
$ kubectl exec -it -n demo  -- bash
  ##  ##      RabbitMQ 3.13.2
  ##  ##
  ##########  Copyright (c) 2007-2024 Broadcom Inc and/or its subsidiaries
  ######  ##
  ##########  Licensed under the MPL 2.0. Website: https://rabbitmq.com

  Erlang:      26.2.5 [jit]
  TLS Library: OpenSSL - OpenSSL 3.1.5 30 Jan 2024
  Release series support status: supported

  Doc guides:  https://www.rabbitmq.com/docs
  Support:     https://www.rabbitmq.com/docs/contact
  Tutorials:   https://www.rabbitmq.com/tutorials
  Monitoring:  https://www.rabbitmq.com/docs/monitoring
  Upgrading:   https://www.rabbitmq.com/docs/upgrade

  Logs: /var/log/rabbitmq/cluster/rabbit@rm-misc-config-0.rm-misc-config-pods.demo.log
        <stdout>
```
So, we can see that that logs are being written to **Logs: /var/log/rabbitmq/cluster**/rabbit@rm-misc-config-0.rm-misc-config-pods.demo.log file.

## Custom Sidecar Containers

Here in this example we will add an extra sidecar container with our RabbitMQ container. Suppose, you are running a KubeDB-managed rabbitmq, and you need to monitor the general logs. We can configure rabbitmq to write those logs in any directory, in the prior example we have configured rabbitmq to write logs to `/var/log/rabbitmq/cluster` directory. In order to export those logs to some remote monitoring solution (such as, Elasticsearch, Logstash, Kafka or Redis) will use a tool like [Filebeat](https://www.elastic.co/beats/filebeat). Filebeat is used to ship logs and files from devices, cloud, containers and hosts. So, it is required to run Filebeat as a sidecar container along with the KubeDB-managed rabbitmq. Here’s a quick demonstration on how to accomplish it.

Firstly, we are going to make our custom filebeat image with our required configuration.
```yaml
filebeat.inputs:
  - type: log
    paths:
      - /var/log/rabbitmq/cluster
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
Now we will deploy our RabbitMQ with custom sidecar container to mount filebeats input directory as a shared directory with rabbitmq's log base directory. Here is the yaml of our RabbitMQ:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq-custom-sidecar
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  podTemplate:
    spec:
      containers:
        - name: rabbitmq
          volumeMounts:
          - mountPath: /var/log/rabbitmq/cluster
            name: log
            readOnly: false
        - name: filebeat
          image: repository_name/custom_filebeat:latest
          volumeMounts:
          - mountPath: /var/log/rabbitmq/cluster
            name: log
            readOnly: true
      volumes:
      - name: log
        emptyDir: {}
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/configuration/rabbitmq-config-sidecar.yaml
rabbitmq.kubedb.com/rabbitmq-custom-sidecar created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `rabbitmq-custom-sidecar-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                        READY   STATUS    RESTARTS      AGE
rabbitmq-custom-sidecar-0   2/2     Running   0             33s

```
Now, Let’s fetch the logs shipped to filebeat console output. The outputs will be generated in json format.

```bash
$ kubectl logs -f -n demo rabbitmq-custom-sidecar-0 -c filebeat
```
We will find the query logs in filebeat console output.
So, we have successfully extracted logs from rabbitmq to our sidecar filebeat container.

## Using Node Selector

Here in this example we will use [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to schedule our RabbitMQ pod to a specific node. Applying nodeSelector to the Pod involves several steps. We first need to assign a label to some node that will be later used by the `nodeSelector` . Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:  

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

Now let's create a RabbitMQ with this new label as nodeSelector. Below is the yaml we are going to apply:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq-node-selector
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  podTemplate:
    spec:
      nodeSelector:
        disktype: ssd
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/configuration/rabbitmq-node-selector.yaml
rabbitmq.kubedb.com/rabbitmq-node-selector created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `rabbitmq-node-selector-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                     READY   STATUS    RESTARTS   AGE
rabbitmq-node-selector-0   1/1     Running   0          60s
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo rabbitmq-node-selector-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo rabbitmq-node-selector-0 -o wide
NAME                     READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
rabbitmq-node-selector-0   1/1     Running   0          3m19s   10.2.1.7   lke212553-307295-5541798e0000   <none>           <none>
```
We can successfully verify that our pod was scheduled to our desired node.

## Using Taints and Tolerations

Here in this example we will use [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to schedule our rabbitmq pod to a specific node and also prevent from scheduling to nodes. Applying taints and tolerations to the Pod involves several steps. Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

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
We can see that our taints were successfully assigned. Now let's try to create a rabbitmq without proper tolerations. Here is the yaml of rabbitmq we are going to create -
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq-without-tolerations
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/configuration/rabbitmq-without-tolerations.yaml
rabbitmq.kubedb.com/rabbitmq-without-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `rabbitmq-without-tolerations-0` has been created and running.

Check that the petset's pod is running or not,
```bash
$ kubectl get pods -n demo
NAME                             READY   STATUS    RESTARTS   AGE
rabbitmq-without-tolerations-0   0/1     Pending   0          3m35s
```
Here we can see that the pod is not running. So let's describe the pod,
```bash
$ kubectl describe pods -n demo rabbitmq-without-tolerations-0 
Name:             rabbitmq-without-tolerations-0
Namespace:        demo
Priority:         0
Service Account:  default
Node:             <none>
Labels:           app.kubernetes.io/component=rabbitmq
                  app.kubernetes.io/instance=rabbitmq-without-tolerations
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=rabbitmqs.kubedb.com
                  apps.kubernetes.io/pod-index=0
                  controller-revision-hash=rabbitmq-without-tolerations-5b85f9cd
                  statefulset.kubernetes.io/pod-name=rabbitmq-without-tolerations-0
Annotations:      <none>
Status:           Pending
IP:               
IPs:              <none>
Controlled By:    PetSet/rabbitmq-without-tolerations
Containers:
  rabbitmq:
    Image:           ghcr.io/appscode-images/rabbitmq:3.13.2@sha256:7f2537e3dc69dae2cebea3500502e6a2b764b42911881e623195eeed32569217
    Ports:           9999/TCP, 9595/TCP
    Host Ports:      0/TCP, 0/TCP
    SeccompProfile:  RuntimeDefault
    Limits:
      memory:  1Gi
    Requests:
      cpu:     500m
      memory:  1Gi
    Mounts:
      /config from rabbitmq-config (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-69qx2 (ro)
Conditions:
  Type           Status
  PodScheduled   False 
Volumes:
  rabbitmq-config:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  rabbitmq-without-tolerations-config
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
Topology Spread Constraints:  kubernetes.io/hostname:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=connection-pooler,app.kubernetes.io/instance=rabbitmq-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=rabbitmqs.kubedb.com
                              topology.kubernetes.io/zone:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=connection-pooler,app.kubernetes.io/instance=rabbitmq-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=rabbitmqs.kubedb.com
Events:
  Type     Reason             Age                   From                Message
  ----     ------             ----                  ----                -------
  Warning  FailedScheduling   5m20s                 default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Warning  FailedScheduling   11s                   default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Normal   NotTriggerScaleUp  13s (x31 over 5m15s)  cluster-autoscaler  pod didn't trigger scale-up:
```
Here we can see that the pod has no tolerations for the tainted nodes and because of that the pod is not able to scheduled.

So, let's add proper tolerations and create another rabbitmq. Here is the yaml we are going to apply,
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq-with-tolerations
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/configuration/rabbitmq-with-tolerations.yaml
rabbitmq.kubedb.com/rabbitmq-with-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `rabbitmq-with-tolerations-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                             READY   STATUS    RESTARTS   AGE
rabbitmq-with-tolerations-0      1/1     Running   0          2m
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo rabbitmq-with-tolerations-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo rabbitmq-with-tolerations-0 -o wide
NAME                        READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
rabbitmq-with-tolerations-0   1/1     Running   0          3m49s   10.2.0.8   lke212553-307295-339173d10000   <none>           <none>
```
We can successfully verify that our pod was scheduled to the node which it has tolerations.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo rm rm-misc-config rabbitmq-custom-sidecar rabbitmq-node-selector rabbitmq-with-tolerations rabbitmq-without-tolerations
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart rabbitmq](/docs/guides/rabbitmq/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your rabbitmq database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md).
- Monitor your rabbitmq database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md).
- Detail concepts of [rabbitmq object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
