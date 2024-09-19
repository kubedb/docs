---
title: Initialize MySQL using Script
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-initialization
    name: Initialization Using Script
    parent: guides-mysql
    weight: 41
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize MySQL using Script

This tutorial will show you how to use KubeDB to initialize a MySQL database with \*.sql, \*.sh and/or \*.sql.gz script.
In this tutorial we will use .sql script stored in GitHub repository [kubedb/mysql-init-scripts](https://github.com/kubedb/mysql-init-scripts).

> Note: The yaml files that are used in this tutorial are stored in [docs/guides/mysql/initialization/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/initialization/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. This tutorial will also use a phpMyAdmin to connect and test MySQL database, once it is running. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  
  $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/initialization/yamls/phpmyadmin.yaml
  deployment.extensions/myadmin created
  service/myadmin created
  
  $ kubectl get pods -n demo 
  NAME                       READY   STATUS    RESTARTS   AGE
  myadmin-66cc8d4c77-wkwht   1/1     Running   0          5m20s
  
  $ kubectl get service -n demo
  NAME      TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
  myadmin   LoadBalancer   10.104.142.213   <pending>     80:31529/TCP     3m14s
  ```

  Then, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. For kind cluster, you can get this URL by running the following command:

  ```bash
  $ kubectl get svc -n demo myadmin -o json | jq '.spec.ports[].nodePort'
  31529
  
  $ kubectl get node -o json | jq '.items[].status.addresses[].address'
  "172.18.0.3"
  "kind-control-plane"
  "172.18.0.4"
  "kind-worker"
  "172.18.0.2"
  "kind-worker2"
  
  # expected url will be:
  url: http://172.18.0.4:31529
  ```
  
## Prepare Initialization Scripts

MySQL supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use `init.sql` script from [mysql-init-scripts](https://github.com/kubedb/mysql-init-scripts) git repository to create a TABLE `kubedb_table` in `mysql` database.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from `init.sql` file. Then, we will provide this ConfigMap as script source in `init.script` of MySQL crd spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo my-init-script \
--from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/mysql-init-scripts/raw/master/init.sql)"
configmap/my-init-script created
```

## Create a MySQL database with Init-Script

Below is the `MySQL` object created in this tutorial.

<ul class="nav nav-tabs" id="definationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active  " id="st-tab" data-toggle="tab" href="#standAlone" role="tab" aria-controls="standAlone" aria-selected="true">Stand Alone</a>
  </li>

  <li class="nav-item">
    <a class="nav-link active" id="gr-tab" data-toggle="tab" href="#groupReplication" role="tab" aria-controls="groupReplication" aria-selected="false">Group Replication</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="ic-tab" data-toggle="tab" href="#innodbCluster" role="tab" aria-controls="innodbCluster" aria-selected="false">Innodb Cluster</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="sc-tab" data-toggle="tab" href="#semisync" role="tab" aria-controls="semisync" aria-selected="false">Semi sync </a>
  </li>

</ul>


<div class="tab-content" id="definationTabContent">
  <div class="tab-pane fade show active" id="groupReplication" role="tabpanel" aria-labelledby="gr-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-init-script
  namespace: demo
spec:
  version: "8.0.35"
  topology:
    mode: GroupReplication
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: my-init-script

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/initialization/yamls/initialize-gr.yaml
mysql.kubedb.com/mysql-init-script created
```

  </div>

  <div class="tab-pane fade" id="innodbCluster" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-init-script
  namespace: demo
spec:
  version: "8.0.31-innodb"
  replicas: 3
  topology:
    mode: InnoDBCluster
    innoDBCluster:
      router:
        replicas: 1
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: my-init-script

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/initialization/yamls/initialize-innodb.yaml
mysql.kubedb.com/mysql-init-script created
```
  </div>

  <div class="tab-pane fade " id="semisync" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-init-script
  namespace: demo
spec:
  version: "8.0.35"
  replicas: 3
  topology:
    mode: SemiSync
    semiSync:
      sourceWaitForReplicaCount: 1
      sourceTimeout: 23h
      errantTransactionRecoveryPolicy: PseudoTransaction
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: my-init-script

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/initialization/yamls/initialize-mysql.yaml
mysql.kubedb.com/mysql-init-script created
```

  </div>

  <div class="tab-pane fade" id="standAlone" role="tabpanel" aria-labelledby="st-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-init-script
  namespace: demo
spec:
  version: "8.0.35"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: my-init-script
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/initialization/yamls/initialize-mysql.yaml
mysql.kubedb.com/mysql-init-script created
```
  </div>

</div>


Here,

- `spec.init.script` specifies a script source used to initialize the database before database server starts. The scripts will be executed alphabetically. In this tutorial, a sample .sql script from the git repository `https://github.com/kubedb/mysql-init-scripts.git` is used to create a test database. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) instead of `ConfigMap`.  The \*.sql, \*sql.gz and/or \*.sh sripts that are stored inside the root folder will be executed alphabetically. The scripts inside child folders will be skipped.

KubeDB operator watches for `MySQL` objects using Kubernetes api. When a `MySQL` object is created, KubeDB operator will create a new PetSet and a Service with the matching MySQL object name. KubeDB operator will also create a governing service for PetSets with the name `kubedb`, if one is not already present. No MySQL specific RBAC roles are required for [RBAC enabled clusters](/docs/setup/README.md#using-yaml).

```bash
$ kubectl dba describe my -n demo mysql-init-scrip
Name:               mysql-init-script
Namespace:          demo
CreationTimestamp:  Thu, 30 Jun 2022 12:21:15 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-init-script","namespace":"demo"},"spec":{"init":{"script"...
Replicas:           1  total
Status:             Provisioning
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  Delete

PetSet:          
  Name:               mysql-init-script
  CreationTimestamp:  Thu, 30 Jun 2022 12:21:15 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mysql-init-script
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:        <none>
  Replicas:           824644789336 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mysql-init-script
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-init-script
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.198.184
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.23:3306

Service:        
  Name:         mysql-init-script-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-init-script
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.23:3306

Auth Secret:
  Name:         mysql-init-script-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql-init-script
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

Init:
  Script Source:
    Volume:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      my-init-script
    Optional:  false

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-init-script","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"my-init-script"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"8.0.35"}}

    Creation Timestamp:  2022-06-30T06:21:15Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mysql-init-script
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysqls.kubedb.com
    Name:                            mysql-init-script
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mysql-init-script
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(mysql-init-script.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  mysql-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  mysql-restore-8.0.21
    Secret:
      Name:   mysql-init-script-auth
    Type:     kubedb.com/mysql
    Version:  8.0.35

Events:
  Type     Reason      Age   From               Message
  ----     ------      ----  ----               -------
  Normal   Successful  10s   KubeDB operator  Successfully created governing service
  Normal   Successful  10s   KubeDB operator  Successfully created service for primary/standalone
  Normal   Successful  10s   KubeDB operator  Successfully created database auth secret
  Normal   Successful  10s   KubeDB operator  Successfully created PetSet
  Normal   Successful  10s   KubeDB operator  Successfully created MySQL
  Normal   Successful  10s   KubeDB operator  Successfully created appbinding


$ kubectl get petset -n demo
NAME                READY   AGE
mysql-init-script   1/1     2m24s

$ kubectl get pvc -n demo
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mysql-init-script-0   Bound    pvc-32a59975-2972-4122-9635-22fe19483145   1Gi        RWO            standard       3m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   REASON   AGE
pvc-32a59975-2972-4122-9635-22fe19483145   1Gi        RWO            Delete           Bound    demo/data-mysql-init-script-0   standard                3m25s

$ kubectl get service -n demo
NAME                    TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
myadmin                 LoadBalancer   10.104.142.213   <pending>     80:31529/TCP   23m
mysql-init-script       ClusterIP      10.103.202.117   <none>        3306/TCP       3m49s
mysql-init-script-pods   ClusterIP      None             <none>        3306/TCP       3m49s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified MySQL object:

```yaml
$ kubectl get my -n demo mysql-init-script -o yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"mysql-init-script","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"my-init-script"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"8.0.35"}}
  creationTimestamp: "2022-06-30T06:21:15Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: mysql-init-script
  namespace: demo
  resourceVersion: "1697522"
  uid: 932c1fe3-6692-4ddc-b4cd-fe34e0d5ebc8
spec:
  allowedReadReplicas:
    namespaces:
      from: Same
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: mysql-init-script-auth
  init:
    initialized: true
    script:
      configMap:
        name: my-init-script
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      affinity:
      ...
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: mysql-init-script
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: Delete
  useAddressType: DNS
  version: 8.0.35
status:
  conditions:
    ...
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready

```

KubeDB operator has created a new Secret called `mysql-init-script-auth` *(format: {mysql-object-name}-auth)* for storing the password for MySQL superuser. This secret contains a `username` key which contains the *username* for MySQL superuser and a `password` key which contains the *password* for MySQL superuser.
If you want to use an existing secret please specify that when creating the MySQL object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`.

Now, you can connect to this database from the phpMyAdmin dashboard using the database pod IP and `mysql` user password.

```bash
$ kubectl get pods mysql-init-script-0 -n demo -o yaml | grep IP
  hostIP: 10.0.2.15
  podIP: 10.244.2.9

$ kubectl get secrets -n demo mysql-init-script-auth -o jsonpath='{.data.\user}' | base64 -d
root

$ kubectl get secrets -n demo mysql-init-script-auth -o jsonpath='{.data.\password}' | base64 -d
1Pc7bwSygrv1MX1Q
```

---
Note: In MySQL: `8.0.14-v1` connection to phpMyAdmin may give error as it is using `caching_sha2_password` and `sha256_password` authentication plugins over `mysql_native_password`. If the error happens do the following for work around. But, It's not recommended to change authentication plugins. See [here](https://stackoverflow.com/questions/49948350/phpmyadmin-on-mysql-8-0) for alternative solutions.

```bash
kubectl exec -it -n demo mysql-quickstart-0 -- mysql -u root --password=1Pc7bwSygrv1MX1Q -e "ALTER USER root IDENTIFIED WITH mysql_native_password BY '1Pc7bwSygrv1MX1Q';"
```

---

Now, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. To log into the phpMyAdmin, use host __`10.244.2.9`__ , username __`root`__ and password __`1Pc7bwSygrv1MX1Q`__.

As you can see here, the initial script has successfully created a table named `kubedb_table` in `mysql` database and inserted three rows of data into that table successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mysql/mysql-init-script -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mysql/mysql-init-script

kubectl delete ns demo
```

## Next Steps

- Initialize [MySQL with Script](/docs/guides/mysql/initialization/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/index.md) to deploy MySQL with KubeDB.
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
