---
title: Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: guides-druid-rotate-auth-guide
    name: Guide
    parent: guides-druid-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Rotate Authentication of Druid

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Druid` user's authentication credentials using a `DruidOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `password`and then reference this secret in the `DruidOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Druid](/docs/guides/druid/concepts/druid.md)
  - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/druid](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Druid

In this section, we are going to deploy a `Druid` database using `KubeDB`.

### Create External Dependency (Deep Storage)

Before proceeding further, we need to prepare deep storage, which is one of the external dependency of Druid and used for storing the segments. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/restart/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

Now, lets go ahead and create a druid database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

Let's create the `Druid` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/quickstart/druid-quickstart.yaml
druid.kubedb.com/druid-quickstart created
```
Now, wait until `druid-quickstart` has status Ready. i.e,

```shell
$ kubectl get druid -n demo
NAME               TYPE                  VERSION   STATUS   AGE
druid-quickstart   kubedb.com/v1alpha2   28.0.1    Ready    5m3s
```

## Verify authentication

````shell
$ kubectl get druid -n demo druid-quickstart -ojson | jq .spec.authSecret.name
"druid-quickstart-auth"
$ kubectl get secret -n demo druid-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎         
$ kubectl get secret -n demo druid-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
e4qcqnS.tt_zFQDa⏎                                                                                                                              
````

## Create RotateAuth DruidOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Druid using operator generated, we have to create a `DruidOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `DruidOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druidops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: druid-quickstart
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `druid-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Druid.

Let's create the `DruidOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/rotate-auth/yamls/Druid-rotate-auth-generated.yaml
 Druidopsrequest.ops.kubedb.com/druidops-rotate-auth-generated created
```
Let's wait for `DruidOpsrequest` to be `Successful`. Run the following command to watch `DruidOpsrequest` CRO
```shell
$ kubectl get Druidopsrequest -n demo
NAME                          TYPE         STATUS       AGE
druidops-rotate-auth-generated   RotateAuth   Successful   6m28s
```
If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Druidopsrequest -n demo druidops-rotate-auth-generated
Name:         druidops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2025-07-15T08:41:30Z
  Generation:          1
  Resource Version:    730862
  UID:                 8e7127b0-5eb5-4f7d-8140-7b2519d6b288
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   druid-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-15T08:41:30Z
    Message:               Druid ops-request has started to rotate auth for druid nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-15T08:41:33Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-15T08:41:34Z
    Message:               Successfully updated druid credential dynamically
    Observed Generation:   1
    Reason:                UpdateCredentialDynamically
    Status:                True
    Type:                  UpdateCredentialDynamically
    Last Transition Time:  2025-07-15T08:41:52Z
    Message:               successfully reconciled the Druid with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-15T08:42:47Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-15T08:41:57Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-historicals-0
    Last Transition Time:  2025-07-15T08:41:57Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-historicals-0
    Last Transition Time:  2025-07-15T08:42:02Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-historicals-0
    Last Transition Time:  2025-07-15T08:42:07Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-middlemanagers-0
    Last Transition Time:  2025-07-15T08:42:07Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-middlemanagers-0
    Last Transition Time:  2025-07-15T08:42:12Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-middlemanagers-0
    Last Transition Time:  2025-07-15T08:42:17Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-brokers-0
    Last Transition Time:  2025-07-15T08:42:17Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-brokers-0
    Last Transition Time:  2025-07-15T08:42:22Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-brokers-0
    Last Transition Time:  2025-07-15T08:42:27Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-routers-0
    Last Transition Time:  2025-07-15T08:42:27Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-routers-0
    Last Transition Time:  2025-07-15T08:42:32Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-routers-0
    Last Transition Time:  2025-07-15T08:42:37Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-coordinators-0
    Last Transition Time:  2025-07-15T08:42:37Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-coordinators-0
    Last Transition Time:  2025-07-15T08:42:42Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-coordinators-0
    Last Transition Time:  2025-07-15T08:42:47Z
    Message:               Successfully completed rotate auth opsRequest
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                              Age   From                         Message
  ----     ------                                                                              ----  ----                         -------
  Normal   Starting                                                                            52m   KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/druidops-rotate-auth-generated
  Normal   Starting                                                                            52m   KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-quickstart
  Normal   Successful                                                                          52m   KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-quickstart for DruidOpsRequest: druidops-rotate-auth-generated
  Normal   UpdatePetSets                                                                       52m   KubeDB Ops-manager Operator  successfully reconciled the Druid with new configure
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0               52m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0             52m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-historicals-0     52m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0            52m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0          52m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0  52m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0                   52m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0                 52m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-brokers-0         52m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-brokers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-routers-0                   51m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-routers-0                 51m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-routers-0         51m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-routers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0              51m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0            51m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-coordinators-0    51m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
  Normal   RestartNodes                                                                        51m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                            51m   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-quickstart
  Normal   Successful                                                                          51m   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-quickstart for DruidOpsRequest: druidops-rotate-auth-generated

```
**Verify Auth is rotated**
```shell
$  kubectl get druid -n demo druid-quickstart -ojson | jq .spec.authSecret.name
"druid-quickstart-auth"
$ kubectl get secret -n demo druid-quickstart-auth -o=jsonpath='{.data.username}' | base64 -d
admin⏎                                                               
$ kubectl get secret -n demo druid-quickstart-auth -o=jsonpath='{.data.password}' | base64 -d
gTJJMdgpKy9U(Eqi⏎                      
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo druid-quickstart-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
admin⏎                                                                                                          
$ kubectl get secret -n demo druid-quickstart-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
e4qcqnS.tt_zFQDa⏎                        
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

> Note: The database `username` is fixed as `admin` and cannot be changed. However, you can update the `password` while keeping the same `username`.
```shell
$ kubectl create secret generic druid-quickstart-auth-user -n demo \
                                             --type=kubernetes.io/basic-auth \
                                             --from-literal=username=admin \
                                             --from-literal=password=testpassword
secret/druid-quickstart-auth-user created
```
Now create a `DruidOpsRequest` with `RotateAuth` type. Below is the YAML of the `DruidOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: druid-quickstart
  authentication:
    secretRef:
      kind: Secret
      name: sample-druid-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `druid-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Druid.
- `spec.authentication.secretRef.name` specifies that we are using `druid-quickstart-auth-user` as `spec.authSecret.name` for authentication.

Let's create the `DruidOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/rotate-auth/yamls/Druid-rotate-auth-user.yaml
Druidopsrequest.ops.kubedb.com/drops-rotate-auth-user created
```
Let’s wait for `DruidOpsRequest` to be Successful. Run the following command to watch `DruidOpsRequest` CRO:

```shell
$ kubectl get drops -n demo
NAME                             TYPE         STATUS       AGE
drops-rotate-auth-user           RotateAuth   Successful   5m32s
druidops-rotate-auth-generated   RotateAuth   Successful   15m

```
We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed.
```shell
$  kubectl describe Druidopsrequest -n demo drops-rotate-auth-user
Name:         drops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2025-07-15T08:51:11Z
  Generation:          1
  Resource Version:    732405
  UID:                 1276e361-6585-4502-a744-d9e21c66b86e
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  sample-druid-auth-user
  Database Ref:
    Name:   druid-quickstart
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-15T08:51:11Z
    Message:               Druid ops-request has started to rotate auth for druid nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-15T08:51:11Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-15T08:51:11Z
    Message:               Successfully updated druid credential dynamically
    Observed Generation:   1
    Reason:                UpdateCredentialDynamically
    Status:                True
    Type:                  UpdateCredentialDynamically
    Last Transition Time:  2025-07-15T08:51:27Z
    Message:               successfully reconciled the Druid with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-15T08:52:22Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-15T08:51:32Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-historicals-0
    Last Transition Time:  2025-07-15T08:51:32Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-historicals-0
    Last Transition Time:  2025-07-15T08:51:37Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-historicals-0
    Last Transition Time:  2025-07-15T08:51:42Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-middlemanagers-0
    Last Transition Time:  2025-07-15T08:51:42Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-middlemanagers-0
    Last Transition Time:  2025-07-15T08:51:47Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-middlemanagers-0
    Last Transition Time:  2025-07-15T08:51:52Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-brokers-0
    Last Transition Time:  2025-07-15T08:51:52Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-brokers-0
    Last Transition Time:  2025-07-15T08:51:57Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-brokers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-brokers-0
    Last Transition Time:  2025-07-15T08:52:02Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-routers-0
    Last Transition Time:  2025-07-15T08:52:02Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-routers-0
    Last Transition Time:  2025-07-15T08:52:07Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-routers-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-routers-0
    Last Transition Time:  2025-07-15T08:52:12Z
    Message:               get pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-quickstart-coordinators-0
    Last Transition Time:  2025-07-15T08:52:12Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-quickstart-coordinators-0
    Last Transition Time:  2025-07-15T08:52:17Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-quickstart-coordinators-0
    Last Transition Time:  2025-07-15T08:52:22Z
    Message:               Successfully completed rotate auth opsRequest
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                              Age   From                         Message
  ----     ------                                                                              ----  ----                         -------
  Normal   Starting                                                                            57m   KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/drops-rotate-auth-user
  Normal   UpdatePetSets                                                                       57m   KubeDB Ops-manager Operator  successfully reconciled the Druid with new configure
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0               57m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0             57m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-historicals-0     57m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0            57m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0          57m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0  57m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-middlemanagers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0                   57m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0                 57m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-brokers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-brokers-0         56m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-brokers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-routers-0                   56m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-routers-0                 56m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-routers-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-routers-0         56m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-routers-0
  Warning  get pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0              56m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0            56m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-quickstart-coordinators-0    56m   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-quickstart-coordinators-0
  Normal   RestartNodes                                                                        56m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                            56m   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-quickstart
  Normal   Successful                                                                          56m   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-quickstart for DruidOpsRequest: drops-rotate-auth-user

```
**Verify auth is rotate**
```shell
$  kubectl get druid -n demo druid-quickstart -ojson | jq .spec.authSecret.name
"druid-quickstart-auth-user"
$ kubectl get secret -n demo druid-quickstart-auth-user -o=jsonpath='{.data.username}' | base64 -d
admin⏎                                                                    
$ kubectl get secret -n demo druid-quickstart-auth-user -o=jsonpath='{.data.password}' | base64 -d
testpassword⏎                                                                                   
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo druid-quickstart-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
admin⏎                                                                                                                                                             
$ kubectl get secret -n demo druid-quickstart-auth-user -o go-template='{{ index .data "password.prev" }}' | base64 -d
gTJJMdgpKy9U(Eqi⏎                                             
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Druidopsrequest druidops-rotate-auth-generated drops-rotate-auth-user -n demo
Druidopsrequest.ops.kubedb.com "druidops-rotate-auth-generated" "drops-rotate-auth-user" deleted
$ kubectl delete secret -n demo  druid-quickstart-auth-user
secret "druid-quickstart-auth-user" deleted
$ kubectl delete secret -n demo  druid-quickstart-auth
secret "druid-quickstart-auth" deleted
```


## Next Steps

- Learn how to use KubeDB to run Apache Druid cluster [here](/docs/guides/druid/README.md).
- Deploy [dedicated topology cluster](/docs/guides/druid/clustering/guide/index.md) for Apache Druid
- Monitor your Druid cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).
- Detail concepts of [DruidVersion object](/docs/guides/druid/concepts/druidversion.md).

[//]: # (- Learn to use KubeDB managed Druid objects using [CLIs]&#40;/docs/guides/druid/cli/cli.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).