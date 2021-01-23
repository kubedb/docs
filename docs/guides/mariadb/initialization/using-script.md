---
title: Initialize MariaDB using Script
menu:
  docs_{{ .version }}:
    identifier: my-using-script-initialization
    name: Using Script
    parent: my-initialization-mariadb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize MariaDB using Script

This tutorial will show you how to use KubeDB to initialize a MariaDB database with \*.sql, \*.sh and/or \*.sql.gz script.
In this tutorial we will use .sql script stored in GitHub repository [kubedb/mariadb-init-scripts](https://github.com/kubedb/mariadb-init-scripts).

> Note: The yaml files that are used in this tutorial are stored in [docs/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. This tutorial will also use a phpMyAdmin to connect and test MariaDB database, once it is running. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  
  $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/quickstart/demo-1.yaml
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

MariaDB supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use `init.sql` script from [mariadb-init-scripts](https://github.com/kubedb/mariadb-init-scripts) git repository to create a TABLE `kubedb_table` in `mariadb` database.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from `init.sql` file. Then, we will provide this ConfigMap as script source in `init.script` of MariaDB crd spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo my-init-script \
--from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/mariadb-init-scripts/raw/master/init.sql)"
configmap/my-init-script created
```

## Create a MariaDB database with Init-Script

Below is the `MariaDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: mariadb-init-script
  namespace: demo
spec:
  version: "8.0.21"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/Initialization/demo-1.yaml
mariadb.kubedb.com/mariadb-init-script created
```

Here,

- `spec.init.script` specifies a script source used to initialize the database before database server starts. The scripts will be executed alphabatically. In this tutorial, a sample .sql script from the git repository `https://github.com/kubedb/mariadb-init-scripts.git` is used to create a test database. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) instead of `ConfigMap`.  The \*.sql, \*sql.gz and/or \*.sh sripts that are stored inside the root folder will be executed alphabatically. The scripts inside child folders will be skipped.

KubeDB operator watches for `MariaDB` objects using Kubernetes api. When a `MariaDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MariaDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present. No MariaDB specific RBAC roles are required for [RBAC enabled clusters](/docs/setup/README.md#using-yaml).

```bash
$ kubectl dba describe my -n demo mariadb-init-scrip
Name:               mariadb-init-script
Namespace:          demo
CreationTimestamp:  Thu, 27 Aug 2020 12:42:04 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"mariadb-init-script","namespace":"demo"},"spec":{"init":{"scriptS...
Replicas:           1  total
Status:             Running
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Halted:              false
Halted:              false
Termination Policy:  Delete

StatefulSet:          
  Name:               mariadb-init-script
  CreationTimestamp:  Thu, 27 Aug 2020 12:42:04 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mariadbs.kubedb.com
                        app.kubernetes.io/instance=mariadb-init-script
  Annotations:        <none>
  Replicas:           824637371096 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mariadb-init-script
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mariadbs.kubedb.com
                  app.kubernetes.io/instance=mariadb-init-script
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.103.202.117
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.2.9:3306

Service:        
  Name:         mariadb-init-script-gvr
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mariadbs.kubedb.com
                  app.kubernetes.io/instance=mariadb-init-script
  Annotations:    service.alpha.kubernetes.io/tolerate-unready-endpoints=true
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   3306/TCP
  Endpoints:    10.244.2.9:3306

Database Secret:
  Name:         mariadb-init-script-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mariadbs.kubedb.com
                  app.kubernetes.io/instance=mariadb-init-script
  Annotations:  <none>
  Type:         Opaque
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
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"mariadb-init-script","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"my-init-script"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"8.0.21"}}

    Creation Timestamp:  2020-08-27T06:43:15Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mariadb-init-script
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mariadb
      app.kubernetes.io/version:     8.0.21
      app.kubernetes.io/name:        mariadbs.kubedb.com
      app.kubernetes.io/instance:               mariadb-init-script
    Name:                            mariadb-init-script
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mariadb-init-script
        Path:    /
        Port:    3306
        Scheme:  mariadb
      URL:       tcp(mariadb-init-script:3306)/
    Secret:
      Name:   mariadb-init-script-auth
    Type:     kubedb.com/mariadb
    Version:  8.0.21

Events:
  Type    Reason      Age   From            Message
  ----    ------      ----  ----            -------
  Normal  Successful  1m    MariaDB operator  Successfully created Service
  Normal  Successful  7s    MariaDB operator  Successfully created StatefulSet
  Normal  Successful  7s    MariaDB operator  Successfully created MariaDB
  Normal  Successful  7s    MariaDB operator  Successfully created appbinding

$ kubectl get statefulset -n demo
NAME                READY   AGE
mariadb-init-script   1/1     2m24s

$ kubectl get pvc -n demo
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mariadb-init-script-0   Bound    pvc-32a59975-2972-4122-9635-22fe19483145   1Gi        RWO            standard       3m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   REASON   AGE
pvc-32a59975-2972-4122-9635-22fe19483145   1Gi        RWO            Delete           Bound    demo/data-mariadb-init-script-0   standard                3m25s

$ kubectl get service -n demo
NAME                    TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
myadmin                 LoadBalancer   10.104.142.213   <pending>     80:31529/TCP   23m
mariadb-init-script       ClusterIP      10.103.202.117   <none>        3306/TCP       3m49s
mariadb-init-script-gvr   ClusterIP      None             <none>        3306/TCP       3m49s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified MariaDB object:

```yaml
$ kubectl get my -n demo mariadb-init-script -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MariaDB","metadata":{"annotations":{},"name":"mariadb-init-script","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"my-init-script"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"8.0.21"}}
  creationTimestamp: "2020-08-27T06:42:04Z"
  finalizers:
  - kubedb.com
  generation: 2
    operation: Update
    time: "2020-08-27T06:43:15Z"
  name: mariadb-init-script
  namespace: demo
  resourceVersion: "11901"
  selfLink: /apis/kubedb.com/v1alpha2/namespaces/demo/mariadbs/mariadb-init-script
  uid: 2903f636-09cb-4299-af5d-c7a2e799ec61
spec:
  authSecret:
    name: mariadb-init-script-auth
  init:
    script:
      configMap:
        name: my-init-script
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
      serviceAccountName: mariadb-init-script
  replicas: 1
  serviceTemplate:
    metadata: {}
    spec: {}
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: Delete
  version: 8.0.21
status:
  observedGeneration: 2
  phase: Running
```

KubeDB operator has created a new Secret called `mariadb-init-script-auth` *(format: {mariadb-object-name}-auth)* for storing the password for MariaDB superuser. This secret contains a `username` key which contains the *username* for MariaDB superuser and a `password` key which contains the *password* for MariaDB superuser.
If you want to use an existing secret please specify that when creating the MariaDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`.

Now, you can connect to this database from the phpMyAdmin dashboard using the database pod IP and and `mariadb` user password.

```bash
$ kubectl get pods mariadb-init-script-0 -n demo -o yaml | grep IP
  hostIP: 10.0.2.15
  podIP: 10.244.2.9

$ kubectl get secrets -n demo mariadb-init-script-auth -o jsonpath='{.data.\user}' | base64 -d
root

$ kubectl get secrets -n demo mariadb-init-script-auth -o jsonpath='{.data.\password}' | base64 -d
1Pc7bwSygrv1MX1Q
```

---
Note: In MariaDB: `8.0.14-v1` connection to phpMyAdmin may give error as it is using `caching_sha2_password` and `sha256_password` authentication plugins over `mariadb_native_password`. If the error happens do the following for work around. But, It's not recommended to change authentication plugins. See [here](https://stackoverflow.com/questions/49948350/phpmyadmin-on-mariadb-8-0) for alternative solutions.

```bash
kubectl exec -it -n demo mariadb-quickstart-0 -- mariadb -u root --password=1Pc7bwSygrv1MX1Q -e "ALTER USER root IDENTIFIED WITH mariadb_native_password BY '1Pc7bwSygrv1MX1Q';"
```

---

Now, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. To log into the phpMyAdmin, use host __`10.244.2.9`__ , username __`root`__ and password __`1Pc7bwSygrv1MX1Q`__.

As you can see here, the initial script has successfully created a table named `kubedb_table` in `mariadb` database and inserted three rows of data into that table successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mariadb/mariadb-init-script -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mariadb/mariadb-init-script

kubectl delete ns demo
```

## Next Steps

- Initialize [MariaDB with Script](/docs/guides/mariadb/initialization/using-script.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mariadb/monitoring/using-prometheus-operator.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mariadb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mariadb/private-registry/using-private-registry.md) to deploy MariaDB with KubeDB.
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb.md).
- Detail concepts of [MariaDBVersion object](/docs/guides/mariadb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
