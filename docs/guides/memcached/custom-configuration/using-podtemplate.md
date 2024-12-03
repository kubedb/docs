---
title: Run Memcached with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: mc-using-podtemplate-configuration
    name: Customize PodTemplate
    parent: custom-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Memcached with Custom PodTemplate

KubeDB supports providing custom configuration for Memcached via [PodTemplate](/docs/guides/memcached/concepts/memcached.md#specpodtemplate). This tutorial will show you how to use KubeDB to run a Memcached database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/memcached](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/memcached) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Memcached database.

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
    - serviceAccountName
    - schedulerName
    - tolerations
    - priorityClassName
    - priority
    - securityContext

Read about the fields in details in [PodTemplate concept](/docs/guides/memcached/concepts/memcached.md#specpodtemplate),

## CRD Configuration

Below is the YAML for the Memcached created in this example. Here, `spec.podTemplate.spec.containers[].env` specifies additional environment variables by users.

In this tutorial, we will register additional two users at starting time of Memcached. So, the fact is any environment variable with having `suffix: USERNAME` and `suffix: PASSWORD` will be key value pairs of username and password and will be registered in the `pool_passwd` file of Memcached. So we can use these users after Memcached initialize without even syncing them.

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: custom-memcached
  namespace: demo
spec:
  version: "1.6.22"
  replicas: 1
  podTemplate:
    spec:
      containers:
        - name: memcacded
          env:
            - name: "Memcached_Key"
              value: KubeDB
            - name: "Memcached_Value"
              value: '123'
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/custom-config/custom-podtemplate.yaml
memcached.kubedb.com/custom-memcached created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `custom-memcached-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                 READY   STATUS    RESTARTS   AGE
custom-memcached-0   1/1     Running   0          30s
```

Now, check if the memcached has started with the custom configuration we have provided. First, we will exec in the pod. Then, we will check if the environment variable is set or not.

```bash
$ kubectl exec -it custom-memcached-0 -n demo memcached -- sh
~ $ echo $Memcached_Key
KubeDB
~ $ echo $Memcached_Value
123
exit
```
So, we can see that the additional environment variables are set correctly. 

## Custom Sidecar Containers

Here in this example we will add an extra sidecar container with our memcached container. So, it is required to run Filebeat as a sidecar container along with the KubeDB-managed Memcached. Here’s a quick demonstration on how to accomplish it.

Firstly, we are going to make our custom filebeat image with our required configuration.
```yaml
filebeat.inputs:
  - type: log
    paths:
      - /tmp/memcached_log/
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
Now we will deploy our memcached with custom sidecar container and will also use the `spec.initConfig` to configure the logs related settings. Here is the yaml of our memcached:
```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcached-custom-sidecar
  namespace: demo
spec:
  version: "1.6.22"
  replicas: 1
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 100m
              memory: 100Mi
        - name: filebeat
          image: evanraisul/custom_filebeat:latest
          resources:
            limits:
              cpu: 300m
              memory: 300Mi
            requests:
              cpu: 300m
              memory: 300Mi
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/custom-config/sidecar-container.yaml
memcached.kubedb.com/-custom-sidecar created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `memcached-custom-sidecar-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                         READY   STATUS    RESTARTS      AGE
memcached-custom-sidecar-0   2/2     Running   0             33s

```

Now, let’s checked the memcached database with the 2 containers with their given resources:

```yaml
kubectl get memcached -n demo memcached-custom-sidecar -oyaml

apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"Memcached","metadata":{"annotations":{},"name":"memcached-custom-sidecar","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","podTemplate":{"spec":{"containers":[{"name":"memcached","resources":{"limits":{"cpu":"100m","memory":"100Mi"},"requests":{"cpu":"100m","memory":"100Mi"}}},{"image":"evanraisul/custom_filebeat:latest","name":"filebeat","resources":{"limits":{"cpu":"300m","memory":"300Mi"},"requests":{"cpu":"300m","memory":"300Mi"}}}]}},"replicas":1,"version":"1.6.22"}}
  creationTimestamp: "2024-12-02T10:59:59Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: memcached-custom-sidecar
  namespace: demo
  resourceVersion: "680005"
  uid: 03d2b334-c5fd-4c9a-b88f-797f9630cec5
spec:
  authSecret:
    name: memcached-custom-sidecar-auth
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
      - name: memcached
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
          requests:
            cpu: 100m
            memory: 100Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          runAsGroup: 999
          runAsNonRoot: true
          runAsUser: 999
          seccompProfile:
            type: RuntimeDefault
      - image: evanraisul/custom_filebeat:latest
        name: filebeat
        resources:
          limits:
            cpu: 300m
            memory: 300Mi
          requests:
            cpu: 300m
            memory: 300Mi
      podPlacementPolicy:
        name: default
      securityContext:
        fsGroup: 999
      serviceAccountName: memcached-custom-sidecar
  replicas: 1
  version: 1.6.22
status:
  conditions:
  - lastTransitionTime: "2024-12-02T10:59:59Z"
    message: 'The KubeDB operator has started the provisioning of Memcached: demo/memcached-custom-sidecar'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2024-12-02T11:00:01Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2024-12-02T11:00:11Z"
    message: 'The Memcached: demo/memcached-custom-sidecar is accepting mcClient requests.'
    observedGeneration: 3
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2024-12-02T11:00:11Z"
    message: 'The Memcached: demo/memcached-custom-sidecar is ready.'
    observedGeneration: 3
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2024-12-02T11:00:11Z"
    message: 'The Memcached: demo/memcached-custom-sidecar is successfully provisioned.'
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 3
  phase: Ready
```

So, we have successfully checked our sidecar filebeat container in Memcached database.

## Using Node Selector

Here in this example we will use [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to schedule our memcached pod to a specific node. Applying nodeSelector to the Pod involves several steps. We first need to assign a label to some node that will be later used by the `nodeSelector` . Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

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

Now let's create a memcached with this new label as nodeSelector. Below is the yaml we are going to apply:
```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcached-node-selector
  namespace: demo
spec:
  version: "1.6.22"
  replicas: 1
  podTemplate:
    spec:
      nodeSelector:
        disktype: ssd
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/custom-config/node-selector.yaml
memcached.kubedb.com/memcached-node-selector created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `memcached-node-selector-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                        READY   STATUS    RESTARTS   AGE
memcached-node-selector-0   1/1     Running   0          60s
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo memcached-node-selector-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo memcached-node-selector-0 -o wide
NAME                        READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
memcached-node-selector-0   1/1     Running   0          3m19s   10.2.1.7   lke212553-307295-5541798e0000   <none>           <none>
```
We can successfully verify that our pod was scheduled to our desired node.

## Using Taints and Tolerations

Here in this example we will use [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to schedule our memcached pod to a specific node and also prevent from scheduling to nodes. Applying taints and tolerations to the Pod involves several steps. Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

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

$ kubectl taint nodes lke212553-307295-5541798e0000 key2=node2:NoSchedule
node/lke212553-307295-5541798e0000 tainted

$ kubectl taint nodes lke212553-307295-5b53c5520000 key3=node3:NoSchedule
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
    "key": "key2",
    "value": "node2"
  }
]
lke212553-307295-5b53c5520000
[
  {
    "effect": "NoSchedule",
    "key": "key3",
    "value": "node3"
  }
]
```
We can see that our taints were successfully assigned. Now let's try to create a memcached without proper tolerations. Here is the yaml of memcached we are going to createc
```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcached-without-tolerations
  namespace: demo
spec:
  version: "1.6.22"
  replicas: 1
  deletionPolicy: WipeOut
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/configuration/memcached-without-tolerations.yaml
memcached.kubedb.com/memcached-without-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `memcached-without-tolerations-0` has been created and running.

Check that the petset's pod is running or not,
```bash
$ kubectl get pods -n demo
NAME                              READY   STATUS    RESTARTS   AGE
memcached-without-tolerations-0   0/1     Pending   0          3m35s
```
Here we can see that the pod is not running. So let's describe the pod,
```bash
$ kubectl describe pods -n demo memcached-without-tolerations-0 
Name:             memcached-without-tolerations-0
Namespace:        demo
Priority:         0
Service Account:  default
Node:             <none>
Labels:           app.kubernetes.io/component=connection-pooler
                  app.kubernetes.io/instance=memcached-without-tolerations
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=memcacheds.kubedb.com
                  apps.kubernetes.io/pod-index=0
                  controller-revision-hash=memcached-without-tolerations-5b85f9cd
                  statefulset.kubernetes.io/pod-name=memcached-without-tolerations-0
Annotations:      <none>
Status:           Pending
IP:               
IPs:              <none>
Controlled By:    PetSet/memcached-without-tolerations
Containers:
  memcached:
    Image:           ghcr.io/appscode-images/memcached:1.6.22-alpine
    Ports:           11211/TCP
    Host Ports:      0/TCP, 0/TCP
    SeccompProfile:  RuntimeDefault
    Limits:
      memory:  1Gi
    Requests:
      cpu:     500m
      memory:  1Gi
    volumeMounts:
      - mountPath: /usr/config/
        name: memcached-config
      - mountPath: /usr/auth/
        name: auth
      - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
        name: kube-api-access-mj7lj
Conditions:
  Type           Status
  PodScheduled   False 
volumes:
  - name: memcached-config
    secret:
      defaultMode: 420
      items:
      - key: memcached.conf
        path: memcached.conf
      secretName: memcd-quickstart-config
  - name: auth
    secret:
      defaultMode: 420
      items:
      - key: authData
        path: authfile
      secretName: mc-auth
  - name: kube-api-access-mj7lj
    projected:
      defaultMode: 420
      sources:
      - serviceAccountToken:
          expirationSeconds: 3607
          path: token
      - configMap:
          items:
          - key: ca.crt
            path: ca.crt
          name: kube-root-ca.crt
      - downwardAPI:
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
            path: namespace
Node-Selectors:               <none>
Tolerations:                  node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                              node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Topology Spread Constraints:  kubernetes.io/hostname:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=connection-pooler,app.kubernetes.io/instance=memcached-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=memcacheds.kubedb.com
                              topology.kubernetes.io/zone:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=connection-pooler,app.kubernetes.io/instance=memcached-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=memcacheds.kubedb.com
Events:
  Type     Reason             Age                   From                Message
  ----     ------             ----                  ----                -------
  Warning  FailedScheduling   5m20s                 default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Warning  FailedScheduling   11s                   default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Normal   NotTriggerScaleUp  13s (x31 over 5m15s)  cluster-autoscaler  pod didn't trigger scale-up:
```
Here we can see that the pod has no tolerations for the tainted nodes and because of that the pod is not able to scheduled.

So, let's add proper tolerations and create another memcached. Here is the yaml we are going to apply,
```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcached-with-tolerations
  namespace: demo
spec:
  version: "1.6.22"
  replicas: 1
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/configuration/with-tolerations.yaml
memcached.kubedb.com/memcached-with-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `memcached-with-tolerations-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                              READY   STATUS    RESTARTS   AGE
memcached-with-tolerations-0      1/1     Running   0          2m
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo memcached-with-tolerations-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo memcached-with-tolerations-0 -o wide
NAME                           READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
memcached-with-tolerations-0   1/1     Running   0          3m49s   10.2.0.8   lke212553-307295-339173d10000   <none>           <none>
```
We can successfully verify that our pod was scheduled to the node which it has tolerations.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pp custom-sidecar node-selector with-tolerations without-tolerations
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart Memcached](/docs/guides/memcached/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Memcached database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/memcached/monitoring/using-prometheus-operator.md).
- Monitor your Memcached database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Memcached object](/docs/guides/memcached/concepts/memcached.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
