---
title: Elasticsearch GitOps Details
description: Elasticsearch GitOps
menu:
  docs_{{ .version }}:
    identifier: gitops-elasticsearch
    name: Guides
    parent: es-gitops
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---



> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Elasticsearch using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create Elasticsearch database and manage updates using GitOps woreslow.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, 
you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).  Pass `--set kubedb-crd-manager.installGitOpsCRDs=true` in the kubedb installation 
process to enable `GitOps` operator.

- You need to install GitOps tools like `ArgoCD` or `FluxCD` and configure with your Git Repository to monitor the Git repository and synchronize the state of the Kubernetes 
cluster with the desired state defined in Git.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```
> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

Registering a Git repository in Argo CD using SSH authentication

```bash
argocd repo add <ssh-repo-url> \
  --ssh-private-key-path <path-to-private-key>
```
Creating an Argo CD Application to deploy resources from the repository into a Kubernetes cluster
```bash
argocd app create <application-name> \
  --repo <repository-url> \
  --path <repository-path> \
  --dest-server <kubernetes-api-server> \
  --dest-namespace <target-namespace>
```

## Create Elasticsearch Database using GitOps

### Create a Elasticsearch GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.2.3
  enableSSL: false
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Create a directory like below,
```bash
$ tree .
├── kubedb
    └── Elasticsearch.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is created in your cluster.

Our `gitops` operator will create an actual `Elasticsearch` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get elasticsearch.gitops.kubedb.com,elasticsearch.kubedb.com -n demo
NAME                                        AGE
elasticsearch.gitops.kubedb.com/es-gitops   20m

NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.2.3   Ready    20m
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Elasticsearch.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=es-gitops'
NAME                                     AGE
petset.apps.k8s.appscode.com/es-gitops   20m

NAME              READY   STATUS    RESTARTS   AGE
pod/es-gitops-0   1/1     Running   0          20m
pod/es-gitops-1   1/1     Running   0          19m

NAME                                           TYPE                       DATA   AGE
secret/es-gitops-a1d7d4                        Opaque                     1      13m
secret/es-gitops-apm-system-cred               kubernetes.io/basic-auth   2      20m
secret/es-gitops-auth                          kubernetes.io/basic-auth   2      20m
secret/es-gitops-beats-system-cred             kubernetes.io/basic-auth   2      20m
secret/es-gitops-ca-cert                       kubernetes.io/tls          2      20m
secret/es-gitops-client-cert                   kubernetes.io/tls          3      20m
secret/es-gitops-config                        Opaque                     1      20m
secret/es-gitops-http-cert                     kubernetes.io/tls          3      20m
secret/es-gitops-kibana-system-cred            kubernetes.io/basic-auth   2      20m
secret/es-gitops-logstash-system-cred          kubernetes.io/basic-auth   2      20m
secret/es-gitops-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      20m
secret/es-gitops-transport-cert                kubernetes.io/tls          3      20m

NAME                       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/es-gitops          ClusterIP   10.43.235.79   <none>        9200/TCP   20m
service/es-gitops-master   ClusterIP   None           <none>        9300/TCP   20m
service/es-gitops-pods     ClusterIP   None           <none>        9200/TCP   20m

NAME                                           TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-gitops   kubedb.com/elasticsearch   8.2.3     20m
```

## Update Elasticsearch Database using GitOps
### Scale Elasticsearch Replicas
Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.2.3
  enableSSL: true
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Update the `replicas` to `3`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` ElasticsearchOpsRequest to update the `Elasticsearch` database replicas. List the 
resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.2.3   Ready    64m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-32p116   HorizontalScaling   Successful   39m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=es-gitops'
NAME          READY   STATUS    RESTARTS   AGE
es-gitops-0   1/1     Running   0          36m
es-gitops-1   1/1     Running   0          16m
es-gitops-2   1/1     Running   0          15m
```


We can also scale down the replicas by updating the `replicas` fields.
### Scale Elasticsearch Database Resources

Before scaling database resouces:
```bash
kubectl get pod -n demo es-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
```

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.2.3 
  enableSSL: false
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
        - name: elasticsearch
          resources:
            limits:
              cpu: 1000m
              memory: 2Gi
            requests:
              cpu: 1000m
              memory: 2Gi

 ```

Resource Requests and Limits are updated to `1000m` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `ElasticsearchOpsRequest` to update the `Elasticsearch` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.2.3   Ready    64m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-injx1l   HorizontalScaling   Successful   15m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-x5mfy0     VerticalScaling     Successful   39m

```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo es-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}
```


### Expand Elasticsearch Volume

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.2.3 
  enableSSL: false
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
        - name: elasticsearch
          resources:
            limits:
              cpu: 1000m
              memory: 2Gi
            requests:
              cpu: 1000m
              memory: 2Gi
```

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` ElasticsearchOpsRequest to update the `Elasticsearch` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get es,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.2.3   Ready    3h1m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-32p116   HorizontalScaling   Successful   157m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-x5mfy0     VerticalScaling     Successful   157m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-sata37     VolumeExpansion     Successful   38m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=es-gitops'
NAME          READY   STATUS    RESTARTS   AGE
es-gitops-0   1/1     Running   0          36m
es-gitops-1   1/1     Running   0          16m
es-gitops-2   1/1     Running   0          15m
```

## Reconfigure Elasticsearch

At first, we will create a secret containing `user.conf` file with required configuration settings.
To know more about this configuration file, check [here](/docs/guides/elasticsearch/configuration/combined-cluster/index.md)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-configuration
  namespace: demo
stringData:
  elasticsearch.yml: |-
    indices.query.bool.max_clause_count: 2048
```

Now, we will add this file to `kubedb/es-configuration.yaml`.

```bash
$ tree .
├── kubedb
│ ├── es-configuration.yaml
│ └── Elasticsearch.yaml
1 directories, 2 files
```

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.2.3
  enableSSL: true
  replicas: 3
  configuration:
    secretName: es-configuration
  podTemplate:
    spec:
      containers:
        - name: elasticsearch
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 100m
              memory: 1Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD`. THe reconfig file is created and  the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` ElasticsearchOpsRequest to update the `Elasticsearch` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.2.3   Ready    3h53m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-32p116   HorizontalScaling   Successful   3h29m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-wj5qyx         Reconfigure         Successful   3m42s
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-lvh38k     VerticalScaling     Successful   99m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-sata37     VolumeExpansion     Successful   90m
```



### Rotate Elasticsearch Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will generate a secret named `es-rotate-auth` with the following content,

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-rotate-auth
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: elastic
  password: elasticsearch-secret
```


Now, we will add this file to `kubedb/es-rotateauth.yaml`.

```bash
$ tree .
├── kubedb
│ ├── es-configuration.yaml
│ ├── es-rotateauth.yaml
│ └── Elasticsearch.yaml
1 directories, 3 files
```

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.2.3
  enableSSL: true
  replicas: 3
  authSecret:
    kind: Secret
    name: es-rotate-auth
  configuration:
    secretName: es-configuration
  podTemplate:
    spec:
      containers:
        - name: elasticsearch
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 100m
              memory: 1Gi
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Change the `authSecret` field to `es-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD`, the authentication file is created  the `Elasticsearch` CR is updated and the authentication file is created in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` ElasticsearchOpsRequest to update the `Elasticsearch` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.2.3   Ready    32m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-kn71nl   HorizontalScaling   Successful   28m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-x7ou3f         Reconfigure         Successful   7m24s
elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-8cgx3b         RotateAuth          Successful   2m34s
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-wyjx4l     VerticalScaling     Successful   21m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-z2e3qb     VolumeExpansion     Successful   17m
```
### Update Version

List Elasticsearch versions using `kubectl get Elasticsearchversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/elasticsearch/update-version/elasticsearch.md).

Let's choose `xpack-8.5.3` in this example.

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.5.3
  enableSSL: false
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  configuration:
    secretName: es-configuration
  authSecret:
    kind: Secret
    name: es-rotate-auth
  podTemplate:
    spec:
      containers:
        - name: elasticsearch
          resources:
            limits:
              cpu: 1000m
              memory: 2Gi
            requests:
              cpu: 1000m
              memory: 2Gi
```

Update the `version` field to `xpack-8.5.3`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` ElasticsearchOpsRequest to update the `Elasticsearch` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,elasticsearch,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.5.3   Ready    54m

NAME                                        AGE
elasticsearch.gitops.kubedb.com/es-gitops   54m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-kn71nl   HorizontalScaling   Successful   50m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-x7ou3f         Reconfigure         Successful   29m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-8cgx3b         RotateAuth          Successful   25m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-versionupdate-z92dz0       UpdateVersion       Successful   17m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-wyjx4l     VerticalScaling     Successful   44m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-z2e3qb     VolumeExpansion     Successful   39m
```


Now, we are going to verify whether the `Elasticsearch`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get Elasticsearch -n demo es-gitops -o=jsonpath='{.spec.version}{"\n"}'
xpack-8.5.3
$ kubectl get petset -n demo es-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/elastic:8.5.3@sha256:705db3750fa6739ddf06e416b52788446b19f79b6b162c0238078790545898d0
$ kubectl get pod -n demo es-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/elastic:8.5.3@sha256:705db3750fa6739ddf06e416b52788446b19f79b6b162c0238078790545898d0
```



### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.5.3
  enableSSL: false
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  configuration:
    secretName: es-configuration
  authSecret:
    kind: Secret
    name: es-rotate-auth
  podTemplate:
    spec:
      containers:
        - name: elasticsearch
          resources:
            limits:
              cpu: 1000m
              memory: 2Gi
            requests:
              cpu: 1000m
              memory: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` ElasticsearchOpsRequest to add the `Elasticsearch` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get es,elasticsearch,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.5.3   Ready    66m

NAME                                        AGE
elasticsearch.gitops.kubedb.com/es-gitops   66m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-kn71nl   HorizontalScaling   Successful   61m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-x7ou3f         Reconfigure         Successful   40m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-restart-sd2424             Restart             Successful   4m10s
elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-8cgx3b         RotateAuth          Successful   36m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-versionupdate-z92dz0       UpdateVersion       Successful   28m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-wyjx4l     VerticalScaling     Successful   55m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-z2e3qb     VolumeExpansion     Successful   50m
```

Verify the monitoring is enabled by checking the prometheus targets.

There are some other fields that will trigger `Restart` ops request.
- `.spec.monitor`
- `.spec.enableSSL` etc.

### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

To add tls, we are going to create an example `Issuer` that will be used to enable SSL/TLS in Elasticsearch. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls es-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/es-ca created
```

Now, Let's create an `Issuer` using the `elasticsearch-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: es-issuer
  namespace: demo
spec:
  ca:
    secretName: elasticsearch-ca
```

Let's add that to our `kubedb/es-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── es-configuration.yaml
│ ├── es-rotateauth.yaml
│ ├── es-secret.yaml
│ ├── es-issuer.yaml
│ └── Elasticsearch.yaml
1 directories, 5 files
```

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-8.5.3
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: es-issuer
    certificates:
    - alias: http
      subject:
        organizations:
        - kubedb.com
      emailAddresses:
      - abc@kubedb.com
  enableSSL: true
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  configuration:
    secretName: es-configuration
  authSecret:
    kind: Secret
    name: es-rotate-auth
  podTemplate:
    spec:
      containers:
        - name: elasticsearch
          resources:
            limits:
              cpu: 1000m
              memory: 2Gi
            requests:
              cpu: 1000m
              memory: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Add `enableSSL: true` and `tls` fields in the spec.
 Commit the changes and push to your Git repository. Your repository has been successfully synchronized with ArgoCD. The `Elasticsearch` CR has been updated, and both the `issuer` and the corresponding `secret` have been created in the cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` ElasticsearchOpsRequest to update the `Elasticsearch` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,elasticsearch,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-8.5.3   Ready    66m

NAME                                        AGE
elasticsearch.gitops.kubedb.com/es-gitops   66m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-kn71nl   HorizontalScaling   Successful   61m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-x7ou3f         Reconfigure         Successful   40m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-restart-sd2424             Restart             Successful   4m10s
elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-8cgx3b         RotateAuth          Successful   36m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-versionupdate-z92dz0       UpdateVersion       Successful   28m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-wyjx4l     VerticalScaling     Successful   55m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-z2e3qb     VolumeExpansion     Successful   50m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfiguretls-r4mx7v      ReconfigureTLS      Successful   9m18s
```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for Elasticsearch.


## Next Steps

- Learn Elasticsearch Scaling
    - [Horizontal Scaling](/docs/guides/elasticsearch/scaling/horizontal/combined.md)
    - [Vertical Scaling](/docs/guides/elasticsearch/scaling/vertical/combined.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/elasticsearch/update-version/elasticsearch.md)
- Monitor your ElasticsearchQL database with KubeDB using [built-in Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)