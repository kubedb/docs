---
title: GitOps PostgreSQL
menu:
  docs_{{ .version }}:
    identifier: pg-using-gitops
    name: GitOps PostgreSQL
    parent: pg-gitops-postgres
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Postgres using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create postgres database and manage updates using GitOps workflow.

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
> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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

## Create Postgres Database using GitOps

### Create a Postgres GitOps CR
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 3
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: postgres
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
    └── postgres.yaml
1 directories, 1 files
```

Now commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is created in your cluster.

Our `gitops` operator will create an actual `Postgres` database CR in the cluster. List the resources created by `gitops` operator in the `demo` namespace.


```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   2m11s

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    2m11s
```

List the resources created by `kubedb` operator created for `kubedb.com/v1` Postgres.

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=ha-postgres'
NAME                                       AGE
petset.apps.k8s.appscode.com/ha-postgres   3m26s

NAME                READY   STATUS    RESTARTS   AGE
pod/ha-postgres-0   2/2     Running   0          3m26s
pod/ha-postgres-1   2/2     Running   0          3m8s
pod/ha-postgres-2   2/2     Running   0          2m50s

NAME                      TYPE                       DATA   AGE
secret/ha-postgres-auth   kubernetes.io/basic-auth   2      3m29s

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/ha-postgres           ClusterIP   10.43.169.122   <none>        5432/TCP,2379/TCP            3m29s
service/ha-postgres-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   3m29s
service/ha-postgres-standby   ClusterIP   10.43.106.75    <none>        5432/TCP                     3m29s

NAME                                             TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/ha-postgres   kubedb.com/postgres   16.6      3m26s
```

## Update Postgres Database using GitOps

### Scale Postgres Database Resources

Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 3
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: postgres
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

Resource Requests and Limits are updated to `700m` CPU and `2Gi` Memory. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.

Now, `gitops` operator will detect the resource changes and create a `PostgresOpsRequest` to update the `Postgres` database. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   13m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    13m

NAME                                                                   TYPE              STATUS        AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l   VerticalScaling   Progressing   2s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the one of the pod,
```bash
$ kubectl get pod -n demo ha-postgres-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "2Gi"
  }
}
```

### Scale Postgres Replicas
Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 5
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: postgres
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

Update the `replicas` to `5`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.
Now, `gitops` operator will detect the replica changes and create a `HorizontalScaling` PostgresOpsRequest to update the `Postgres` database replicas. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   21m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    21m

NAME                                                                     TYPE                STATUS        AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-horizontalscaling-wvxu5x   HorizontalScaling   Progressing   6s
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l     VerticalScaling     Successful    7m54s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the number of pods,
```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=ha-postgres'
NAME            READY   STATUS    RESTARTS   AGE
ha-postgres-0   2/2     Running   0          9m4s
ha-postgres-1   2/2     Running   0          10m
ha-postgres-2   2/2     Running   0          9m44s
ha-postgres-3   2/2     Running   0          2m58s
ha-postgres-4   2/2     Running   0          2m23s
```

We can also scale down the replicas by updating the `replicas` fields.

### Exapand Postgres Volume

Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  replicas: 5
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: postgres
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

Update the `storage.resources.requests.storage` to `10Gi`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.

Now, `gitops` operator will detect the volume changes and create a `VolumeExpansion` PostgresOpsRequest to update the `Postgres` database volume. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   27m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    27m

NAME                                                                     TYPE                STATUS       AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-horizontalscaling-wvxu5x   HorizontalScaling   Successful   6m
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l     VerticalScaling     Successful   13m
postgresopsrequest.ops.kubedb.com/ha-postgres-volumeexpansion-2j5x5g     VolumeExpansion     Progressing  2s
```

After Ops Request becomes `Successful`, We can validate the changes by checking the pvc size,
```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=ha-postgres'
NAME                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-ha-postgres-0   Bound    pvc-061f3622-234f-4f91-b4d1-b81aa8739503   10Gi       RWO            standard       <unset>                 30m
data-ha-postgres-1   Bound    pvc-045fc563-fb4e-416c-a9c2-b20c96532978   10Gi       RWO            standard       <unset>                 30m
data-ha-postgres-2   Bound    pvc-a0f1d8fd-a677-4407-80b1-104b9f7b4cd1   10Gi       RWO            standard       <unset>                 30m
data-ha-postgres-3   Bound    pvc-060b6fab-0c2d-4935-b31b-2866be68dd6f   10Gi       RWO            standard       <unset>                 8m58s
data-ha-postgres-4   Bound    pvc-8149b579-a40f-4cd8-ac37-6a2401fd7807   10Gi       RWO            standard       <unset>                 8m23s
```

## Reconfigure Postgres

At first, we will create `user.conf` file containing required configuration settings.
To know more about this configuration file, check [here](/docs/guides/postgres/configuration/using-config-file.md)
```ini
$ cat user.conf
max_connections=200
shared_buffers=256MB
```

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo pg-configuration --from-file=./user.conf
secret/pg-configuration created
```

Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  configSecret:
    name: pg-configuration
  replicas: 5
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: postgres
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

Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.

Now, `gitops` operator will detect the configuration changes and create a `Reconfigure` PostgresOpsRequest to update the `Postgres` database configuration. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   36m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    36m

NAME                                                                     TYPE                STATUS        AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-horizontalscaling-wvxu5x   HorizontalScaling   Successful    15m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfigure-i4r23j         Reconfigure         Progressing   1s
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l     VerticalScaling     Successful    23m
```

After Ops Request becomes `Succesful`, lets check these parameters,

```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql
psql (16.1)
Type "help" for help.

postgres=# show max_connections;
 max_connections 
-----------------
 200
(1 row)

postgres=# show shared_buffers;
 shared_buffers 
----------------
 256MB
(1 row)
```
You can check the other pods same way.
So we have configured custom parameters.

> We can also reconfigure the parameters creating another secret and reference the secret in the `configSecret` field. Also you can remove the `configSecret` field to use the default parameters.

### Rotate Postgres Auth

To do that, create a `kubernetes.io/basic-auth` type k8s secret with the new username and password.
```bash
$ kubectl create secret generic -n demo pg-rotate-auth --type=kubernetes.io/basic-auth --from-literal=username=postgres --from-literal=password=pgpassword
```

Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  authSecret:
    name: pg-rotate-auth
  configSecret:
    name: pg-configuration
  replicas: 5
  version: "16.6"
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: postgres
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

Change the `authSecret` field to `pg-rotate-auth`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.

Now, `gitops` operator will detect the auth changes and create a `RotateAuth` PostgresOpsRequest to update the `Postgres` database auth. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   44m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    44m

NAME                                                                     TYPE                STATUS        AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-horizontalscaling-wvxu5x   HorizontalScaling   Successful    22m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfigure-i4r23j         Reconfigure         Successful    7m25s
postgresopsrequest.ops.kubedb.com/ha-postgres-rotate-auth-zot83x         RotateAuth          Progressing   2s
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l     VerticalScaling     Successful    30m
```

After Ops Request becomes `Successful`, We can validate the changes connecting postgres with new credentials.
```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql -U postgres -W
Password: <new-password>
psql (16.6)
Type "help" for help.

postgres=# 
```

### TLS configuration

We can add, rotate or remove TLS configuration using `gitops`.

To add tls, we are going to create an example `Issuer` that will be used to enable SSL/TLS in Postgres. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls postgres-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/postgres-ca created
```

Now, Let's create an `Issuer` using the `postgres-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: pg-issuer
  namespace: demo
spec:
  ca:
    secretName: postgres-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/reconfigure-tls/issuer.yaml
issuer.cert-manager.io/pg-issuer created
```

```bash
$ kubectl get issuer -n demo
NAME        READY   AGE
pg-issuer   True    11s
```
Issuer is ready(true).

Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  authSecret:
    name: pg-rotate-auth
  configSecret:
    name: pg-configuration
  replicas: 5
  version: "16.6"
  storageType: Durable
  sslMode: verify-full
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pg-issuer
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
      - name: postgres
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

Add `sslMode` and `tls` fields in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.

Now, `gitops` operator will detect the tls changes and create a `ReconfigureTLS` PostgresOpsRequest to update the `Postgres` database tls. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   3h17m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    3h17m

NAME                                                                     TYPE                STATUS        AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-horizontalscaling-wvxu5x   HorizontalScaling   Successful    176m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfigure-i4r23j         Reconfigure         Successful    161m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfiguretls-91fseg      ReconfigureTLS      Progressing   4s
postgresopsrequest.ops.kubedb.com/ha-postgres-rotate-auth-zot83x         RotateAuth          Successful    153m
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l     VerticalScaling     Successful    3h4m
```

After Ops Request becomes `Successful`, We can validate the changes connecting postgres with new credentials.
```bash
$ kubectl exec -it -n demo ha-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
ha-postgres-0:/$ psql -h ha-postgres.demo.svc -U postgres -d "sslmode=verify-full sslrootcert=/tls/certs/client/ca.crt sslcert=/tls/certs/client/client.crt sslkey=/tls/certs/client/client.key"
psql (13.13)
SSL connection (protocol: TLSv1.3, cipher: TLS_AES_256_GCM_SHA384, bits: 256, compression: off)
Type "help" for help.

postgres=# 
```

> We can also rotate the certificates updating `.spec.tls.certificates` field. Also you can remove the `.spec.tls` field to remove tls for postgres.

### Update Version

List postgres versions using `kubectl get postgresversion` and choose desired version that is compatible for upgrade from current version. Check the version constraints and ops request [here](/docs/guides/postgres/update-version/versionupgrading/index.md).

Let's choose `17.4` in this example.

Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  authSecret:
    name: pg-rotate-auth
  configSecret:
    name: pg-configuration
  replicas: 5
  version: "17.4"
  storageType: Durable
  sslMode: verify-full
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pg-issuer
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
      - name: postgres
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

Update the `version` field to `17.4`. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.

Now, `gitops` operator will detect the version changes and create a `VersionUpdate` PostgresOpsRequest to update the `Postgres` database version. List the resources created by `gitops` operator in the `demo` namespace.

```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   3h25m

NAME                              VERSION   STATUS   AGE
postgres.kubedb.com/ha-postgres   16.6      Ready    3h25m

NAME                                                                     TYPE                STATUS        AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-horizontalscaling-wvxu5x   HorizontalScaling   Successful    3h3m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfigure-i4r23j         Reconfigure         Successful    168m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfiguretls-91fseg      ReconfigureTLS      Successful    7m33s
postgresopsrequest.ops.kubedb.com/ha-postgres-rotate-auth-zot83x         RotateAuth          Successful    161m
postgresopsrequest.ops.kubedb.com/ha-postgres-versionupdate-1wxgt9       UpdateVersion       Progressing   4s
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l     VerticalScaling     Successful    3h11m
```


Now, we are going to verify whether the `Postgres`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get postgres -n demo ha-postgres -o=jsonpath='{.spec.version}{"\n"}'
17.4

$ kubectl get petset -n demo ha-postgres -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/postgres:17.4-alpine

$ kubectl get pod -n demo ha-postgres-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/postgres:17.4-alpine
```

### Enable Monitoring

If you already don't have a Prometheus server running, deploy one following tutorial from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/operator/README.md#deploy-prometheus-server).

Update the `postgres.yaml` with the following, 
```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Postgres
metadata:
  name: ha-postgres
  namespace: demo
spec:
  authSecret:
    name: pg-rotate-auth
  configSecret:
    name: pg-configuration
  replicas: 5
  version: "17.4"
  storageType: Durable
  sslMode: verify-full
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: pg-issuer
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
      - name: postgres
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

Add `monitor` field in the spec. Commit the changes and push to your Git repository. Your repository is synced with `ArgoCD` and the `Postgres` CR is updated in your cluster.

Now, `gitops` operator will detect the monitoring changes and create a `Restart` PostgresOpsRequest to add the `Postgres` database monitoring. List the resources created by `gitops` operator in the `demo` namespace.
```bash
$ kubectl get postgreses.gitops.kubedb.com,postgreses.kubedb.com,postgresopsrequest -n demo
NAME                                     AGE
postgres.gitops.kubedb.com/ha-postgres   3h34m

NAME                              VERSION   STATUS     AGE
postgres.kubedb.com/ha-postgres   16.6      NotReady   3h34m

NAME                                                                     TYPE                STATUS       AGE
postgresopsrequest.ops.kubedb.com/ha-postgres-horizontalscaling-wvxu5x   HorizontalScaling   Successful   3h13m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfigure-i4r23j         Reconfigure         Successful   177m
postgresopsrequest.ops.kubedb.com/ha-postgres-reconfiguretls-91fseg      ReconfigureTLS      Successful   16m
postgresopsrequest.ops.kubedb.com/ha-postgres-restart-nhjk9u             Restart             Progressing  2s
postgresopsrequest.ops.kubedb.com/ha-postgres-rotate-auth-zot83x         RotateAuth          Successful   170m
postgresopsrequest.ops.kubedb.com/ha-postgres-versionupdate-1wxgt9       UpdateVersion       Successful   9m30s
postgresopsrequest.ops.kubedb.com/ha-postgres-verticalscaling-i0kr1l     VerticalScaling     Successful   3h21m
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

```bash

## Next Steps

- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
