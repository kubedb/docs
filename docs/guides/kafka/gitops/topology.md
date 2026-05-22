---
title: Topology GitOps
menu:
  docs_{{ .version }}:
    identifier: kf-gitops-topology
    name: Guides
    parent: kf-gitops
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Kafka using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create Kafka database and manage updates using GitOps workflow.

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
> Note: YAML files used in this tutorial are stored in [docs/examples/kafka](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

## Create Kafka Database using GitOps

### Create a Kafka GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: Standard
    controller:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: Standard
  storageType: Durable
  deletionPolicy: WipeOut
```

Create a directory like below,
```bash
$ tree .
├── kubedb
    └── Kafka.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Kafka` CR is created in your cluster.

Our `gitops` operator will create an actual `Kafka` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get kafka.gitops.kubedb.com,kafka.kubedb.com -n demo
NAME                                        AGE
kafka.gitops.kubedb.com/kafka-gitops        62m

NAME                                 VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops         3.9.0     Ready    62m
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Kafka.

```bash
$  kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=kafka-gitops'
NAME                                                        AGE
petset.apps.k8s.appscode.com/kafka-gitops-broker            62m
petset.apps.k8s.appscode.com/kafka-gitops-controller        62m

NAME                                 READY   STATUS    RESTARTS   AGE
pod/kafka-gitops-broker-0            1/1     Running   0          7m42s
pod/kafka-gitops-broker-1            1/1     Running   0          6m52s
pod/kafka-gitops-controller-0        1/1     Running   0          6m3s
pod/kafka-gitops-controller-1        1/1     Running   0          5m13s

NAME                            TYPE                       DATA   AGE
secret/kafka-gitops-auth   kubernetes.io/basic-auth        2      62m

NAME                             TYPE        CLUSTER-IP   EXTERNAL-IP     PORT(S)                       AGE
service/kafka-gitops-pods       ClusterIP     None         <none>        9092/TCP,9093/TCP,29092/TCP   62m

NAME                                                   TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-gitops   kubedb.com/kafka        3.9.0     62m
```

## Update Kafka Database using GitOps

### Scale Kafka Database Resources

Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
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
        storageClassName: longhorn
    controller:
      podTemplate:
        spec:
          containers:
            - name: kafka
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
        storageClassName: longhorn
  storageType: Durable
  deletionPolicy: WipeOut
 ```

The resource requests and limits for the topology broker have been updated to `1536Mi` of memory, and the controller’s memory allocation has also been increased accordingly. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Kafka` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `KafkaOpsRequest` to update the `Kafka` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,kafka,kfops -n demo
NAME                                 VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops        3.9.0     Ready    64m

NAME                                        AGE
kafka.gitops.kubedb.com/kafka-gitops        64m

NAME                                                                      TYPE              STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-c2ejz2   VerticalScaling        Successful   10m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo Kafka-gitops-broker-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1540Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
$ kubectl get pod -n demo Kafka-gitops-controller-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1540Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
```

### Scale Kafka Replicas

Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: longhorn
    controller:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: longhorn
  storageType: Durable
  deletionPolicy: WipeOut
```

Update the replicas count for both the broker and controller to 3. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Kafka` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` KafkaOpsRequest to update the `Kafka` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,kafka,kfops -n demo
NAME                            TYPE            VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops   kubedb.com/v1   3.9.0     Ready    22h

NAME                                   AGE
kafka.gitops.kubedb.com/kafka-gitops   22h

NAME                                                                    TYPE                STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-horizontalscaling-j0wni6   HorizontalScaling   Successful   13m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-tfkvi8     VerticalScaling     Successful   8m29s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$  kubectl get pod -n demo -l 'app.kubernetes.io/instance=kafka-gitops'
NAME                      READY     STATUS    RESTARTS   AGE
kafka-gitops-broker-0       1/1     Running   0          34m
kafka-gitops-broker-1       1/1     Running   0          33m
kafka-gitops-broker-2       1/1     Running   0          33m
kafka-gitops-controller-0   1/1     Running   0          32m
kafka-gitops-controller-1   1/1     Running   0          31m
kafka-gitops-controller-2   1/1     Running   0          31m
```

We can also scale down the replicas by updating the `replicas` fields.

### Expand Kafka Volume

Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
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
            - name: kafka
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
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

Set the `storage.resources.requests.storage` for both the broker and controller to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Kafka` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` KafkaOpsRequest to update the `Kafka` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,kafka,kfops -n demo
NAME                            VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops   3.9.0     Ready    6m51s

NAME                                   AGE
kafka.gitops.kubedb.com/kafka-gitops   6m51s

NAME                                                                   TYPE                STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-horizontalscaling-i7l7rn   HorizontalScaling   Successful   112m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-mwqdzx     VerticalScaling     Successful   117m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-volumeexpansion-7aweww     VolumeExpansion     Successful   4m30s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$  kubectl get pvc -n demo -l 'app.kubernetes.io/instance=kafka-gitops'
NAME                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
kafka-gitops-data-kafka-gitops-broker-0       Bound    pvc-a00ef6d5-44a6-40ce-8131-680b9d24d982   2Gi        RWO            longhorn       <unset>                 7m20s
kafka-gitops-data-kafka-gitops-broker-1       Bound    pvc-141798c5-9000-480f-a997-85a5743f63e2   2Gi        RWO            longhorn       <unset>                 7m4s
kafka-gitops-data-kafka-gitops-broker-2       Bound    pvc-819ba56b-4fda-4361-9f3e-e18258e2de7e   2Gi        RWO            longhorn       <unset>                 6m46s
kafka-gitops-data-kafka-gitops-controller-0   Bound    pvc-5a6cc06e-60ff-450a-875e-b84f75358f67   2Gi        RWO            longhorn       <unset>                 7m20s
kafka-gitops-data-kafka-gitops-controller-1   Bound    pvc-7cae3a1d-0efc-48a2-8953-12f4338a9602   2Gi        RWO            longhorn       <unset>                 7m4s
kafka-gitops-data-kafka-gitops-controller-2   Bound    pvc-22e0a63f-58c1-4a03-9493-e670498723db   2Gi        RWO            longhorn       <unset>                 6m46s
```

## Reconfigure Kafka

At first, we will create a secret containing `user.conf` file with required configuration settings.
To know more about this configuration file, check [here](/docs/guides/kafka/configuration/kafka-topology.md)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kf-reconfig
  namespace: demo
stringData:
  server.properties: |-
    log.retention.hours=125    
```

Now, we will add this file to `kubedb/kf-configuration.yaml`.

```bash
$ tree .
├── kubedb
│ ├── kf-configuration.yaml
│ └── Kafka.yaml
1 directories, 2 files
```

Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  configSecret:
    name: kf-reconfig
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
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
            - name: kafka
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
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

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD`, the reconfiguration file is created and  the `Kafka` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` KafkaOpsRequest to update the `Kafka` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,kafka,kfops -n demo
NAME                            VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops   3.9.0     Ready    19m

NAME                                   AGE
kafka.gitops.kubedb.com/kafka-gitops   19m

NAME                                                                   TYPE                STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-horizontalscaling-i7l7rn   HorizontalScaling   Successful   124m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfigure-fszhk8         Reconfigure         Successful   8m44s
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-mwqdzx     VerticalScaling     Successful   130m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-volumeexpansion-7aweww     VolumeExpansion     Successful   16m
```



> We can also reconfigure the parameters creating another secret and reference the secret in the `configSecret` field. Also you can remove the `configSecret` field to use the default parameters.

### Rotate Kafka Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will create a secret named `kf-rotate-auth ` with the following content,

```bash
$ apiVersion: v1
kind: Secret
metadata:
  name: kf-rotate-auth
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: kafka
  password: kafka-secret
```


Now, we will add this file to `kubedb/kf-rotateauth.yaml`.

```bash
$ tree .
├── kubedb
│ ├── kf-configuration.yaml
│ ├── kf-rotateauth.yaml
│ └── Kafka.yaml
1 directories, 3 files
```


Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  authSecret:
    kind: Secret
    name: kf-rotate-auth
  configuration:
    secretName: kf-reconfig
  version: 3.9.0
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
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
            - name: kafka
              resources:
                limits:
                  memory: 1536Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
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

Change the `authSecret` field to `kf-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD`, the authentication file is created and  the `Kafka` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` KafkaOpsRequest to update the `Kafka` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,kafka,kfops -n demo
NAME                            VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops   3.9.0     Ready    17h

NAME                                   AGE
kafka.gitops.kubedb.com/kafka-gitops   17h

NAME                                                                   TYPE                STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-horizontalscaling-i7l7rn   HorizontalScaling   Successful   18h
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfigure-fszhk8         Reconfigure         Successful   16h
kafkaopsrequest.ops.kubedb.com/kafka-gitops-rotate-auth-pkb3t1         RotateAuth          Successful   13m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-mwqdzx     VerticalScaling     Successful   18h
kafkaopsrequest.ops.kubedb.com/kafka-gitops-volumeexpansion-7aweww     VolumeExpansion     Successful   17h
```


### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

To add tls, we are going to create an example `Issuer` that will be used to enable SSL/TLS in Kafka. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

Now we are going to create a `ca-secret` using the certificate files that we have just generated.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kafka-ca
  namespace: demo
type: Opaque
data:
  ca.crt: <base64-encoded-ca.crt>
```

Now, Let's create an `Issuer` using the `Kafka-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: kf-issuer
  namespace: demo
spec:
  ca:
    secretName: kafka-ca
```

Let's add that to our `kubedb/kf-issuer.yaml` and kubedb/kf-secret.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── kf-configuration.yaml
│ ├── kf-rotateauth.yaml
│ ├── kf-secret.yaml
│ ├── kf-issuer.yaml
│ └── Kafka.yaml
1 directories, 5 files
```

Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  version: 3.9.0
  authSecret:
    kind: Secret
    name: kf-rotate-auth
  configSecret:
    name: kf-reconfig-topo
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: kf-issuer
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: longhorn
    controller:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: longhorn
  storageType: Durable
  deletionPolicy: WipeOut
```

Add `sslMode` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository has been successfully synchronized with ArgoCD. The `Kafka` CR has been updated, and both the `issuer` and the corresponding `secret` have been created in the cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` KafkaOpsRequest to update the `Kafka` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,kafka,kfops,pods -n demo
NAME                            VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops   3.9.0     Ready    67m

NAME                                   AGE
kafka.gitops.kubedb.com/kafka-gitops   67m

NAME                                                                   TYPE                STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-horizontalscaling-rqmqe5   HorizontalScaling   Successful   27m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfigure-bqek6h         Reconfigure         Successful   27m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfiguretls-wkax2u      ReconfigureTLS      Successful   27m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-rotate-auth-y2vwx4         RotateAuth          Successful   27m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-c4llju     VerticalScaling     Successful   27m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-volumeexpansion-9e85tf     VolumeExpansion     Successful   27m
```


> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for Kafka.

### Update Version

List Kafka versions using `kubectl get kafkaversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/kafka/update-version/update-version.md).

Let's choose `4.0.0` in this example.

Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  version: 4.0.0
  authSecret:
    kind: Secret
    name: kf-rotate-auth
  configSecret:
    name: kf-reconfig-topo
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: kf-issuer
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: longhorn
    controller:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: longhorn
  storageType: Durable
  deletionPolicy: WipeOut
```

Update the `version` field to `4.0.0`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Kafka` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` KafkaOpsRequest to update the `Kafka` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get kf,kafka,kfops,pods -n demo
NAME                            VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops   4.0.0     Ready    72m

NAME                                   AGE
kafka.gitops.kubedb.com/kafka-gitops   72m

NAME                                                                   TYPE                STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-horizontalscaling-rqmqe5   HorizontalScaling   Successful   32m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfigure-bqek6h         Reconfigure         Successful   32m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfiguretls-wkax2u      ReconfigureTLS      Successful   32m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-rotate-auth-y2vwx4         RotateAuth          Successful   32m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-versionupdate-6z70bp       UpdateVersion       Successful   4m16s
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-c4llju     VerticalScaling     Successful   32m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-volumeexpansion-9e85tf     VolumeExpansion     Successful   32m
```


Now, we are going to verify whether the `Kafka`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get Kafka -n demo kafka-gitops -o=jsonpath='{.spec.version}{"\n"}'
4.0.0
$ kubectl get petset -n demo kafka-gitops-broker -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/kafka:4.0.0@sha256:62fb3652bc7672a74582d8d0abb7d0090155a237b7cf21bdb3837c3dba107010
$ kubectl get pod -n demo kafka-gitops-broker-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/kafka:4.0.0@sha256:62fb3652bc7672a74582d8d0abb7d0090155a237b7cf21bdb3837c3dba107010
$ kubectl get pod -n demo kafka-gitops-controller-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/kafka:4.0.0@sha256:62fb3652bc7672a74582d8d0abb7d0090155a237b7cf21bdb3837c3dba107010
$ kubectl get petset -n demo kafka-gitops-controller -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/kafka:4.0.0@sha256:62fb3652bc7672a74582d8d0abb7d0090155a237b7cf21bdb3837c3dba107010
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `Kafka.yaml` with the following,
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Kafka
metadata:
  name: kafka-gitops
  namespace: demo
spec:
  version: 4.0.0
  authSecret:
    kind: Secret
    name: kf-rotate-auth
  configSecret:
    name: kf-reconfig-topo
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: kf-issuer
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9091
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  topology:
    broker:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: longhorn
    controller:
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1540Mi
                requests:
                  cpu: 500m
                  memory: 1536Mi
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 2Gi
        storageClassName: longhorn
  storageType: Durable
  deletionPolicy: WipeOut
```

Add `monitor` field in the `spec`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Kafka` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` KafkaOpsRequest to add the `Kafka` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get kf,kafka,kfops -n demo
NAME                            VERSION   STATUS   AGE
kafka.kubedb.com/kafka-gitops   4.0.0     Ready    90m

NAME                                   AGE
kafka.gitops.kubedb.com/kafka-gitops   90m

NAME                                                                   TYPE                STATUS       AGE
kafkaopsrequest.ops.kubedb.com/kafka-gitops-horizontalscaling-rqmqe5   HorizontalScaling   Successful   49m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfigure-bqek6h         Reconfigure         Successful   49m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-reconfiguretls-wkax2u      ReconfigureTLS      Successful   49m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-restart-kzxwxa             Restart             Successful   12m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-rotate-auth-y2vwx4         RotateAuth          Successful   49m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-versionupdate-6z70bp       UpdateVersion       Successful   22m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-verticalscaling-c4llju     VerticalScaling     Successful   49m
kafkaopsrequest.ops.kubedb.com/kafka-gitops-volumeexpansion-9e85tf     VolumeExpansion     Successful   49m
```

Verify the monitoring is enabled by checking the prometheus targets.



## Next Steps

- Learn Kafka Scaling
    - [Horizontal Scaling](/docs/guides/kafka/scaling/horizontal-scaling/topology.md)
    - [Vertical Scaling](/docs/guides/kafka/scaling/vertical-scaling/topology.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/kafka/update-version/overview.md)
- Monitor your KafkaQL database with KubeDB using [built-in Prometheus](/docs/guides/kafka/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
