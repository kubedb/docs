---
title: MariaDB GitOps
description: MariaDB GitOps Overview
menu:
  docs_{{ .version }}:
    identifier: md-gitops
    name: MariaDB GitOps
    parent: guides-mariadb-gitops
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---



> New to KubeDB? Please start [here](/docs/README.md).

# GitOps MariaDB using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create MariaDB database and manage updates using GitOps workflow.

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
> Note: YAML files used in this tutorial are stored in [docs/examples/mariadb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mariadb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

## Create MariaDB Database using GitOps

### Create a MariaDB GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "11.8.5"
  replicas: 3
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
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
    └── MariaDB.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is created in your cluster.

Our `gitops` operator will create an actual `MariaDB` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get MariaDB.gitops.kubedb.com,MariaDB.kubedb.com -n demo
NAME                                       AGE
mariadb.gitops.kubedb.com/mariadb-gitops   22m

NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   11.8.5    Ready    22m
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` MariaDB.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=mariadb-gitops'
NAME                                          AGE
petset.apps.k8s.appscode.com/mariadb-gitops   22m

NAME                   READY   STATUS    RESTARTS   AGE
pod/mariadb-gitops-0   2/2     Running   0          22m
pod/mariadb-gitops-1   2/2     Running   0          22m
pod/mariadb-gitops-2   2/2     Running   0          22m

NAME                         TYPE                       DATA   AGE
secret/mariadb-gitops-auth   kubernetes.io/basic-auth   2      22m

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/mariadb-gitops        ClusterIP   10.43.215.232   <none>        3306/TCP   22m
service/mariadb-gitops-pods   ClusterIP   None            <none>        3306/TCP   22m

NAME                                                TYPE                 VERSION   AGE
appbinding.appcatalog.appscode.com/mariadb-gitops   kubedb.com/mariadb   11.8.5    22m
```

## Update MariaDB Database using GitOps

### Scale MariaDB Database Resources

Before scaling database resouces:

```shell
$ kubectl get pod -n demo mariadb-gitops-0 -o json | jq '.spec.containers[0].resources'
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
Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "11.8.5"
  replicas: 3
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
 ```

Resource Requests and Limits are updated to `600m` CPU and `1.2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `MariaDBOpsRequest` to update the `MariaDB` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   11.8.5    Ready    39m

NAME                                                                     TYPE              STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268   VerticalScaling   Successful   9m17s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo mariadb-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1288490188800m"
  }
}
```

### Scale MariaDB Replicas
Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "11.8.5"
  replicas: 5
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Update the `replicas` to `5`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` MariaDBOpsRequest to update the `MariaDB` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   11.8.5    Ready    107m

NAME                                                                       TYPE                STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-horizontalscaling-m7iex7   HorizontalScaling   Successful   63m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268     VerticalScaling     Successful   76m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=mariadb-gitops'
NAME               READY   STATUS    RESTARTS   AGE
mariadb-gitops-0   2/2     Running   0          76m
mariadb-gitops-1   2/2     Running   0          75m
mariadb-gitops-2   2/2     Running   0          73m
mariadb-gitops-3   2/2     Running   0          63m
mariadb-gitops-4   2/2     Running   0          62m
```

We can also scale down the replicas by updating the `replicas` fields.

### Expand MariaDB Volume

Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "11.8.5"
  replicas: 5
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` MariaDBOpsRequest to update the `MariaDB` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   11.8.5    Ready    111m

NAME                                                                       TYPE                STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-horizontalscaling-m7iex7   HorizontalScaling   Successful   68m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268     VerticalScaling     Successful   81m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-volumeexpansion-01m39b     VolumeExpansion     Successful   115s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=mariadb-gitops'
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-mariadb-gitops-0   Bound    pvc-f1a49c53-d095-43a7-b62f-e91a5a9f5496   2Gi        RWO            standard       <unset>                 116m
data-mariadb-gitops-1   Bound    pvc-79d99636-e0d7-413f-a6a6-559e908ed817   2Gi        RWO            standard       <unset>                 116m
data-mariadb-gitops-2   Bound    pvc-f66ccf56-63b6-4ac6-a7f9-09f573466266   2Gi        RWO            standard       <unset>                 116m
data-mariadb-gitops-3   Bound    pvc-b7744652-611a-4c10-a06f-c93f030fb90f   2Gi        RWO            standard       <unset>                 72m
data-mariadb-gitops-4   Bound    pvc-dee1c10d-0456-4935-9306-0d86c3db54d0   2Gi        RWO            standard       <unset>                 71m
```

## Reconfigure MariaDB

At first, we will create a secret containing `user.conf` file with required configuration settings.
To know more about this configuration file, check [here](/docs/guides/mariadb/reconfigure/overview/index.md)
```yaml
apiVersion: v1
stringData:
  user.cnf: |
    [mysqld]
    max_connections = 200
    read_buffer_size = 1048576
kind: Secret
metadata:
  name: md-configuration
  namespace: demo
type: Opaque
```

Now, we will add this file to `kubedb /md-configuration.yaml`.

```bash
$ tree .
├── kubedb
│ ├── md-configuration.yaml
│ └── MariaDB.yaml
1 directories, 2 files
```

Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "11.8.5"
  replicas: 5
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: md-configuration
```

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` MariaDBOpsRequest to update the `MariaDB` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   11.8.5    Ready    17h

NAME                                                                       TYPE                STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-horizontalscaling-m7iex7   HorizontalScaling   Successful   19h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfigure-1leaj8         Reconfigure         Successful   17h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268     VerticalScaling     Successful   19h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-volumeexpansion-01m39b     VolumeExpansion     Successful   18h
```


### Rotate MariaDB Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will do that using gitops, create the file `kubedb /md-auth.yaml` with the following content,

```bash
$ kubectl create secret generic mdauth -n demo \
                                  --type=kubernetes.io/basic-auth \
                                  --from-literal=username=root \
                                  --from-literal=password=md-secret
secret/mdauth created
```



Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "11.8.5"
  replicas: 5
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: md-configuration
  authSecret:
    kind: Secret
    name: mdauth
```

Change the `authSecret` field to `mdauth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` MariaDBOpsRequest to update the `MariaDB` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   11.8.5    Ready    18h

NAME                                                                       TYPE                STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-horizontalscaling-m7iex7   HorizontalScaling   Successful   19h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfigure-1leaj8         Reconfigure         Successful   18h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-rotate-auth-1xy3d7         RotateAuth          Successful   12m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268     VerticalScaling     Successful   19h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-volumeexpansion-01m39b     VolumeExpansion     Successful   18h
```


### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mariadb/O=kubedb"
Generating a RSA private key
...........................................................................+++++
........................................................................................................+++++
writing new private key to './ca.key'
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls md-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/md-ca created
```

Now, we are going to create an `Issuer` using the `md-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: md-issuer
  namespace: demo
spec:
  ca:
    secretName: md-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}//docs/guides/mariadb/reconfigure-tls/cluster/examples/issuer.yaml
issuer.cert-manager.io/md-issuer created
```

Let's add that to our `kubedb /md-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── md-configuration.yaml
│ ├── md-issuer.yaml
│ └── MariaDB.yaml
1 directories, 3 files
```

Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "11.8.5"
  replicas: 5
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: md-configuration
  authSecret:
    kind: Secret
    name: mdauth
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: md-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
```

Add `requireSSL` as `true` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` MariaDBOpsRequest to update the `MariaDB` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   11.8.5    Ready    18h

NAME                                                                       TYPE                STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-horizontalscaling-m7iex7   HorizontalScaling   Successful   19h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfigure-1leaj8         Reconfigure         Successful   18h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfiguretls-ftyfdq      ReconfigureTLS      Successful   8m20s
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-rotate-auth-1xy3d7         RotateAuth          Successful   35m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268     VerticalScaling     Successful   20h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-volumeexpansion-01m39b     VolumeExpansion     Successful   18h
```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for MariaDB.

### Update Version

List MariaDB versions using `kubectl get MariaDBversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/mariadb/update-version/overview/index.md).

Let's choose `12.1.2` in this example.

Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "12.1.2"
  replicas: 5
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: md-configuration
  authSecret:
    kind: Secret
    name: mdauth
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: md-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
```

Update the `version` field to `12.1.2`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` MariaDBOpsRequest to update the `MariaDB` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kkubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   12.1.2    Ready    18h

NAME                                                                       TYPE                STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-horizontalscaling-m7iex7   HorizontalScaling   Successful   20h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfigure-1leaj8         Reconfigure         Successful   18h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfiguretls-ftyfdq      ReconfigureTLS      Successful   24m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-rotate-auth-1xy3d7         RotateAuth          Successful   51m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-versionupdate-ksli15       UpdateVersion       Successful   10m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268     VerticalScaling     Successful   20h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-volumeexpansion-01m39b     VolumeExpansion     Successful   19h
```


Now, we are going to verify whether the `MariaDB`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get MariaDB -n demo mariadb-gitops -o=jsonpath='{.spec.version}{"\n"}'
12.1.2
$ kubectl get petset -n demo mariadb-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/mariadb:12.1.2-noble@sha256:843852d8651b3f321896a4a91f8118605d988d70703e520927c8d2c9313aded4
$ kubectl get pod -n demo mariadb-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/mariadb:12.1.2-noble@sha256:843852d8651b3f321896a4a91f8118605d988d70703e520927c8d2c9313aded4
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `MariaDB.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MariaDB
metadata:
  name: mariadb-gitops
  namespace: demo
spec:
  version: "12.1.2"
  replicas: 5
  podTemplate:
    spec:
      containers:
      - name: mariadb
        resources:
          requests:
            memory: "1.2Gi"
            cpu: "0.6"
          limits:
            memory: "1.2Gi"
            cpu: "0.6"
  storageType: Durable
  deletionPolicy: WipeOut
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  configuration:
    secretName: md-configuration
  authSecret:
    kind: Secret
    name: mdauth
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: md-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MariaDB` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` MariaDBOpsRequest to add the `MariaDB` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get md,mariadbopsrequest -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb-gitops   12.1.2    Ready    19h

NAME                                                                       TYPE                STATUS       AGE
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-horizontalscaling-m7iex7   HorizontalScaling   Successful   20h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfigure-1leaj8         Reconfigure         Successful   19h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-reconfiguretls-ftyfdq      ReconfigureTLS      Successful   40m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-restart-e7rv89             Restart             Successful   8m26s
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-rotate-auth-1xy3d7         RotateAuth          Successful   67m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-versionupdate-ksli15       UpdateVersion       Successful   26m
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-verticalscaling-rjs268     VerticalScaling     Successful   20h
mariadbopsrequest.ops.kubedb.com/mariadb-gitops-volumeexpansion-01m39b     VolumeExpansion     Successful   19h
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

- Learn MariaDB Scaling
    - [Horizontal Scaling](/docs/guides/mariadb/scaling/horizontal-scaling/overview/index.md)
    - [Vertical Scaling](/docs/guides/mariadb/scaling/vertical-scaling/overview/index.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/mariadb/update-version/overview/index.md)
- Monitor your MariaDBQL database with KubeDB using [built-in Prometheus](/docs/guides/mariadb/monitoring/prometheus-operator/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
