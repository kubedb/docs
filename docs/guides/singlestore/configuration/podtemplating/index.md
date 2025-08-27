---
title: Run SingleStore with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-configuration-using-podtemplate
    name: Customize PodTemplate
    parent: guides-sdb-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run SingleStore with Custom PodTemplate

KubeDB supports providing custom configuration for SingleStore via [PodTemplate](/docs/guides/singlestore/concepts/singlestore.md#spec.topology). This tutorial will show you how to use KubeDB to run a SingleStore database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/singlestore/configuration/podtemplating/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for `leaf` and `aggregator` pod through `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate`. KubeDB operator will pass the information provided in `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate` to the `aggregator` and `leaf` PetSet created for SingleStore database.

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

Read about the fields in details in [PodTemplate concept](/docs/guides/singlestore/concepts/singlestore.md#spectopology),

## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## CRD Configuration

Below is the YAML for the SingleStore created in this example. Here, [`spec.topology.aggregator/leaf.podTemplate.spec.args`](/docs/guides/mysql/concepts/database/index.md#specpodtemplatespecargs) provides extra arguments.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-misc-config
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
            args:
              - --character-set-server=utf8mb4
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"     
              args:
                - --character-set-server=utf8mb4              
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls/sdb-misc-config.yaml
singlestore.kubedb.com/sdb-misc-config created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-misc-config-aggregator-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                           READY   STATUS    RESTARTS   AGE
sdb-misc-config-aggregator-0   2/2     Running   0          4m51s
sdb-misc-config-leaf-0         2/2     Running   0          4m48s
sdb-misc-config-leaf-1         2/2     Running   0          4m30s
```

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sdb-misc-config-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sdb-misc-config-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 311
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> SHOW VARIABLES LIKE 'char%';
+--------------------------+------------------------------------------------------+
| Variable_name            | Value                                                |
+--------------------------+------------------------------------------------------+
| character_set_client     | utf8mb4                                              |
| character_set_connection | utf8mb4                                              |
| character_set_database   | utf8mb4                                              |
| character_set_filesystem | binary                                               |
| character_set_results    | utf8mb4                                              |
| character_set_server     | utf8mb4                                              |
| character_set_system     | utf8                                                 |
| character_sets_dir       | /opt/memsql-server-8.7.10-95e2357384/share/charsets/ |
+--------------------------+------------------------------------------------------+
8 rows in set (0.00 sec)

singlestore> exit
Bye

```

Here we can see the character_set_server value is utf8mb4.  

## Custom Sidecar Containers

Here in this example we will add an extra sidecar container with our SingleStore cluster. This below example configuration allows you to run a SingleStore instance alongside a simple Nginx sidecar container, which can be used for HTTP requests, logging, or as a reverse proxy. Adjust the configuration as needed to fit your application's architecture.

Firstly, we are going to create a sample configmap for the nginx configuration. Here is the yaml of ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-config-map
  namespace: demo
data:
  default.conf: |
    server {
        listen 80;
        location / {
            proxy_pass http://localhost:9000;
        }
    }
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls/nginx-config-map.yaml
configmap/nginx-config-map created
```

Now we will deploy our singlestore with custom sidecar container. Here is the yaml of singlestore,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-custom-sidecar
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"
            - name: sidecar
              image: nginx:alpine
              ports:
                - containerPort: 80
              volumeMounts:
                - name: nginx-config
                  mountPath: /etc/nginx/conf.d
          volumes:
            - name: nginx-config
              configMap:
                name: nginx-config-map
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Here,

- Primary Container: The main singlestore container runs the SingleStore database, configured with specific resource limits and requests.

- Sidecar Container: The sidecar container runs Nginx, a lightweight web server. It's configured to listen on port 80 and is intended to proxy requests to the SingleStore database.

- Volume Mounts: The sidecar container mounts a volume for Nginx configuration from a ConfigMap, which allows you to customize Nginx's behavior.

- Volumes: A volume is defined to link the ConfigMap nginx-config-map to the Nginx configuration directory.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls/sdb-custom-sidecar.yaml
singlestore.kubedb.com/sdb-custom-sidecar created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see the pods has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                              READY   STATUS    RESTARTS   AGE
sdb-custom-sidecar-aggregator-0   3/3     Running   0          3m17s
sdb-custom-sidecar-leaf-0         2/2     Running   0          3m14s
sdb-custom-sidecar-leaf-1         2/2     Running   0          2m59s
```

Now check the logs of sidecar container,

```bash
$ kubectl logs -f -n demo sdb-custom-sidecar-aggregator-0 -c sidecar
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
/docker-entrypoint.sh: Launching /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
10-listen-on-ipv6-by-default.sh: info: can not modify /etc/nginx/conf.d/default.conf (read-only file system?)
/docker-entrypoint.sh: Sourcing /docker-entrypoint.d/15-local-resolvers.envsh
/docker-entrypoint.sh: Launching /docker-entrypoint.d/20-envsubst-on-templates.sh
/docker-entrypoint.sh: Launching /docker-entrypoint.d/30-tune-worker-processes.sh
/docker-entrypoint.sh: Configuration complete; ready for start up
2024/10/29 07:43:11 [notice] 1#1: using the "epoll" event method
2024/10/29 07:43:11 [notice] 1#1: nginx/1.27.2
2024/10/29 07:43:11 [notice] 1#1: built by gcc 13.2.1 20240309 (Alpine 13.2.1_git20240309) 
2024/10/29 07:43:11 [notice] 1#1: OS: Linux 6.8.0-47-generic
2024/10/29 07:43:11 [notice] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576
2024/10/29 07:43:11 [notice] 1#1: start worker processes
2024/10/29 07:43:11 [notice] 1#1: start worker process 21
2024/10/29 07:43:11 [notice] 1#1: start worker process 22
2024/10/29 07:43:11 [notice] 1#1: start worker process 23
2024/10/29 07:43:11 [notice] 1#1: start worker process 24
2024/10/29 07:43:11 [notice] 1#1: start worker process 25
2024/10/29 07:43:11 [notice] 1#1: start worker process 26
2024/10/29 07:43:11 [notice] 1#1: start worker process 27
2024/10/29 07:43:11 [notice] 1#1: start worker process 28
2024/10/29 07:43:11 [notice] 1#1: start worker process 29
2024/10/29 07:43:11 [notice] 1#1: start worker process 30
2024/10/29 07:43:11 [notice] 1#1: start worker process 31
2024/10/29 07:43:11 [notice] 1#1: start worker process 32
```
So, we have successfully deploy sidecar container in KubeDB manage SingleStore.

## Using Node Selector

Here in this example we will use [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) to schedule our singlestore pod to a specific node. Applying nodeSelector to the Pod involves several steps. We first need to assign a label to some node that will be later used by the `nodeSelector` . Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

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

Now let's create a singlestore with this new label as nodeSelector. Below is the yaml we are going to apply:
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-node-selector
  namespace: demo
spec:
  version: "8.7.10"
  podTemplate:
    spec:
      nodeSelector:
        disktype: ssd
  deletionPolicy: WipeOut
  licenseSecret:
    name: license-secret
  storage:
    storageClassName: "longhorn"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  storageType: Durable
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls/sdb-node-selector.yaml
singlestore.kubedb.com/sdb-node-selector created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-node-selector-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                     READY   STATUS    RESTARTS   AGE
sdb-node-selector-0      1/1     Running   0          60s
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo sdb-node-selector-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo sdb-node-selector-0 -o wide
NAME                     READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
sdb-node-selector-0      1/1     Running   0          3m19s   10.2.1.7   lke212553-307295-5541798e0000   <none>           <none>
```
We can successfully verify that our pod was scheduled to our desired node.

## Using Taints and Tolerations

Here in this example we will use [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to schedule our singlestore pod to a specific node and also prevent from scheduling to nodes. Applying taints and tolerations to the Pod involves several steps. Let’s find what nodes exist in your cluster. To get the name of these nodes, you can run:

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
We can see that our taints were successfully assigned. Now let's try to create a singlestore without proper tolerations. Here is the yaml of singlestore we are going to createc
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-without-tolerations
  namespace: demo
spec:
  deletionPolicy: WipeOut
  licenseSecret:
    name: license-secret
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  storageType: Durable
  version: 8.7.10
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls/sdb-without-tolerations.yaml
singlestore.kubedb.com/sdb-without-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-without-tolerations-0` has been created and running.

Check that the petset's pod is running or not,
```bash
$ kubectl get pods -n demo
NAME                           READY   STATUS    RESTARTS   AGE
sdb-without-tolerations-0      0/1     Pending   0          3m35s
```
Here we can see that the pod is not running. So let's describe the pod,
```bash
$ kubectl describe pods -n demo sdb-without-tolerations-0 
Name:             sdb-without-tolerations-0
Namespace:        demo
Priority:         0
Service Account:  sdb-without-tolerations
Node:             ashraful/192.168.0.227
Start Time:       Tue, 29 Oct 2024 15:44:22 +0600
Labels:           app.kubernetes.io/component=database
                  app.kubernetes.io/instance=sdb-without-tolerations
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=singlestores.kubedb.com
                  apps.kubernetes.io/pod-index=0
                  controller-revision-hash=sdb-without-tolerations-6449dc959b
                  kubedb.com/petset=standalone
                  statefulset.kubernetes.io/pod-name=sdb-without-tolerations-0
Annotations:      <none>
Status:           Running
IP:               10.42.0.122
IPs:
  IP:           10.42.0.122
Controlled By:  PetSet/sdb-without-tolerations
Init Containers:
  singlestore-init:
    Container ID:    containerd://382a8cca4103e609c0a763f65db11e89ca38fe4b982dd6f03c18eb33c083998c
    Image:           ghcr.io/kubedb/singlestore-init:8.7.10-v1@sha256:7f8a60b45c9a402c5a3de56a266e06a70db1feeff1c28a506e485e60afc7f5fa
    Image ID:        ghcr.io/kubedb/singlestore-init@sha256:7f8a60b45c9a402c5a3de56a266e06a70db1feeff1c28a506e485e60afc7f5fa
    Port:            <none>
    Host Port:       <none>
    SeccompProfile:  RuntimeDefault
    State:           Terminated
      Reason:        Completed
      Exit Code:     0
      Started:       Tue, 29 Oct 2024 15:44:31 +0600
      Finished:      Tue, 29 Oct 2024 15:44:31 +0600
    Ready:           True
    Restart Count:   0
    Limits:
      memory:  512Mi
    Requests:
      cpu:        200m
      memory:     512Mi
    Environment:  <none>
    Mounts:
      /scripts from init-scripts (rw)
      /var/lib/memsql from data (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-htm2z (ro)
Containers:
  singlestore:
    Container ID:    containerd://b52ae6c34300ea23b60ce91fbbc6a01a1fd71bb7a3de6fea97d9a726ca280e55
    Image:           singlestore/cluster-in-a-box:alma-8.7.10-95e2357384-4.1.0-1.17.14@sha256:6b1b66b57e11814815a43114ab28db407428662af4c7d1c666c14a3f53c5289f
    Image ID:        docker.io/singlestore/cluster-in-a-box@sha256:6b1b66b57e11814815a43114ab28db407428662af4c7d1c666c14a3f53c5289f
    Ports:           3306/TCP, 8081/TCP
    Host Ports:      0/TCP, 0/TCP
    SeccompProfile:  RuntimeDefault
    Args:
      /scripts/standalone-run.sh
    State:          Running
      Started:      Tue, 29 Oct 2024 15:44:32 +0600
    Ready:          True
    Restart Count:  0
    Limits:
      memory:  2Gi
    Requests:
      cpu:     500m
      memory:  2Gi
    Environment:
      ROOT_USERNAME:        <set to the key 'username' in secret 'sdb-without-tolerations-root-cred'>  Optional: false
      ROOT_PASSWORD:        <set to the key 'password' in secret 'sdb-without-tolerations-root-cred'>  Optional: false
      SINGLESTORE_LICENSE:  <set to the key 'password' in secret 'license-secret'>                     Optional: false
      LICENSE_KEY:          <set to the key 'password' in secret 'license-secret'>                     Optional: false
      HOST_IP:               (v1:status.hostIP)
    Mounts:
      /scripts from init-scripts (rw)
      /var/lib/memsql from data (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-htm2z (rotate-auth)
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
    ClaimName:  data-sdb-without-tolerations-0
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
Topology Spread Constraints:  kubernetes.io/hostname:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=database,app.kubernetes.io/instance=sdb-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=singlestores.kubedb.com,kubedb.com/petset=standalone
                              topology.kubernetes.io/zone:ScheduleAnyway when max skew 1 is exceeded for selector app.kubernetes.io/component=database,app.kubernetes.io/instance=sdb-without-tolerations,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=singlestores.kubedb.com,kubedb.com/petset=standalone
Events:
  Type     Reason             Age                   From                Message
  ----     ------             ----                  ----                -------
  Warning  FailedScheduling   5m20s                 default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Warning  FailedScheduling   11s                   default-scheduler   0/3 nodes are available: 1 node(s) had untolerated taint {key1: node1}, 1 node(s) had untolerated taint {key1: node2}, 1 node(s) had untolerated taint {key1: node3}. preemption: 0/3 nodes are available: 3 Preemption is not helpful for scheduling.
  Normal   NotTriggerScaleUp  13s (x31 over 5m15s)  cluster-autoscaler  pod didn't trigger scale-up:
```
Here we can see that the pod has no tolerations for the tainted nodes and because of that the pod is not able to scheduled.

So, let's add proper tolerations and create another singlestore. Here is the yaml we are going to apply,
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-with-tolerations
  namespace: demo
spec:
  podTemplate:
    spec:
      tolerations:
      - key: "key1"
        operator: "Equal"
        value: "node1"
        effect: "NoSchedule"
  deletionPolicy: WipeOut
  licenseSecret:
    name: license-secret
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  storageType: Durable
  version: 8.7.10
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls/sdb-with-tolerations.yaml
singlestore.kubedb.com/sdb-with-tolerations created
```
Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-with-tolerations-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pods -n demo
NAME                           READY   STATUS    RESTARTS   AGE
sdb-with-tolerations-0         1/1     Running   0          2m
```
As we see the pod is running, you can verify that by running `kubectl get pods -n demo sdb-with-tolerations-0 -o wide` and looking at the “NODE” to which the Pod was assigned.
```bash
$ kubectl get pods -n demo sdb-with-tolerations-0 -o wide
NAME                        READY   STATUS    RESTARTS   AGE     IP         NODE                            NOMINATED NODE   READINESS GATES
sdb-with-tolerations-0      1/1     Running   0          3m49s   10.2.0.8   lke212553-307295-339173d10000   <none>           <none>
```
We can successfully verify that our pod was scheduled to the node which it has tolerations.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete singlestore -n demo sdb-misc-config

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart SingleStore](/docs/guides/singlestore/quickstart/quickstart.md) with KubeDB Operator.
- Initialize [SingleStore with Script](/docs/guides/singlestore/initialization).
- Detail concepts of [SingleStore object](/docs/guides/singlestore/concepts/singlestore.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
