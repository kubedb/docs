---
title: Virtual Secret Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-virtualsecret
    name: VirtualSecret Pgpool
    parent: pp-vs
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Virtual Secrets For pgpool: Secure Kubernetes Secrets
KubeDB's Virtual Secrets feature enhances the security of your database credentials by allowing you to use external secret management systems instead of storing sensitive information directly 
in Kubernetes Secrets. This guide will walk you through the steps to set up and use Virtual Secrets with your pgpool database in KubeDB.

## Virtual Secrets Design
`Virtual Secrets` extends Kubernetes by introducing a new `Secret` resource under the `virtual-secrets.dev` API group. From a user perspective, it behaves similarly to the native Kubernetes Secret
resource, providing familiar workflows for managing sensitive data. Unlike standard Kubernetes Secrets, Virtual Secrets does not store secret data in `etcd`. Instead, it securely stores the 
actual secret data in an `external secret manager`, ensuring enhanced security and compliance.

The Virtual Secret resource is structured into two distinct components:

- **Secret Data**– The sensitive information itself, stored externally to protect against unauthorized access.

- **Secret Metadata** – Non-sensitive information retained within the Kubernetes cluster to improve performance and support standard API operations.

This design ensures a seamless Kubernetes experience while providing enterprise-grade security for managing secrets.

## Prerequisites
Before you begin, ensure you have the following prerequisites in place:
- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- A running vault server with kubeVault operator installed. Follow the installation guide [here](https://kubevault.com/articles/how-to-use-hashicorp-vault-in-kubernetes-using-kubevault/).

- You should be familiar with the following `KubeDB` concepts:
    - [pgpool](/docs/guides/pgpool/concepts/pgpool.md)


To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
## How to use Virtual Secrets
### Install Virtual Secrets Server

First, install the virtual-secret-server which is a custom api server for the `secrets.virtual-secrets.dev` resource.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/virtual-secrets-server --version=v2025.3.14
$ helm upgrade -i virtual-secrets-server appscode/virtual-secrets-server \
    --version=v2025.3.14 -n kubevault --create-namespace 
```

### Deploy and Configure Vault Server

Before we create a custom `Secret`, we need to deploy a vault server where the secret data will be stored. Also, it needs to be configured to grant necessary permissions like create, update,
read, list, delete and delete in a kv secret engine named `virtual-secrets.dev` to the virtual-secrets-server.

Now let’s configure the vault server with following commands:

```shell
# enable kv secret engine in the path virtual-secrets.dev
$ vault secrets enable -path=virtual-secrets.dev -version=2 kv
Success! Enabled the kv secrets engine at: virtual-secrets.dev/


# creates a policy with the permission to create, update, read, list and delete  
$ vault policy write virtual-secrets-policy - <<EOF
path "virtual-secrets.dev/*" {
capabilities = ["create", "update", "read", "list", "delete"]
}
EOF
Success! Uploaded policy: virtual-secrets-policy


# binds this policy with a service account of the virtual-secrets server
$ vault write auth/kubernetes/role/virtual-secrets-role \
    bound_service_account_names=virtual-secrets-server \
    bound_service_account_namespaces=kubevault \
    policies="virtual-secrets-policy"
Success! Data written to: auth/kubernetes/role/virtual-secrets-role
```

### Create SecretStore
We need to create another resource called `SecretStore` which will contain the connection information to the external secret manager where the secrets will be stored.

```yaml
apiVersion: config.virtual-secrets.dev/v1alpha1
kind: SecretStore
metadata:
  name: vault
spec:
  vault:
    url: http://vault.vault-demo.svc:8200
    roleName: virtual-secrets-role
```
```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/vault/secretstore.yaml
secretstore.config.virtual-secrets.dev/vault configured
```

Here,

- `spec.vault` - section describes the connection information for vault.
- `spec.vault.url` - contains the connection url to the vault server.
- `spec.vault.roleName` - contains the role name we specified when binding the policy to the service account earlier.

> **Note:** `spec.aws`, `spec.azure` and `spec.gcp` can be used to specify the connection information of the corresponding secret manager.


### Create Virtual Secret
Now, we are going to create a `Virtual Secret` resource that will store the pgpool credentials in the vault server.
```yaml
apiVersion: virtual-secrets.dev/v1alpha1
kind: Secret
metadata:
  name: virtual-secret
  namespace: demo
stringData:
  username: pgpool
  password: virtual-secret
secretStoreName: vault
```
Here,

- `secretStoreName` - specifies the SecretStore we just created.
- Other than that, everything else is similar to a core Kubernetes Secret.
Let’s go ahead and apply the Secret,
```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/vault/pp_vs.yaml
secret.virtual-secrets.dev/virtual-secret created
```

Let's  list the Secrets to see if it is created or not,

```shell
kubectl get secrets.virtual-secrets.dev -n demo
NAME             TYPE                       DATA   AGE
virtual-secret   Opaque                     2      2d19h
```

We can also get the whole definition of the `Secret`,

```shell
$  kubectl get secrets.virtual-secrets.dev -n demo virtual-secret -oyaml
apiVersion: virtual-secrets.dev/v1alpha1
data:
  password: dmlydHVhbC1zZWNyZXQ=
  username: cGdwb29s
kind: Secret
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"virtual-secrets.dev/v1alpha1","kind":"Secret","metadata":{"annotations":{},"name":"virtual-secret","namespace":"demo"},"secretStoreName":"vault","stringData":{"password":"virtual-secret","username":"pgpool"}}
  creationTimestamp: "2026-02-26T09:07:06Z"
  generation: 2
  name: virtual-secret
  namespace: demo
  resourceVersion: "687530"
  uid: fb756118-3dbf-46b6-ac24-fa5cded478bc
secretStoreName: vault
type: Opaque
```

We can see that this `Secret`actually behaves identical of the core `Secret`. But the data is not stored in the `etcd` and it is way more secure than using the native `k8s Secret`.

#### Check server secret existence in Vault

We will connect to the Vault by using Vault CLI. Therefore, we need to export the necessary environment variables and port-forward the service.

In one terminal port-forward the vault server service,
```shell
$ kubectl port-forward -n demo service/vault 8200
Forwarding from 127.0.0.1:8200 -> 8200
Forwarding from [::1]:8200 -> 8200
```
```shell
$ export VAULT_ADDR=http://127.0.0.1:8200
$ export VAULT_TOKEN=(kubectl vault root-token get vaultserver vault -n demo --value-only)
$ vault kv get virtual-secrets.dev/demo/virtual-secret
================ Secret Path ================
virtual-secrets.dev/data/demo/virtual-secret

======= Metadata =======
Key                Value
---                -----
created_time       2025-12-30T11:13:54.455334654Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1

====== Data ======
Key         Value
---         -----
password    virtual-secret
username    default
```
We can see that the secret data is stored in the `virtual-secrets.dev/demo/virtual-secret` path where,

- `virtual-secret.dev` is the secret engine name.
- `demo` is the namespace.
- `virtual-secret`is the name of the secret.

### Mount Virtual Secret in pgpool
`Secrets` are not that useful if we can not mount them to pods. We can mount the virtual secrets using [Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/) .

Virtual Secrets comes with a custom provider of `Secrets Store CSI Driver`, named `secrets-store-csi-driver-provider-virtual-secrets` which leverages `virtual-secrets-server` to read secret
data from virtual secrets and uses the `Secrets Store CSI Driver` to mount those into to the pods.

Let’s go ahead and install `Secrets Store CSI Driver` and `secrets-store-csi-driver-provider-virtual-secrets` into our cluster,

```shell
$ helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
$ helm install csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver --namespace kube-system
$ helm search repo appscode/secrets-store-csi-driver-provider-virtual-secrets --version=v2025.3.14
$ helm upgrade -i secrets-store-csi-driver-provider-virtual-secrets appscode/secrets-store-csi-driver-provider-virtual-secrets -n kube-system --create-namespace --version=v2025.3.14
```

If both of them are deployed we should see two new pods in the `kube-system` namespace.

```shell
$ kubectl get pods -n kube-system
NAME                                                      READY   STATUS    RESTARTS      AGE
coredns-695cbbfcb9-r6v7j                                  1/1     Running   1 (36h ago)   2d18h
csi-secrets-store-secrets-store-csi-driver-qzq8z          3/3     Running   3 (36h ago)   2d
local-path-provisioner-546dfc6456-lpdp4                   1/1     Running   1 (36h ago)   2d18h
secrets-store-csi-driver-provider-virtual-secrets-mdw84   1/1     Running   1 (36h ago)   47h
```
The `Secrets Store CSI Driver` uses a custom resource named `SecretProviderClass` to mount the secret. Let’s go ahead and create that,

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: virtual-secret
  namespace: demo
spec:
  provider: virtual-secrets
  parameters:
    secretName: virtual-secret
```

Here,

- `spec.provider` - specifies the provider for Secrets Store CSI Driver to communicate and use.
- `spec.parameters.secretName` - specifies the name of the virtual secret we want to mount.

> **Note:**  We can also call the mount subresource of the virtual secret to create the SecretProviderClass for us.
The namespace and the name of SecretProviderClass should be same as the Virtual Secret it is being used for.

 Let’s create the SecretProviderClass,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/vault/secretProviderClass.yaml
secretproviderclass.secrets-store.csi.x-k8s.io/virtual-secret created
```

With the custom secret created, the authentication configured and role created, the `provider-virtual-secrets` extension installed and the `SecretProviderClass` defined it is finally time to
create a pod that mounts the desired secret.

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: webapp
  namespace: demo
spec:
  containers:
    - image: jweissig/app:0.0.1
      name: webapp
      volumeMounts:
        - name: virtual-secrets-store
          mountPath: "/mnt/virtual-secrets"
          readOnly: true
  volumes:
    - name: virtual-secrets-store
      csi:
        driver: secrets-store.csi.k8s.io
        readOnly: true
        volumeAttributes:
          secretProviderClass: "virtual-secret"
```
Here,

- In `spec.volumes[0]`, a volume with name `virtual-secrets-store` with necessary configs is specified.
- In `spec.containers[0].volumeMounts`, the volume is referred to be mounted in the `/mnt/virtual-secrets` path.


Let’s create the pod,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/vault/webapp.yaml
pod/webapp created
```

If we get the pod we will see that it will get to the `Running` state after some period,
```shell
$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
webapp              1/1     Running   0          6m45s
```

Now, check the secret data written to the file system at /mnt/virtual-secrets on the webapp pod.

```shell
$  kubectl exec -n demo webapp -- cat /mnt/virtual-secrets/username
pgbouncer⏎                    

$ kubectl exec -n demo webapp -- cat /mnt/virtual-secrets/password
virtual-secret⏎ 
```

The value displayed matches the username and password value for the custom secret named `virtual-secret` we created earlier.

## Get PostgreSQL Server ready 

Pgpool is a middleware for PostgreSQL. Therefore you will need to have a PostgreSQL server up and running for Pgpool to connect to.

Luckily PostgreSQL is readily available in KubeDB as CRD and can easily be deployed using this guide [here](/docs/guides/postgres/quickstart/quickstart.md). But by default this will create a PostgreSQL server with `max_connections=100`, but we need more than 100 connections for our Pgpool to work as expected. 

Pgpool requires at least `2*num_init_children*max_pool*spec.replicas` connections in PostgreSQL server. So use [this](https://kubedb.com/docs/v2024.4.27/guides/postgres/configuration/using-config-file/) to create a PostgreSQL server with custom `max_connections`.

In this tutorial, we will use a PostgreSQL named `quick-postgres` in the `demo` namespace.

KubeDB creates all the necessary resources including services, secrets, and appbindings to get this server up and running. A default database `postgres` is created in `quick-postgres`. Database secret `quick-postgres-auth` holds this user's username and password. Following is the yaml file for it.

KubeDB creates all the necessary resources including services, secrets, and appbindings to get this server up and running. A default database `postgres` is created in `pp-demo`. Database secret `pp-demo-auth` holds this user's username and password. Following is the yaml file for it.

### Use Virtual Secrets with pgpool
Virtual Secrets is integrated with KubeDB from the v2025.3.24 and it can be used to store KubeDB’s database credential. Now, the support has been added for `pgpool`.
We can proceed with deploying a `pgpool` which will use `virtual-secrets` to create custom secret for the database authentication credential.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-vs
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: pg-demo
    namespace: demo
  sslMode: disable
  clientAuthMode: md5
  syncUsers: true
  deletionPolicy: WipeOut
  authSecret:
    kind: secret
    apiGroup: "virtual-secrets.dev"
    secretStoreName: vault
    name: virtual-secret
```
Here,

- `spec.authSecret.apiGroup`- specifies that we want to use virtual secrets instead of native k8s secret.
- `spec.authSecret.secretStoreName` - specifies the `SecretStore` resource that contains the connection information for external secret store to store the secret data.

We can now apply the pgpool custom resource,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/pp_vs.yaml
pgpool.kubedb.com/pp-vs created
``` 
Now, wait until `pp-vs` has status `Ready`. i.e. ,
```shell
$ kubectl get pp -n demo
NAME    VERSION   STATUS   AGE
pp-vs   4.5.0     Ready    41m
```

Now, lets go ahead and check what secret it is using,
```shell
$ kubectl get secrets.virtual-secrets.dev -n demo
NAME             TYPE     DATA   AGE
virtual-secret   Opaque   2      1d
```

We can see that the pgpool user password is stored in the vault server as named ```virtual-secret``` . Now let’s go ahead and connect to the database using the password to check whether it is working or not.
```bash
$ export PGPASSWORD='virtual-secret'
$ psql --host=localhost --port=9999 --username=postgres postgres
psql (16.11 (Ubuntu 16.11-0ubuntu0.24.04.1), server 18.2)
WARNING: psql major version 16, server major version 18.
         Some psql features might not work.
Type "help" for help.

postgres=# SELECT datname FROM pg_database;
  datname  
-----------
 postgres
 template1
 template0
(3 rows)

```
We can see that we are able to connect to the database and create a database and a table successfully.

## Cleanup
To clean up the resources created in this guide, run the following commands:
```bash
$ kubectl delete pp -n demo pp-vs
pgpool.kubedb.com "pp-vs" deleted
$ kubectl delete pod webapp -n demo
$ kubectl delete secretproviderclass -n demo virtual-secret
$ kubectl delete ns demo
$ helm uninstall virtual-secrets-server -n kubevault
$ helm uninstall secrets-store-csi-driver-provider-virtual-secrets -n kube-system
$ helm uninstall csi-secrets-store -n kube-system
```
If you want to uninstall the `KubeVault`, run:
```bash
$ helm uninstall kubevault --namespace kubevault
```
