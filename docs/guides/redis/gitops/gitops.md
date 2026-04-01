---
title: Redis GitOps Guides
menu:
  docs_{{ .version }}:
    identifier: rd-gitops-guides
    name: Gitops Redis
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
> Note: YAML files used in this tutorial are stored in [docs/examples/Redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "300Mi"
            cpu: "200m"
          limits:
            memory: "300Mi"
            cpu: "200m"
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
redis.gitops.kubedb.com/rd-gitops   2m19s

NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    2m18s
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Redis.

```bash
$  kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                                            AGE
petset.apps.k8s.appscode.com/rd-gitops-shard0   5m53s
petset.apps.k8s.appscode.com/rd-gitops-shard1   5m51s
petset.apps.k8s.appscode.com/rd-gitops-shard2   5m49s

NAME                     READY   STATUS    RESTARTS   AGE
pod/rd-gitops-shard0-0   1/1     Running   0          5m52s
pod/rd-gitops-shard0-1   1/1     Running   0          5m35s
pod/rd-gitops-shard1-0   1/1     Running   0          5m50s
pod/rd-gitops-shard1-1   1/1     Running   0          5m31s
pod/rd-gitops-shard2-0   1/1     Running   0          5m49s
pod/rd-gitops-shard2-1   1/1     Running   0          5m31s

NAME                      TYPE                       DATA   AGE
secret/rd-gitops-auth     kubernetes.io/basic-auth   2      5m55s
secret/rd-gitops-b3b686   Opaque                     1      5m55s

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)              AGE
service/rd-gitops        ClusterIP   10.43.156.126   <none>        6379/TCP             5m55s
service/rd-gitops-pods   ClusterIP   None            <none>        6379/TCP,16379/TCP   5m55s

NAME                                           TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/rd-gitops   kubedb.com/redis   8.0.4     5m49s
```

## Update Redis Database using GitOps

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
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
  podTemplate:
    spec:
      containers:
      - name: redis
        resources:
          requests:
            memory: "300Mi"
            cpu: "200m"
          limits:
            memory: "300Mi"
            cpu: "200m"
  deletionPolicy: WipeOut
 ```
Update the `replicas` to `3`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` RedisOpsRequest to update the `Redis` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    55m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   55m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-ule15j   HorizontalScaling   Successful   16m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kkubectl get pod -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                 READY   STATUS    RESTARTS   AGE
rd-gitops-shard0-0   1/1     Running   0          55m
rd-gitops-shard0-1   1/1     Running   0          55m
rd-gitops-shard0-2   1/1     Running   0          16m
rd-gitops-shard1-0   1/1     Running   0          55m
rd-gitops-shard1-1   1/1     Running   0          55m
rd-gitops-shard1-2   1/1     Running   0          16m
rd-gitops-shard2-0   1/1     Running   0          55m
rd-gitops-shard2-1   1/1     Running   0          55m
rd-gitops-shard2-2   1/1     Running   0          15m
```

We can also scale down the replicas by updating the `replicas` fields.

### Scale Redis Database Resources

before Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo rd-gitops-shard0-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
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
    storageClassName: "longhorn"
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


Resource Requests and Limits are updated to `700m` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

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

kubectl get pod -n demo rd-gitops-shard0-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "500Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "500Mi"
  }
}


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
    storageClassName: "longhorn"
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

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` RedisOpsRequest to update the `Redis` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    17h

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   17h

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-ule15j   HorizontalScaling   Successful   16h
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-lliwo8     VerticalScaling     Successful   16h
redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-3twy3x     VolumeExpansion     Successful   2m23s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-rd-gitops-shard0-0   Bound    pvc-77634b30-7ccc-4391-a768-54d4e3c1907c   2Gi        RWO            longhorn       <unset>                 17h
data-rd-gitops-shard0-1   Bound    pvc-e4b23a61-5de7-4e5c-b4b2-0722bffc7464   2Gi        RWO            longhorn       <unset>                 17h
data-rd-gitops-shard0-2   Bound    pvc-c3ffc095-248b-4091-bfaa-a1de401e3e36   2Gi        RWO            longhorn       <unset>                 16h
data-rd-gitops-shard1-0   Bound    pvc-62e7a6c0-ddd5-4281-8878-9ec42e68585d   2Gi        RWO            longhorn       <unset>                 17h
data-rd-gitops-shard1-1   Bound    pvc-0e9ca41e-8b25-4ea1-865a-17aba48f619c   2Gi        RWO            longhorn       <unset>                 17h
data-rd-gitops-shard1-2   Bound    pvc-c1e6806d-bb6c-4521-94c6-8e47ec6ad36e   2Gi        RWO            longhorn       <unset>                 16h
data-rd-gitops-shard2-0   Bound    pvc-0954e736-e699-46f4-9dfb-33b23f65e351   2Gi        RWO            longhorn       <unset>                 17h
data-rd-gitops-shard2-1   Bound    pvc-1fd1321a-ff04-4c81-b903-80473598ffac   2Gi        RWO            longhorn       <unset>                 17h
data-rd-gitops-shard2-2   Bound    pvc-8458eb68-930b-4817-8577-24616079e9f7   2Gi        RWO            longhorn       <unset>                 16h
```

## Reconfigure Redis

At first, we will create a secret containing `user.conf` file with required configuration settings.
To know more about this configuration file, check [here](/docs/guides/Redis/configuration/Redis-combined.md)
```yaml
apiVersion: v1
stringData:
  redis.conf: |
    maxclients 500
kind: Secret
metadata:
  name: rd-new-configuration
  namespace: demo
type: Opaque
```

Now, we will add this file to `kubedb /rd-configuration.yaml`.

```bash
$ tree .
├── kubedb
│ ├── rd-configuration.yaml
│ └── Redis.yaml
1 directories, 2 files
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
  configuration:
    secretName: rd-new-configuration
  storage:
    resources:
      requests:
        storage: "2Gi"
    storageClassName: "longhorn"
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

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` RedisOpsRequest to update the `Redis` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    22m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   22m

NAME                                                          TYPE          STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-5v9956   Reconfigure   Successful   15m
```



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
    storageClassName: "longhorn"
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
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    33m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   33m

NAME                                                          TYPE          STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-5v9956   Reconfigure   Successful   26m
redisopsrequest.ops.kubedb.com/rd-gitops-rotate-auth-oz83g9   RotateAuth    Successful   7m39s
```


### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

To add tls, we are going to create an example `Issuer` that will be used to enable SSL/TLS in Redis. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls redis-ca \
                                       --cert=ca.crt \
                                       --key=ca.key \
                                       --namespace=demo
secret/redis-ca created
```

Now, Let's create an `Issuer` using the `Redis-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: rd-issuer
  namespace: demo
spec:
  ca:
    secretName: redis-ca
```

Let's add that to our `kubedb /rd-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── rd-configuration.yaml
│ ├── rd-issuer.yaml
│ └── Redis.yaml
1 directories, 4 files
```

Update the `Redis.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Redis
metadata:
  name: rd-gitops
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Redis
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: Standard
    controller:
      podTemplate:
        spec:
          containers:
            - name: Redis
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: Standard
  storageType: Durable
  deletionPolicy: WipeOut
```

Add `sslMode` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` RedisOpsRequest to update the `Redis` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get rd,Redis,rdops,pods -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Redis.kubedb.com/rd-gitops   kubedb.com/v1   3.9.0     Ready    41m

NAME                                 AGE
Redis.gitops.kubedb.com/rd-gitops   75m

NAME                                                               TYPE              STATUS       AGE
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-ukj41o       Reconfigure       Successful   5d18h
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfiguretls-r4mx7v    ReconfigureTLS    Successful   9m18s
Redisopsrequest.ops.kubedb.com/rd-gitops-rotate-auth-43ris8       RotateAuth        Successful   5d1h
Redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   5d19h

```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for Redis.

### Update Version

List Redis versions using `kubectl get Redisversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/Redis/update-version/update-version.md).

Let's choose `4.0.0` in this example.

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
    storageClassName: "longhorn"
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

Update the `version` field to `17.4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` RedisOpsRequest to update the `Redis` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    14m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   14m

NAME                                                            TYPE            STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-versionupdate-xvnekk   UpdateVersion   Successful   11m
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
  version: 4.0.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Redis
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: Standard
    controller:
      podTemplate:
        spec:
          containers:
            - name: Redis
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: Standard
  storageType: Durable
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9091
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` RedisOpsRequest to add the `Redis` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get Redises.gitops.kubedb.com,Redises.kubedb.com,Redisopsrequest -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Redis.kubedb.com/rd-gitops   kubedb.com/v1   4.0.0     Ready    5h12m

NAME                                 AGE
Redis.gitops.kubedb.com/rd-gitops   5h12m

NAME                                                               TYPE              STATUS       AGE
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-ukj41o       Reconfigure       Successful   6d
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfiguretls-r4mx7v    ReconfigureTLS    Successful   5h42m
Redisopsrequest.ops.kubedb.com/rd-gitops-restart-ljpqih           Restart           Successful   3m51s
Redisopsrequest.ops.kubedb.com/rd-gitops-rotate-auth-43ris8       RotateAuth        Successful   5d7h
Redisopsrequest.ops.kubedb.com/rd-gitops-versionupdate-wyn2dp     UpdateVersion     Successful   5h16m
Redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   6d

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

[//]: # (- Learn Redis [GitOps]&#40;/docs/guides/Redis/concepts/Redis-gitops.md&#41;)
- Learn Redis Scaling
    - [Horizontal Scaling](/docs/guides/Redis/scaling/horizontal-scaling/combined.md)
    - [Vertical Scaling](/docs/guides/Redis/scaling/vertical-scaling/combined.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/Redis/update-version/overview.md)
- Monitor your RedisQL database with KubeDB using [built-in Prometheus](/docs/guides/Redis/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
