---
title: Redis GitOps Guides
menu:
  docs_{{ .version }}:
    identifier: rd-gitops-guides
    name: GitOps Redis
    parent: rd-gitops
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Redis using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create Redis database and manage updates using GitOps worrdlow.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).  Pass `--set kubedb-crd-manager.installGitOpsCRDs=true` in the kubedb installation process to enable `GitOps` operator.

- You need to install GitOps tools like `ArgoCD` or `FluxCD` and configure with your Git Repository to monitor the Git repository and synchronize the state of the Kubernetes cluster with the desired state defined in Git.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```
> Note: YAML files used in this tutorial are stored in [docs/examples/Redis](/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

We are going to use `ArgoCD` in this tutorial. You can install `ArgoCD` in your cluster by following the steps [here](https://argo-cd.readthedocs.io/en/stable/getting_started/). Also, you need to install `argocd` CLI in your local machine. You can install `argocd` CLI by following the steps [here](https://argo-cd.readthedocs.io/en/stable/cli_installation/).

## Creating Apps via CLI

### For Public Repository
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace>
```

### For Private Repository
#### Using HTTPS
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace> --username <username> --password <github-token>
```

#### Using SSH
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace> --ssh-private-key-path ~/.ssh/id_rsa
```

## Create Redis Database using GitOps

### Create a Redis GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 8.0.4
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

Create a directory like below,
```bash
$ tree .
├── kubedb
    └── Redis.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is created in your cluster.

Our `gitops` operator will create an actual `Redis` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get redis.gitops.kubedb.com,redis.kubedb.com -n demo
NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   8m37s

NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    8m37s
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Redis.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                                            AGE
petset.apps.k8s.appscode.com/rd-gitops-shard0   8m58s
petset.apps.k8s.appscode.com/rd-gitops-shard1   8m55s
petset.apps.k8s.appscode.com/rd-gitops-shard2   8m52s

NAME                     READY   STATUS    RESTARTS   AGE
pod/rd-gitops-shard0-0   1/1     Running   0          8m57s
pod/rd-gitops-shard0-1   1/1     Running   0          7m52s
pod/rd-gitops-shard1-0   1/1     Running   0          8m54s
pod/rd-gitops-shard1-1   1/1     Running   0          7m52s
pod/rd-gitops-shard2-0   1/1     Running   0          8m52s
pod/rd-gitops-shard2-1   1/1     Running   0          7m52s

NAME                      TYPE                       DATA   AGE
secret/rd-gitops-9aa2fa   Opaque                     1      9m4s
secret/rd-gitops-auth     kubernetes.io/basic-auth   2      9m4s

NAME                     TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)              AGE
service/rd-gitops        ClusterIP   10.43.90.210   <none>        6379/TCP             9m
service/rd-gitops-pods   ClusterIP   None           <none>        6379/TCP,16379/TCP   9m4s

NAME                                           TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/rd-gitops   kubedb.com/redis   8.0.4     8m53s
```

## Update Redis Database using GitOps


### TLS configuration



We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Redis. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=redis/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls redis-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: redis-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: redis-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/tls/issuer.yaml
issuer.cert-manager.io/redis-ca-issuer created
```

Let's add that to our `kubedb /rd-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── rd-configuration.yaml
│ ├── rd-issuer.yaml
│ └── redis.yaml
1 directories, 3 files
```

Update the `redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 7.4.1
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: rd-issuer
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

Add  `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` ElasticsearchOpsRequest to update the `Elasticsearch` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   7.4.1     Ready    15m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   15m

NAME                                                             TYPE             STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-reconfiguretls-qcdjjd   ReconfigureTLS   Successful   9m47s
```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for Elasticsearch.


### Scale Redis Replicas


Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 8.0.4
  mode: Cluster
  cluster:
    shards: 3
    replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
 ```
Update the `replicas` to `3`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` RedisOpsRequest to update the `Redis` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    19m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   19m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-4ecw03   HorizontalScaling   Successful   4m2s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                 READY   STATUS    RESTARTS   AGE
rd-gitops-shard0-0   1/1     Running   0          20m
rd-gitops-shard0-1   1/1     Running   0          19m
rd-gitops-shard0-2   1/1     Running   0          4m26s
rd-gitops-shard1-0   1/1     Running   0          20m
rd-gitops-shard1-1   1/1     Running   0          19m
rd-gitops-shard1-2   1/1     Running   0          4m6s
rd-gitops-shard2-0   1/1     Running   0          20m
rd-gitops-shard2-1   1/1     Running   0          19m
rd-gitops-shard2-2   1/1     Running   0          3m46s
```

We can also scale down the replicas by updating the `replicas` fields.

### Scale Redis Database Resources

Before the Ops Request reaches the `Successful` state, the configured memory limits are as follows:

```bash
$ kubectl get pod -n demo rd-gitops-shard0-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```


Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 8.0.4
  mode: Cluster
  cluster:
    shards: 3
    replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "1.5Gi"
            cpu: "1000m"
          limits:
            memory: "1.5Gi"
            cpu: "1000m"
```


Resource Requests and Limits are updated to `1000m` CPU and `1.5Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `RedisOpsRequest` to update the `Redis` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    17h

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   17h

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-ule15j   HorizontalScaling   Successful   16h
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-lliwo8     VerticalScaling     Successful   16h
```

```bash 
$ kubectl get pod -n demo rd-gitops-shard0-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "1",
    "memory": "1536Mi"
  }
}
```


### Expand Redis Volume

Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 8.0.4
  mode: Cluster
  cluster:
    shards: 3
    replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "2Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "1.5Gi"
            cpu: "1000m"
          limits:
            memory: "1.5Gi"
            cpu: "1000m"
```

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` RedisOpsRequest to update the `Redis` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    39m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   39m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-4ecw03   HorizontalScaling   Successful   24m
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-r0oosa     VerticalScaling     Successful   17m
redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-0ubdaw     VolumeExpansion     Successful   7m31s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-rd-gitops-shard0-0   Bound    pvc-97afbd7c-5887-4381-bef8-4584411b4ed5   2Gi        RWO            standard       <unset>                 39m
data-rd-gitops-shard0-1   Bound    pvc-461c6122-94ba-4b96-805b-1bf2619a10d4   2Gi        RWO            standard       <unset>                 38m
data-rd-gitops-shard0-2   Bound    pvc-90fc4131-a272-44bd-bfa2-3ffc5c093178   2Gi        RWO            standard       <unset>                 24m
data-rd-gitops-shard1-0   Bound    pvc-38855b1c-4e1f-485f-93c5-5848eed1465d   2Gi        RWO            standard       <unset>                 39m
data-rd-gitops-shard1-1   Bound    pvc-14b89fe8-8681-4aba-af37-035da020c3cb   2Gi        RWO            standard       <unset>                 38m
data-rd-gitops-shard1-2   Bound    pvc-6d3d1e55-689e-4884-a3cc-ed6080e48cf0   2Gi        RWO            standard       <unset>                 23m
data-rd-gitops-shard2-0   Bound    pvc-e6b9fa56-40b2-45aa-8ebd-ab360522a294   2Gi        RWO            standard       <unset>                 39m
data-rd-gitops-shard2-1   Bound    pvc-6fdfd395-d2b8-45df-8369-f9897e3d678c   2Gi        RWO            standard       <unset>                 38m
data-rd-gitops-shard2-2   Bound    pvc-1570c37b-63da-456c-af59-b60ed544e651   2Gi        RWO            standard       <unset>                 23m
```


### Update Version

List Redis versions using `kubectl get Redisversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/redis/update-version/cluster.md).

Let's choose `7.4.1` in this example.

Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 7.4.1
  mode: Cluster
  cluster:
    shards: 3
    replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "2Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "1.5Gi"
            cpu: "1000m"
          limits:
            memory: "1.5Gi"
            cpu: "1000m"
```

Update the `version` field to `7.4.1`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` RedisOpsRequest to update the `Redis` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   7.4.1     Ready    21m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   21m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-4ecw03   HorizontalScaling   Successful   144m
redisopsrequest.ops.kubedb.com/rd-gitops-versionupdate-wbsjct       UpdateVersion       Successful   12m
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-r0oosa     VerticalScaling     Successful   137m
redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-0ubdaw     VolumeExpansion     Successful   127m
```
## Reconfigure Redis


At first, we will create `redis.conf` file containing required configuration settings.

```ini
$ cat redis.conf
maxclients 2000
```
Here, `maxclients` is set to `500`, whereas the default value is `10000`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo rd-custom-config --from-file=./redis.conf
secret/rd-custom-config created
```

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo mg-custom-config --from-file=./mongod.conf
secret/mg-custom-config created
```


Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 7.4.1
  mode: Cluster
  cluster:
    shards: 3
    replicas: 3
  storageType: Durable
  configuration:
    secretName: rd-custom-config
  storage:
    resources:
      requests:
        storage: "2Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "1.5Gi"
            cpu: "1000m"
          limits:
            memory: "1.5Gi"
            cpu: "1000m"
```

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` RedisOpsRequest to update the `Redis` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   7.4.1     Ready    51m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   51m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-4ecw03   HorizontalScaling   Successful   3h42m
redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-uc97bo         Reconfigure         Successful   37m
redisopsrequest.ops.kubedb.com/rd-gitops-versionupdate-wbsjct       UpdateVersion       Successful   91m
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-r0oosa     VerticalScaling     Successful   3h35m
redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-0ubdaw     VolumeExpansion     Successful   3h26m
```

We can also reconfigure the parameters creating another secret and reference the secret in the `configuration.secretName` field. Also you can remove the `configuration.secretName` field to use the default parameters.



### Rotate Redis Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will create a secret named `rd-rotate-auth ` with the following content,

```bash
$ kubectl create secret generic rd-rotate-auth -n demo \
        --type=kubernetes.io/basic-auth \
        --from-literal=username=redis \
        --from-literal=password=redis-secret
secret/rd-rotate-auth created
```



Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 8.0.4
  mode: Cluster
  cluster:
    shards: 3
    replicas: 3
  storageType: Durable
  authSecret:
    kind: Secret
    name: rd-rotate-auth
  configuration:
    secretName: rd-new-configuration
  storage:
    resources:
      requests:
        storage: "2Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "500Mi"
            cpu: "500m"
          limits:
            memory: "500Mi"
            cpu: "500m"
  deletionPolicy: WipeOut
```

Change the `authSecret` field to `rd-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` RedisOpsRequest to update the `Redis` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   7.4.1     Ready    77m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   77m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-4ecw03   HorizontalScaling   Successful   4h9m
redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-uc97bo         Reconfigure         Successful   63m
redisopsrequest.ops.kubedb.com/rd-gitops-rotate-auth-2l0psh         RotateAuth          Successful   22m
redisopsrequest.ops.kubedb.com/rd-gitops-versionupdate-wbsjct       UpdateVersion       Successful   117m
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-r0oosa     VerticalScaling     Successful   4h2m
redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-0ubdaw     VolumeExpansion     Successful   3h52m
```
### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 7.4.1
  mode: Cluster
  cluster:
    shards: 3
    replicas: 3
  storageType: Durable
  authSecret:
    kind: Secret
    name: rd-rotate-auth
  configuration:
    secretName: rd-custom-config
  # tls:
  #   issuerRef:
  #     apiGroup: "cert-manager.io"
  #     kind: Issuer
  #     name: redis-ca-issuer
  storage:
    resources:
      requests:
        storage: "2Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "1.5Gi"
            cpu: "1000m"
          limits:
            memory: "1.5Gi"
            cpu: "1000m"
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` RedisOpsRequest to add the `Redis` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   7.4.1     Ready    117m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   117m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-4ecw03   HorizontalScaling   Successful   4h48m
redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-uc97bo         Reconfigure         Successful   103m
redisopsrequest.ops.kubedb.com/rd-gitops-restart-pj9zrj             Restart             Successful   8m56s
redisopsrequest.ops.kubedb.com/rd-gitops-rotate-auth-2l0psh         RotateAuth          Successful   62m
redisopsrequest.ops.kubedb.com/rd-gitops-versionupdate-wbsjct       UpdateVersion       Successful   157m
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-r0oosa     VerticalScaling     Successful   4h41m
redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-0ubdaw     VolumeExpansion     Successful   4h32m
```

Verify the monitoring is enabled by checking the prometheus targets.

There are some other fields that will trigger `Restart` ops request.
- `.spec.monitor`
- `.spec.spec.archiver`
- `.spec.remoteReplica`
- `spec.replication`
- `.spec.standbyMode`
- `.spec.streamingMode`
- `.spec.enforceGroup`
- `.spec.sslMode` etc.


## Next Steps


- Learn Redis Scaling
    - [Horizontal Scaling](/docs/guides/redis/scaling/horizontal-scaling/cluster.md)
    - [Vertical Scaling](/docs/guides/redis/scaling/horizontal-scaling/cluster.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/redis/update-version/cluster.md)
- Monitor your RedisQL database with KubeDB using [built-in Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
