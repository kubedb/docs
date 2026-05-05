---
title: GitOps MySQL 
menu:
  docs_{{ .version }}:
    identifier: mysql-gitops
    name: Guide
    parent: guides-mysql-gitops
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps MySQL using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create MySQL database and manage updates using GitOps workflow.

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
> Note: YAML files used in this tutorial are stored in [docs/examples/MySQL](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/MySQL) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

## Create MySQL Database using GitOps

### Create a MySQL GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.4.0"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Create a directory like below,
```bash
$ tree .
├── kubedb
    └── MySQL.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is created in your cluster.

Our `gitops` operator will create an actual `MySQL` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   76m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.4.0     Ready    76m
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` MySQL.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=my-gitops'
NAME                                     AGE
petset.apps.k8s.appscode.com/my-gitops   78m

NAME              READY   STATUS    RESTARTS   AGE
pod/my-gitops-0   2/2     Running   0          78m
pod/my-gitops-1   2/2     Running   0          75m
pod/my-gitops-2   2/2     Running   0          75m

NAME                    TYPE                       DATA   AGE
secret/my-gitops-auth   kubernetes.io/basic-auth   2      78m

NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/my-gitops           ClusterIP   10.43.236.155   <none>        3306/TCP   78m
service/my-gitops-pods      ClusterIP   None            <none>        3306/TCP   78m
service/my-gitops-standby   ClusterIP   10.43.239.55    <none>        3306/TCP   78m

NAME                                           TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/my-gitops   kubedb.com/mysql   9.4.0     78m
```

## Update MySQL Database using GitOps

### Scale MySQL Database Resources


Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.4.0"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Resource requests have been updated to `700m` CPU and `1536Mi` memory, and the memory limits have been set to `1.5Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `MySQLOpsRequest` to update the `MySQL` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   178m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.4.0     Ready    178m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-mw8s6j     VerticalScaling     Successful   144m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo my-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "1536Mi"
  }
}
```

### Scale MySQL Replicas
Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.4.0"
  replicas: 4
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Update the `replicas` to `4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` MySQLOpsRequest to update the `MySQL` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$   kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   101m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.4.0     Ready    101m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-h542j4   HorizontalScaling   Successful   5m1s
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-zbtqnv     VerticalScaling     Successful   19m

```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=my-gitops'
NAME          READY   STATUS    RESTARTS   AGE
my-gitops-0   2/2     Running   0          15m
my-gitops-1   2/2     Running   0          19m
my-gitops-2   2/2     Running   0          17m
my-gitops-3   2/2     Running   0          5m37s
```

We can also scale down the replicas by updating the `replicas` fields.

### Exapand MySQL Volume

Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.4.0"
  replicas: 4
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Update the `storage.resources.requests.storage` to `2Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` MySQLOpsRequest to update the `MySQL` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   104m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.4.0     Ready    104m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-h542j4   HorizontalScaling   Successful   8m23s
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-zbtqnv     VerticalScaling     Successful   22m
mysqlopsrequest.ops.kubedb.com/my-gitops-volumeexpansion-tzncw1     VolumeExpansion     Successful   112s

```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=my-gitops'
NAME               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-my-gitops-0   Bound    pvc-c926c69f-ff5e-4f93-be78-1fc1d39da25c   2Gi        RWO            standard       <unset>                 105m
data-my-gitops-1   Bound    pvc-8bfcf4d8-37fb-4543-96c5-7a656270967a   2Gi        RWO            standard       <unset>                 101m
data-my-gitops-2   Bound    pvc-238989bc-0a53-4db8-a3c2-2ce77aee4042   2Gi        RWO            standard       <unset>                 101m
data-my-gitops-3   Bound    pvc-5345bc6a-baad-461a-a1fe-108e75c32a11   2Gi        RWO            standard       <unset>                8m41s
```

## Reconfigure MySQL

Before, create opsrequest  parameters were,
```shell
kubectl exec -it -n demo my-gitops-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-5.1$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 277
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 151   |
+-----------------+-------+
1 row in set (0.01 sec)

mysql> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 131072 |
+------------------+--------+
1 row in set (0.00 sec)

```
At first, we will create a secret containing `user.conf` file with required configuration settings.
To know more about this configuration file, check [here](/docs/guides/mysql/configuration/using-config-file.md)
```yaml
apiVersion: v1
stringData:
  user.cnf: |
    [mysqld]
    max_connections = 200
    read_buffer_size = 1048575
kind: Secret
metadata:
  name: my-configuration
  namespace: demo
type: Opaque
```

Now, we will add this file to `kubedb/my-configuration.yaml`.

```bash
$ tree .
├── kubedb
│ ├── my-configuration.yaml
│ └── MySQL.yaml
1 directories, 2 files
```

Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.6.0"
  replicas: 4
  configSecret:
    name: my-config
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` MySQLOpsRequest to update the `MySQL` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   126m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.4.0     Ready    126m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-h542j4   HorizontalScaling   Successful   30m
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfigure-yskbt5         Reconfigure         Successful   18m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-zbtqnv     VerticalScaling     Successful   44m
mysqlopsrequest.ops.kubedb.com/my-gitops-volumeexpansion-tzncw1     VolumeExpansion     Successful   23m
```

After Ops Request becomes `Succesful`, lets check these parameters,

```bash
$ kubectl exec -it -n demo my-gitops-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-5.1$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3115
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 200   |
+-----------------+-------+
1 row in set (0.00 sec)

mysql> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 1044480 |
+------------------+---------+

```
You can check the other pods same way.
So we have configured custom parameters.

> We can also reconfigure the parameters creating another secret and reference the secret in the `configuration.secretName` field. Also you can remove the `configuration` field to use the default parameters.

### Rotate MySQL Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will do that using gitops, create the file `kubedb/my-auth.yaml` with the following content,

```yaml
apiVersion: v1
data:
  password: bXlwYXNzd29yZA==
  username: cm9vdA==
kind: Secret
metadata:
  name: myauth
  namespace: demo
type: kubernetes.io/basic-auth
```

File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── my-auth.yaml
│ ├── my-configuration.yaml
│ └── MySQL.yaml
1 directories, 3 files
```

Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.6.0"
  replicas: 4
  configSecret:
    name: my-config
  authSecret:
    kind: Secret
    name: myauth  
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Change the `authSecret.name` field to `myauth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` MySQLOpsRequest to update the `MySQL` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   19h

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.4.0     Ready    19h

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-h542j4   HorizontalScaling   Successful   18h
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfigure-b5s92r         Reconfigure         Successful   60m
mysqlopsrequest.ops.kubedb.com/my-gitops-rotate-auth-q2z2vf         RotateAuth          Successful   24m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-zbtqnv     VerticalScaling     Successful   18h
mysqlopsrequest.ops.kubedb.com/my-gitops-volumeexpansion-tzncw1     VolumeExpansion     Successful   18h
```

After Ops Request becomes `Successful`, We can validate the changes connecting MySQL with new credentials.
```bash
$ kubectl exec -it -n demo my-gitops-0 -c mysql -- bash
bash-5.1$  mysql -uroot -p"mypassword"
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 808
Server version: 9.4.0 MySQL Community Server - GPL

Copyright (c) 2000, 2025, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.001 sec)
```

### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

First, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mysql/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls my-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/my-ca created
```

Now, we are going to create an `Issuer` using the `my-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mysql-issuer
  namespace: demo
spec:
  ca:
    secretName: my-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls/issuer.yaml
issuer.cert-manager.io/mysql-issuer created
```

Let's add that to our `kubedb/my-issuer.yaml` file. File structure will look like this,
```bash
$ tree .
├── kubedb
│ ├── my-auth.yaml
│ ├── my-configuration.yaml
│ ├── my-issuer.yaml
│ └── MySQL.yaml
1 directories, 4 files
```

Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.4.0"
  replicas: 4
  configSecret:
    name: my-config
  authSecret:
    kind: Secret
    name: myauth  
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Add `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` MySQLOpsRequest to update the `MySQL` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   20h

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.4.0     Ready    20h

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-h542j4   HorizontalScaling   Successful   18h
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfigure-b5s92r         Reconfigure         Successful   74m
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfiguretls-8kwaw9      ReconfigureTLS      Successful   9m51s
mysqlopsrequest.ops.kubedb.com/my-gitops-rotate-auth-q2z2vf         RotateAuth          Successful   38m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-zbtqnv     VerticalScaling     Successful   18h
mysqlopsrequest.ops.kubedb.com/my-gitops-volumeexpansion-tzncw1     VolumeExpansion     Successful   18h
```

After Ops Request becomes `Successful`, We can validate the changes connecting MySQL with new credentials.
```bash
$ kubectl exec -it -n demo my-gitops-0 -c mysql -- bash
bash-5.1$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 505
Server version: 9.4.0 MySQL Community Server - GPL

Copyright (c) 2000, 2025, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW VARIABLES LIKE '%ssl%';
+---------------------------------------------------+-----------------------------+
| Variable_name                                     | Value                       |
+---------------------------------------------------+-----------------------------+
| admin_ssl_ca                                      |                             |
| admin_ssl_capath                                  |                             |
| admin_ssl_cert                                    |                             |
| admin_ssl_cipher                                  |                             |
| admin_ssl_crl                                     |                             |
| admin_ssl_crlpath                                 |                             |
| admin_ssl_key                                     |                             |
| clone_ssl_ca                                      | /etc/mysql/certs/ca.crt     |
| clone_ssl_cert                                    | /etc/mysql/certs/server.crt |
| clone_ssl_key                                     | /etc/mysql/certs/server.key |
| group_replication_recovery_ssl_ca                 | /etc/mysql/certs/ca.crt     |
| group_replication_recovery_ssl_capath             |                             |
| group_replication_recovery_ssl_cert               | /etc/mysql/certs/server.crt |
| group_replication_recovery_ssl_cipher             |                             |
| group_replication_recovery_ssl_crl                |                             |
| group_replication_recovery_ssl_crlpath            |                             |
| group_replication_recovery_ssl_key                | /etc/mysql/certs/server.key |
| group_replication_recovery_ssl_verify_server_cert | OFF                         |
| group_replication_recovery_use_ssl                | ON                          |
| group_replication_ssl_mode                        | VERIFY_CA                   |
| mysqlx_ssl_ca                                     |                             |
| mysqlx_ssl_capath                                 |                             |
| mysqlx_ssl_cert                                   |                             |
| mysqlx_ssl_cipher                                 |                             |
| mysqlx_ssl_crl                                    |                             |
| mysqlx_ssl_crlpath                                |                             |
| mysqlx_ssl_key                                    |                             |
| performance_schema_show_processlist               | OFF                         |
| ssl_ca                                            | /etc/mysql/certs/ca.crt     |
| ssl_capath                                        | /etc/mysql/certs            |
| ssl_cert                                          | /etc/mysql/certs/server.crt |
| ssl_cipher                                        |                             |
| ssl_crl                                           |                             |
| ssl_crlpath                                       |                             |
| ssl_fips_mode                                     | OFF                         |
| ssl_key                                           | /etc/mysql/certs/server.key |
| ssl_session_cache_mode                            | ON                          |
| ssl_session_cache_timeout                         | 300                         |
+---------------------------------------------------+-----------------------------+
38 rows in set (0.016 sec)

mysql> SHOW VARIABLES LIKE '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | OFF   |
+--------------------------+-------+
1 row in set (0.002 sec)
```

> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for MySQL.

### Update Version

List MySQL versions using `kubectl get MySQLversion` and choose desired version that is compatible for umyrade from current version. Check the version constraints and ops request [here](/docs/guides/mysql/update-version/versionumyrading/index.md).

Let's choose `9.6.0` in this example.

Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.6.0"
  replicas: 4
  configSecret:
    name: my-config
  authSecret:
    kind: Secret
    name: myauth  
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Update the `version` field to `9.6.0`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` MySQLOpsRequest to update the `MySQL` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   22h

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.6.0     Ready    22h

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-h542j4   HorizontalScaling   Successful   20h
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfigure-b5s92r         Reconfigure         Successful   3h40m
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfiguretls-8kwaw9      ReconfigureTLS      Successful   156m
mysqlopsrequest.ops.kubedb.com/my-gitops-rotate-auth-q2z2vf         RotateAuth          Successful   3h5m
mysqlopsrequest.ops.kubedb.com/my-gitops-versionupdate-bskr89       UpdateVersion       Successful   142m
mysqlopsrequest.ops.kubedb.com/my-gitops-versionupdate-g2n3y9       UpdateVersion       Successful   132m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-zbtqnv     VerticalScaling     Successful   21h
mysqlopsrequest.ops.kubedb.com/my-gitops-volumeexpansion-tzncw1     VolumeExpansion     Successful   20h
```


Now, we are going to verify whether the `MySQL`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get MySQL -n demo my-gitops -o=jsonpath='{.spec.version}{"\n"}'
9.6.0
$ kubectl get petset -n demo my-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/mysql:9.6.0-oracle@sha256:16e6b7b93df8aa255d3886ff33c2d78093d1cd2346522d14bf1b9cc0ad03a460
$ kubectl get pod -n demo my-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/mysql:9.6.0-oracle@sha256:16e6b7b93df8aa255d3886ff33c2d78093d1cd2346522d14bf1b9cc0ad03a460
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  version: "9.6.0"
  replicas: 4
  configSecret:
    name: my-config
  authSecret:
    kind: Secret
    name: myauth  
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: mysql
        resources:
          limits:
            memory: 1.5Gi
          requests:
            cpu: 700m
            memory: 1.5Gi
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` MySQLOpsRequest to add the `MySQL` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   22h

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.6.0     Ready    22h

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-h542j4   HorizontalScaling   Successful   21h
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfigure-b5s92r         Reconfigure         Successful   3h57m
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfiguretls-8kwaw9      ReconfigureTLS      Successful   172m
mysqlopsrequest.ops.kubedb.com/my-gitops-restart-lb1wyu             Restart             Successful   9m36s
mysqlopsrequest.ops.kubedb.com/my-gitops-rotate-auth-q2z2vf         RotateAuth          Successful   3h21m
mysqlopsrequest.ops.kubedb.com/my-gitops-versionupdate-bskr89       UpdateVersion       Successful   158m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-zbtqnv     VerticalScaling     Successful   21h
mysqlopsrequest.ops.kubedb.com/my-gitops-volumeexpansion-tzncw1     VolumeExpansion     Successful   21h
```

Verify the monitoring is enabled by checking the prometheus targets.

There are some other fields that will trigger `Restart` ops request.
- `.spec.monitor`
- `.spec.spec.archiver`
- `.spec.remoteReplica`
- `.spec.leaderElection`
- `spec.replication`
- `.spec.standbyMode`
- `.spec.streamingMode`
- `.spec.enforceGroup`
- `.spec.sslMode` etc.


## Next Steps

- Learn MySQL [GitOps](/docs/guides/mysql/concepts/MySQL-gitops.md)
- Learn MySQL Scaling 
  - [Horizontal Scaling](/docs/guides/mysql/scaling/horizontal-scaling/overview/index.md)
  - [Vertical Scaling](/docs/guides/mysql/scaling/vertical-scaling/overview/index.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/mysql/update-version/versionumyrading/index.md)
- Monitor your MySQL database with KubeDB using [built-in Prometheus](/docs/guides/mysql/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
