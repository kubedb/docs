---
title: Git Repository Pgpool Initialization
menu:
  docs_{{ .version }}:
    identifier: pp-git-repo-pgpool
    name: Git Repository
    parent: pp-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# Initialization Pgpool from a Git Repository
This guide demonstrates how to use KubeDB to initialize a Pgpool database with initialization scripts (.sql, .sh, .js and/or .sql.gz) stored in a public or private Git repository.
To fetch the repository contents, KubeDB uses a sidecar container called [git-sync](https://github.com/kubernetes/git-sync).
In this example, we will initialize Pgpool using a `.sh` script from the GitHub repository [kubedb/pgpool-init-scripts](https://github.com/kubedb/pgbouncer-pgpool-init-scripts/).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Pgpool](/docs/guides/pgpool/concepts/pgpool.md)
    - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)
    - [Updating Overview](/docs/guides/pgpool/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Pgpool](/docs/examples/pgpool) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md),but you have to set `password` as  `qrDy;GnX4QsKQ0UL`.

### Prepare Pgpool

Now, we are going to deploy a `Pgpool` with version `4.4.5`.

## From Public Git Repository

KubeDB implements a `Pgpool ` Custom Resource Definition (CRD) to define the specification of a Pgpool database.
To initialize the database from a public Git repository, you need to specify the required arguments for the `git-sync` sidecar container within the Pgpool resource specification.
The following YAML manifest shows an example `Pgpool ` object configured with `git-sync`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: postgres
    namespace: demo
  configuration:
    inline:
      pgpool.conf: |
          num_init_children : 6
          max_pool : 65
          child_life_time : 400
  deletionPolicy: WipeOut
  init:
    script:
      scriptPath: "current"
      git:
        args:
          - --repo=https://github.com/kubedb/pgbouncer-pgpool-init-scripts
          - --depth=1
          - --add-user=true
          - --period=60s
          - --one-time
```
```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/initialization/git-sync/git-sync-public.yaml
Pgpool .kubedb.com/pgpool created
```
Here,

- `.spec.init.script.git.args` specifies the arguments for the `git-sync` container.
  The `git-sync` container has one required flags:
- `--repo`  – specifies the remote Git repository to sync.

> To know more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync).

Now, wait until `pgpool` has status `Ready`. i.e,

```bash
$ kkubectl get Pgpool -n demo
NAME     TYPE                  VERSION   STATUS   AGE
pgpool   kubedb.com/v1alpha2   4.4.5     Ready    4m
```

Next, we will connect to the Pgpool database and verify the data inserted from the `*.sql` script stored in the Git repository.

```bash
$kubectl exec -it -n demo pgpool-0 -- sh
Defaulted container "pgpool" out of: pgpool, git-sync (init)
/ $ export PGPASSWORD="qrDy;GnX4QsKQ0UL"
/ $ psql -U postgres -d postgres -h localhost -p <db container port>
psql (17.6, server 13.13)
Type "help" for help.

postgres=# \dt
                   List of relations
 Schema |           Name            | Type  |  Owner   
--------+---------------------------+-------+----------
 public | kubedb_write_check_pgpool | table | postgres
 public | my_table                  | table | postgres
(2 rows)

```
`my_table` is created by the `init-script.sh` script stored in the Git repository.
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
$ kubectl create secret generic -n demo <secret_name> \
    --from-file=ssh=$HOME/.ssh/id_rsa \
    --from-file=known_hosts=/tmp/known_hosts
```

The following YAML manifest provides an example of a `Pgpool ` resource configured to use `git-sync` with a private Git repository:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: postgres
    namespace: demo
  configuration:
    inline:
      pgpool.conf: |
          num_init_children : 6
          max_pool : 65
          child_life_time : 400
  deletionPolicy: WipeOut
  init:
    script:
      scriptPath: "current"
      git:
        args:
          # use --ssh for private repository
          - --ssh
          - --repo=<private_git_repo_ssh_url>
          - --depth=1
          - --period=60s
          - --link=current
          - --root=/init-script-from-git
          # terminate after successful sync
          - --one-time
        authSecret:
          name: git-creds
        # run as git sync user
        securityContext:
          runAsUser: 65533
```


Here, replace `<private_git_repo_ssh_url>` with your private Git repository's SSH URL.


The `git-sync` container has two required flags:
- `--repo`  – specifies the remote Git repository to sync.
- `--root`  – specifies the working directory where the repository will be cloned.
- `spec.init.git.authSecret` specifies the secret containing the `SSH` key.
- `<private_git_repo_ssh_url>` with your private Git repository's SSH URL.
- `spec.init.script.scriptPath` – specifies the path within the repository and folder where the initialization scripts are located.
  for more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync/blob/master/docs/ssh.md)

Once the database reaches the `Ready` state, you can verify the data using the method described above.
```bash
$ kubectl get Pgpool -n demo
NAME     TYPE                  VERSION   STATUS   AGE
pgpool   kubedb.com/v1alpha2   4.4.5     Ready    5m23s

```
```bash
$ kubectl exec -it -n demo pgpool-0 -- sh
Defaulted container "pgpool" out of: pgpool, git-sync (init)
/ $ export PGPASSWORD="qrDy;GnX4QsKQ0UL"
/ $ psql -U postgres -d postgres -h localhost -p <db container port>
psql (17.6, server 13.13)
Type "help" for help.

postgres=# \dt
                   List of relations
 Schema |           Name            | Type  |  Owner   
--------+---------------------------+-------+----------
 public | kubedb_write_check_pgpool | table | postgres
 public | my_table                  | table | postgres
(2 rows)

```
`my_table` is created by the `init-script.sh` script stored in the Git repository.

### 2. Using Username and Personal Access Token(PAT)

First, create a `Personal Access Token (PAT)` on your Git host server with the required permissions to access the repository.
Then create a Kubernetes secret using the `Personal Access Token (PAT)`:
> Here, you can use any key name instead of `git-pat` to store the token in the secret.
```bash
$ kubectl create secret generic -n demo git-pat \
    --from-literal=github-pat=<ghp_yourpersonalaccesstoken>
```

Now, create a `Pgpool ` resource that references the secret created above.
The following YAML manifest shows an example:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: postgres
    namespace: demo
  configuration:
    inline:
      pgpool.conf: |
          num_init_children : 6
          max_pool : 65
          child_life_time : 400
  deletionPolicy: WipeOut
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
```



Here,

- `--credential`Provides authentication information for accessing a private Git repository over HTTPS.
- `<private_git_repo_http_url>` with your private Git repository's HTTPS URL.

OOnce the database reaches the `Ready` state, you can verify the data using the method described above.
```bash
$ kubectl get Pgpool -n demo
NAME     TYPE                  VERSION   STATUS   AGE
pgpool   kubedb.com/v1alpha2   4.4.5     Ready    3m32s

```
```bash
$ kubectl exec -it -n demo pgpool-0 -- sh
Defaulted container "pgpool" out of: pgpool, git-sync (init)
/ $ export PGPASSWORD="qrDy;GnX4QsKQ0UL"
/ $ psql -U postgres -d postgres -h localhost -p 9999
psql (17.6, server 13.13)
Type "help" for help.

postgres=# \dt
                   List of relations
 Schema |           Name            | Type  |  Owner   
--------+---------------------------+-------+----------
 public | kubedb_write_check_pgpool | table | postgres
 public | my_table                  | table | postgres
(2 rows)
```
`my_table` is created by the `init-script.sh` script stored in the Private Git repository.

## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete Pgpool -n demo pgpool
$ kubectl delete secret -n demo git-pat git-creds
$ kubectl delete ns demo
```