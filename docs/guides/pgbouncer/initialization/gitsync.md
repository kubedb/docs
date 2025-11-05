---
title: Initialize PgBouncer From Git Repository
menu:
  docs_{{ .version }}:
    identifier: guides-PgBouncer -gitsync
    name: Git Repository
    parent: pb-initialization-pgbouncer
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialization PgBouncer from a Git Repository
This guide demonstrates how to use KubeDB to initialize a PgBouncer database with initialization scripts (.sql, .sh, .js and/or .sql.gz) stored in a public or private Git repository.
To fetch the repository contents, KubeDB uses a sidecar container called [git-sync](https://github.com/kubernetes/git-sync).
In this example, we will initialize PgBouncer using a `.sql` script from the GitHub repository [kubedb/PgBouncer -init-scripts](https://github.com/kubedb/PgBouncer -init-scripts).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
    - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)
    - [Updating Overview](/docs/guides/pgbouncer/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgbouncer](/docs/examples/pgbouncer) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare PgBouncer

Now, we are going to deploy a `PgBouncer` with version `1.24.0`.

## From Public Git Repository

KubeDB implements a `PgBouncer ` Custom Resource Definition (CRD) to define the specification of a PgBouncer database.
To initialize the database from a public Git repository, you need to specify the required arguments for the `git-sync` sidecar container within the PgBouncer resource specification.
The following YAML manifest shows an example `PgBouncer ` object configured with `git-sync`:

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-demo
  namespace: demo
spec:
  version: "1.24.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "postgres-demo"
      namespace: demo
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
  init:
    script:
      scriptPath: "sync-test"
      git:
        args:
          - --repo=<desired repo>
          - --depth=1
          - --add-user=true
          - --period=60s
          - --one-time
        securityContext:
          runAsUser: 999
```
```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/PgBouncer /initialization/yamls/git-sync-public.yaml
PgBouncer .kubedb.com/sample-PgBouncer created
```
Here,

-`.spec.init.git.securityContext.runAsUser` the init container git_sync runs with user 999.
- `.spec.init.script.git.args` specifies the arguments for the `git-sync` container.
The `git-sync` container has one required flags:
- `--repo`  – specifies the remote Git repository to sync.

Here, the value of the `--link` argument must match the value of `spec.init.script.scriptPath`.
The `--link` argument creates a symlink that always points to the latest synced data.

> To know more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync).

Now, wait until `sample-PgBouncer ` has status `Ready`. i.e,

```bash
$ kubectl get PgBouncer -n demo 
NAME           VERSION   STATUS   AGE
sample-PgBouncer   9.1.0     Ready    42m
```

Next, we will connect to the PgBouncer database and verify the data inserted from the `*.sql` script stored in the Git repository.

```bash
$  kubectl get secrets -n demo sample-PgBouncer -auth -o jsonpath='{.data.\username}' | base64 -d
root⏎  
$ kubectl get secrets -n demo sample-PgBouncer -auth -o jsonpath='{.data.\password}' | base64 -d
lYeFGCL~mklFUN1b⏎ 
$  kubectl exec -it -n demo sample-PgBouncer -0 -- bash
Defaulted container "PgBouncer " out of: PgBouncer , PgBouncer -init (init), git-sync (init)
bash-5.1$ PgBouncer -uroot -p"lYeFGCL~mklFUN1b"  
PgBouncer : [Warning] Using a password on the command line interface can be insecure.
Welcome to the PgBouncer monitor.  Commands end with ; or \g.
Your PgBouncer connection id is 156
Server version: 9.1.0 PgBouncer Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PgBouncer > show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| PgBouncer              |
| performance_schema |
| sys                |
+--------------------+
5 rows in set (0.02 sec)

PgBouncer > use PgBouncer ;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed

PgBouncer > select * from kubedb_table;
+----+-------+
| id | name  |
+----+-------+
|  1 | name1 |
|  2 | name2 |
|  3 | name3 |
+----+-------+
3 rows in set (0.00 sec)
PgBouncer > select * from kubedb_table;
+----+-------+
| id | name  |
+----+-------+
|  1 | name1 |
|  2 | name2 |
|  3 | name3 |
+----+-------+
3 rows in set (0.00 sec)
```
## From Private Git Repository

### 1. Using SSH Key

Git-sync supports using SSH protocol for pulling git content.

First, Obtain the host keys for your git server:

```bash
$ ssh-keyscan $YOUR_GIT_HOST > /tmp/known_hosts
```

> `$YOUR_GIT_HOST` refers to the hostname of your Git server. <br>
> For example, if you are using GitHub, set the variable as follows:
> ```bash
> export YOUR_GIT_HOST=github.com
> ```

Use the `kubectl create secret` command to create a secret from your local SSH key and known hosts file.
This secret will be used by git-sync to authenticate with the Git repository.

>Here, we are using the default SSH key file located at `$HOME/.ssh/id_rsa`. If your SSH key is stored in a different location, please update the command accordingly. Also you can use any name instead of `git-creds` to create the secret.

```bash
$ kubectl create secret generic -n demo git-creds \
    --from-file=ssh=$HOME/.ssh/id_rsa \
    --from-file=known_hosts=/tmp/known_hosts
```

The following YAML manifest provides an example of a `PgBouncer ` resource configured to use `git-sync` with a private Git repository:

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer 
metadata:
  name: sample-PgBouncer 
  namespace: demo
spec:
  init:
    script:
      scriptPath: "current"
      git:
        args:
          # update with your private repository    
          - --repo=<private_git_repo_ssh_url>
          - --link=current
          - --root=/git
          # terminate after one successful sync
          - --one-time
        authSecret:
          # the name of the secret created above
          name: git-creds
        # run as git sync user 
        securityContext:
          runAsUser: 65533
  version: "9.1.0"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/PgBouncer /initialization/yamls/git-sync-ssh.yaml
PgBouncer .kubedb.com/sample-PgBouncer created
```

Here,
- `.spec.init.git.securityContext.runAsUser: 65533` ensure the container runs as the dedicated non-root `git-sync` user.
- `.spec.init.git.authSecret` specifies the secret containing the `SSH` key.

Once the database reaches the `Ready` state, you can verify the data using the method described above.

### 2. Using Username and Personal Access Token(PAT)

First, create a `Personal Access Token (PAT)` on your Git host server with the required permissions to access the repository.
Then create a Kubernetes secret using the `Personal Access Token (PAT)`:
> Here, you can use any key name instead of `git-pat` to store the token in the secret.
```bash
$ kubectl create secret generic -n demo git-pat \
    --from-literal=github-pat=<ghp_yourpersonalaccesstoken>
```

Now, create a `PgBouncer ` resource that references the secret created above.
The following YAML manifest shows an example:

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer 
metadata:
  name: sample-PgBouncer 
  namespace: demo
spec:
  init:
    script:
      scriptPath: "current"
      git:
        args:
          # update with your private repository    
          - --repo=<private_git_repo_http_url>
          - --link=current
          - --root=/git
          - --credential={"url":"https://github.com","username":"<username>","password-file":"/etc/git-secret/github-pat"}
          # terminate after one successful sync
          - --one-time
        authSecret:
            # the name of the secret created above
          name: git-pat
        # run as git sync user 
        securityContext:
          runAsUser: 65533
  version: "9.1.0"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/PgBouncer /initialization/yamls/git-sync-pat.yaml
PgBouncer .kubedb.com/sample-PgBouncer created
```

Once the database reaches the `Ready` state, you can verify the data using the method described above.


## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete PgBouncer -n demo sample-PgBouncer 
$ kubectl delete ns demo
```