---
title: Run Druid with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: guides-druid-configuration-podtemplating
    name: Customize PodTemplate
    parent: guides-druid-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Druid with Custom PodTemplate

KubeDB supports providing custom configuration for Druid via [PodTemplate](/docs/guides/druid/concepts/druid.md#spec.topology). This tutorial will show you how to use KubeDB to run a Druid database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/druid/configuration/podtemplating/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/druid/configuration/podtemplating/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for `leaf` and `aggregator` pod through `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate`. KubeDB operator will pass the information provided in `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate` to the `aggregator` and `leaf` PetSet created for Druid database.
KubeDB allows providing a template for all the druid pods through `spec.topology.<node-name>.podTemplate`. KubeDB operator will pass the information provided in `spec.topology.<node-name>.podTemplate` to the corresponding PetSet created for Druid database.

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

Read about the fields in details in [PodTemplate concept](/docs/guides/druid/concepts/druid.md#spectopology),


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

## CRD Configuration

Below is the YAML for the Druid created in this example. Here, [`spec.topology.aggregator/leaf.podTemplate.spec.args`](/docs/guides/mysql/concepts/database/index.md#specpodtemplatespecargs) provides extra arguments.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: 28.0.1
  configSecret:
    name: config-secret
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    coordinators:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"
    brokers:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: druid
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"
    routers:
      replicas: 1
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/podtemplating/yamls/druid-cluster.yaml
druid.kubedb.com/druid-cluster created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that `druid-cluster` is in `Ready` state.
```bash
$ kubectl get druid -n demo
NAME            TYPE                  VERSION   STATUS   AGE
druid-cluster   kubedb.com/v1alpha2   28.0.1    Ready    6m5s
```

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=druid-cluster
NAME                             READY   STATUS    RESTARTS   AGE
druid-cluster-brokers-0          1/1     Running   0          7m2s
druid-cluster-coordinators-0     1/1     Running   0          7m9s
druid-cluster-historicals-0      1/1     Running   0          7m7s
druid-cluster-middlemanagers-0   1/1     Running   0          7m5s
druid-cluster-routers-0          1/1     Running   0          7m
```

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl get pod -n demo druid-cluster-coordinators-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "2Gi"
  }
}

$ kubectl get pod -n demo druid-cluster-brokers-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "2Gi"
  }
}
```

Here we can see the containers of the both `coordinators` and `brokers` have the resources we have specified in the manifest.

## Using Node Selector

Here in this example we will use [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to schedule our druid pod to a specific node. Applying nodeSelector to the Pod involves several steps. We first need to assign a label to some node that will be later used by the `nodeSelector` . Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

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

Now let's create a druid with this new label as nodeSelector. Below is the yaml we are going to apply:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-node-selector
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
    coordinators:
      podTemplate:
        spec:
          nodeSelector:
            disktype: ssd
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/podtemplating/yamls/druid-node-selector.yaml
druid.kubedb.com/druid-node-selector created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that the `druid-node-selector` instance is in `Ready` state.

```bash
$ kubectl get druid -n demo
NAME                  TYPE                  VERSION   STATUS   AGE
druid-node-selector   kubedb.com/v1alpha2   28.0.1    Ready    54m
```
You can verify that by running `kubectl get pods -n demo druid-node-selector-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo druid-node-selector-cooridnators-0 -o wide
NAME                                    READY   STATUS    RESTARTS   AGE     IP            NODE                         NOMINATED NODE   READINESS GATES
druid-node-selector-cooridnators-0      1/1     Running   0          3m19s   10.2.1.7   lke212553-307295-5541798e0000   <none>           <none>
```
We can successfully verify that our pod was scheduled to our desired node.

## Using Taints and Tolerations

Here in this example we will use [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to schedule our druid pod to a specific node and also prevent from scheduling to nodes. Applying taints and tolerations to the Pod involves several steps. Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

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
We can see that our taints were successfully assigned. Now let's try to create a druid without proper tolerations. Here is the yaml of druid we are going to create.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-without-tolerations
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/podtemplating/yamls/druid-without-tolerations.yaml
druid.kubedb.com/druid-without-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `druid-without-tolerations-0` has been created and running.

Check that the petset's pod is running or not,
```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=druid-without-tolerations
NAME                                          READY   STATUS    RESTARTS   AGE
druid-without-tolerations-brokers-0           0/1     Pending   0          3m35s
druid-without-tolerations-cooridnators-0      0/1     Pending   0          3m35s
druid-without-tolerations-historicals-0       0/1     Pending   0          3m35s
druid-without-tolerations-middlemanager-0     0/1     Pending   0          3m35s
druid-without-tolerations-routers-0           0/1     Pending   0          3m35s
```
Here we can see that the pod is not running. So let's describe the pod,
```bash
$ kubectl describe pods -n demo druid-without-tolerations-coordinators-0 
Name:             druid-without-tolerations-coordinators-0
Namespace:        demo
Priority:         0
Service Account:  default
Node:             kind-control-plane/172.18.0.2
Start Time:       Wed, 13 Nov 2024 11:59:06 +0600
Labels:           app.kubernetes.io/component=database
                  app.kubernetes.io/instance=druid-without-tolerations
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=druids.kubedb.com
                  apps.kubernetes.io/pod-index=0
                  controller-revision-hash=druid-without-tolerations-coordinators-65c8c99fc7
                  kubedb.com/role=coordinators
                  statefulset.kubernetes.io/pod-name=druid-without-tolerations-coordinators-0
Annotations:      <none>
Status:           Running
IP:               10.244.0.53
IPs:
  IP:           10.244.0.53
Controlled By:  PetSet/druid-without-tolerations-coordinators
Init Containers:
  init-druid:
    Container ID:   containerd://62c9a2053d619dded2085e354cd2c0dfa238761033cc0483c824c1ed8ee4c002
    Image:          ghcr.io/kubedb/druid-init:28.0.1@sha256:ed87835bc0f89dea923fa8e3cf1ef209e3e41cb93944a915289322035dcd8a91
    Image ID:       ghcr.io/kubedb/druid-init@sha256:ed87835bc0f89dea923fa8e3cf1ef209e3e41cb93944a915289322035dcd8a91
    Port:           <none>
    Host Port:      <none>
    State:          Terminated
      Reason:       Completed
      Exit Code:    0
      Started:      Wed, 13 Nov 2024 11:59:07 +0600
      Finished:     Wed, 13 Nov 2024 11:59:07 +0600
    Ready:          True
    Restart Count:  0
    Limits:
      memory:  512Mi
    Requests:
      cpu:     200m
      memory:  512Mi
    Environment:
      DRUID_METADATA_TLS_ENABLE:    false
      DRUID_METADATA_STORAGE_TYPE:  MySQL
    Mounts:
      /opt/druid/conf from main-config-volume (rw)
      /opt/druid/extensions/mysql-metadata-storage from mysql-metadata-storage (rw)
      /tmp/config/custom-config from custom-config (rw)
      /tmp/config/operator-config from operator-config-volume (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-9t5kp (ro)
Containers:
  druid:
    Container ID:  containerd://3a52f120ca09f90fcdc062c94bf404964add7a5b6ded4a372400267a9d0fd598
    Image:         ghcr.io/appscode-images/druid:28.0.1@sha256:d86e424233ec5a120c1e072cf506fa169868fd9572bbb9800a85400f0c879dec
    Image ID:      ghcr.io/appscode-images/druid@sha256:d86e424233ec5a120c1e072cf506fa169868fd9572bbb9800a85400f0c879dec
    Port:          8081/TCP
    Host Port:     0/TCP
    Command:
      /druid.sh
      coordinator
    State:          Running
      Started:      Wed, 13 Nov 2024 11:59:09 +0600
    Ready:          True
    Restart Count:  0
    Limits:
      cpu:     600m
      memory:  2Gi
    Requests:
      cpu:     600m
      memory:  2Gi
    Environment:
      DRUID_ADMIN_PASSWORD:             <set to the key 'password' in secret 'druid-without-tolerations-auth'>  Optional: false
      DRUID_METADATA_STORAGE_PASSWORD:  VHJ6!hFuT8WDjcyy
      DRUID_ZK_SERVICE_PASSWORD:        VHJ6!hFuT8WDjcyy
    Mounts:
      /opt/druid/conf from main-config-volume (rw)
      /opt/druid/extensions/mysql-metadata-storage from mysql-metadata-storage (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-9t5kp (ro)

Conditions:
  Type                        Status
  PodReadyToStartContainers   True 
  Initialized                 True 
  Ready                       True 
  ContainersReady             True 
  PodScheduled                True 
Volumes:
  data:
    Type:       PersistentVolumeClaim (a reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  data-druid-without-tolerations-0
    ReadOnly:   false
  init-scripts:
    Type:       EmptyDir (a temporary directory that shares a pod's lifetime)
    Medium:     
    SizeLimit:  <unset>
  kube-api-access-htm2z:
    Type:                     Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:   3607
    ConfigMapName:            kube-root-ca.crt
    ConfigMapOptional:        <nil>
    DownwardAPI:              true
QoS Class:                    Burstable
Node-Selectors:               <none>
Tolerations:                  node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                              node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Topology Spread Constraints:  kubernetes.io/hostname:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,kubedb.com/petset=standalone
                              topology.kubernetes.io/zone:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,kubedb.com/petset=standalone
Events:
  Type     Reason             Age                   From                Message
  ----     ------             ----                  ----                -------
  Warning  FailedScheduling   5m20s                 default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Warning  FailedScheduling   11s                   default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Normal   NotTriggerScaleUp  13s (x31 over 5m15s)  cluster-autoscaler  pod didn't trigger scale-up:
```
Here we can see that the pod has no tolerations for the tainted nodes and because of that the pod is not able to scheduled.

So, let's add proper tolerations and create another druid. Here is the yaml we are going to apply,
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
      podTemplate:
        spec:
          tolerations:
            - key: "key1"
              operator: "Equal"
              value: "node1"
              effect: "NoSchedule"
      replicas: 1
    coordinators:
      podTemplate:
        spec:
          tolerations:
            - key: "key1"
              operator: "Equal"
              value: "node1"
              effect: "NoSchedule"
      replicas: 1
    brokers:
      podTemplate:
        spec:
          tolerations:
            - key: "key1"
              operator: "Equal"
              value: "node1"
              effect: "NoSchedule"
      replicas: 1
    historicals:
      podTemplate:
        spec:
          tolerations:
            - key: "key1"
              operator: "Equal"
              value: "node1"
              effect: "NoSchedule"
      replicas: 1
    middleManagers:
      podTemplate:
        spec:
          tolerations:
            - key: "key1"
              operator: "Equal"
              value: "node1"
              effect: "NoSchedule"
      replicas: 1
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/podtemplating/yamls/druid-with-tolerations.yaml
druid.kubedb.com/druid-with-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `druid-with-tolerations-0` has been created.

Check that the petset's pod is running

```bash
$ $ kubectl get pods -n demo -l app.kubernetes.io/instance=druid-cluster

NAME                                      READY   STATUS    RESTARTS   AGE
druid-with-tolerations-brokers-0          1/1     Running   0          164m
druid-with-tolerations-coordinators-0     1/1     Running   0          164m
druid-with-tolerations-historicals-0      1/1     Running   0          164m
druid-with-tolerations-middlemanagers-0   1/1     Running   0          164m
druid-with-tolerations-routers-0          1/1     Running   0          164m
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo druid-with-tolerations-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo druid-with-tolerations-coordinators-0 -o wide
NAME                                       READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
druid-with-tolerations-coordinators-0      1/1     Running   0          3m49s   10.2.0.8   lke212553-307295-339173d10000   <none>           <none>
```
We can successfully verify that our pod was scheduled to the node which it has tolerations.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete druid -n demo druid-with-tolerations

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart Druid](/docs/guides/druid/quickstart/guide/index.md) with KubeDB Operator.
- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
