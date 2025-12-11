---
title: Initialize Redis gitsync
menu:
  docs_{{ .version }}:
    identifier: rd-gitsync-initialization
    name: Git Repository
    parent: rd-initialization-redis
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialization Redis from a Git Repository
This guide demonstrates how to use KubeDB to initialize a Redis database with initialization scripts ( .sh, and/or .sh.gz) stored in a public or private Git repository.
To fetch the repository contents, KubeDB uses a sidecar container called [git-sync](https://github.com/kubernetes/git-sync).
In this example, we will initialize Redis using a `.sh` script from the GitHub repository [kubedb/redis-init-scripts](https://github.com/kubedb/redis-init-scripts).

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
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-demo
  namespace: demo
spec:
  version: "8.2.2"
  mode: Standalone
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  init:
    script:
      scriptPath: "redis-init-scripts"
      git:
        args:
          - --repo=https://github.com/kubedb/redis-init-scripts
          - --depth=1
          - --period=60s
          - --one-time
        securityContext:
          runAsUser: 999
  terminationPolicy: Delete

```
```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/initialization/git-sync-public.yaml
Redis.kubedb.com/redis-demo created
```

The `git-sync` container has two required flags:
- `--repo`  – specifies the remote Git repository to sync.
- `spec.init.script.scriptPath` – specifies the path within the repository where the initialization scripts are located.
- - `.spec.init.git.securityContext.runAsUser: 999` ensure the container runs as the dedicated non-root `git-sync` user.
> To know more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync).

Now, wait until `redis-demo` has status `Ready`. i.e,

```bash
$ kubectl get Redis -n demo
NAME             VERSION   STATUS   AGE
redis-demo        8.2.2    Ready    5m
```

Next, we will connect to the Redis database and verify the data inserted from the `*.sh` script stored in the Git repository.

```bash
$ kubectl exec -n demo -it redis-demo-0 -- bash
Defaulted container "redis" out of: redis, redis-init (init)

# Inside the pod
root@redis-demo-0:/data# redis-cli
127.0.0.1:6379> GET user:1:name
"John Doe"
127.0.0.1:6379> GET user:1:email
"john@example.com"
127.0.0.1:6379> GET user:2:name
"Jane Smith"
127.0.0.1:6379> GET user:3:name
"Bob Johnson"
127.0.0.1:6379> GET user:4:name
"Alice Williams"
127.0.0.1:6379> GET user:5:name
"Charlie Brown"
127.0.0.1:6379> GET user:6:name
"Diana Prince"
127.0.0.1:6379> KEYS user:*
 1) "user:1:name"
 2) "user:1:email"
 3) "user:2:name"
 4) "user:2:email"
 5) "user:3:name"
 6) "user:3:email"
 7) "user:4:name"
 8) "user:4:email"
 9) "user:5:name"
10) "user:5:email"
11) "user:6:name"
12) "user:6:email"
127.0.0.1:6379> QUIT
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
> Here, we are using the default SSH key file located at `$HOME/.ssh/id_rsa`. If your SSH key is stored in a different location, please update the command accordingly. Also, you can use any name instead of `git-creds` to create the secret.
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
  name: redis-demo
  namespace: demo
spec:
  version: 8.2.2
  mode: Cluster
  init:
    script:
      scriptPath: "redis_script.git"
      git:
        args:
          # use --ssh for private repository
          - --ssh
          - --repo=<private_git_repo_ssh_url>
          - --depth=1
          - --period=60s
          - --root=/init-script-from-git
          # terminate after successful sync
          - --one-time
        authSecret:
            # the name of the secret created above
          name: git-creds
        # got credentials from the secret
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```


```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/initialization/git-sync-ssh.yaml
Redis.kubedb.com/redis-demo created
```

Here, replace `<private_git_repo_ssh_url>` with your private Git repository's SSH URL.


The `git-sync` container has two required flags:
- `--repo`  – specifies the remote Git repository to sync.
- `--root`  – specifies the working directory where the repository will be cloned.
- `spec.init.git.authSecret` specifies the secret containing the `SSH` key.
- `spec.init.script.scriptPath` – specifies the path within the repository where the initialization scripts are located.
for more about `git-sync` configuration visit this [link](https://github.com/kubernetes/git-sync/blob/master/docs/ssh.md)

Once the database reaches the `Ready` state, you can verify the data using the method described above.
```shell
kubectl get redis -n demo 
NAME         VERSION   STATUS   AGE
redis-demo   8.2.2     Ready    48m

```
```shell
$ kubectl exec -n demo -it redis-demo-shard0-0 -- bash
Defaulted container "redis" out of: redis, redis-init (init), git-sync (init)
redis@redis-demo-shard0-0:/data$ redis-cli -c
127.0.0.1:6379>  get user:1:name
-> Redirected to slot [12440] located at 10.42.0.241:6379
"John Doe"
10.42.0.241:6379> exit
```

### 2. Using Username and Personal Access Token(PAT)

First, create a `Personal Access Token (PAT)` on your Git host server with the required permissions to access the repository.
Then create a Kubernetes secret using the `Personal Access Token (PAT)`:
> Here, you can use any key name instead of `git-pat` to store the token in the secret.
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
  name: redis-demo
  namespace: demo
spec:
  version: 8.2.2
  mode: Cluster
  init:
    script:
      scriptPath: "redis_script.git"
      git:
        args:
          - --repo=<private_git_repo_http_url>
          - --depth=1
          - --period=60s
          - --root=/init-script-from-git
          - --credential={"url":"https://github.com","username":"<username>","password-file":"/etc/git-secret/github-pat"}
          # terminate after successful sync
          - --one-time
        authSecret:
          # the name of the secret created above
          name: git-pat
        # run as git credentials user
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/initialization/git-sync-pat.yaml
Redis.kubedb.com/redis-demo created
```
Here,

- `--credential`Provides authentication information for accessing a private Git repository over HTTPS.
- `<private_git_repo_http_url>` with your private Git repository's HTTPS URL.
Once the database reaches the `Ready` state, you can verify the data using the method described above. Let's check:
```shell
kubectl get redis -n demo 
NAME         VERSION   STATUS   AGE
redis-demo   8.2.2     Ready    48m

```
```shell
$ kubectl exec -n demo -it redis-demo-shard0-0 -- bash
Defaulted container "redis" out of: redis, redis-init (init), git-sync (init)
redis@redis-demo-shard0-0:/data$ redis-cli -c
127.0.0.1:6379>  get user:1:name
-> Redirected to slot [12440] located at 10.42.0.241:6379
"John Doe"
10.42.0.241:6379> exit
```

## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete Redis -n demo redis-demo
$ kubectl delete ns demo
```