---
title: Elasticsearch Gitops Overview
description: Elasticsearch Gitops Overview
menu:
  docs_{{ .version }}:
    identifier: es-gitops-overview
    name: Overview
    parent: es-gitops
    weight: 10
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
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace> --ssh-private-key-path ~/.ssh/id_rsa
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
  enableSSL: true
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
$ kubectl get Elasticsearch.gitops.kubedb.com,Elasticsearch.kubedb.com -n demo
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
        storage: 1Gi
  deletionPolicy: WipeOut
 ```

Resource Requests and Limits are updated to `500m` CPU and `1Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

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
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "1Gi"
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
  enableSSL: true
  replicas: 3
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
stringData:
  user.conf: |
    max_connections=200
    shared_buffers=256MB    
kind: Secret
metadata:
  name: es-configuration
  namespace: demo
type: Opaque
```

Now, we will add this file to `kubedb /es-configuration.yaml`.

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

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

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

```bash
$ kubectl create secret generic es-rotate-auth -n demo \
                                      --type=kubernetes.io/basic-auth \
                                      --from-literal=username=elastic \
                                      --from-literal=password=elasticsearch-secret
secret/es-rotate-auth created

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

Change the `authSecret` field to `es-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` ElasticsearchOpsRequest to update the `Elasticsearch` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-9.1.9   Ready    18m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-wiezxp   HorizontalScaling   Successful   13m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-i9puam         Reconfigure         Successful   13m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-xisot8         RotateAuth          Successful   4m38s
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-7dyz6e     VerticalScaling     Successful   13m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-pptef3     VolumeExpansion     Successful   13m
```


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
$ kubectl create secret tls Elasticsearch-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/Elasticsearch-ca created
```

Now, Let's create an `Issuer` using the `Elasticsearch-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: es-issuer
  namespace: demo
spec:
  ca:
    secretName: Elasticsearch-ca
```

Let's add that to our `kubedb /es-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── es-configuration.yaml
│ ├── es-issuer.yaml
│ └── Elasticsearch.yaml
1 directories, 4 files
```

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Elasticsearch
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
            - name: Elasticsearch
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

Add `sslMode` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` ElasticsearchOpsRequest to update the `Elasticsearch` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get es,Elasticsearch,esops,pods -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Elasticsearch.kubedb.com/es-gitops   kubedb.com/v1   3.9.0     Ready    41m

NAME                                 AGE
Elasticsearch.gitops.kubedb.com/es-gitops   75m

NAME                                                               TYPE              STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-ukj41o       Reconfigure       Successful   5d18h
Elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfiguretls-r4mx7v    ReconfigureTLS    Successful   9m18s
Elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-43ris8       RotateAuth        Successful   5d1h
Elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   5d19h

```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for Elasticsearch.

### Update Version

List Elasticsearch versions using `kubectl get Elasticsearchversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/elasticsearch/update-version/update-version.md).

Let's choose `4.0.0` in this example.

Update the `Elasticsearch.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: es-gitops
  namespace: demo
spec:
  version: xpack-9.2.3
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

Update the `version` field to `17.4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Elasticsearch` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` ElasticsearchOpsRequest to update the `Elasticsearch` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get es,Elasticsearch,esops -n demo
NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-9.2.3   Ready    64m

NAME                                        AGE
elasticsearch.gitops.kubedb.com/es-gitops   64m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-os8fon   HorizontalScaling   Successful   58m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-hrbrs1         Reconfigure         Successful   58m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-s0tkka         RotateAuth          Successful   58m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-versionupdate-bn3rk4       UpdateVersion       Successful   31m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-vhzkxl     VerticalScaling     Successful   58m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-bzqd4r     VolumeExpansion     Successful   58m

```


Now, we are going to verify whether the `Elasticsearch`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get Elasticsearch -n demo es-gitops -o=jsonpath='{.spec.version}{"\n"}'
xpack-9.2.3

$ kubectl get petset -n demo es-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/elastic:9.2.3@sha256:714f35c53d333fef7e673079d05ce80440caf2c20ca3dc9b3e366728527760bb

$ kubectl get pod -n demo es-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/elastic:9.2.3@sha256:714f35c53d333fef7e673079d05ce80440caf2c20ca3dc9b3e366728527760bb
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
  version: xpack-9.2.3
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
$ kubectl get Elasticsearches.gitops.kubedb.com,Elasticsearches.kubedb.com,Elasticsearchopsrequest -n demo
NAME                                        AGE
elasticsearch.gitops.kubedb.com/es-gitops   106m

NAME                                 VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-gitops   xpack-9.2.3   Ready    106m

NAME                                                                        TYPE                STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-gitops-horizontalscaling-os8fon   HorizontalScaling   Successful   101m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-reconfigure-hrbrs1         Reconfigure         Successful   101m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-restart-7v0e29             Restart             Successful   15m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-rotate-auth-s0tkka         RotateAuth          Successful   101m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-versionupdate-bn3rk4       UpdateVersion       Successful   73m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-verticalscaling-vhzkxl     VerticalScaling     Successful   101m
elasticsearchopsrequest.ops.kubedb.com/es-gitops-volumeexpansion-bzqd4r     VolumeExpansion     Successful   101m

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

[//]: # (- Learn Elasticsearch [GitOps]&#40;/docs/guides/elasticsearch/concepts/Elasticsearch-gitops.md&#41;)
- Learn Elasticsearch Scaling
    - [Horizontal Scaling](/docs/guides/elasticsearch/scaling/horizontal/combined.md)
    - [Vertical Scaling](/docs/guides/elasticsearch/scaling/vertical/combined.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/elasticsearch/update-version/overview.md)
- Monitor your ElasticsearchQL database with KubeDB using [built-in Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)