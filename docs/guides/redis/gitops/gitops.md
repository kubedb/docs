---
title: Redis GitOps Guides
menu:
  docs_{{ .version }}:
    identifier: rd-gitops-guides
    name: Gitops Redis
    parent: rd-gitops-Redis
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
              cpu: "100m"
              memory: "100Mi"
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
redis.gitops.kubedb.com/rd-gitops   8m46s

NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    8m46s

```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Redis.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                                            AGE
petset.apps.k8s.appscode.com/rd-gitops-shard0   9m38s
petset.apps.k8s.appscode.com/rd-gitops-shard1   9m36s
petset.apps.k8s.appscode.com/rd-gitops-shard2   9m34s

NAME                     READY   STATUS    RESTARTS   AGE
pod/rd-gitops-shard0-0   1/1     Running   0          9m37s
pod/rd-gitops-shard0-1   1/1     Running   0          9m14s
pod/rd-gitops-shard1-0   1/1     Running   0          9m35s
pod/rd-gitops-shard1-1   1/1     Running   0          9m14s
pod/rd-gitops-shard2-0   1/1     Running   0          9m34s
pod/rd-gitops-shard2-1   1/1     Running   0          9m14s

NAME                      TYPE                       DATA   AGE
secret/rd-gitops-9a6ce0   Opaque                     1      9m40s
secret/rd-gitops-auth     kubernetes.io/basic-auth   2      65m

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)              AGE
service/rd-gitops        ClusterIP   10.43.193.228   <none>        6379/TCP             9m40s
service/rd-gitops-pods   ClusterIP   None            <none>        6379/TCP,16379/TCP   9m40s

NAME                                           TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/rd-gitops   kubedb.com/redis   8.0.4     9m34s

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
    shards: 4
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
              cpu: "100m"
              memory: "100Mi"
  deletionPolicy: WipeOut
 ```
Update the `replicas` to `3`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` RedisOpsRequest to update the `Redis` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    12m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   12m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-1nblpl   HorizontalScaling   Successful   10m

```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$  kubectl get pod -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                 READY   STATUS    RESTARTS   AGE
rd-gitops-shard0-0   1/1     Running   0          16m
rd-gitops-shard0-1   1/1     Running   0          16m
rd-gitops-shard0-2   1/1     Running   0          13m
rd-gitops-shard1-0   1/1     Running   0          16m
rd-gitops-shard1-1   1/1     Running   0          16m
rd-gitops-shard1-2   1/1     Running   0          12m
rd-gitops-shard2-0   1/1     Running   0          16m
rd-gitops-shard2-1   1/1     Running   0          16m
rd-gitops-shard2-2   1/1     Running   0          12m
rd-gitops-shard3-0   1/1     Running   0          14m
rd-gitops-shard3-1   1/1     Running   0          14m
rd-gitops-shard3-2   1/1     Running   0          11m

```

We can also scale down the replicas by updating the `replicas` fields.

### Scale Redis Database Resources

before Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo rd-gitops-shard0-0 -o json | jq '.spec.containers[0].resources'
{
  "requests": {
    "cpu": "100m",
    "memory": "100Mi"
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
    shards: 4
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
              cpu: "200m"
              memory: "200Mi"
  deletionPolicy: WipeOut
```


Resource Requests and Limits are updated to `700m` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `RedisOpsRequest` to update the `Redis` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                         VERSION   STATUS   AGE
redis.kubedb.com/rd-gitops   8.0.4     Ready    28m

NAME                                AGE
redis.gitops.kubedb.com/rd-gitops   28m

NAME                                                                TYPE                STATUS       AGE
redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-1nblpl   HorizontalScaling   Successful   26m
redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-or3shk     VerticalScaling     Successful   10m

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

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` RedisOpsRequest to update the `Redis` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Redis.kubedb.com/rd-gitops   kubedb.com/v1   3.9.0     Ready    23m

NAME                                 AGE
Redis.gitops.kubedb.com/rd-gitops   23m

NAME                                                                 TYPE                STATUS       AGE
Redisopsrequest.ops.kubedb.com/rd-gitops-horizontalscaling-j0wni6   HorizontalScaling   Successful   13m
Redisopsrequest.ops.kubedb.com/rd-gitops-verticalscaling-tfkvi8     VerticalScaling     Successful   8m29s
Redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-41xthr     VolumeExpansion     Successful   19m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=rd-gitops'
NAME                                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
rd-gitops-data-rd-gitops-broker-0       Bound    pvc-2afd4835-5686-492b-be93-c6e040e0a6c6   2Gi        RWO            Standard       <unset>                 3h39m
rd-gitops-data-rd-gitops-broker-1       Bound    pvc-aaf994cc-6b04-4c37-80d5-5e966dad8487   2Gi        RWO            Standard       <unset>                 3h39m
rd-gitops-data-rd-gitops-controller-0   Bound    pvc-82d2b233-203d-4df2-a0fd-ecedbc0825b7   2Gi        RWO            Standard       <unset>                 3h39m
rd-gitops-data-rd-gitops-controller-1   Bound    pvc-91852c29-ab1a-48ad-9255-a0b15d5a7515   2Gi        RWO            Standard       <unset>                 3h39m

```

## Reconfigure Redis

At first, we will create a secret containing `user.conf` file with required configuration settings.
To know more about this configuration file, check [here](/docs/guides/Redis/configuration/Redis-combined.md)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-rd-combined-custom-config
  namespace: demo
stringData:
  server.properties: |-
    log.retention.hours=125    
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
  configSecret:
    name: new-rd-combined-custom-config
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

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` RedisOpsRequest to update the `Redis` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Redis.kubedb.com/rd-gitops   kubedb.com/v1   3.9.0     Ready    74m

NAME                                 AGE
Redis.gitops.kubedb.com/rd-gitops   74m

NAME                                                               TYPE              STATUS       AGE
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-ukj41o       Reconfigure       Successful   24m
Redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   70m

```



> We can also reconfigure the parameters creating another secret and reference the secret in the `configSecret` field. Also you can remove the `configSecret` field to use the default parameters.

### Rotate Redis Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will create a secret named `rd-rotate-auth ` with the following content,

```bash
kubectl create secret generic rd-rotate-auth -n demo \
--type=kubernetes.io/basic-auth \
--from-literal=username=Redis \
--from-literal=password=Redis-secret
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
  authSecret:
    kind: Secret
    name: rd-rotate-auth
  configuration:
    secretName: new-rd-combined-custom-config
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

Change the `authSecret` field to `rd-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` RedisOpsRequest to update the `Redis` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get rd,redis,redisopsrequest -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Redis.kubedb.com/rd-gitops   kubedb.com/v1   3.9.0     Ready    7m11s

NAME                                 AGE
Redis.gitops.kubedb.com/rd-gitops   7m11s

NAME                                                               TYPE              STATUS       AGE
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-ukj41o       Reconfigure       Successful   17h
Redisopsrequest.ops.kubedb.com/rd-gitops-rotate-auth-43ris8       RotateAuth        Successful   28m
Redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   17h

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
$ kubectl create secret tls Redis-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/Redis-ca created
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
    secretName: Redis-ca
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
  deletionPolicy: WipeOut
```

Update the `version` field to `17.4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Redis` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` RedisOpsRequest to update the `Redis` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get rd,redis,redisopsrequest -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Redis.kubedb.com/rd-gitops   kubedb.com/v1   4.0.0     Ready    3h47m

NAME                                 AGE
Redis.gitops.kubedb.com/rd-gitops   3h47m

NAME                                                               TYPE              STATUS       AGE
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfigure-ukj41o       Reconfigure       Successful   5d22h
Redisopsrequest.ops.kubedb.com/rd-gitops-reconfiguretls-r4mx7v    ReconfigureTLS    Successful   4h16m
Redisopsrequest.ops.kubedb.com/rd-gitops-rotate-auth-43ris8       RotateAuth        Successful   5d6h
Redisopsrequest.ops.kubedb.com/rd-gitops-versionupdate-wyn2dp     UpdateVersion     Successful   3h51m
Redisopsrequest.ops.kubedb.com/rd-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   5d23h
```


Now, we are going to verify whether the `Redis`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get Redis -n demo rd-gitops -o=jsonpath='{.spec.version}{"\n"}'
4.0.0

$ kubectl get petset -n demo rd-gitops-broker -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/Redis:4.0.0@sha256:42a79fe8f14b00b1c76d135bbbaf7605b8c66f45cf3eb749c59138f6df288b31

$  kubectl get pod -n demo rd-gitops-broker-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/Redis:4.0.0@sha256:42a79fe8f14b00b1c76d135bbbaf7605b8c66f45cf3eb749c59138f6df288b31
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
