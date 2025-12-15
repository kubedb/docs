---
title: Initialize MariaDB From Git Repository
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-initialization-gitsync
    name: Git Repository
    parent: guides-mariadb-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialization MariaDB from a Git Repository
This guide demonstrates how to use KubeDB to initialize a MariaDB database with initialization scripts (.sql, .sh, and/or .sql.gz) stored in a public or private Git repository.
To fetch the repository contents, KubeDB uses a sidecar container called [git-sync](https://github.com/kubernetes/git-sync).
In this example, we will initialize MariaDB using a `.sql` script from the GitHub repository [kubedb/mysql-init-scripts](https://github.com/kubedb/mysql-init-scripts).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster by following the steps [here](/docs/setup/README.md).

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.
```bash
$ kubectl create ns demo
namespace/demo created
```

## From Public Git Repository

KubeDB implements a `MariaDB` Custom Resource Definition (CRD) to define the specification of a MariaDB database.
To initialize the database from a public Git repository, you need to specify the required arguments for the `git-sync` sidecar container within the MariaDB resource specification.
The following YAML manifest shows an example `MariaDB` object configured with `git-sync`: 

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  init:
   script:
     scriptPath: "current"
     git:
       args:
       - --repo=https://github.com/kubedb/mysql-init-scripts
       - --link=current
       - --root=/root
       # terminate after one successful sync
       - --one-time 
  version: "10.5.23"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```
```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/initialization/git-sync-public.yaml
mariadb.kubedb.com/sample-mariadb created
```

The `git-sync` container has two required flags: 
 - `--repo`  – specifies the remote Git repository to sync.
 - `--root`  – specifies the working directory where the repository will be cloned.

Here, the value of the `--link` argument must match the value of `spec.init.script.scriptPath`. 
The `--link` argument creates a symlink that always points to the latest synced data. 

> To know more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync).

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo
NAME             VERSION   STATUS   AGE
sample-mariadb   10.5.23   Ready    5m
```

Next, we will connect to the MariaDB database and verify the data inserted from the `*.sql` script stored in the Git repository.

```bash
kubectl exec -it -n demo sample-mariadb-0 -- bash
Defaulted container "mariadb" out of: mariadb, mariadb-init (init), git-sync (init)

mysql@sample-mariadb-0:/$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 6
Server version: 10.5.23-MariaDB-1:10.5.23+maria~ubu2004 mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> use mysql;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
MariaDB [mysql]> select * from kubedb_table;
+----+-------+
| id | name  |
+----+-------+
|  1 | name1 |
|  2 | name2 |
|  3 | name3 |
+----+-------+
3 rows in set (0.000 sec)

MariaDB [mysql]> select * from demo_table;
+----+-------+
| id | name  |
+----+-------+
|  1 | name1 |
|  2 | name2 |
|  3 | name3 |
+----+-------+
3 rows in set (0.000 sec)
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

```bash
$ kubectl create secret generic -n demo git-creds \
    --from-file=ssh=$HOME/.ssh/id_rsa \
    --from-file=known_hosts=/tmp/known_hosts
```

The following YAML manifest provides an example of a `MariaDB` resource configured to use `git-sync` with a private Git repository:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  init:
   script:
     scriptPath: "current"
     git:
       args:
       # update with your private repository    
       - --repo=<your-ssh-repo-url>
       - --link=current
       - --root=/root
       # terminate after one successful sync
       - --one-time 
       authSecret:
         name: git-creds
       # run as git sync user 
       securityContext:
         runAsUser: 65533
  version: "10.5.23"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/initialization/git-sync-ssh.yaml
mariadb.kubedb.com/sample-mariadb created
```

Here, 
- `.spec.init.git.securityContext.runAsUser: 65533` ensure the container runs as the dedicated non-root `git-sync` user.
- `.spec.init.git.authSecret` specifies the secret containing the `SSH` key.

Once the database reaches the `Ready` state, you can verify the data using the method described above.

### 2. Using Username and Personal Access Token(PAT)

First, create a `Personal Access Token (PAT)` on your Git host server with the required permissions to access the repository.
Then create a Kubernetes secret using the `Personal Access Token (PAT)`:

```bash
$ kubectl create secret generic -n demo git-pat \
    --from-literal=github-pat=<ghp_yourpersonalaccesstoken>
```

Now, create a `MariaDB` resource that references the secret created above. 
The following YAML manifest shows an example:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  init:
   script:
     scriptPath: "current"
     git:
       args:
       # update with your private repository    
       - --repo=<your-https-repo-url>
       - --link=current
       - --root=/root
       - --credential={"url":"https://github.com","username":"<username>","password-file":"/etc/git-secret/github-pat"}
       # terminate after one successful sync
       - --one-time 
       authSecret:
         name: git-pat
       # run as git sync user 
       securityContext:
         runAsUser: 65533
  version: "10.5.23"
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/initialization/git-sync-pat.yaml
mariadb.kubedb.com/sample-mariadb created
```

Once the database reaches the `Ready` state, you can verify the data using the method described above.


## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
$ kubectl delete ns demo
```