---
title: GitOps MySQL
menu:
  docs_{{ .version }}:
    identifier: my-using-gitops
    name: GitOps MySQL
    parent: my-gitops-MySQL
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
  replicas: 3
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: MySQL
        resources:
          limits:
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 1Gi
  storage:
    storageClassName: "standard"
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
    └── MySQL.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is created in your cluster.

Our `gitops` operator will create an actual `MySQL` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$  kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   172m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.1.0     Ready    172m
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` MySQL.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=my-gitops'
NAME                                     AGE
petset.apps.k8s.appscode.com/my-gitops   174m

NAME              READY   STATUS    RESTARTS   AGE
pod/my-gitops-0   2/2     Running   0          133m
pod/my-gitops-1   2/2     Running   0          129m
pod/my-gitops-2   2/2     Running   0          132m

NAME                    TYPE                       DATA   AGE
secret/my-gitops-auth   kubernetes.io/basic-auth   2      174m

NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/my-gitops           ClusterIP   10.43.39.209    <none>        3306/TCP   174m
service/my-gitops-pods      ClusterIP   None            <none>        3306/TCP   174m
service/my-gitops-standby   ClusterIP   10.43.212.101   <none>        3306/TCP   174m

NAME                                           TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/my-gitops   kubedb.com/mysql   9.1.0     174m
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
  replicas: 3
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: MySQL
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 700m
            memory: 2Gi
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Resource Requests and Limits are updated to `700m` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `MySQLOpsRequest` to update the `MySQL` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   178m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.1.0     Ready    178m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-mw8s6j     VerticalScaling     Successful   144m
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$  kubectl get pod -n demo my-gitops-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
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
  replicas: 5
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: MySQL
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 700m
            memory: 2Gi
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Update the `replicas` to `5`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` MySQLOpsRequest to update the `MySQL` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$  kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   178m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.1.0     Ready    178m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-s9lyzc   HorizontalScaling   Successful   117m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-mw8s6j     VerticalScaling     Successful   144m

```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=my-gitops'
NAME          READY   STATUS    RESTARTS   AGE
my-gitops-0   2/2     Running   0          144m
my-gitops-1   2/2     Running   0          140m
my-gitops-2   2/2     Running   0          143m

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
  replicas: 5
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: MySQL
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 700m
            memory: 2Gi
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Update the `storage.resources.requests.storage` to `10Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` MySQLOpsRequest to update the `MySQL` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                     AGE
MySQL.gitops.kubedb.com/my-gitops   27m

NAME                              VERSION   STATUS   AGE
MySQL.kubedb.com/my-gitops   16.6      Ready    27m

NAME                                                                     TYPE                STATUS       AGE
MySQLopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-wvxu5x   HorizontalScaling   Successful   6m
MySQLopsrequest.ops.kubedb.com/my-gitops-verticalscaling-i0kr1l     VerticalScaling     Successful   13m
MySQLopsrequest.ops.kubedb.com/my-gitops-volumeexpansion-2j5x5g     VolumeExpansion     Progressing  2s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=my-gitops'
NAME                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-my-gitops-0   Bound    pvc-061f3622-234f-4f91-b4d1-b81aa8739503   10Gi       RWO            standard       <unset>                 30m
data-my-gitops-1   Bound    pvc-045fc563-fb4e-416c-a9c2-b20c96532978   10Gi       RWO            standard       <unset>                 30m
data-my-gitops-2   Bound    pvc-a0f1d8fd-a677-4407-80b1-104b9f7b4cd1   10Gi       RWO            standard       <unset>                 30m
data-my-gitops-3   Bound    pvc-060b6fab-0c2d-4935-b31b-2866be68dd6f   10Gi       RWO            standard       <unset>                 8m58s
data-my-gitops-4   Bound    pvc-8149b579-a40f-4cd8-ac37-6a2401fd7807   10Gi       RWO            standard       <unset>                 8m23s
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
To know more about this configuration file, check [here](/docs/guides/MySQL/configuration/using-config-file.md)
```yaml
apiVersion: v1
stringData:
  user.conf: |
    max_connections=200
    shared_buffers=256MB
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
  configSecret:
    name: my-configuration
  replicas: 5
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: MySQL
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 700m
            memory: 2Gi
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` MySQLOpsRequest to update the `MySQL` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                     AGE
MySQL.gitops.kubedb.com/my-gitops   36m

NAME                              VERSION   STATUS   AGE
MySQL.kubedb.com/my-gitops   16.6      Ready    36m

NAME                                                                     TYPE                STATUS        AGE
MySQLopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-wvxu5x   HorizontalScaling   Successful    15m
MySQLopsrequest.ops.kubedb.com/my-gitops-reconfigure-i4r23j         Reconfigure         Progressing   1s
MySQLopsrequest.ops.kubedb.com/my-gitops-verticalscaling-i0kr1l     VerticalScaling     Successful    23m
```

After Ops Request becomes `Succesful`, lets check these parameters,

```bash
$  kubectl exec -it -n demo my-gitops-0 -- bash
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
| read_buffer_size | 1048576 |
+------------------+---------+
1 row in set (0.01 sec)
```
You can check the other pods same way.
So we have configured custom parameters.

> We can also reconfigure the parameters creating another secret and reference the secret in the `configSecret` field. Also you can remove the `configSecret` field to use the default parameters.

### Rotate MySQL Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.

We will do that using gitops, create the file `kubedb/my-auth.yaml` with the following content,

```yaml
apiVersion: v1
data:
  password: cGdwYXNzd29yZA==
  username: cG9zdGdyZXM=
kind: Secret
metadata:
  name: my-rotate-auth
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
  version: "9.1.0"
  replicas: 3
  authSecret:
    kind: Secret
    name: mysql-quickstart-auth-user
  configuration:
    secretName: my-configuration
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  podTemplate:
    spec:
      containers:
        - name: mysql
          resources:
            limits:
              cpu: 1000m
              memory: 1.5Gi
            requests:
              cpu: 500m
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

Change the `authSecret` field to `my-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` MySQLOpsRequest to update the `MySQL` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   5h2m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.1.0     Ready    5h2m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-s9lyzc   HorizontalScaling   Successful   4h1m
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfigure-q4yrg0         Reconfigure         Successful   4h28m
mysqlopsrequest.ops.kubedb.com/my-gitops-rotate-auth-5lg5o3         RotateAuth          Successful   10m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-mw8s6j     VerticalScaling     Successful   4h28m

```

After Ops Request becomes `Successful`, We can validate the changes connecting MySQL with new credentials.
```bash
$ kubectl exec -it -n demo my-gitops-0 -c mysql -- bash
bash-5.1$  mysql -uroot -p"Mysql2"
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 347
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql>  SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

```

### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

To add tls, we are going to create an example `Issuer` that will be used to enable SSL/TLS in MySQL. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls mysql-ca \
                                       --cert=ca.crt \
                                       --key=ca.key \
                                       --namespace=demo
secret/mysql-ca created
```

Now, Let's create an `Issuer` using the `MySQL-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: my-issuer
  namespace: demo
spec:
  ca:
    secretName: MySQL-ca
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
  version: "9.1.0"
  replicas: 3
  authSecret:
    kind: Secret
    name: mysql-quickstart-auth-user
  configuration:
    secretName: my-configuration
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: my-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  podTemplate:
    spec:
      containers:
        - name: mysql
          resources:
            limits:
              cpu: 1000m
              memory: 1.5Gi
            requests:
              cpu: 500m
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

Add `sslMode` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` MySQLOpsRequest to update the `MySQL` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                AGE
mysql.gitops.kubedb.com/my-gitops   6h7m

NAME                         VERSION   STATUS   AGE
mysql.kubedb.com/my-gitops   9.1.0     Ready    6h7m

NAME                                                                TYPE                STATUS       AGE
mysqlopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-s9lyzc   HorizontalScaling   Successful   5h6m
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfigure-q4yrg0         Reconfigure         Successful   5h33m
mysqlopsrequest.ops.kubedb.com/my-gitops-reconfiguretls-39ir77      ReconfigureTLS      Successful   4m39s
mysqlopsrequest.ops.kubedb.com/my-gitops-rotate-auth-5lg5o3         RotateAuth          Successful   74m
mysqlopsrequest.ops.kubedb.com/my-gitops-verticalscaling-mw8s6j     VerticalScaling     Successful   5h33m

```

After Ops Request becomes `Successful`, We can validate the changes connecting MySQL with new credentials.
```bash
$  kubectl exec -it -n demo my-gitops-0 -c mysql -- bash
bash-5.1$  mysql -uroot -p"Mysql2"
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 347
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql>  SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> command terminated with exit code 137
banusree@bonusree-datta-PC ~ [SIGKILL]> kubectl exec -it -n demo my-gitops-0 -c mysql -- bash
bash-5.1$ ls /etc/mysql/certs/
ca.crt	client.crt  client.key	server.crt  server.key
bash-5.1$ mysql -u${MYSQL_ROOT_USERNAME} -p{MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'root'@'localhost' (using password: YES)
bash-5.1$  mysql -uroot -p"Mysql2"
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 201
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

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
38 rows in set (0.00 sec)

mysql> SHOW VARIABLES LIKE '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | OFF   |
+--------------------------+-------+
1 row in set (0.00 sec)


```

> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for MySQL.

### Update Version

List MySQL versions using `kubectl get MySQLversion` and choose desired version that is compatible for umyrade from current version. Check the version constraints and ops request [here](/docs/guides/MySQL/update-version/versionumyrading/index.md).

Let's choose `17.4` in this example.

Update the `MySQL.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-gitops
  namespace: demo
spec:
  authSecret:
    kind: Secret
    name: my-rotate-auth
  configSecret:
    name: my-configuration
  replicas: 5
  version: "17.4"
  storageType: Durable
  sslMode: verify-full
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: my-issuer
      kind: Issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  podTemplate:
    spec:
      containers:
      - name: MySQL
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 700m
            memory: 2Gi
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Update the `version` field to `17.4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` MySQLOpsRequest to update the `MySQL` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                     AGE
MySQL.gitops.kubedb.com/my-gitops   3h25m

NAME                              VERSION   STATUS   AGE
MySQL.kubedb.com/my-gitops   16.6      Ready    3h25m

NAME                                                                     TYPE                STATUS        AGE
MySQLopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-wvxu5x   HorizontalScaling   Successful    3h3m
MySQLopsrequest.ops.kubedb.com/my-gitops-reconfigure-i4r23j         Reconfigure         Successful    168m
MySQLopsrequest.ops.kubedb.com/my-gitops-reconfiguretls-91fseg      ReconfigureTLS      Successful    7m33s
MySQLopsrequest.ops.kubedb.com/my-gitops-rotate-auth-zot83x         RotateAuth          Successful    161m
MySQLopsrequest.ops.kubedb.com/my-gitops-versionupdate-1wxgt9       UpdateVersion       Progressing   4s
MySQLopsrequest.ops.kubedb.com/my-gitops-verticalscaling-i0kr1l     VerticalScaling     Successful    3h11m
```


Now, we are going to verify whether the `MySQL`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get MySQL -n demo my-gitops -o=jsonpath='{.spec.version}{"\n"}'
17.4

$ kubectl get petset -n demo my-gitops -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/MySQL:17.4-alpine

$ kubectl get pod -n demo my-gitops-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/MySQL:17.4-alpine
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
  authSecret:
    kind: Secret
    name: my-rotate-auth
  configSecret:
    name: my-configuration
  replicas: 5
  version: "17.4"
  storageType: Durable
  sslMode: verify-full
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: my-issuer
      kind: Issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  podTemplate:
    spec:
      containers:
      - name: MySQL
        resources:
          limits:
            memory: 2Gi
          requests:
            cpu: 700m
            memory: 2Gi
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `MySQL` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` MySQLOpsRequest to add the `MySQL` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get mysql.gitops.kubedb.com,mysql.kubedb.com,MySQLopsrequest -n demo
NAME                                     AGE
MySQL.gitops.kubedb.com/my-gitops   3h34m

NAME                              VERSION   STATUS     AGE
MySQL.kubedb.com/my-gitops   16.6      NotReady   3h34m

NAME                                                                     TYPE                STATUS       AGE
MySQLopsrequest.ops.kubedb.com/my-gitops-horizontalscaling-wvxu5x   HorizontalScaling   Successful   3h13m
MySQLopsrequest.ops.kubedb.com/my-gitops-reconfigure-i4r23j         Reconfigure         Successful   177m
MySQLopsrequest.ops.kubedb.com/my-gitops-reconfiguretls-91fseg      ReconfigureTLS      Successful   16m
MySQLopsrequest.ops.kubedb.com/my-gitops-restart-nhjk9u             Restart             Progressing  2s
MySQLopsrequest.ops.kubedb.com/my-gitops-rotate-auth-zot83x         RotateAuth          Successful   170m
MySQLopsrequest.ops.kubedb.com/my-gitops-versionupdate-1wxgt9       UpdateVersion       Successful   9m30s
MySQLopsrequest.ops.kubedb.com/my-gitops-verticalscaling-i0kr1l     VerticalScaling     Successful   3h21m
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

- Learn MySQL [GitOps](/docs/guides/MySQL/concepts/MySQL-gitops.md)
- Learn MySQL Scaling 
  - [Horizontal Scaling](/docs/guides/MySQL/scaling/horizontal-scaling/overview/index.md)
  - [Vertical Scaling](/docs/guides/MySQL/scaling/vertical-scaling/overview/index.md)
- Learn Version Update Ops Request and Constraints [here](/docs/guides/MySQL/update-version/versionumyrading/index.md)
- Monitor your MySQL database with KubeDB using [built-in Prometheus](/docs/guides/MySQL/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
