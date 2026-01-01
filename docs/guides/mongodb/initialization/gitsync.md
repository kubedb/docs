---
title: Initialize MongoDB  Git Repository
menu:
  docs_{{ .version }}:
    identifier: mg-gitsync-mongodb
    name: Git Repository
    parent: mg-initialization-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialization MongoDB from a Git Repository
This guide demonstrates how to use KubeDB to initialize a MongoDB database with initialization scripts (.js, and/or .sh) stored in a public or private Git repository.
To fetch the repository contents, KubeDB uses a sidecar container called [git-sync](https://github.com/kubernetes/git-sync).
In this example, we will initialize MongoDB using a `.js` script from the GitHub repository [kubedb/mongodb-init-scripts](https://github.com/kubedb/mongodb-init-scripts).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster by following the steps [here](/docs/setup/README.md).

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.
```bash
$ kubectl create ns demo
namespace/demo created
```

## From Public Git Repository

KubeDB implements a `MongoDB` Custom Resource Definition (CRD) to define the specification of a MongoDB database.
To initialize the database from a public Git repository, you need to specify the required arguments for the `git-sync` sidecar container within the MongoDB resource specification.
The following YAML manifest shows an example `MongoDB` object configured with `git-sync`:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-git
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
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut

```
```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/initialization/git-sync-public.yaml
MongoDB.kubedb.com/mg-git created
```

The `git-sync` container has two required flags:
- `--repo`  – specifies the remote Git repository to sync.
- `--root`  – specifies the root directory inside the container where the repository will be synced.



> To know more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync).

Now, wait until `mg-git` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME   VERSION   STATUS   AGE
mg-git     8.0.4     Ready    49m

```

Next, we will connect to the MongoDB database and verify the data inserted from the `*.js` script stored in the Git repository.

```bash
$  kubectl exec -it -n demo mg-git-0 -- bash
Defaulted container "mongodb" out of: mongodb, copy-config (init), git-sync (init)
mongodb@mg-git-0:/$ mongosh -u root -p $<your_mongodb_root_password>
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

>Here, we are using the default SSH key file located at `$HOME/.ssh/id_rsa`. If your SSH key is stored in a different location, please update the command accordingly. Also, you can use any name instead of `git-creds` to create the secret.

```bash
$ kubectl create secret generic -n demo git-creds \
    --from-file=ssh=$HOME/.ssh/id_rsa \
    --from-file=known_hosts=/tmp/known_hosts
```

The following YAML manifest provides an example of a `MongoDB` resource configured to use `git-sync` with a private Git repository:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-git-ssh
  namespace: demo
spec:
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
          - --root=/git
          # terminate after successful sync
          - --one-time
        authSecret:
          name: git-creds
        # run as git sync user 
        securityContext:
          runAsUser: 65533
  podTemplate:
    spec:
      # permission for reading ssh key
      securityContext:
        fsGroup: 65533
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



Here,
- `repo` with your private Git repository's SSH URL.
- `.spec.init.git.securityContext.runAsUser: 65533` ensure the container runs as the dedicated non-root `git-sync` user.
- `.spec.init.git.authSecret` specifies the secret containing the `SSH` key.

Once the database reaches the `Ready` state, you can verify the data using the method described above.

Next, we will connect to the MongoDB database and verify the data inserted from the `*.js` script stored in the Git repository.

```bash
$  kubectl exec -it -n demo mg-git-ssh-0 -- bash
Defaulted container "mongodb" out of: mongodb, copy-config (init), git-sync (init)
mongodb@mg-git-ssh-0:/$ mongosh -u root -p 'tQ;c(ykM_T_EbLKS'
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

> Here, you can use any key name instead of `git-pat` to store the token in the secret.

```bash
$ kubectl create secret generic -n demo git-pat \
    --from-literal=github-pat=<ghp_yourpersonalaccesstoken>
```

Now, create a `MongoDB` resource that references the secret created above.
The following YAML manifest shows an example:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-git-pat
  namespace: demo
spec:
  init:
    script:
      scriptPath: "current"
      git:
        args:
          # use --ssh for private repository
          - --repo=<private_git_repo_http_url>
          - --depth=1
          - --period=60s
          - --link=current
          - --root=/git
          - --credential={"url":"https://github.com","username":"<username>","password-file":"/etc/git-secret/github-pat"}
          # terminate after successful sync
          - --one-time
        authSecret:
          # the name of the secret created above
          name: git-pat
        # run as git sync user 
        securityContext:
          runAsUser: 65533
  podTemplate:
    spec:
      # permission for reading ssh key
      securityContext:
        fsGroup: 65533
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


The `git-sync` container has two required flags:
- `--repo`  – specifies the remote Git repository to sync.
- `--root`  – specifies the working directory where the repository will be cloned.
- `spec.init.git.authSecret` specifies the secret containing the `Personal Access Token (PAT)`.'
- `spec.init.git.args.--credential` specifies the credential information for accessing the private Git repository.

Here, the value of the `--link` argument must match the value of `spec.init.script.scriptPath`.
The `--link` argument creates a symlink that always points to the latest synced data.
Once the database reaches the `Ready` state, you can verify the data using the method described above.
```yaml
$ kubectl get mg -n demo
NAME          VERSION   STATUS   AGE
mg-git-pat     8.0.4     Ready    38m
```

## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete MongoDB -n demo mg-git mg-git-ssh mg-git-pat
$ kubectl delete secret -n demo git-pat git-creds
$ kubectl delete ns demo
```