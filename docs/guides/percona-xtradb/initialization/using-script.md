---
title: Initialize PerconaXtraDB using Script
menu:
  docs_{{ .version }}:
    identifier: px-using-script-initialization
    name: Using Script
    parent: px-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize PerconaXtraDB using Script

This tutorial will show you how to use KubeDB to initialize a PerconaXtraDB database with \*.sql, \*.sh and/or \*.sql.gz script.
In this tutorial we will use .sql script stored in GitHub repository [kubedb/percona-xtradb-init-scripts](https://github.com/kubedb/percona-xtradb-init-scripts).

> Note: The yaml files that are used in this tutorial are stored in [docs/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB CLI on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Prepare Initialization Scripts

PerconaXtraDB supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use `init.sql` script from [percona-xtradb-init-scripts](https://github.com/kubedb/percona-xtradb-init-scripts) git repository to create a TABLE `kubedb_table` in the `mysql` database.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from `init.sql` file. Then, we will provide this ConfigMap as script source in `.spec.init.script` of PerconaXtraDB object spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo px-init-script \
--from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/percona-xtradb-init-scripts/raw/master/init.sql)"
configmap/px-init-script created
```

## Create a PerconaXtraDB database with Init-Script

Below is the `PerconaXtraDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: px-init-script
  namespace: demo
spec:
  version: "5.7"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  init:
    script:
      configMap:
        name: px-init-script
  terminationPolicy: DoNotTerminate
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/px-init-script.yaml
perconaxtradb.kubedb.com/px-init-script created
```

Here,

- `.spec.init.script` specifies a script source used to initialize the database before database server starts. The scripts will be executed alphabetically. In this tutorial, a sample `.sql` script from the git repository `https://github.com/kubedb/percona-xtradb-init-scripts.git` is used to create a test database. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) instead of `ConfigMap`. The \*.sql, \*sql.gz and/or \*.sh scripts that are stored inside the directory `/docker-entrypoint-initdb.d` will be executed alphabetically. The scripts inside child folders will be skipped.

KubeDB operator watches for `PerconaXtraDB` objects using Kubernetes API. When a `PerconaXtraDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching PerconaXtraDB object name. KubeDB operator will also create a governing service for StatefulSets with the name ``<percona-xtradb-object-name>-gvr`, if one is not already present.

```bash
$ kubectl dba describe px -n demo px-init-script
Name:         px-init-script
Namespace:    demo
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"kubedb.com/v1alpha2","kind":"PerconaXtraDB","metadata":{"annotations":{},"name":"px-init-script","namespace":"demo"},"spec"...
API Version:  kubedb.com/v1alpha2
Kind:         PerconaXtraDB
Metadata:
  Creation Timestamp:  2020-01-09T13:45:43Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  64559
  Self Link:         /apis/kubedb.com/v1alpha2/namespaces/demo/perconaxtradbs/px-init-script
  UID:               d7ad081e-8b2d-41d1-aae3-6141a01a66f1
Spec:
  Database Secret:
    Secret Name:  px-init-script-auth
  Init:
    Script Source:
      Config Map:
        Name:  px-init-script
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Readiness Probe:
        Exec:
          Command:
            bash
            -c
            export MYSQL_PWD="${MYSQL_ROOT_PASSWORD}"
ping_resp=$(mysqladmin -uroot ping)
if [[ "$ping_resp" != "mysqld is alive" ]]; then
    echo "[ERROR] server is not ready. PING_RESPONSE: $ping_resp"
    exit 1
fi

        Initial Delay Seconds:  30
        Period Seconds:         10
      Resources:
  Replicas:  1
  Service Template:
    Metadata:
    Spec:
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         50Mi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    DoNotTerminate
  Version:  5.7
Status:
  Observed Generation:  2
  Phase:                Running
Events:
  Type    Reason      Age    From                    Message
  ----    ------      ----   ----                    -------
  Normal  Successful  4m50s  PerconaXtraDB operator  Successfully created Service
  Normal  Successful  4m16s  PerconaXtraDB operator  Successfully created StatefulSet demo/px-init-script
  Normal  Successful  4m16s  PerconaXtraDB operator  Successfully created PerconaXtraDB
  Normal  Successful  4m16s  PerconaXtraDB operator  Successfully created appbinding
  Normal  Successful  4m16s  PerconaXtraDB operator  Successfully patched StatefulSet demo/px-init-script
  Normal  Successful  4m16s  PerconaXtraDB operator  Successfully patched PerconaXtraDB

$ kubectl get statefulset -n demo
NAME             READY   AGE
px-init-script   1/1     5m28s

$ kubectl get pvc -n demo
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-px-init-script-0   Bound    pvc-5188bbf9-6a71-4f00-a27c-d9590d7c71f4   50Mi       RWO            standard       19m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                              STORAGECLASS   REASON   AGE
pvc-5188bbf9-6a71-4f00-a27c-d9590d7c71f4   50Mi       RWO            Delete           Bound    demo/data-px-init-script-0         standard                19m

$ kubectl get service -n demo
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
px-init-script       ClusterIP   10.97.113.212   <none>        3306/TCP   6m47s
px-init-script-gvr   ClusterIP   None            <none>        3306/TCP   6m47s
```

KubeDB operator sets the `.status.phase` to `Running` once the database is successfully created. Run the following command to see the modified PerconaXtraDB object:

```bash
$ kubectl get px -n demo px-init-script -o yaml
```

Output:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"PerconaXtraDB","metadata":{"annotations":{},"name":"px-init-script","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"px-init-script"}}},"replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"50Mi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"DoNotTerminate","updateStrategy":{"type":"RollingUpdate"},"version":"5.7"}}
  creationTimestamp: "2020-01-09T13:45:43Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: px-init-script
  namespace: demo
  resourceVersion: "64559"
  selfLink: /apis/kubedb.com/v1alpha2/namespaces/demo/perconaxtradbs/px-init-script
  uid: d7ad081e-8b2d-41d1-aae3-6141a01a66f1
spec:
  authSecret:
    name: px-init-script-auth
  init:
    script:
      configMap:
        name: px-init-script
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      readinessProbe:
        exec:
          command:
          - bash
          - -c
          - |
            export MYSQL_PWD="${MYSQL_ROOT_PASSWORD}"
            ping_resp=$(mysqladmin -uroot ping)
            if [[ "$ping_resp" != "mysqld is alive" ]]; then
                echo "[ERROR] server is not ready. PING_RESPONSE: $ping_resp"
                exit 1
            fi
        initialDelaySeconds: 30
        periodSeconds: 10
      resources: {}
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: DoNotTerminate
  version: "5.7"
status:
  observedGeneration: 2
  phase: Running
```

## Connect with PerconaXtraDB database

KubeDB operator has created a new Secret called `px-init-script-auth` *(format: {percona-xtradb-object-name}-auth)* for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **"username"** for `mysql` superuser and a `password` key which contains the **"password"** for the superuser.

If you want to use an existing secret please specify that when creating the PerconaXtraDB object using `.spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys (`username` and `password`) in `.data` section and also make sure of using `root` as value of `username` key. For more details see [here](/docs/guides/percona-xtradb/concepts/percona-xtradb.md#specdatabasesecret).

Now, you can connect to this database using the database pod IP and and `root` user password.

```bash
$ kubectl get pods px-init-script-0 -n demo -o yaml | grep podIP
  podIP: 10.244.2.52

$ kubectl get secrets -n demo px-init-script-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secrets -n demo px-init-script-auth -o jsonpath='{.data.password}' | base64 -d
B0BMhl1ECz1C0uIN
```

To connect you just need to specify the host name for the database we created (either PodIP or the fully-qualified-domain-name for that Pod using the governing service named <percona-xtradb-object-name>-gvr or the fully-qualified-domain-name for the database Service with matching name of PerconaXtraDB object) by --host flag.

> Do not worry about the warning messages in the following output. Those are coming for providing a password on the command line

```bash
# connect to the server
$ kubectl exec -it -n demo px-init-script-0 -- mysql -u root --password=B0BMhl1ECz1C0uIN --host=px-init-script.demo.svc -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# read
$ kubectl exec -it -n demo px-init-script-0 -- mysql -u root --password=B0BMhl1ECz1C0uIN --host=px-init-script.demo.svc -e "SHOW databases;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
+--------------------+

$ kubectl exec -it -n demo px-init-script-0 -- mysql -u root --password=B0BMhl1ECz1C0uIN --host=px-init-script.demo.svc -e "SHOW tables IN mysql;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---------------------------+
| Tables_in_mysql           |
+---------------------------+
| ...                       |
| kubedb_table              |
| ...                       |
+---------------------------+

$ kubectl exec -it -n demo px-init-script-0 -- mysql -u root --password=B0BMhl1ECz1C0uIN --host=px-init-script.demo.svc -e "SELECT * FROM mysql.kubedb_table;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+
| id | name  |
+----+-------+
|  1 | name1 |
|  2 | name2 |
|  3 | name3 |
+----+-------+
```

As you can see here, the initial script has successfully created a table named `kubedb_table` in `mysql` database and inserted three rows of data into that table successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo px/px-init-script -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo px/px-init-script

kubectl delete ns demo
```

## Next Steps

- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/percona-xtradb/monitoring/using-prometheus-operator.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-config-file.md).
- How to use [custom rbac resource](/docs/guides/percona-xtradb/custom-rbac/using-custom-rbac.md) for PerconaXtraDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/backup/stash.md).
- Detail concepts of [PerconaXtraDB object](/docs/guides/percona-xtradb/concepts/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/guides/percona-xtradb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
