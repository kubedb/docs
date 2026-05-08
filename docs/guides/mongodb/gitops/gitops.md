---
title: MongoDB GitOps 
description: MongoDB GitOps 
menu:
  docs_{{ .version }}:
    identifier: mg-gitops
    name: Guide
    parent: mg-gitops-mongodb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---



> New to KubeDB? Please start [here](/docs/README.md).

# GitOps MongoDB using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create MongoDB database and manage updates using GitOps workflow.

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
> Note: YAML files used in this tutorial are stored in [docs/examples/MongoDB](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

## Create MongoDB Database using GitOps

### Create a MongoDB GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.10"
  replicaSet: 
    name: "replicaset"
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Create a directory like below,
```bash
$ tree .
├── kubedb
    └── MongoDB.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is created in your cluster.

Our `gitops` operator will create an actual `MongoDB` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$  kubectl get MongoDB.gitops.kubedb.com,MongoDB.kubedb.com -n demo
NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   33m

NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.10    Ready    33m

```

List the resources created by `kubedb` operator created for `kubedb.com/v1` MongoDB.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=mg-gitops'
NAME                                     AGE
petset.apps.k8s.appscode.com/mg-gitops   34m

NAME              READY   STATUS    RESTARTS   AGE
pod/mg-gitops-0   2/2     Running   0          34m
pod/mg-gitops-1   2/2     Running   0          33m

NAME                    TYPE                       DATA   AGE
secret/mg-gitops-auth   kubernetes.io/basic-auth   2      34m
secret/mg-gitops-key    Opaque                     1      34m

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/mg-gitops        ClusterIP   10.43.206.183   <none>        27017/TCP   34m
service/mg-gitops-pods   ClusterIP   None            <none>        27017/TCP   34m

NAME                                           TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mg-gitops   kubedb.com/mongodb   8.0.10    34m

```

## Update MongoDB Database using GitOps

### Scale MongoDB Database Resources

Before scaling database resouces:

```shell
kubectl get pod -n demo mg-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "800m",
    "memory": "1536Mi"
  }
}
```
Update the `MongoDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.10"
  replicaSet: 
    name: "replicaset"
  replicas: 2
  podTemplate:
   spec:
     containers:
     - name: mongodb
       resources:
         limits:
           memory: 2Gi
         requests:
           cpu: 1000m
           memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
 ```

Resource Requests and Limits are updated from `800m` to `1000m` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `MongoDBOpsRequest` to update the `MongoDB` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.10    Ready    13m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   13m

NAME                                                                TYPE              STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-verticalscaling-ojwxpm   VerticalScaling   Successful   4m35s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo mg-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}

```

### Scale MongoDB Replicas
Update the `MongoDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.10"
  replicaSet: 
    name: "replicaset"
  replicas: 3
  podTemplate:
   spec:
     containers:
     - name: mongodb
       resources:
         limits:
           memory: 2Gi
         requests:
           cpu: 1000m
           memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Update the `replicas` to `3`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` MongoDBOpsRequest to update the `MongoDB` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.10    Ready    18m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   18m

NAME                                                                  TYPE                STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-horizontalscaling-n8xx64   HorizontalScaling   Successful   4m2s
mongodbopsrequest.ops.kubedb.com/mg-gitops-verticalscaling-ojwxpm     VerticalScaling     Successful   9m5s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=mg-gitops'
NAME          READY   STATUS    RESTARTS   AGE
mg-gitops-0   2/2     Running   0          8m37s
mg-gitops-1   2/2     Running   0          9m22s
mg-gitops-2   2/2     Running   0          4m34s
```

We can also scale down the replicas by updating the `replicas` fields.

### Expand MongoDB Volume

Update the `MongoDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.10"
  replicaSet: 
    name: "replicaset"
  replicas: 3
  podTemplate:
   spec:
     containers:
     - name: mongodb
       resources:
         limits:
           memory: 2Gi
         requests:
           cpu: 1000m
           memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` MongoDBOpsRequest to update the `MongoDB` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.10    Ready    21m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   21m

NAME                                                                  TYPE                STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-horizontalscaling-n8xx64   HorizontalScaling   Successful   7m24s
mongodbopsrequest.ops.kubedb.com/mg-gitops-verticalscaling-ojwxpm     VerticalScaling     Successful   12m
mongodbopsrequest.ops.kubedb.com/mg-gitops-volumeexpansion-8441ym     VolumeExpansion     Successful   2m10s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=mg-gitops'
NAME                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
datadir-mg-gitops-0   Bound    pvc-cea7fe6a-dd75-4e81-99d3-9ab2867c6650   2Gi        RWO            longhorn       <unset>                 22m
datadir-mg-gitops-1   Bound    pvc-bcd63bd2-b3b8-4fb8-8c35-5f6e40031f61   2Gi        RWO            longhorn       <unset>                 21m
datadir-mg-gitops-2   Bound    pvc-2535f213-28fb-41ef-bdfd-7fbe91859c81   2Gi        RWO            longhorn       <unset>               7m56s
```

## Reconfigure MongoDB

At first, we will create `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 10000
```
Here, `maxIncomingConnections` is set to `10000`, whereas the default value is `65536`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo mg-custom-config --from-file=./mongod.conf
secret/mg-custom-config created
```


Update the `MongoDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.10"
  replicaSet: 
    name: "replicaset"
  replicas: 3
  podTemplate:
   spec:
     containers:
     - name: mongodb
       resources:
         limits:
           memory: 2Gi
         requests:
           cpu: 1000m
           memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: mg-custom-config
```

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` MongoDBOpsRequest to update the `MongoDB` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.10    Ready    32m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   32m

NAME                                                                  TYPE                STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-horizontalscaling-n8xx64   HorizontalScaling   Successful   17m
mongodbopsrequest.ops.kubedb.com/mg-gitops-reconfigure-djow20         Reconfigure         Successful   6m7s
mongodbopsrequest.ops.kubedb.com/mg-gitops-verticalscaling-ojwxpm     VerticalScaling     Successful   22m
mongodbopsrequest.ops.kubedb.com/mg-gitops-volumeexpansion-8441ym     VolumeExpansion     Successful   12m
```

We can also reconfigure the parameters creating another secret and reference the secret in the `configuration.secretName` field. Also you can remove the `configuration` field to use the default parameters.

### Rotate Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will do that using gitops, create the file `kubedb/mg-auth.yaml` with the following content,

```bash
apiVersion: v1
kind: Secret
metadata:
  name: mgauth
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: root
  password: mongodb-secret
```

Let's add that to our `kubedb/mg-auth.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── mg-configuration.yaml
│ ├── mg-auth.yaml
│ └── mongodb.yaml
1 directories, 3 files
```



Update the `MongoDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.10"
  replicaSet: 
    name: "replicaset"
  replicas: 3
  podTemplate:
   spec:
     containers:
     - name: mongodb
       resources:
         limits:
           memory: 2Gi
         requests:
           cpu: 1000m
           memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: mg-custom-config
  authSecret:
    kind: Secret
    name: mgauth
```

Add the secret name in `authSecret` field. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` MongoDBOpsRequest to update the `MongoDB` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.10    Ready    41m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   41m

NAME                                                                  TYPE                STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-horizontalscaling-n8xx64   HorizontalScaling   Successful   27m
mongodbopsrequest.ops.kubedb.com/mg-gitops-reconfigure-djow20         Reconfigure         Successful   15m
mongodbopsrequest.ops.kubedb.com/mg-gitops-rotate-auth-u75ihg         RotateAuth          Successful   3m10s
mongodbopsrequest.ops.kubedb.com/mg-gitops-verticalscaling-ojwxpm     VerticalScaling     Successful   32m
mongodbopsrequest.ops.kubedb.com/mg-gitops-volumeexpansion-8441ym     VolumeExpansion     Successful   22m
```

### Update Version

List MongoDB versions using `kubectl get mgversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/mongodb/update-version/replicaset.md).

Let's choose `8.0.17` in this example.

Update the `MongoDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
 name: mg-gitops
 namespace: demo
spec:
 version: "8.0.17"
 replicaSet:
   name: "replicaset"
 replicas: 3
 podTemplate:
   spec:
     containers:
     - name: mongodb
       resources:
         limits:
           memory: 1Gi
         requests:
           cpu: "1"
           memory: 1Gi
 storageType: Durable
 storage:
   storageClassName: longhorn
   accessModes:
   - ReadWriteOnce
   resources:
     requests:
       storage: 1Gi
```

Update the `version` field to `8.0.17`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` MongoDBOpsRequest to update the `MongoDB` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.17    Ready    46m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   46m

NAME                                                                  TYPE                STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-horizontalscaling-n8xx64   HorizontalScaling   Successful   31m
mongodbopsrequest.ops.kubedb.com/mg-gitops-reconfigure-djow20         Reconfigure         Successful   20m
mongodbopsrequest.ops.kubedb.com/mg-gitops-rotate-auth-u75ihg         RotateAuth          Successful   7m19s
mongodbopsrequest.ops.kubedb.com/mg-gitops-versionupdate-kkc2gc       UpdateVersion       Successful   2m38s
mongodbopsrequest.ops.kubedb.com/mg-gitops-verticalscaling-ojwxpm     VerticalScaling     Successful   36m
mongodbopsrequest.ops.kubedb.com/mg-gitops-volumeexpansion-8441ym     VolumeExpansion     Successful   26m
```


Now, we are going to verify whether the `MongoDB`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get MongoDB -n demo mg-gitops -o=jsonpath='{.spec.version}{"\n"}'
8.0.17
$ kubectl get petset -n demo mg-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/mongo:8.0.17@sha256:b3e1ae71bd7df56b3497527f2b08549bfccb532d9e26df6d4a1331a71cd085db
$ kubectl get pod -n demo mg-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/mongo:8.0.17@sha256:b3e1ae71bd7df56b3497527f2b08549bfccb532d9e26df6d4a1331a71cd085db
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `MongoDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.17"
  replicaSet: 
    name: "replicaset"
  replicas: 3
  podTemplate:
    spec:
      containers:
      - name: mongodb
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 1000m
            memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: mg-custom-config
  authSecret:
    kind: Secret
    name: mgauth
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` MongoDBOpsRequest to add the `MongoDB` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.17    Ready    52m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   52m

NAME                                                                  TYPE                STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-horizontalscaling-n8xx64   HorizontalScaling   Successful   38m
mongodbopsrequest.ops.kubedb.com/mg-gitops-reconfigure-djow20         Reconfigure         Successful   26m
mongodbopsrequest.ops.kubedb.com/mg-gitops-restart-ykgxj3             Restart             Successful   3m45s
mongodbopsrequest.ops.kubedb.com/mg-gitops-rotate-auth-u75ihg         RotateAuth          Successful   13m
mongodbopsrequest.ops.kubedb.com/mg-gitops-versionupdate-kkc2gc       UpdateVersion       Successful   9m15s
mongodbopsrequest.ops.kubedb.com/mg-gitops-verticalscaling-ojwxpm     VerticalScaling     Successful   43m
mongodbopsrequest.ops.kubedb.com/mg-gitops-volumeexpansion-8441ym     VolumeExpansion     Successful   32m
```

Verify the monitoring is enabled by checking the prometheus targets.



### TLS configuration

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in MongoDB. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mongo/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls mongo-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mongo-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: mongo-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/tls/issuer.yaml
issuer.cert-manager.io/mongo-ca-issuer created
```

Let's add that to our `kubedb/mg-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── mg-configuration.yaml
│ ├── mg-auth.yaml
│ ├── mg-issuer.yaml
│ └── mongodb.yaml
1 directories, 4 files
```

Update the `mongodb.yaml` with the following,

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-gitops
  namespace: demo
spec:
  version: "8.0.17"
  replicaSet: 
    name: "replicaset"
  replicas: 3
  podTemplate:
   spec:
     containers:
     - name: mongodb
       resources:
         limits:
           memory: 2Gi
         requests:
           cpu: 1000m
           memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: mg-custom-config
  authSecret:
    kind: Secret
    name: mgauth
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  tls:
    issuerRef:
      name: mg-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - mongo
          organizationalUnits:
            - client
```
Add `sslMode` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MongoDB` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` ElasticsearchOpsRequest to update the `MongoDB` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mg,mongodb,mgops -n demo
NAME                           VERSION   STATUS   AGE
mongodb.kubedb.com/mg-gitops   8.0.17    Ready    20m

NAME                                  AGE
mongodb.gitops.kubedb.com/mg-gitops   20m

NAME                                                               TYPE             STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-gitops-reconfiguretls-2pzvw4   ReconfigureTLS   Successful   10m
```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for MongoDB.

## Next Steps

- Learn MongoDB Scaling
    - [Horizontal Scaling](/docs/guides/mongodb/scaling/horizontal-scaling/replicaset.md)
    - [Vertical Scaling](/docs/guides/mongodb/scaling/vertical-scaling/replicaset.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/mongodb/update-version/replicaset.md)
- Monitor your ElasticsearchQL database with KubeDB using [built-in Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
