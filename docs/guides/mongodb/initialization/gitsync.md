---
title: Initialize Redis From Git Repository
menu:
  docs_{{ .version }}:
    identifier: guides-Redis-initialization-gitsync
    name: Git Repository
    parent: guides-Redis-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialization Redis from a Git Repository
This guide demonstrates how to use KubeDB to initialize a Redis database with initialization scripts (.sql, .sh, and/or .sql.gz) stored in a public or private Git repository.
To fetch the repository contents, KubeDB uses a sidecar container called [git-sync](https://github.com/kubernetes/git-sync).
In this example, we will initialize Redis using a `.sql` script from the GitHub repository [kubedb/mysql-init-scripts](https://github.com/kubedb/mysql-init-scripts).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster by following the steps [here](/docs/setup/README.md).

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.
```bash
$ kubectl create ns demo
namespace/demo created
```

## From Public Git Repository

KubeDB implements a `Redis` Custom Resource Definition (CRD) to define the specification of a Redis database.
To initialize the database from a public Git repository, you need to specify the required arguments for the `git-sync` sidecar container within the Redis resource specification.
The following YAML manifest shows an example `Redis` object configured with `git-sync`:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: rs
  namespace: demo
spec:
  init:
    script:
      scriptPath: "current"
      git:
        args:
          - --repo=https://github.com/kubedb/git-sync-demo.git
          - --depth=1
          - --period=60s
          - --link=current
          - --root=/git
          # terminate after successful sync
          - --one-time
  version: "8.0.4"
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut

```
```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Redis/initialization/git-sync-public.yaml
Redis.kubedb.com/sample-Redis created
```

The `git-sync` container has two required flags:
- `--repo`  – specifies the remote Git repository to sync.
- `--root`  – specifies the root directory inside the container where the repository will be synced.



> To know more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync).

Now, wait until `sample-Redis` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME   VERSION   STATUS   AGE
rs     8.0.4     Ready    49m

```

Next, we will connect to the Redis database and verify the data inserted from the `*.sql` script stored in the Git repository.

```bash
$  kubectl exec -it -n demo rs-0 -- bash
Defaulted container "mongodb" out of: mongodb, copy-config (init), git-sync (init)
mongodb@rs-0:/$ mongosh -u root -p 'tQ;c(ykM_T_EbLKS'
Current Mongosh Log ID:	6900695d231f7a9e99ce5f46
Connecting to:		mongodb://<credentials>@127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.5.8
Using MongoDB:		8.0.4
Using Mongosh:		2.5.8

For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/

------
   The server generated these startup warnings when booting
   2025-10-28T06:12:45.927+00:00: Using the XFS filesystem is strongly recommended with the WiredTiger storage engine. See http://dochub.mongodb.org/core/prodnotes-filesystem
   2025-10-28T06:12:46.608+00:00: For customers running the current memory allocator, we suggest changing the contents of the following sysfsFile
   2025-10-28T06:12:46.608+00:00: For customers running the current memory allocator, we suggest changing the contents of the following sysfsFile
   2025-10-28T06:12:46.608+00:00: We suggest setting the contents of sysfsFile to 0.
   2025-10-28T06:12:46.608+00:00: Your system has glibc support for rseq built in, which is not yet supported by tcmalloc-google and has critical performance implications. Please set the environment variable GLIBC_TUNABLES=glibc.pthread.rseq=0
   2025-10-28T06:12:46.608+00:00: vm.max_map_count is too low
------

test> use kubedb
... 
switched to db kubedb
kubedb> show collections
... 
people
kubedb> db.people.find()
... 
[
  {
    _id: ObjectId('69005edc696fc9db26ce5f47'),
    firstname: 'kubernetes',
    lastname: 'database'
  }
]
kubedb> exit
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

The following YAML manifest provides an example of a `Redis` resource configured to use `git-sync` with a private Git repository:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: sample-Redis
  namespace: demo
spec:
  init:
   script:
     scriptPath: "current"
     git:
       args:
       # update with your private repository    
       - --repo=git@github.com:refat75/mysql-init-scripts.git
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
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Redis/initialization/git-sync-ssh.yaml
Redis.kubedb.com/sample-Redis created
```

Here,
- `.spec.init.git.securityContext.runAsUser: 65533` ensure the container runs as the dedicated non-root `git-sync` user.
- `.spec.init.git.authSecret` specifies the secret containing the `SSH` key.

Once the database reaches the `Ready` state, you can verify the data using the method described above.

Next, we will connect to the Redis database and verify the data inserted from the `*.sql` script stored in the Git repository.

```bash
$  kubectl exec -it -n demo rs-0 -- bash
Defaulted container "mongodb" out of: mongodb, copy-config (init), git-sync (init)
mongodb@rs-0:/$ mongosh -u root -p 'tQ;c(ykM_T_EbLKS'
Current Mongosh Log ID:	6900695d231f7a9e99ce5f46
Connecting to:		mongodb://<credentials>@127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.5.8
Using MongoDB:		8.0.4
Using Mongosh:		2.5.8

For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/

------
   The server generated these startup warnings when booting
   2025-10-28T06:12:45.927+00:00: Using the XFS filesystem is strongly recommended with the WiredTiger storage engine. See http://dochub.mongodb.org/core/prodnotes-filesystem
   2025-10-28T06:12:46.608+00:00: For customers running the current memory allocator, we suggest changing the contents of the following sysfsFile
   2025-10-28T06:12:46.608+00:00: For customers running the current memory allocator, we suggest changing the contents of the following sysfsFile
   2025-10-28T06:12:46.608+00:00: We suggest setting the contents of sysfsFile to 0.
   2025-10-28T06:12:46.608+00:00: Your system has glibc support for rseq built in, which is not yet supported by tcmalloc-google and has critical performance implications. Please set the environment variable GLIBC_TUNABLES=glibc.pthread.rseq=0
   2025-10-28T06:12:46.608+00:00: vm.max_map_count is too low
------

test> use kubedb
... 
switched to db kubedb
kubedb> show collections
... 
people
kubedb> db.people.find()
... 
[
  {
    _id: ObjectId('69005edc696fc9db26ce5f47'),
    firstname: 'kubernetes',
    lastname: 'database'
  }
]
kubedb> exit
```
### 2. Using Username and Personal Access Token(PAT)

First, create a `Personal Access Token (PAT)` on your Git host server with the required permissions to access the repository.
Then create a Kubernetes secret using the `Personal Access Token (PAT)`:

```bash
$ kubectl create secret generic -n demo git-pat \
    --from-literal=github-pat=<ghp_yourpersonalaccesstoken>
```

Now, create a `Redis` resource that references the secret created above.
The following YAML manifest shows an example:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: sample-Redis
  namespace: demo
spec:
  init:
   script:
     scriptPath: "current"
     git:
       args:
       # update with your private repository    
       - --repo=https://github.com/refat75/mysql-init-scripts.git
       - --link=current
       - --root=/root
       - --credential={"url":"https://github.com","username":"refat75","password-file":"/etc/git-secret/github-pat"}
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
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Redis/initialization/git-sync-pat.yaml
Redis.kubedb.com/sample-Redis created
```

Once the database reaches the `Ready` state, you can verify the data using the method described above.


## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete Redis -n demo sample-Redis
$ kubectl delete ns demo
```