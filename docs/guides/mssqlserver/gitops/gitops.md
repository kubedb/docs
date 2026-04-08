---
title: Mssqlserver Gitops
description: Mssqlserver Gitops
menu:
  docs_{{ .version }}:
    identifier: mssql-gitops
    name: MSSQL Gitops
    parent: mssqlserver-gitops
    weight: 10
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
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace> --ssh-private-key-path ~/.ssh/id_rsa
```

## Create Mssqlserver Database using GitOps

### Create a Mssqlserver GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MSSQLServer
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: "2022-cu12"
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
              value: Evaluation # Change it 
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
    └── Mssqlserver.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is created in your cluster.

Our `gitops` operator will create an actual `Mssqlserver` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get Mssqlserver.gitops.kubedb.com,Mssqlserver.kubedb.com -n demo
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
Update the `Mssqlserver.yaml` with the following,
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
              value: Evaluation # Change it 
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

Update the `replicas` to `5`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` mssqlserverOpsRequest to update the `Mssqlserver` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,Mssqlserver,msops -n demo
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
Update the `Mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Mssqlserver
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Mssqlserver
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: local-path
    controller:
      podTemplate:
        spec:
          containers:
            - name: Mssqlserver
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: local-path
  storageType: Durable
  deletionPolicy: WipeOut
 ```

Resource Requests and Limits are updated to `700m` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `mssqlserverOpsRequest` to update the `Mssqlserver` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,Mssqlserver,msops -n demo
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

Update the `Mssqlserver.yaml` with the following,
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
              value: Evaluation # Change it 
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
$ kubectl get ms,Mssqlserver,msops -n demo
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

At first, we will create a secret containing `user.conf` file with required configuration settings.
To know more about this configuration file, check [here](/docs/guides/Mssqlserver/configuration/Mssqlserver-combined.md)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-kf-combined-custom-config
  namespace: demo
stringData:
  server.properties: |-
    log.retention.hours=125    
```

Now, we will add this file to `kubedb /kf-configuration.yaml`.

```bash
$ tree .
├── kubedb
│ ├── kf-configuration.yaml
│ └── Mssqlserver.yaml
1 directories, 2 files
```

Update the `Mssqlserver.yaml` with the following,
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
              value: Evaluation # Change it 
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
$ kubectl get ms,Mssqlserver,msops -n demo
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



> We can also reconfigure the parameters creating another secret and reference the secret in the `configSecret` field. Also you can remove the `configSecret` field to use the default parameters.

### Rotate Mssqlserver Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will do that using gitops, create the file `kubedb /kf-auth.yaml` with the following content,

```bash
$ kubectl create secret generic mssqlserver-auth -n demo \
  --type=kubernetes.io/basic-auth \
  --from-literal=username=sa \
  --from-literal=password=Mssqlserver2
secret/mssqlserver-auth created
```



Update the `Mssqlserver.yaml` with the following,
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
              value: Evaluation # Change it 
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

Change the `authSecret` field to `kf-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` mssqlserverOpsRequest to update the `Mssqlserver` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,Mssqlserver,msops -n demo
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


### TLS configuration


Update the `Mssqlserver.yaml` with the following,
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
    clientTLS: true
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation # Change it 
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

Convert `spec.tls.clientTLS` to `true`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` mssqlserverOpsRequest to update the `Mssqlserver` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get kf,Mssqlserver,kfops,pods -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   3.9.0     Ready    41m

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   75m

NAME                                                               TYPE              STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfigure-ukj41o       Reconfigure       Successful   5d18h
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfiguretls-r4mx7v    ReconfigureTLS    Successful   9m18s
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-43ris8       RotateAuth        Successful   5d1h
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   5d19h

```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for Mssqlserver.

### Update Version

List Mssqlserver versions using `kubectl get Elasticsearchversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/Mssqlserver/update-version/update-version.md).

Let's choose `4.0.0` in this example.

Update the `Mssqlserver.yaml` with the following,
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
              value: Evaluation # Change it 
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

Update the `version` field to `17.4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` mssqlserverOpsRequest to update the `Mssqlserver` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get ms,Mssqlserver,msops -n demo
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
$ kubectl get Mssqlserver -n demo mssql-gitops -o=jsonpath='{.spec.version}{"\n"}'
2022-cu22

$ kubectl get petset -n demo mssql-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mcr.microsoft.com/mssql/server:2022-CU22-ubuntu-22.04@sha256:db9a8fe3098b7e8bbde41106bdc7caee942e97124e5fdb71b872ca208de3092d

$ kubectl get pod -n demo mssql-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mcr.microsoft.com/mssql/server:2022-CU22-ubuntu-22.04@sha256:db9a8fe3098b7e8bbde41106bdc7caee942e97124e5fdb71b872ca208de3092d
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `Mssqlserver.yaml` with the following,
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
              value: Evaluation # Change it 
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

[//]: # (- Learn Mssqlserver [GitOps]&#40;/docs/guides/Mssqlserver/concepts/Mssqlserver-gitops.md&#41;)
- Learn Mssqlserver Scaling
    - [Horizontal Scaling](/docs/guides/Mssqlserver/scaling/horizontal-scaling/combined.md)
    - [Vertical Scaling](/docs/guides/Mssqlserver/scaling/vertical-scaling/combined.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/Mssqlserver/update-version/overview.md)
- Monitor your ElasticsearchQL database with KubeDB using [built-in Prometheus](/docs/guides/Mssqlserver/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
