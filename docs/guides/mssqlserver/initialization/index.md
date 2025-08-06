---
title: Initialize SQL Server using Script
menu:
  docs_{{ .version }}:
    identifier: ms-initialization
    name: Initialization Using Script
    parent: guides-mssqlserver
    weight: 41
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# Initialize Microsoft SQL Server using Script

This tutorial will show you how to use KubeDB to initialize a MSSQLServer database with `*.sql`, `*.sh` or `*.sql.gz` script.

In this tutorial, we will use .sql script stored in GitHub repository [kubedb/mssqlserver-init-scripts](https://github.com/kubedb/mssqlserver-init-scripts).

> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).


## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```
  
## Prepare Initialization Scripts

MSSQLServer supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use `init.sql` script from [mssqlserver-init-scripts](https://github.com/kubedb/mssqlserver-init-scripts) git repository to create a database named `mssql` and a table named `kubedb_init` in that database. 

We will use a ConfigMap as a script source. You can use any Kubernetes supported [volumes](https://kubernetes.io/docs/concepts/storage/volumes) as a script source.

At first, we will create a ConfigMap with `init.sql` file. Then, we will provide this ConfigMap as script source in `init.script` of the MSSQLServer CR spec.

Let's create a ConfigMap with the `init.sql` initialization script,

```bash
$ kubectl create configmap -n demo mssql-init-scripts \
--from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/mssqlserver-init-scripts/raw/master/init.sql)"
configmap/mssql-init-scripts created
```

## Create a MSSQLServer database with Init-Script

Below is the `MSSQLServer` object created in this tutorial.

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
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ms-init
  namespace: demo
spec:
  version: "2022-cu19"
  storageType: Durable
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
  init:
    script:
      configMap:
        name: mssql-init-scripts
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/initialization/yamls/initialize-gr.yaml
mssqlserver.kubedb.com/ms-init created
```
  </div>

  



  <div class="tab-pane fade" id="innodbCluster" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: ms-ag-init
  namespace: demo
spec:
  version: "2022-cu19"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb
        - mssql
      secondaryAccessMode: "All"
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
  init:
    script:
      configMap:
        name: mssql-init-scripts
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/initialization/yamls/initialize-innodb.yaml
mssqlsever.kubedb.com/ms-ag-init created
```
  </div>

</div>





Here,
- `spec.init.script` specifies a script source used to initialize the database. The scripts will be executed alphabetically. In this tutorial, a sample .sql script from the git repository `https://github.com/kubedb/mssqlserver-init-scripts.git` is used to create a test database named `mssql`. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) instead of `ConfigMap`.  The \*.sql, \*sql.gz and/or \*.sh scripts that are stored inside the folder will be executed alphabetically. The scripts inside child folders will be skipped.

KubeDB operator watches for `MSSQLServer` objects using Kubernetes API. When a `MSSQLServer` object is created, KubeDB operator will create a PetSet and Services, Secrets, and other necessary resouces for this `MSSQLServer` Database.

```bash
$ kubectl dba describe my -n demo MSSQLServer-init-scrip
Name:               ms-init
Namespace:          demo
CreationTimestamp:  Thu, 30 Jun 2022 12:21:15 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"MSSQLServer","metadata":{"annotations":{},"name":"ms-init","namespace":"demo"},"spec":{"init":{"script"...
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
  Name:               ms-init
  CreationTimestamp:  Thu, 30 Jun 2022 12:21:15 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=ms-init
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=MSSQLServers.kubedb.com
  Annotations:        <none>
  Replicas:           824644789336 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         ms-init
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=ms-init
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=MSSQLServers.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.198.184
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.23:3306

Service:        
  Name:         ms-init-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=ms-init
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=MSSQLServers.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.23:3306

Auth Secret:
  Name:         ms-init-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=ms-init
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=MSSQLServers.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

Init:
  Script Source:
    Volume:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      mssql-init-scripts
    Optional:  false

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"MSSQLServer","metadata":{"annotations":{},"name":"ms-init","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"mssql-init-scripts"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"9.1.0"}}

    Creation Timestamp:  2022-06-30T06:21:15Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    ms-init
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        MSSQLServers.kubedb.com
    Name:                            ms-init
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    ms-init
        Path:    /
        Port:    3306
        Scheme:  MSSQLServer
      URL:       tcp(ms-init.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  MSSQLServer-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  MSSQLServer-restore-8.0.21
    Secret:
      Name:   ms-init-auth
    Type:     kubedb.com/MSSQLServer
    Version:  9.1.0

Events:
  Type     Reason      Age   From               Message
  ----     ------      ----  ----               -------
  Normal   Successful  10s   KubeDB operator  Successfully created governing service
  Normal   Successful  10s   KubeDB operator  Successfully created service for primary/standalone
  Normal   Successful  10s   KubeDB operator  Successfully created database auth secret
  Normal   Successful  10s   KubeDB operator  Successfully created PetSet
  Normal   Successful  10s   KubeDB operator  Successfully created MSSQLServer
  Normal   Successful  10s   KubeDB operator  Successfully created appbinding


$ kubectl get petset -n demo
NAME                READY   AGE
ms-init   1/1     2m24s

$ kubectl get pvc -n demo
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-ms-init-0   Bound    pvc-32a59975-2972-4122-9635-22fe19483145   1Gi        RWO            standard       3m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   REASON   AGE
pvc-32a59975-2972-4122-9635-22fe19483145   1Gi        RWO            Delete           Bound    demo/data-ms-init-0   standard                3m25s

$ kubectl get service -n demo
NAME                    TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
myadmin                 LoadBalancer   10.104.142.213   <pending>     80:31529/TCP   23m
ms-init       ClusterIP      10.103.202.117   <none>        3306/TCP       3m49s
ms-init-pods   ClusterIP      None             <none>        3306/TCP       3m49s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified MSSQLServer object:

```yaml
$ kubectl get my -n demo ms-init -o yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MSSQLServer","metadata":{"annotations":{},"name":"ms-init","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"mssql-init-scripts"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"9.1.0"}}
  creationTimestamp: "2022-06-30T06:21:15Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: ms-init
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
    name: ms-init-auth
  init:
    initialized: true
    script:
      configMap:
        name: mssql-init-scripts
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: ms-init
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
  version: 9.1.0
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

KubeDB operator has created a new Secret called `ms-init-auth` *(format: {MSSQLServer-object-name}-auth)* for storing the password for MSSQLServer superuser. This secret contains a `username` key which contains the *username* for MSSQLServer superuser and a `password` key which contains the *password* for MSSQLServer superuser.
If you want to use an existing secret please specify that when creating the MSSQLServer object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`.

Now, you can connect to this database from the phpMyAdmin dashboard using the database pod IP and `MSSQLServer` user password.

```bash
$ kubectl get pods ms-init-0 -n demo -o yaml | grep IP
  hostIP: 10.0.2.15
  podIP: 10.244.2.9

$ kubectl get secrets -n demo ms-init-auth -o jsonpath='{.data.\user}' | base64 -d
root

$ kubectl get secrets -n demo ms-init-auth -o jsonpath='{.data.\password}' | base64 -d
1Pc7bwSygrv1MX1Q
```

---
Note: In MSSQLServer: `8.0.14-v1` connection to phpMyAdmin may give error as it is using `caching_sha2_password` and `sha256_password` authentication plugins over `MSSQLServer_native_password`. If the error happens do the following for work around. But, It's not recommended to change authentication plugins. See [here](https://stackoverflow.com/questions/49948350/phpmyadmin-on-MSSQLServer-8-0) for alternative solutions.

```bash
kubectl exec -it -n demo MSSQLServer-quickstart-0 -- MSSQLServer -u root --password=1Pc7bwSygrv1MX1Q -e "ALTER USER root IDENTIFIED WITH MSSQLServer_native_password BY '1Pc7bwSygrv1MX1Q';"
```

---

Now, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. To log into the phpMyAdmin, use host __`10.244.2.9`__ , username __`root`__ and password __`1Pc7bwSygrv1MX1Q`__.

As you can see here, the initial script has successfully created a table named `kubedb_table` in `MSSQLServer` database and inserted three rows of data into that table successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo MSSQLServer/ms-init -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo MSSQLServer/ms-init

kubectl delete ns demo
```

## Next Steps

- Initialize [MSSQLServer with Script](/docs/guides/mssqlserver/initialization/index.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mssqlserver/monitoring/prometheus-operator/index.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mssqlserver/monitoring/builtin-prometheus/index.md).
- Use [private Docker registry](/docs/guides/mssqlserver/private-registry/index.md) to deploy MSSQLServer with KubeDB.
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/database/index.md).
- Detail concepts of [MSSQLServerVersion object](/docs/guides/mssqlserver/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
