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
mssqlserver.gitops.kubedb.com/mssql-gitops   16m

NAME                                  VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssql-gitops   2022-cu12   Ready    16m
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Mssqlserver.

```bash
$  kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=mssql-gitops'
NAME                                        AGE
petset.apps.k8s.appscode.com/mssql-gitops   16m

NAME                 READY   STATUS    RESTARTS   AGE
pod/mssql-gitops-0   2/2     Running   0          16m
pod/mssql-gitops-1   2/2     Running   0          14m
pod/mssql-gitops-2   2/2     Running   0          14m

NAME                                TYPE                       DATA   AGE
secret/mssql-gitops-449c85          Opaque                     1      17m
secret/mssql-gitops-auth            kubernetes.io/basic-auth   2      17m
secret/mssql-gitops-client-cert     kubernetes.io/tls          4      17m
secret/mssql-gitops-dbm-login       kubernetes.io/basic-auth   1      17m
secret/mssql-gitops-endpoint-cert   kubernetes.io/tls          3      17m
secret/mssql-gitops-master-key      kubernetes.io/basic-auth   1      17m
secret/mssql-gitops-server-cert     kubernetes.io/tls          3      17m

NAME                             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
service/mssql-gitops             ClusterIP   10.43.7.188    <none>        1433/TCP            17m
service/mssql-gitops-pods        ClusterIP   None           <none>        1433/TCP,5022/TCP   17m
service/mssql-gitops-secondary   ClusterIP   10.43.186.58   <none>        1433/TCP            17m

NAME                                              TYPE                     VERSION     AGE
appbinding.appcatalog.appscode.com/mssql-gitops   kubedb.com/mssqlserver   2022-cu12   16m
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
  version: "2022-cu12"
  replicas: 5
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

Update the `replicas` to `5`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` ElasticsearchOpsRequest to update the `Mssqlserver` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,Mssqlserver,kfops -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   3.9.0     Ready    22h

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   22h

NAME                                                                 TYPE                STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-j0wni6   HorizontalScaling   Successful   13m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-tfkvi8     VerticalScaling     Successful   8m29s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$  kubectl get pod -n demo -l 'app.kubernetes.io/instance=mssql-gitops'
NAME                      READY   STATUS    RESTARTS   AGE
mssql-gitops-broker-0       1/1     Running   0          34m
mssql-gitops-broker-1       1/1     Running   0          33m
mssql-gitops-broker-2       1/1     Running   0          33m
mssql-gitops-controller-0   1/1     Running   0          32m
mssql-gitops-controller-1   1/1     Running   0          31m
mssql-gitops-controller-2   1/1     Running   0          31m
```

We can also scale down the replicas by updating the `replicas` fields.

### Scale Mssqlserver Database Resources

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

Now, `gitops` operator will detect the resource changes and create a `ElasticsearchOpsRequest` to update the `Mssqlserver` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,Mssqlserver,kfops -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   3.9.0     Ready    22h

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   22h

NAME                                                                   TYPE              STATUS        AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-i0kr1l   VerticalScaling       Successful     2s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo mssql-gitops-broker-0 -o json | jq '.spec.containers[0].resources'
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


### Expand Mssqlserver Volume

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
            - name: Mssqlserver
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

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` ElasticsearchOpsRequest to update the `Mssqlserver` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,Mssqlserver,kfops -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   3.9.0     Ready    23m

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   23m

NAME                                                                 TYPE                STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-horizontalscaling-j0wni6   HorizontalScaling   Successful   13m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-verticalscaling-tfkvi8     VerticalScaling     Successful   8m29s
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-41xthr     VolumeExpansion     Successful   19m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=mssql-gitops'
NAME                                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
mssql-gitops-data-mssql-gitops-broker-0       Bound    pvc-2afd4835-5686-492b-be93-c6e040e0a6c6   2Gi        RWO            Standard       <unset>                 3h39m
mssql-gitops-data-mssql-gitops-broker-1       Bound    pvc-aaf994cc-6b04-4c37-80d5-5e966dad8487   2Gi        RWO            Standard       <unset>                 3h39m
mssql-gitops-data-mssql-gitops-controller-0   Bound    pvc-82d2b233-203d-4df2-a0fd-ecedbc0825b7   2Gi        RWO            Standard       <unset>                 3h39m
mssql-gitops-data-mssql-gitops-controller-1   Bound    pvc-91852c29-ab1a-48ad-9255-a0b15d5a7515   2Gi        RWO            Standard       <unset>                 3h39m

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
kind: Mssqlserver
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  configSecret:
    name: new-kf-combined-custom-config
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Mssqlserver
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
            - name: Mssqlserver
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

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` ElasticsearchOpsRequest to update the `Mssqlserver` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,Mssqlserver,kfops -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   3.9.0     Ready    74m

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   74m

NAME                                                               TYPE              STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfigure-ukj41o       Reconfigure       Successful   24m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   70m

```



> We can also reconfigure the parameters creating another secret and reference the secret in the `configSecret` field. Also you can remove the `configSecret` field to use the default parameters.

### Rotate Mssqlserver Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will do that using gitops, create the file `kubedb /kf-auth.yaml` with the following content,

```bash
kubectl create secret generic kf-rotate-auth -n demo \
--type=kubernetes.io/basic-auth \
--from-literal=username=Mssqlserver \
--from-literal=password=Mssqlserver-secret
secret/kf-rotate-auth created

```



Update the `Mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Mssqlserver
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  authSecret:
    kind: Secret
    name: kf-rotate-auth
  configSecret:
    name: new-kf-combined-custom-config
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Mssqlserver
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
            - name: Mssqlserver
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

Change the `authSecret` field to `kf-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` ElasticsearchOpsRequest to update the `Mssqlserver` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get kf,Mssqlserver,kfops -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   3.9.0     Ready    7m11s

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   7m11s

NAME                                                               TYPE              STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfigure-ukj41o       Reconfigure       Successful   17h
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-43ris8       RotateAuth        Successful   28m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   17h

```


### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

To add tls, we are going to create an example `Issuer` that will be used to enable SSL/TLS in Mssqlserver. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. As well as we need to install [csi-driver-cacerts](https://github.com/kubeops/csi-driver-cacerts) which will be used to add self-signed ca certificates to the OS trusted certificate store (eg, /etc/ssl/certs/ca-certificates.crt)


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
$ kubectl create secret tls Mssqlserver-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/Mssqlserver-ca created
```

Now, Let's create an `Issuer` using the `Mssqlserver-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: kf-issuer
  namespace: demo
spec:
  ca:
    secretName: Mssqlserver-ca
```

Let's add that to our `kubedb /kf-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── kf-configuration.yaml
│ ├── kf-issuer.yaml
│ └── Mssqlserver.yaml
1 directories, 4 files
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
            - name: Mssqlserver
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

Add `sslMode` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` ElasticsearchOpsRequest to update the `Mssqlserver` database tls. List the resources created by `gitops` operator in the `demo` namespace.

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
kind: Mssqlserver
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: 4.0.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Mssqlserver
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
            - name: Mssqlserver
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

Update the `version` field to `17.4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` ElasticsearchOpsRequest to update the `Mssqlserver` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,Mssqlserver,kfops -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   4.0.0     Ready    3h47m

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   3h47m

NAME                                                               TYPE              STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfigure-ukj41o       Reconfigure       Successful   5d22h
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfiguretls-r4mx7v    ReconfigureTLS    Successful   4h16m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-43ris8       RotateAuth        Successful   5d6h
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-versionupdate-wyn2dp     UpdateVersion     Successful   3h51m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   5d23h
```


Now, we are going to verify whether the `Mssqlserver`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get Mssqlserver -n demo mssql-gitops -o=jsonpath='{.spec.version}{"\n"}'
4.0.0

$ kubectl get petset -n demo mssql-gitops-broker -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/Mssqlserver:4.0.0@sha256:42a79fe8f14b00b1c76d135bbbaf7605b8c66f45cf3eb749c59138f6df288b31

$  kubectl get pod -n demo mssql-gitops-broker-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/Mssqlserver:4.0.0@sha256:42a79fe8f14b00b1c76d135bbbaf7605b8c66f45cf3eb749c59138f6df288b31
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `Mssqlserver.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Mssqlserver
metadata:
  name: mssql-gitops
  namespace: demo
spec:
  version: 4.0.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: Mssqlserver
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
            - name: Mssqlserver
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

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Mssqlserver` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` ElasticsearchOpsRequest to add the `Mssqlserver` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get Elasticsearches.gitops.kubedb.com,Elasticsearches.kubedb.com,Elasticsearchopsrequest -n demo
NAME                          TYPE            VERSION   STATUS   AGE
Mssqlserver.kubedb.com/mssql-gitops   kubedb.com/v1   4.0.0     Ready    5h12m

NAME                                 AGE
Mssqlserver.gitops.kubedb.com/mssql-gitops   5h12m

NAME                                                               TYPE              STATUS       AGE
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfigure-ukj41o       Reconfigure       Successful   6d
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-reconfiguretls-r4mx7v    ReconfigureTLS    Successful   5h42m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-restart-ljpqih           Restart           Successful   3m51s
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-rotate-auth-43ris8       RotateAuth        Successful   5d7h
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-versionupdate-wyn2dp     UpdateVersion     Successful   5h16m
Elasticsearchopsrequest.ops.kubedb.com/mssql-gitops-volumeexpansion-41xthr   VolumeExpansion   Successful   6d

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
