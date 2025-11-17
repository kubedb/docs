---
title: Initialize PgpoolFrom Git Repository
menu:
  docs_{{ .version }}:
    identifier: guides-Pgpool-gitsync
    name: Git Repository
    parent: pb-initialization-Pgpool
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialization Pgpoolfrom a Git Repository
This guide demonstrates how to use KubeDB to initialize a Pgpooldatabase with initialization scripts (.sql, .sh, .js and/or .sql.gz) stored in a public or private Git repository.
To fetch the repository contents, KubeDB uses a sidecar container called [git-sync](https://github.com/kubernetes/git-sync).
In this example, we will initialize Pgpoolusing a `.sql` script from the GitHub repository [kubedb/Pgpool-init-scripts](https://github.com/kubedb/Pgpool-init-scripts).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Pgpool](/docs/guides/pgpool/concepts/Pgpool.md)
    - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)
    - [Updating Overview](/docs/guides/pgpool/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Pgpool](/docs/examples/Pgpool) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare Pgpool

Now, we are going to deploy a `Pgpool` with version `1.24.0`.

## From Public Git Repository

KubeDB implements a `Pgpool` Custom Resource Definition (CRD) to define the specification of a Pgpooldatabase.
To initialize the database from a public Git repository, you need to specify the required arguments for the `git-sync` sidecar container within the Pgpoolresource specification.
The following YAML manifest shows an example `Pgpool` object configured with `git-sync`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-demo
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: quick-postgres
    namespace: demo
  initConfig:
    pgpoolConfig:
      num_init_children : 6
      max_pool : 65
      child_life_time : 400
  deletionPolicy: WipeOut
  init:
    script:
      scriptPath: "pgbouncer-init-scripts/init"
      git:
        args:
          - --repo=https://github.com/kubedb/pgbouncer-init-scripts
          - --depth=1
          - --add-user=true
          - --period=60s
          - --one-time

```
```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/pgpool/initialization/yamls/git-sync-public.yaml
Pgpool.kubedb.com/sample-Pgpoolcreated
```
Here,

-`.spec.init.git.securityContext.runAsUser` the init container git_sync runs with user 999.
- `.spec.init.script.git.args` specifies the arguments for the `git-sync` container.
  The `git-sync` container has one required flags:
- `--repo`  – specifies the remote Git repository to sync.

Here, the value of the `--link` argument must match the value of `spec.init.script.scriptPath`.
The `--link` argument creates a symlink that always points to the latest synced data.

> To know more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync).

Now, wait until `sample-Pgpool` has status `Ready`. i.e,

```bash
$ kubectl get Pgpool-n demo 
NAME           VERSION   STATUS   AGE
sample-Pgpool  9.1.0     Ready    42m
```

Next, we will connect to the Pgpooldatabase and verify the data inserted from the `*.sql` script stored in the Git repository.

```bash
$   kubectl exec -it -n demo pgpool-demo-0 -- sh
Defaulted container "pgpool" out of: pgpool, git-sync (init)
/ $ cd init-scripts/
/init-scripts $ export PGPASSWORD="qrDy;GnX4QsKQ0UL"
/init-scripts $ psql -U postgres -d postgres -h localhost -p 9999
psql (17.6, server 13.13)
Type "help" for help.

postgres=# \dt
                   List of relations
 Schema |           Name            | Type  |  Owner   
--------+---------------------------+-------+----------
 public | kubedb_write_check_pgpool | table | postgres
 public | my_table                  | table | postgres
(2 rows)

postgres=# 

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

The following YAML manifest provides an example of a `Pgpool` resource configured to use `git-sync` with a private Git repository:

```yaml
apiVersion: kubedb.com/v1
kind: Pgpool
metadata:
  name: sample-Pgpool
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
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/pgpool/initialization/yamls/git-sync-ssh.yaml
Pgpool.kubedb.com/sample-Pgpoolcreated
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

Now, create a `Pgpool` resource that references the secret created above.
The following YAML manifest shows an example:

```yaml
apiVersion: kubedb.com/v1
kind: Pgpool
metadata:
  name: sample-Pgpool
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
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/pgpool/initialization/yamls/git-sync-pat.yaml
Pgpool.kubedb.com/sample-Pgpoolcreated
```

Once the database reaches the `Ready` state, you can verify the data using the method described above.


## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete Pgpool-n demo sample-Pgpool
$ kubectl delete ns demo
```