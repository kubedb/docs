---
title: Mssqlserver GitOps
description: Mssqlserver GitOps
menu:
  docs_{{ .version }}:
    identifier: mssql-gitops
    name: Guide
    parent: mssqlserver-gitops
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---



> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Mssqlserver using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create Mssqlserver database and manage updates using GitOps workflow.

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
> Note: YAML files used in this tutorial are stored in [docs/examples/Mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

## Create Mssqlserver Database using GitOps
First, an issuer needs to be created, even if TLS is not enabled for SQL Server. The issuer will be used to configure the TLS-enabled Wal-G proxy server, which is required for the SQL Server backup and restore operations.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=MSSQLServer/O=kubedb"
```
- Create a secret using the certificate files we have just generated,
```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```
Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```
Create a directory like below,
```bash
$ tree .
├── kubedb
    └── ms-issuer.yaml
1 directories, 1 files

Now, we are going to deploy a `MSSQLServer` availability group with version `2022-cu12`.


### Create a Mssqlserver GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 4
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Create a directory like below,

```bash
$ tree .
├── kubedb
│ ├──ms-issuer.yaml
│ ├──mssql.yaml
1 directories, 2 files
```
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is created in your cluster.

Our `gitops` operator will create an actual `Mssqlserver` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get mssqlserver.gitops.kubedb.com,mssqlserver.kubedb.com -n demo
NAME                                         AGE
mssqlserver.gitops.kubedb.com/mssql-gitops   19h

NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu19   Ready    19h
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Mssqlserver.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=mssql-gitops'
NAME                                                AGE
petset.apps.k8s.appscode.com/mssql-gitops           19h
petset.apps.k8s.appscode.com/mssql-gitops-arbiter   19h

NAME                         READY   STATUS    RESTARTS   AGE
pod/mssql-gitops-0           2/2     Running   0          19h
pod/mssql-gitops-1           2/2     Running   0          19h
pod/mssql-gitops-2           2/2     Running   0          19h
pod/mssql-gitops-3           2/2     Running   0          19h
pod/mssql-gitops-arbiter-0   1/1     Running   0          19h

NAME                                TYPE                       DATA   AGE
secret/mssql-gitops-218872          Opaque                     1      19h
secret/mssql-gitops-auth            kubernetes.io/basic-auth   2      19h
secret/mssql-gitops-client-cert     kubernetes.io/tls          4      19h
secret/mssql-gitops-dbm-login       kubernetes.io/basic-auth   1      19h
secret/mssql-gitops-endpoint-cert   kubernetes.io/tls          3      19h
secret/mssql-gitops-master-key      kubernetes.io/basic-auth   1      19h
secret/mssql-gitops-server-cert     kubernetes.io/tls          3      19h

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/mssql-gitops             ClusterIP   10.43.242.226   <none>        1433/TCP            19h
service/mssql-gitops-pods        ClusterIP   None            <none>        1433/TCP,5022/TCP   19h
service/mssql-gitops-secondary   ClusterIP   10.43.53.184    <none>        1433/TCP            19h

NAME                                              TYPE                     VERSION     AGE
appbinding.appcatalog.appscode.com/mssql-gitops   kubedb.com/mssqlserver   2022-cu19   19h
```

## Update Mssqlserver Database using GitOps

### Scale Mssqlserver Replicas
Update the `mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu19"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Scale down the `replicas` to `3`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` mssqlserverOpsRequest to update the `Mssqlserver` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,mssqlserver,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu19   Ready    19h

NAME                                         AGE
mssqlserver.gitops.kubedb.com/mssql-gitops   19h

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   15m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=mssql-gitops'
NAME             READY   STATUS    RESTARTS   AGE
mssql-gitops-0   2/2     Running   0          19h
mssql-gitops-1   2/2     Running   0          19h
mssql-gitops-2   2/2     Running   0          19h
```

We can also scale down the replicas by updating the `replicas` fields.

### Scale Mssqlserver Database Resources

Before the Ops Request reaches the `Successful` state, the configured memory limits are as follows:

```bash
$ kubectl get pod -n demo mssql-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "1500m",
    "memory": "2Gi"
  }
}
```
Update the `mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu19"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
          resources:
            requests:
              memory: "2Gi"
              cpu: "700m"
            limits:
              cpu: 2
              memory: "2Gi"
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
 ```

Resource Requests and Limits are updated to `2` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `mssqlserverOpsRequest` to update the `Mssqlserver` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,mssqlserver,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu19   Ready    19h

NAME                                         AGE
mssqlserver.gitops.kubedb.com/mssql-gitops   19h

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   25m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-yi3db5     VerticalScaling     Successful   5m2s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo mssql-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "2",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "2Gi"
  }
}

```


### Expand Mssqlserver Volume

Update the `mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu19"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
          resources:
            requests:
              memory: "2Gi"
              cpu: "700m"
            limits:
              cpu: 2
              memory: "2Gi"
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` mssqlserverOpsRequest to update the `Mssqlserver` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,mssqlserver,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu19   Ready    20h

NAME                                         AGE
mssqlserver.gitops.kubedb.com/mssql-gitops   20h

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   51m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-yi3db5     VerticalScaling     Successful   30m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-rsa80j     VolumeExpansion     Successful   15m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=mssql-gitops'
NAME                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-mssql-gitops-0   Bound    pvc-481caac3-7849-422c-9d8f-704ef82b3bc6   2Gi        RWO            longhorn       <unset>                 20h
data-mssql-gitops-1   Bound    pvc-49ea234d-e7e2-4a3b-9372-b430d72fee5e   2Gi        RWO            longhorn       <unset>                 19h
data-mssql-gitops-2   Bound    pvc-c72d4562-81d2-405b-ae8d-52816e59767f   2Gi        RWO            longhorn       <unset>                 19h
```

## Reconfigure Mssqlserver

At first, we create a configuration file 
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ms-custom-config
  namespace: demo
type: Opaque
stringData:
  mssql.conf: |
    [memory]
    memorylimitmb = 2048
```


Now, we will add this file to `kubedb/ms-configuration.yaml`.

```bash
$ tree .
├── kubedb
│ ├──ms-issuer.yaml
│ ├──ms-configuration.yaml
│ └──mssql.yaml
1 directories, 3 files
```

Update the `mssqlserver.yaml` with `spec.configuration.secretName` as the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu19"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
          resources:
            requests:
              memory: "2Gi"
              cpu: "700m"
            limits:
              cpu: 2
              memory: "2Gi"
  configuration:
    secretName: ms-custom-config
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` mssqlserverOpsRequest to update the `Mssqlserver` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,mssqlserver,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu19   Ready    20h

NAME                                         AGE
mssqlserver.gitops.kubedb.com/mssql-gitops   20h

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   71m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-reconfigure-6i2hvt         Reconfigure         Successful   6m24s
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-yi3db5     VerticalScaling     Successful   50m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-rsa80j     VolumeExpansion     Successful   35m
```



> We can also reconfigure the parameters creating another secret and reference the secret in the `configuration.secretName` field. Also you can remove the `configuration` field to use the default parameters.

### Rotate Mssqlserver Auth

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,
> Note: The `username` must be fixed as `sa`. The `password` must include uppercase letters, lowercase letters, and numbers
We will do that using gitops, create the file `kubedb/ms-auth.yaml` with the following content,

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mssqlserver-quickstart-auth-user
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: sa
  password: Mssqlserver2
```
Let's add that to our `kubedb/md-auth.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├──ms-issuer.yaml
│ ├── ms-configuration.yaml
│ ├── ms-auth.yaml
│ └── mssql.yaml
1 directories, 4 files
```

Update the `mssql.yaml` ading `authsecret` as the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu19"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
          resources:
            requests:
              memory: "2Gi"
              cpu: "700m"
            limits:
              cpu: 2
              memory: "2Gi"
  configuration:
    secretName: ms-custom-config
  authSecret:
    kind: Secret
    name: mssqlserver-auth
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Add the `authSecret.kind` and `authSecret.name` field to `mssqlserver-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster and the authentication file is created in you cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` mssqlserverOpsRequest to update the `Mssqlserver` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,mssqlserver,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu19   Ready    21h

NAME                                         AGE
mssqlserver.gitops.kubedb.com/mssql-gitops   21h

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   153m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-reconfigure-6i2hvt         Reconfigure         Successful   88m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-otytes         RotateAuth          Successful   77m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-yi3db5     VerticalScaling     Successful   133m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-rsa80j     VolumeExpansion     Successful   117m
```


### Update Version

List Mssqlserver versions using `kubectl get msversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/mssqlserver/update-version/overview.md).

Let's choose `2022-cu22` in this example.

Update the `mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu22"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
          resources:
            requests:
              memory: "2Gi"
              cpu: "700m"
            limits:
              cpu: 2
              memory: "2Gi"
  configuration:
    secretName: ms-custom-config
  authSecret:
    kind: Secret
    name: mssqlserver-auth
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Update the `version` field to `2022-cu22`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` mssqlserverOpsRequest to update the `Mssqlserver` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,mssqlserver,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu22   Ready    22h

NAME                                         AGE
mssqlserver.gitops.kubedb.com/mssql-gitops   22h

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   3h21m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-reconfigure-6i2hvt         Reconfigure         Successful   136m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-otytes         RotateAuth          Successful   125m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-versionupdate-mlq0kn       UpdateVersion       Successful   13m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-yi3db5     VerticalScaling     Successful   3h1m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-rsa80j     VolumeExpansion     Successful   165m
```


Now, we are going to verify whether the `Mssqlserver`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get mssqlserver -n demo mssql-gitops -o=jsonpath='{.spec.version}{"\n"}'
2022-cu22

$ kubectl get petset -n demo mssql-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mcr.microsoft.com/mssql/server:2022-CU22-ubuntu-22.04@sha256:db9a8fe3098b7e8bbde41106bdc7caee942e97124e5fdb71b872ca208de3092d

$ kubectl get pod -n demo mssql-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mcr.microsoft.com/mssql/server:2022-CU22-ubuntu-22.04@sha256:db9a8fe3098b7e8bbde41106bdc7caee942e97124e5fdb71b872ca208de3092d
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu22"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
          resources:
            requests:
              memory: "2Gi"
              cpu: "700m"
            limits:
              cpu: 2
              memory: "2Gi"
  configuration:
    secretName: ms-custom-config
  authSecret:
    kind: Secret
    name: mssqlserver-auth
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  monitor:
   agent: prometheus.io/operator
   prometheus:
     exporter:
       port: 9399
       securityContext:
         allowPrivilegeEscalation: false
         capabilities:
           drop:
             - ALL
         runAsGroup: 10001
         runAsNonRoot: true
         runAsUser: 10001
         seccompProfile:
           type: RuntimeDefault
     serviceMonitor:
       interval: 100s
       labels:
         release: prometheus
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` mssqlserverOpsRequest to add the `Mssqlserver` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get ms,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu22   Ready    11m

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   5h9m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-reconfigure-6i2hvt         Reconfigure         Successful   4h4m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-restart-hoi536             Restart             Successful   6m8s
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-otytes         RotateAuth          Successful   3h53m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-versionupdate-mlq0kn       UpdateVersion       Successful   120m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-yi3db5     VerticalScaling     Successful   4h48m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-rsa80j     VolumeExpansion     Successful   4h33m
```

Verify the monitoring is enabled by checking the prometheus targets.


### TLS configuration

First install [csi-driver-cacerts](https://github.com/kubeops/csi-driver-cacerts) which will be used to add self-signed ca certificates to the OS trusted certificate store (eg, /etc/ssl/certs/ca-certificates.crt)
Update the `mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu22"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: true
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation 
          resources:
            requests:
              memory: "2Gi"
              cpu: "700m"
            limits:
              cpu: 2
              memory: "2Gi"
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
  monitor:
   agent: prometheus.io/operator
   prometheus:
     exporter:
       port: 9399
       securityContext:
         allowPrivilegeEscalation: false
         capabilities:
           drop:
             - ALL
         runAsGroup: 10001
         runAsNonRoot: true
         runAsUser: 10001
         seccompProfile:
           type: RuntimeDefault
     serviceMonitor:
       interval: 100s
       labels:
         release: prometheus
```

Convert `spec.tls.clientTLS` to `true`. Commit the changes and push to your Git repository. Your repository has been successfully synchronized with ArgoCD. The `MSSQLserver` CR has been updated in the cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` mssqlserverOpsRequest to update the `Mssqlserver` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,msops -n demo
NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu22   Ready    20m

NAME                                                                         TYPE                STATUS       AGE
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-28njbi   HorizontalScaling   Successful   23h
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-reconfigure-6i2hvt         Reconfigure         Successful   22h
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-reconfiguretls-hq78lx      ReconfigureTLS      Successful   12m
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-restart-hoi536             Restart             Successful   18h
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-otytes         RotateAuth          Successful   22h
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-versionupdate-mlq0kn       UpdateVersion       Successful   20h
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-yi3db5     VerticalScaling     Successful   23h
mssqlserveropsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-rsa80j     VolumeExpansion     Successful   23h
```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can change the value of the `.spec.tls.clientTLS` field  for Mssqlserver.

## Next Steps

- Learn Mssqlserver Scaling
    - [Horizontal Scaling](/docs/guides/mssqlserver/scaling/horizontal-scaling/overview.md)
    - [Vertical Scaling](/docs/guides/mssqlserver/scaling/vertical-scaling/overview.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/mssqlserver/update-version/overview.md)
- Monitor your MSSQLServer database with KubeDB using [built-in Prometheus](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
