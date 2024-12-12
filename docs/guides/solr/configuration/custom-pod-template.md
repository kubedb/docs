---
title: Run Solr with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: sl-custom-pod-template
    name: Customize PodTemplate
    parent: sl-custom-config
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Solr with Custom PodTemplate

KubeDB supports providing custom configuration for Solr via [PodTemplate](/docs/guides/solr/concepts/solr.md#spectopology). This tutorial will show you how to use KubeDB to run a Solr database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/solr/configuration/podtemplating/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/Solr/configuration/podtemplating/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for `leaf` and `aggregator` pod through `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate`. KubeDB operator will pass the information provided in `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate` to the `aggregator` and `leaf` PetSet created for Solr database.

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

Read about the fields in details in [PodTemplate concept](/docs/guides/solr/concepts/solr.md#spectopology),


## CRD Configuration

Below is the YAML for the Solr created in this example. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-misc-config
  namespace: demo
spec:
  version: "9.6.1"
  topology:
    data:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: "solr"
              resources:
                requests:
                  cpu: "900m"
                limits:
                  cpu: "900m"
                  memory: "2.5Gi"
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    overseer:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: "solr"
              resources:
                requests:
                  cpu: "900m"
                limits:
                  cpu: "900m"
                  memory: "2.5Gi"
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    coordinator:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: "solr"
              resources:
                requests:
                  cpu: "900m"
                limits:
                  cpu: "900m"
                  memory: "2.5Gi"
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/configuration/sl-custom-podtemplate.yaml
Solr.kubedb.com/solr-misc-config created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-misc-config-aggregator-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=solr-misc-config
NAME                             READY   STATUS    RESTARTS   AGE
solr-misc-config-coordinator-0   1/1     Running   0          3m30s
solr-misc-config-data-0          1/1     Running   0          3m35s
solr-misc-config-overseer-0      1/1     Running   0          3m33s
```

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl get pod -n demo solr-misc-config-coordinator-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "900m",
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "2560Mi"
  }
}

$ kubectl get pod -n demo solr-misc-config-data-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "900m",
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "2560Mi"
  }
}

$ kubectl get pod -n demo solr-misc-config-overseer-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "900m",
    "memory": "2560Mi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "2560Mi"
  }
}

```


## Using Node Selector

Here in this example we will use [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to schedule our Solr pod to a specific node. Applying nodeSelector to the Pod involves several steps. We first need to assign a label to some node that will be later used by the `nodeSelector` . Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

```bash
$ kubectl get nodes
NAME                                    STATUS   ROLES    AGE    VERSION
gke-pritam-default-pool-c682fe6e-59x3   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-c682fe6e-rbtx   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-c682fe6e-spdb   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-cc96ce9b-049h   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-cc96ce9b-b8p8   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-cc96ce9b-vbpc   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-dadbf4db-5fv5   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-dadbf4db-5vkv   Ready    <none>   110m   v1.30.5-gke.1443001
gke-pritam-default-pool-dadbf4db-p039   Ready    <none>   110m   v1.30.5-gke.1443001
```
As you see, we have nine nodes in the cluster.

Let’s say we want pods to schedule to nodes with key `topology.gke.io/zone` and value `us-central1-b`
```bash
$ kubectl get nodes -n demo -l topology.gke.io/zone=us-central1-b
NAME                                    STATUS   ROLES    AGE    VERSION
gke-pritam-default-pool-c682fe6e-59x3   Ready    <none>   118m   v1.30.5-gke.1443001
gke-pritam-default-pool-c682fe6e-rbtx   Ready    <none>   118m   v1.30.5-gke.1443001
gke-pritam-default-pool-c682fe6e-spdb   Ready    <none>   118m   v1.30.5-gke.1443001
```

As you see, the gke-pritam-default-pool-c682fe6e-59x3 now has a new label topology.gke.io/zone=us-central1-b. To see all labels attached to the node, you can also run:
```bash
$ kubectl describe nodes gke-pritam-default-pool-c682fe6e-59x3
Name:               gke-pritam-default-pool-c682fe6e-59x3
Roles:              <none>
Labels:             beta.kubernetes.io/arch=amd64
                    beta.kubernetes.io/instance-type=e2-standard-2
                    beta.kubernetes.io/os=linux
                    cloud.google.com/gke-boot-disk=pd-balanced
                    cloud.google.com/gke-container-runtime=containerd
                    cloud.google.com/gke-cpu-scaling-level=2
                    cloud.google.com/gke-logging-variant=DEFAULT
                    cloud.google.com/gke-max-pods-per-node=110
                    cloud.google.com/gke-memory-gb-scaling-level=8
                    cloud.google.com/gke-nodepool=default-pool
                    cloud.google.com/gke-os-distribution=cos
                    cloud.google.com/gke-provisioning=standard
                    cloud.google.com/gke-stack-type=IPV4
                    cloud.google.com/machine-family=e2
                    cloud.google.com/private-node=false
                    disktype=ssd
                    failure-domain.beta.kubernetes.io/region=us-central1
                    failure-domain.beta.kubernetes.io/zone=us-central1-b
                    kubernetes.io/arch=amd64
                    kubernetes.io/hostname=gke-pritam-default-pool-c682fe6e-59x3
                    kubernetes.io/os=linux
                    node.kubernetes.io/instance-type=e2-standard-2
                    topology.gke.io/zone=us-central1-b
                    topology.kubernetes.io/region=us-central1
                    topology.kubernetes.io/zone=us-central1-b
```

Now let's create a Solr with this new label as nodeSelector. Below is the yaml we are going to apply:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-custom-nodeselector
  namespace: demo
spec:
  version: 9.6.1
  replicas: 2
  podTemplate:
    spec:
      nodeSelector:
        topology.gke.io/zone: us-central1-b
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi

```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/configuration/sl-custom-nodeselector.yaml
solr.kubedb.com/solr-node-selector created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-node-selector-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=solr-custom-nodeselector
NAME                         READY   STATUS    RESTARTS   AGE
solr-custom-nodeselector-0   1/1     Running   0          3m18s
solr-custom-nodeselector-1   1/1     Running   0          2m54s
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo sdb-node-selector-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=solr-custom-nodeselector -owide
NAME                         READY   STATUS    RESTARTS   AGE     IP          NODE                                    NOMINATED NODE   READINESS GATES
solr-custom-nodeselector-0   1/1     Running   0          3m52s   10.12.7.7   gke-pritam-default-pool-c682fe6e-spdb   <none>           <none>
solr-custom-nodeselector-1   1/1     Running   0          3m28s   10.12.8.9   gke-pritam-default-pool-c682fe6e-59x3   <none>           <none>
```
We can successfully verify that our pod was scheduled to our desired node.

## Using Taints and Tolerations

Here in this example we will use [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to schedule our Solr pod to a specific node and also prevent from scheduling to nodes. Applying taints and tolerations to the Pod involves several steps. Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

```bash
$ kubectl get nodes
NAME                                    STATUS   ROLES    AGE    VERSION
gke-pritam-default-pool-c682fe6e-59x3   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-c682fe6e-rbtx   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-c682fe6e-spdb   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-cc96ce9b-049h   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-cc96ce9b-b8p8   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-cc96ce9b-vbpc   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-dadbf4db-5fv5   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-dadbf4db-5vkv   Ready    <none>   123m   v1.30.5-gke.1443001
gke-pritam-default-pool-dadbf4db-p039   Ready    <none>   123m   v1.30.5-gke.1443001
```
As you see, we have nine nodes in the cluster

Next, we are going to taint these nodes.
```bash
$ kubectl taint nodes gke-pritam-default-pool-c682fe6e-59x3 key1=node1:NoSchedule
node/gke-pritam-default-pool-c682fe6e-59x3 tainted
$ kubectl taint nodes gke-pritam-default-pool-c682fe6e-rbtx key1=node2:NoSchedule
node/gke-pritam-default-pool-c682fe6e-rbtx tainted
$ kubectl taint nodes gke-pritam-default-pool-c682fe6e-spdb key1=node3:NoSchedule
node/gke-pritam-default-pool-c682fe6e-spdb tainted
$ kubectl taint nodes gke-pritam-default-pool-cc96ce9b-049h key1=node4:NoSchedule
node/gke-pritam-default-pool-cc96ce9b-049h tainted
$ kubectl taint nodes gke-pritam-default-pool-cc96ce9b-b8p8 key1=node5:NoSchedule
node/gke-pritam-default-pool-cc96ce9b-b8p8 tainted
$ kubectl taint nodes gke-pritam-default-pool-cc96ce9b-vbpc key1=node6:NoSchedule
node/gke-pritam-default-pool-cc96ce9b-vbpc tainted
$ kubectl taint nodes gke-pritam-default-pool-dadbf4db-5fv5 key1=node7:NoSchedule
node/gke-pritam-default-pool-dadbf4db-5fv5 tainted
$ kubectl taint nodes gke-pritam-default-pool-dadbf4db-5vkv key1=node8:NoSchedule
node/gke-pritam-default-pool-dadbf4db-5vkv tainted
$ kubectl taint nodes gke-pritam-default-pool-dadbf4db-p039 key1=node9:NoSchedule
node/gke-pritam-default-pool-dadbf4db-p039 tainted
```
Let's see our tainted nodes here,
```bash
$ kubectl get nodes -o json | jq -r '.items[] | select(.spec.taints != null) | .metadata.name, .spec.taints'
gke-pritam-default-pool-c682fe6e-59x3
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node1"
  }
]
gke-pritam-default-pool-c682fe6e-rbtx
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node2"
  }
]
gke-pritam-default-pool-c682fe6e-spdb
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node3"
  }
]
gke-pritam-default-pool-cc96ce9b-049h
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node4"
  }
]
gke-pritam-default-pool-cc96ce9b-b8p8
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node5"
  }
]
gke-pritam-default-pool-cc96ce9b-vbpc
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node6"
  }
]
gke-pritam-default-pool-dadbf4db-5fv5
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node7"
  }
]
gke-pritam-default-pool-dadbf4db-5vkv
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node8"
  }
]
gke-pritam-default-pool-dadbf4db-p039
[
  {
    "effect": "NoSchedule",
    "key": "key1",
    "value": "node9"
  }
]
```
We can see that our taints were successfully assigned. Now let's try to create a Solr without proper tolerations. Here is the yaml of Solr we are going to createc
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-without-toleration
  namespace: demo
spec:
  version: 9.6.1
  replicas: 2
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/configuration/solr-without-tolerations.yaml
solr.kubedb.com/solr-without-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-without-tolerations-0` has been created and running.

Check that the petset's pod is running or not,
```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=solr-without-toleration
NAME                        READY   STATUS    RESTARTS   AGE
solr-without-toleration-0   0/1     Pending   0          64s
```
Here we can see that the pod is not running. So let's describe the pod,
```bash
$ kubectl describe pod -n demo solr-without-toleration-0
Name:             solr-without-toleration-0
Namespace:        demo
Priority:         0
Service Account:  default
Node:             <none>
Labels:           app.kubernetes.io/component=database
                  app.kubernetes.io/instance=solr-without-toleration
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=solrs.kubedb.com
                  apps.kubernetes.io/pod-index=0
                  controller-revision-hash=solr-without-toleration-7d4b4d4bcc
                  coordinator=set
                  data=set
                  overseer=set
                  statefulset.kubernetes.io/pod-name=solr-without-toleration-0
Annotations:      cloud.google.com/cluster_autoscaler_unhelpable_since: 2024-11-13T12:32:48+0000
                  cloud.google.com/cluster_autoscaler_unhelpable_until: Inf
Status:           Pending
IP:               
IPs:              <none>
Controlled By:    PetSet/solr-without-toleration
Init Containers:
  init-solr:
    Image:           ghcr.io/kubedb/solr-init:9.6.1@sha256:dbbee5c25da5666a90fbb5d90d146c3a8d54f04eefacd779b59a248c0972ef15
    Port:            <none>
    Host Port:       <none>
    SeccompProfile:  RuntimeDefault
    Limits:
      memory:  512Mi
    Requests:
      cpu:     200m
      memory:  512Mi
    Environment:
      SOLR_JAVA_MEM:           -Xms1g -Xmx3g
      SOLR_HOME:               /var/solr
      SOLR_PORT:               8983
      SOLR_NODE_PORT:          8983
      SOLR_logS_DIR:           /var/solr/logs
      SOLR_log_LEVEL:          DEBUG
      SOLR_PORT_ADVERTISE:     8983
      CLUSTER_NAME:            solr-without-toleration
      POD_HOSTNAME:            solr-without-toleration-0 (v1:metadata.name)
      POD_NAME:                solr-without-toleration-0 (v1:metadata.name)
      POD_IP:                   (v1:status.podIP)
      GOVERNING_SERVICE:       solr-without-toleration-pods
      POD_NAMESPACE:           demo (v1:metadata.namespace)
      SOLR_HOST:               $(POD_NAME).$(GOVERNING_SERVICE).$(POD_NAMESPACE)
      ZK_HOST:                 zoo-0.zoo-pods.demo.svc.cluster.local:2181,zoo-1.zoo-pods.demo.svc.cluster.local:2181,zoo-2.zoo-pods.demo.svc.cluster.local:2181/demosolr-without-toleration
      ZK_SERVER:               zoo-0.zoo-pods.demo.svc.cluster.local:2181,zoo-1.zoo-pods.demo.svc.cluster.local:2181,zoo-2.zoo-pods.demo.svc.cluster.local:2181
      ZK_CHROOT:               /demosolr-without-toleration
      SOLR_MODULES:            
      JAVA_OPTS:               
      CONNECTION_SCHEME:       http
      SECURITY_ENABLED:        true
      SOLR_USER:               <set to the key 'username' in secret 'solr-without-toleration-auth'>  Optional: false
      SOLR_PASSWORD:           <set to the key 'password' in secret 'solr-without-toleration-auth'>  Optional: false
      SOLR_OPTS:               -DhostPort=$(SOLR_NODE_PORT) -Dsolr.autoSoftCommit.maxTime=1000 -DzkACLProvider=org.apache.solr.common.cloud.DigestZkACLProvider -DzkCredentialsInjector=org.apache.solr.common.cloud.VMParamsZkCredentialsInjector -DzkCredentialsProvider=org.apache.solr.common.cloud.DigestZkCredentialsProvider -DzkDigestPassword=nwbnmVwBoJhW)eft -DzkDigestReadonlyPassword=7PxFSc)z~DWLL)Tt -DzkDigestReadonlyUsername=zk-digest-readonly -DzkDigestUsername=zk-digest
      ZK_CREDS_AND_ACLS:       -DzkACLProvider=org.apache.solr.common.cloud.DigestZkACLProvider -DzkCredentialsInjector=org.apache.solr.common.cloud.VMParamsZkCredentialsInjector -DzkCredentialsProvider=org.apache.solr.common.cloud.DigestZkCredentialsProvider -DzkDigestPassword=nwbnmVwBoJhW)eft -DzkDigestReadonlyPassword=7PxFSc)z~DWLL)Tt -DzkDigestReadonlyUsername=zk-digest-readonly -DzkDigestUsername=zk-digest
      SOLR_ZK_CREDS_AND_ACLS:  -DzkACLProvider=org.apache.solr.common.cloud.DigestZkACLProvider -DzkCredentialsInjector=org.apache.solr.common.cloud.VMParamsZkCredentialsInjector -DzkCredentialsProvider=org.apache.solr.common.cloud.DigestZkCredentialsProvider -DzkDigestPassword=nwbnmVwBoJhW)eft -DzkDigestReadonlyPassword=7PxFSc)z~DWLL)Tt -DzkDigestReadonlyUsername=zk-digest-readonly -DzkDigestUsername=zk-digest
    Mounts:
      /temp-config from default-config (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-n9nxh (ro)
      /var/security from auth-config (rw)
      /var/solr from slconfig (rw)
Containers:
  solr:
    Image:           ghcr.io/appscode-images/solr:9.6.1@sha256:b625c7e8c91c8070b23b367cc03a736f2c5c2cb9cfd7981f72c461e57df800a1
    Port:            8983/TCP
    Host Port:       0/TCP
    SeccompProfile:  RuntimeDefault
    Limits:
      memory:  2Gi
    Requests:
      cpu:     900m
      memory:  2Gi
    Environment:
      SOLR_JAVA_MEM:           -Xms1g -Xmx3g
      SOLR_HOME:               /var/solr
      SOLR_PORT:               8983
      SOLR_NODE_PORT:          8983
      SOLR_logS_DIR:           /var/solr/logs
      SOLR_log_LEVEL:          DEBUG
      SOLR_PORT_ADVERTISE:     8983
      CLUSTER_NAME:            solr-without-toleration
      POD_HOSTNAME:            solr-without-toleration-0 (v1:metadata.name)
      POD_NAME:                solr-without-toleration-0 (v1:metadata.name)
      POD_IP:                   (v1:status.podIP)
      GOVERNING_SERVICE:       solr-without-toleration-pods
      POD_NAMESPACE:           demo (v1:metadata.namespace)
      SOLR_HOST:               $(POD_NAME).$(GOVERNING_SERVICE).$(POD_NAMESPACE)
      ZK_HOST:                 zoo-0.zoo-pods.demo.svc.cluster.local:2181,zoo-1.zoo-pods.demo.svc.cluster.local:2181,zoo-2.zoo-pods.demo.svc.cluster.local:2181/demosolr-without-toleration
      ZK_SERVER:               zoo-0.zoo-pods.demo.svc.cluster.local:2181,zoo-1.zoo-pods.demo.svc.cluster.local:2181,zoo-2.zoo-pods.demo.svc.cluster.local:2181
      ZK_CHROOT:               /demosolr-without-toleration
      SOLR_MODULES:            
      JAVA_OPTS:               
      CONNECTION_SCHEME:       http
      SECURITY_ENABLED:        true
      SOLR_USER:               <set to the key 'username' in secret 'solr-without-toleration-auth'>  Optional: false
      SOLR_PASSWORD:           <set to the key 'password' in secret 'solr-without-toleration-auth'>  Optional: false
      SOLR_OPTS:               -DhostPort=$(SOLR_NODE_PORT) -Dsolr.autoSoftCommit.maxTime=1000 -DzkACLProvider=org.apache.solr.common.cloud.DigestZkACLProvider -DzkCredentialsInjector=org.apache.solr.common.cloud.VMParamsZkCredentialsInjector -DzkCredentialsProvider=org.apache.solr.common.cloud.DigestZkCredentialsProvider -DzkDigestPassword=nwbnmVwBoJhW)eft -DzkDigestReadonlyPassword=7PxFSc)z~DWLL)Tt -DzkDigestReadonlyUsername=zk-digest-readonly -DzkDigestUsername=zk-digest
      ZK_CREDS_AND_ACLS:       -DzkACLProvider=org.apache.solr.common.cloud.DigestZkACLProvider -DzkCredentialsInjector=org.apache.solr.common.cloud.VMParamsZkCredentialsInjector -DzkCredentialsProvider=org.apache.solr.common.cloud.DigestZkCredentialsProvider -DzkDigestPassword=nwbnmVwBoJhW)eft -DzkDigestReadonlyPassword=7PxFSc)z~DWLL)Tt -DzkDigestReadonlyUsername=zk-digest-readonly -DzkDigestUsername=zk-digest
      SOLR_ZK_CREDS_AND_ACLS:  -DzkACLProvider=org.apache.solr.common.cloud.DigestZkACLProvider -DzkCredentialsInjector=org.apache.solr.common.cloud.VMParamsZkCredentialsInjector -DzkCredentialsProvider=org.apache.solr.common.cloud.DigestZkCredentialsProvider -DzkDigestPassword=nwbnmVwBoJhW)eft -DzkDigestReadonlyPassword=7PxFSc)z~DWLL)Tt -DzkDigestReadonlyUsername=zk-digest-readonly -DzkDigestUsername=zk-digest
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-n9nxh (ro)
      /var/solr from slconfig (rw)
      /var/solr/data from solr-without-toleration-data (rw)
Conditions:
  Type           Status
  PodScheduled   False 
Volumes:
  solr-without-toleration-data:
    Type:       PersistentVolumeClaim (a reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  solr-without-toleration-data-solr-without-toleration-0
    ReadOnly:   false
  default-config:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  solr-without-toleration-config
    Optional:    false
  slconfig:
    Type:       EmptyDir (a temporary directory that shares a pod's lifetime)
    Medium:     
    SizeLimit:  <unset>
  auth-config:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  solr-without-toleration-auth-config
    Optional:    false
  kube-api-access-n9nxh:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    ConfigMapOptional:       <nil>
    DownwardAPI:             true
QoS Class:                   Burstable
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type     Reason             Age                  From                Message
  ----     ------             ----                 ----                -------
  Normal   NotTriggerScaleUp  106s                 cluster-autoscaler  pod didn't trigger scale-up:
  Warning  FailedScheduling   104s (x2 over 106s)  default-scheduler   0/9 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}, 1 node(s) had untolerated taint {key1: node4}, 1 node(s) had untolerated taint {key1: node5}, 1 node(s) had untolerated taint {key1: node6}, 1 node(s) had untolerated taint {key1: node7}, 1 node(s) had untolerated taint {key1: node8}, 1 node(s) had untolerated taint {key1: node9}. preemption: 0/9 nodes are available: 9 Preemption is not helpful for scheduling.
```
Here we can see that the pod has no tolerations for the tainted nodes and because of that the pod is not able to scheduled.

So, let's add proper tolerations and create another Solr. Here is the yaml we are going to apply,
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-with-toleration
  namespace: demo
spec:
  version: 9.6.1
  replicas: 2
  podTemplate:
    spec:
      tolerations:
        - key: "key1"
          operator: "Equal"
          value: "node7"
          effect: "NoSchedule"
        - key: "key1"
          operator: "Equal"
          value: "node8"
          effect: "NoSchedule"
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/Solr/configuration/solr-with-tolerations.yaml
solr.kubedb.com/solr-with-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-with-tolerations-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=solr-with-toleration
NAME                     READY   STATUS    RESTARTS   AGE
solr-with-toleration-0   1/1     Running   0          2m12s
solr-with-toleration-1   1/1     Running   0          80s
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo sdb-with-tolerations-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pod -n demo -l app.kubernetes.io/instance=solr-with-toleration -owide
NAME                     READY   STATUS    RESTARTS   AGE     IP          NODE                                    NOMINATED NODE   READINESS GATES
solr-with-toleration-0   1/1     Running   0          2m37s   10.12.3.7   gke-pritam-default-pool-dadbf4db-5fv5   <none>           <none>
solr-with-toleration-1   1/1     Running   0          105s    10.12.5.5   gke-pritam-default-pool-dadbf4db-5vkv   <none>           <none>
```
We can successfully verify that our pod was scheduled to the node which it has tolerations.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete solr -n demo solr-misc-config solr-without-toleration solr-with-toleration

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).


## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).