---
title: ProxySQL Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-rotate-auth-guide
    name: Guide
    parent: guides-proxysql-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of ProxySQL

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate the authentication credentials of a `ProxySQL` server using a `ProxySQLOpsRequest`. These credentials are used for the ProxySQL admin interface. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a secret of type `kubernetes.io/basic-auth` containing the desired `password` and then reference this secret in the `ProxySQLOpsRequest` CR.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/index.md)
  - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest/index.md)
  - [Rotate Authentication Overview](/docs/guides/proxysql/rotate-auth/overview.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in the [docs/examples/proxysql/rotate-auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/proxysql/rotate-auth) folder in the GitHub repository kubedb/docs.

## Prepare MySQL Backend

ProxySQL needs a backend to proxy the traffic to. In this tutorial we are going to use a KubeDB managed MySQL Group Replication as the backend. Below is the `MySQL` object we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.4.8"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/rotate-auth/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's wait for the MySQL to be `Ready`.

```bash
$ kubectl get mysql -n demo
NAME           VERSION   STATUS   AGE
mysql-server   8.4.8     Ready    4m
```

## Deploy ProxySQL

Now, we are going to create a `ProxySQL` cluster with the MySQL server above as the backend. Below is the `ProxySQL` object we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "3.0.1-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: mysql-server
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/rotate-auth/sample-proxysql.yaml
proxysql.kubedb.com/proxy-server created
```

Let's wait for the ProxySQL to be `Ready`.

```bash
$ kubectl get proxysql -n demo
NAME           VERSION        STATUS   AGE
proxy-server   3.0.1-debian   Ready    3m
```

## Verify Authentication

The user can verify whether they are authorized by connecting to the ProxySQL admin interface. To do this, the user needs the `username` and `password`. Below is an example showing how to retrieve the credentials from the secret.

```bash
$ kubectl get proxysql -n demo proxy-server -ojson | jq .spec.authSecret.name
"proxy-server-auth"
$ kubectl get secret -n demo proxy-server-auth -o jsonpath='{.data.username}' | base64 -d
cluster
$ kubectl get secret -n demo proxy-server-auth -o jsonpath='{.data.password}' | base64 -d
Hj5Tq0bWmZ2xKc9R
```

Now, you can exec into the pod `proxy-server-0` and connect to the admin interface using the `username` and `password`.

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -ucluster -pHj5Tq0bWmZ2xKc9R -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 312
Server version: 8.4.8 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select variable_name, variable_value from global_variables where variable_name in ('admin-cluster_username', 'admin-cluster_password');
+------------------------+------------------+
| variable_name          | variable_value   |
+------------------------+------------------+
| admin-cluster_username | cluster          |
| admin-cluster_password | Hj5Tq0bWmZ2xKc9R |
+------------------------+------------------+
2 rows in set (0.001 sec)
```

If you can connect to the admin interface and run queries, it means the secret is working correctly.

## Create RotateAuth ProxySQLOpsRequest

#### 1. Using Operator Generated Credentials:

In order to rotate the authentication of ProxySQL using operator generated credentials, we have to create a `ProxySQLOpsRequest` CR with `RotateAuth` type. Below is the YAML of the `ProxySQLOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: proxyops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  proxyRef:
    name: proxy-server
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.proxyRef.name` specifies that we are performing rotate authentication operation on `proxy-server`.
- `spec.type` specifies that we are performing `RotateAuth` on ProxySQL.

Let's create the `ProxySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/rotate-auth/rotate-auth-generated.yaml
proxysqlopsrequest.ops.kubedb.com/proxyops-rotate-auth-generated created
```

Let's wait for `ProxySQLOpsRequest` to be `Successful`. Run the following command to watch `ProxySQLOpsRequest` CRO,

```bash
$ kubectl get proxysqlopsrequest -n demo
NAME                             TYPE         STATUS       AGE
proxyops-rotate-auth-generated   RotateAuth   Successful   2m
```

If we describe the `ProxySQLOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe proxysqlopsrequest -n demo proxyops-rotate-auth-generated
Name:         proxyops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ProxySQLOpsRequest
Metadata:
  Creation Timestamp:  2026-10-06T08:12:31Z
  Generation:          1
  Resource Version:    184512
  UID:                 5b7c2f3e-91a4-4d0b-bf27-6e3d8a1c9f40
Spec:
  Apply:  IfReady
  Proxy Ref:
    Name:   proxy-server
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2026-10-06T08:12:31Z
    Message:               Controller has started to Progress the ProxySQLOpsRequest: demo/proxyops-rotate-auth-generated
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-10-06T08:12:34Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2026-10-06T08:12:39Z
    Message:               Successfully reconciled ProxySQL with new auth credentials for ProxySQLOpsRequest: demo/proxyops-rotate-auth-generated
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-10-06T08:14:02Z
    Message:               Successfully restarted ProxySQL pods for ProxySQLOpsRequest: demo/proxyops-rotate-auth-generated
    Observed Generation:   1
    Reason:                RestartPodsSucceeded
    Status:                True
    Type:                  Restart
    Last Transition Time:  2026-10-06T08:14:03Z
    Message:               Controller has successfully rotated authentication for ProxySQL demo/proxy-server
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age    From                         Message
  ----    ------      ----   ----                         -------
  Normal  Starting    2m     KubeDB Ops-manager Operator  Started processing for ProxySQLOpsRequest: demo/proxyops-rotate-auth-generated
  Normal  Starting    2m     KubeDB Ops-manager Operator  Pausing ProxySQL databse: demo/proxy-server
  Normal  Successful  2m     KubeDB Ops-manager Operator  Successfully paused ProxySQL database: demo/proxy-server for ProxySQLOpsRequest: proxyops-rotate-auth-generated
  Normal  Starting    2m     KubeDB Ops-manager Operator  Restarting Pod: demo/proxy-server-0
  Normal  Starting    110s   KubeDB Ops-manager Operator  Restarting Pod: demo/proxy-server-1
  Normal  Starting    80s    KubeDB Ops-manager Operator  Restarting Pod: demo/proxy-server-2
  Normal  Starting    50s    KubeDB Ops-manager Operator  Resuming ProxySQL database: demo/proxy-server
  Normal  Successful  50s    KubeDB Ops-manager Operator  Successfully resumed ProxySQL database: demo/proxy-server
  Normal  Successful  50s    KubeDB Ops-manager Operator  Controller has successfully rotated ProxySQL authentication
```

**Verify Auth is rotated**

```bash
$ kubectl get proxysql -n demo proxy-server -ojson | jq .spec.authSecret.name
"proxy-server-auth"
$ kubectl get secret -n demo proxy-server-auth -o jsonpath='{.data.username}' | base64 -d
cluster
$ kubectl get secret -n demo proxy-server-auth -o jsonpath='{.data.password}' | base64 -d
pN8vLw3sYe6QaD1u
```

Let's verify if we can connect to the admin interface using the new credentials.

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -ucluster -ppN8vLw3sYe6QaD1u -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 14
Server version: 8.4.8 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select variable_name, variable_value from global_variables where variable_name in ('admin-cluster_username', 'admin-cluster_password');
+------------------------+------------------+
| variable_name          | variable_value   |
+------------------------+------------------+
| admin-cluster_username | cluster          |
| admin-cluster_password | pN8vLw3sYe6QaD1u |
+------------------------+------------------+
2 rows in set (0.001 sec)
```

Also, there will be two more new keys in the secret that store the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo proxy-server-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
cluster
$ kubectl get secret -n demo proxy-server-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
Hj5Tq0bWmZ2xKc9R
```

Let's confirm that the previous credentials no longer work.

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -ucluster -pHj5Tq0bWmZ2xKc9R -h127.0.0.1 -P6032
ERROR 1045 (28000): ProxySQL Error: Access denied for user 'cluster'@'127.0.0.1' (using password: YES)
```

The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.

#### 2. Using User Created Credentials

At first, we need to create a secret with `kubernetes.io/basic-auth` type using custom password. Below is the command to create a secret with `kubernetes.io/basic-auth` type,

> Note: The `username` must be fixed as `cluster` and the `password` must be alphanumeric, otherwise the `ProxySQLOpsRequest` will fail.

```bash
$ kubectl create secret generic proxy-server-auth-user -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=cluster \
                --from-literal=password=ProxySQL2026
secret/proxy-server-auth-user created
```

Now create a `ProxySQLOpsRequest` with `RotateAuth` type. Below is the YAML of the `ProxySQLOpsRequest` that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: proxyops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  proxyRef:
    name: proxy-server
  authentication:
    secretRef:
      kind: Secret
      name: proxy-server-auth-user
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.proxyRef.name` specifies that we are performing rotate authentication operation on `proxy-server`.
- `spec.type` specifies that we are performing `RotateAuth` on ProxySQL.
- `spec.authentication.secretRef.name` specifies that we use `proxy-server-auth-user` as the new auth secret of ProxySQL.

Let's create the `ProxySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/rotate-auth/rotate-auth-user.yaml
proxysqlopsrequest.ops.kubedb.com/proxyops-rotate-auth-user created
```

Let's wait for `ProxySQLOpsRequest` to be `Successful`. Run the following command to watch `ProxySQLOpsRequest` CRO:

```bash
$ kubectl get proxysqlopsrequest -n demo
NAME                             TYPE         STATUS       AGE
proxyops-rotate-auth-generated   RotateAuth   Successful   12m
proxyops-rotate-auth-user        RotateAuth   Successful   2m
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded. If we describe the `ProxySQLOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe proxysqlopsrequest -n demo proxyops-rotate-auth-user
Name:         proxyops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ProxySQLOpsRequest
Metadata:
  Creation Timestamp:  2026-10-06T08:22:47Z
  Generation:          1
  Resource Version:    186930
  UID:                 a3e91d5c-0f7b-4c62-8d14-27b6e5f0c8a1
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Kind:  Secret
      Name:  proxy-server-auth-user
  Proxy Ref:
    Name:   proxy-server
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2026-10-06T08:22:47Z
    Message:               Controller has started to Progress the ProxySQLOpsRequest: demo/proxyops-rotate-auth-user
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-10-06T08:22:50Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2026-10-06T08:22:55Z
    Message:               Successfully reconciled ProxySQL with new auth credentials for ProxySQLOpsRequest: demo/proxyops-rotate-auth-user
    Observed Generation:   1
    Reason:                UpdatePetSetsSucceeded
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-10-06T08:24:16Z
    Message:               Successfully restarted ProxySQL pods for ProxySQLOpsRequest: demo/proxyops-rotate-auth-user
    Observed Generation:   1
    Reason:                RestartPodsSucceeded
    Status:                True
    Type:                  Restart
    Last Transition Time:  2026-10-06T08:24:17Z
    Message:               Controller has successfully rotated authentication for ProxySQL demo/proxy-server
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age    From                         Message
  ----    ------      ----   ----                         -------
  Normal  Starting    2m     KubeDB Ops-manager Operator  Started processing for ProxySQLOpsRequest: demo/proxyops-rotate-auth-user
  Normal  Starting    2m     KubeDB Ops-manager Operator  Pausing ProxySQL databse: demo/proxy-server
  Normal  Successful  2m     KubeDB Ops-manager Operator  Successfully paused ProxySQL database: demo/proxy-server for ProxySQLOpsRequest: proxyops-rotate-auth-user
  Normal  Starting    2m     KubeDB Ops-manager Operator  Restarting Pod: demo/proxy-server-0
  Normal  Starting    110s   KubeDB Ops-manager Operator  Restarting Pod: demo/proxy-server-1
  Normal  Starting    80s    KubeDB Ops-manager Operator  Restarting Pod: demo/proxy-server-2
  Normal  Starting    50s    KubeDB Ops-manager Operator  Resuming ProxySQL database: demo/proxy-server
  Normal  Successful  50s    KubeDB Ops-manager Operator  Successfully resumed ProxySQL database: demo/proxy-server
  Normal  Successful  50s    KubeDB Ops-manager Operator  Controller has successfully rotated ProxySQL authentication
```

**Verify Auth is rotated**

```bash
$ kubectl get proxysql -n demo proxy-server -ojson | jq .spec.authSecret.name
"proxy-server-auth-user"
$ kubectl get secret -n demo proxy-server-auth-user -o jsonpath='{.data.username}' | base64 -d
cluster
$ kubectl get secret -n demo proxy-server-auth-user -o jsonpath='{.data.password}' | base64 -d
ProxySQL2026
```

Let's verify if we can connect to the admin interface using the new credentials.

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -ucluster -pProxySQL2026 -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 11
Server version: 8.4.8 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select variable_name, variable_value from global_variables where variable_name in ('admin-cluster_username', 'admin-cluster_password');
+------------------------+----------------+
| variable_name          | variable_value |
+------------------------+----------------+
| admin-cluster_username | cluster        |
| admin-cluster_password | ProxySQL2026   |
+------------------------+----------------+
2 rows in set (0.001 sec)
```

Also, there will be two more new keys in the secret that store the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo proxy-server-auth-user -o go-template='{{ index .data "username.prev" }}' | base64 -d
cluster
$ kubectl get secret -n demo proxy-server-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
pN8vLw3sYe6QaD1u
```

Let's confirm that the previous credentials no longer work.

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -ucluster -ppN8vLw3sYe6QaD1u -h127.0.0.1 -P6032
ERROR 1045 (28000): ProxySQL Error: Access denied for user 'cluster'@'127.0.0.1' (using password: YES)
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete proxysqlopsrequest -n demo proxyops-rotate-auth-generated proxyops-rotate-auth-user
proxysqlopsrequest.ops.kubedb.com "proxyops-rotate-auth-generated" deleted
proxysqlopsrequest.ops.kubedb.com "proxyops-rotate-auth-user" deleted
$ kubectl delete proxysql -n demo proxy-server
proxysql.kubedb.com "proxy-server" deleted
$ kubectl delete mysql -n demo mysql-server
mysql.kubedb.com "mysql-server" deleted
$ kubectl delete secret -n demo proxy-server-auth-user proxy-server-auth
secret "proxy-server-auth-user" deleted
secret "proxy-server-auth" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Detail concepts of [ProxySQL object](/docs/guides/proxysql/concepts/proxysql/index.md).
- Detail concepts of [ProxySQLOpsRequest object](/docs/guides/proxysql/concepts/opsrequest/index.md).
- Want to set up a ProxySQL cluster? Check how to configure a [ProxySQL Cluster](/docs/guides/proxysql/clustering/proxysql-cluster/index.md).
- Monitor your ProxySQL with KubeDB using [built-in Prometheus](/docs/guides/proxysql/monitoring/builtin-prometheus/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
