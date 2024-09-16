---
title: Initialize Redis using Script
menu:
  docs_{{ .version }}:
    identifier: rd-using-script-initialization
    name: Using Script
    parent: rd-initialization-redis
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize Redis using Script

This tutorial will show you how to use KubeDB to initialize a Redis database with shell or lua script.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Initialization Scripts

Redis supports initialization with `.sh` and `.lua` files. In this tutorial, we will use `init.sh` script to insert data inside `kubedb` DB.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from `init.sh` file. Then, we will provide this ConfigMap as script source in `init.script` of Redis crd spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo rd-init-script --from-literal=init.sh="redis-cli set hello world"
configmap/rd-init-script created
```

## Create a Redis database with Init-Script

Below is the `Redis` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-init-script
  namespace: demo
spec:
  version: 7.2.3
  disableAuth: false
  storageType: Durable
  init:
    script:
      projected:
        sources:
          - configMap:
              name: redis-init-script
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/initialization/demo-1.yaml
redis.kubedb.com/rd-init-script created
```

Here,

- `spec.init.script` specifies a script source used to initialize the database before database server starts. Bash script with `.sh` extension and lua script with `.lua` extension are supported.

KubeDB operator watches for `Redis` objects using Kubernetes api. When a `Redis` object is created, KubeDB operator will create a new PetSet and a Service with the matching Redis object name. KubeDB operator will also create a governing service for PetSets with the name `<redis-crd-name>-gvr`, if one is not already present. No Redis specific RBAC roles are required for [RBAC enabled clusters](/docs/setup/README.md#using-yaml).

```bash
$ kubectl  describe rd -n demo rd-init-script
Name:         rd-init-script
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Redis
Metadata:
  Creation Timestamp:  2024-08-06T05:59:40Z
  Finalizers:
    kubedb.com
  Generation:        3
  Resource Version:  133452
  UID:               ce6268b0-82b8-4ee5-a8fb-a6d2e1550a66
Spec:
  Allowed Schemas:
    Namespaces:
      From:  Same
  Auth Secret:
    Name:  rd-init-script-auth
  Auto Ops:
  Coordinator:
    Resources:
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Init:
    Initialized:  true
    Script:
      Projected:
        Sources:
          Config Map:
            Name:  redis-init-script
  Mode:            Standalone
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Affinity:
        Pod Anti Affinity:
          Preferred During Scheduling Ignored During Execution:
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    rd-init-script
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        redises.kubedb.com
              Namespaces:
                demo
              Topology Key:  kubernetes.io/hostname
            Weight:          100
            Pod Affinity Term:
              Label Selector:
                Match Labels:
                  app.kubernetes.io/instance:    rd-init-script
                  app.kubernetes.io/managed-by:  kubedb.com
                  app.kubernetes.io/name:        redises.kubedb.com
              Namespaces:
                demo
              Topology Key:  failure-domain.beta.kubernetes.io/zone
            Weight:          50
      Container Security Context:
        Allow Privilege Escalation:  false
        Capabilities:
          Drop:
            ALL
        Run As Group:     999
        Run As Non Root:  true
        Run As User:      999
        Seccomp Profile:
          Type:  RuntimeDefault
      Resources:
        Limits:
          Memory:  1Gi
        Requests:
          Cpu:     500m
          Memory:  1Gi
      Security Context:
        Fs Group:            999
      Service Account Name:  rd-init-script
  Replicas:                  1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    WipeOut
  Version:               7.2.3
Status:
  Conditions:
    Last Transition Time:  2024-08-06T05:59:40Z
    Message:               The KubeDB operator has started the provisioning of Redis: demo/rd-init-script
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-08-06T05:59:49Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-08-06T06:00:22Z
    Message:               The Redis: demo/rd-init-script is ready.
    Observed Generation:   3
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-08-06T06:00:32Z
    Message:               The Redis: demo/rd-init-script is accepting rdClient requests.
    Observed Generation:   3
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-08-06T06:00:15Z
    Message:               The Redis: demo/rd-init-script is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     2
  Phase:                   Ready
Events:
  Type    Reason      Age   From             Message
  ----    ------      ----  ----             -------
  Normal  Successful  82s   KubeDB Operator  Successfully created governing service
  Normal  Successful  82s   KubeDB Operator  Successfully created Service
  Normal  Successful  79s   KubeDB Operator  Successfully created appbinding


$ kubectl get petset -n demo
NAME              READY   AGE
rd-init-script   1/1     30s

$ kubectl get pvc -n demo
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-rd-init-script-0   Bound    pvc-31dbab22-09af-4eeb-b032-1df287d9e579   1Gi        RWO            standard       2m17s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS   REASON   AGE
pvc-31dbab22-09af-4eeb-b032-1df287d9e579   1Gi        RWO            Delete           Bound    demo/data-rd-init-script-0   standard                2m37s


$ kubectl get service -n demo
NAME                  TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
rd-init-script        ClusterIP   10.96.3.28   <none>        6379/TCP   3m11s
rd-init-script-pods   ClusterIP   None         <none>        6379/TCP   3m11s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified Redis object:

```bash
$ kubectl get rd -n demo rd-init-script -o yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Redis","metadata":{"annotations":{},"name":"rd-init-script","namespace":"demo"},"spec":{"disableAuth":false,"init":{"script":{"projected":{"sources":[{"configMap":{"name":"redis-init-script"}}]}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"WipeOut","version":"7.2.3"}}
  creationTimestamp: "2024-08-06T05:59:40Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: rd-init-script
  namespace: demo
  resourceVersion: "133452"
  uid: ce6268b0-82b8-4ee5-a8fb-a6d2e1550a66
spec:
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: rd-init-script-auth
  autoOps: {}
  coordinator:
    resources: {}
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  init:
    initialized: true
    script:
      projected:
        sources:
        - configMap:
            name: redis-init-script
  mode: Standalone
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: rd-init-script
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: redises.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: rd-init-script
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: redises.kubedb.com
              namespaces:
              - demo
              topologyKey: failure-domain.beta.kubernetes.io/zone
            weight: 50
      containerSecurityContext:
        allowPrivilegeEscalation: false
        capabilities:
          drop:
          - ALL
        runAsGroup: 999
        runAsNonRoot: true
        runAsUser: 999
        seccompProfile:
          type: RuntimeDefault
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      securityContext:
        fsGroup: 999
      serviceAccountName: rd-init-script
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: WipeOut
  version: 7.2.3
status:
  conditions:
  - lastTransitionTime: "2024-08-06T05:59:40Z"
    message: 'The KubeDB operator has started the provisioning of Redis: demo/rd-init-script'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2024-08-06T05:59:49Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2024-08-06T06:00:22Z"
    message: 'The Redis: demo/rd-init-script is ready.'
    observedGeneration: 3
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2024-08-06T06:00:32Z"
    message: 'The Redis: demo/rd-init-script is accepting rdClient requests.'
    observedGeneration: 3
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2024-08-06T06:00:15Z"
    message: 'The Redis: demo/rd-init-script is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `rd-init-script-auth` *(format: {redis-object-name}-auth)* for storing the password for Redis superuser. This secret contains a `username` key which contains the *username* for Redis superuser and a `password` key which contains the *password* for Redis superuser.
If you want to use an existing secret please specify that when creating the Redis object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`.

```bash
$ kubectl get secrets -n demo rd-init-script-auth -o yaml
apiVersion: v1
data:
  password: STRMTl9fVjJuaDlsdndhcg==
  username: ZGVmYXVsdA==
kind: Secret
metadata:
  creationTimestamp: "2024-08-06T05:59:40Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: rd-init-script
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: redises.kubedb.com
  name: rd-init-script-auth
  namespace: demo
  resourceVersion: "133291"
  uid: ece22594-6c5f-4428-ac0f-5f2d2690785f
type: kubernetes.io/basic-auth
```

Now, you can connect to this database through redis cli. In this tutorial, we are connecting to the Redis server from inside the pod.

```bash
$ kubectl get secrets -n demo rd-init-script-auth -o jsonpath='{.data.\username}' | base64 -d
default

$ kubectl get secrets -n demo rd-init-script-auth -o jsonpath='{.data.\password}' | base64 -d
I4LN__V2nh9lvwar

$ kubectl exec -it rd-init-script-0 -n demo -- bash

Defaulted container "redis" out of: redis, redis-init (init)
redis@rd-init-script-0:/data$ 
redis@rd-init-script-0:/data$ redis-cli get hello
"world"
redis@rd-init-script-0:/data$ exit
exit
```

As you can see here, the initial script has successfully created a database named `kubedb` and inserted data into that database successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo rd/rd-init-script -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo rd/rd-init-script

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/redis/backup/stash/overview/index.md) Redis databases using Stash.
- Initialize [Redis with Script](/docs/guides/redis/initialization/using-script.md).
- Monitor your Redis database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [redisVersion object](/docs/guides/redis/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
