# Rotate Authentication of Hazelcast
**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Hazelcast` user's authentication credentials using a `HazelcastOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential.
2. **User Defined:** The user can create their own credentials by defining a Secret of type
   `kubernetes.io/basic-auth` containing the desired `password`, and then reference this Secret in the
   `HazelcastOpsRequest` CR.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB in your cluster following the steps [here](/docs/setup/README.md) and make sure install with helm command including `--set global.featureGates.Hazelcast=true` to ensure Hazelcast CRDs.

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

 ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY     VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete            WaitForFirstConsumer   false                  6h22m
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```
## Find Available HazelcastVersion

When you have installed KubeDB, it has created `HazelcastVersion` CR for all supported Hazelcast versions. Check it by using the `kubectl get hazelcastversions` command. You can also use `hzversion` shorthand instead of `hazelcastversions`.

```bash
$ kubectl get hzversion
NAME    VERSION   DB_IMAGE                               DEPRECATED   AGE
5.5.2   5.5.2     hazelcast/hazelcast-enterprise:5.5.2                3m52s
5.5.6   5.5.6     hazelcast/hazelcast-enterprise:5.5.6                3m52s
```
## Create a Hazelcast server

KubeDB implements a `Hazelcast` CRD to define the specification of a Hazelcast server.
Before deploying hazelcast we need to create license secret since we are running enterprise version of hazelcast.


```bash
kubectl create secret generic hz-license-key -n demo --from-literal=licenseKey=TrialLicense#10Nodes#eyJhbGxvd2VkTmF0aXZlTWVtb3J5U2l6ZSI6MTAwLCJhbGxvd2VkTnVtYmVyT2ZOb2RlcyI6MTAsImFsbG93ZWRUaWVyZWRTdG9yZVNpemUiOjAsImFsbG93ZWRUcGNDb3JlcyI6MCwiY3JlYXRpb25EYXRlIjoxNzQ4ODQwNDc3LjYzOTQ0NzgxNiwiZXhwaXJ5RGF0ZSI6MTc1MTQxNDM5OS45OTk5OTk5OTksImZlYXR1cmVzIjpbMCwyLDMsNCw1LDYsNyw4LDEwLDExLDEzLDE0LDE1LDE3LDIxLDIyXSwiZ3JhY2VQZXJpb2QiOjAsImhhemVsY2FzdFZlcnNpb24iOjk5LCJvZW0iOmZhbHNlLCJ0cmlhbCI6dHJ1ZSwidmVyc2lvbiI6IlY3In0=.6PYD6i-hejrJ5Czgc3nYsmnwF7mAI-78E8LFEuYp-lnzXh_QLvvsYx4ECD0EimqcdeG2J5sqUI06okLD502mCA==
secret/hz-license-key created
```

Below is the `Hazelcast` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hazelcast-quickstart
  namespace: demo
spec:
  deletionPolicy: WipeOut
  licenseSecret:
    name: hz-license-key
  replicas: 2
  version: 5.5.2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```
Let's create the Hazelcast CR that is shown above:
```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/hazelcast/quickstart/overview/yamls/hazelcast.yaml
hazelcast.kubedb.com/hazelcast-sample created
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get hz -n demo hazelcast-quickstart -ojson | jq .spec.authSecret.name
"hazelcast-quickstart-auth"
$ kubectl get secret -n demo hazelcast-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎          
$ kubectl get secret -n demo hazelcast-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
pp5rmyri3A2SskRi⏎                      
````
Now, you can exec into the pod `hazelcast-quickstart-0` and run a REST api using `username` and `password`
```bash

$ kubectl exec -it -n demo hazelcast-quickstart-0 -c hazelcast -- curl -u admin:'0TdsoNJez9zjJddh' http://localhost:5701/hazelcast/rest/cluster

{"members":[{"address":"[10.244.0.21]:5701","liteMember":false,"localMember":true,"uuid":"f6c9c447-7abd-4254-9a52-3457f1e85713","memberVersion":"5.5.2"},{"address":"[10.244.0.23]:5701","liteMember":false,"localMember":false,"uuid":"9490ac0d-6d0c-437d-898c-c6c6aa81402e","memberVersion":"5.5.2"}],"connectionCount":1,"allConnectionCount":2}⏎     

```
If you can access the map and retrieve values using the REST API, it means the secrets are working correctly.

## Create RotateAuth HazelcastOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Hazelcast using operator generated, we have to create a `RabbitMQOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `HazelcastOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: hazelcast-quickstart
  apply: IfReady
```
Here,
- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `hazelcast-quickstart` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Hazelcast.

Let's create the `HazelcastOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/hazelcast/rotate-auth/rotate-auth-generated.yaml
 hazelcastopsrequest.ops.kubedb.com/hzops-rotate-auth-generated created
```
Let's wait for `HazelcastOpsrequest` to be `Successful`. Run the following command to watch `HazelcastOpsrequest` CRO
```shell
 $ kubectl get Hazelcastopsrequest -n demo
NAME                          TYPE         STATUS       AGE
hzops-rotate-auth-generated   RotateAuth   Successful   2m32s
```
If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Hazelcastopsrequest -n demo hzops-rotate-auth-generated 
Name:         hzops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-18T06:55:53Z
  Generation:          1
  Resource Version:    369136
  UID:                 9db7e41a-57ca-45f9-bff3-0d11374cffb6
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hazelcast-quickstart
  Type:    RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-18T06:55:53Z
    Message:               hazelcast ops request has started to rotate auth for rmq nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-18T06:55:56Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-18T06:56:06Z
    Message:               successfully reconciled the hazelcast with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-18T06:57:06Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-18T06:56:16Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-quickstart-0
    Last Transition Time:  2025-08-18T06:56:16Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-quickstart-1
    Last Transition Time:  2025-08-18T06:56:16Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-18T06:56:26Z
    Message:               running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hazelcast-quickstart-0
    Last Transition Time:  2025-08-18T06:56:56Z
    Message:               running pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hazelcast-quickstart-1
    Last Transition Time:  2025-08-18T06:57:06Z
    Message:               Successfully completed reconfigure Ignite
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           4m16s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-rotate-auth-generated
  Normal   Starting                                                           4m16s  KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hazelcast-quickstart
  Normal   Successful                                                         4m16s  KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hazelcast-quickstart for HazelcastOpsRequest: hzops-rotate-auth-generated
  Normal   UpdateStatefulSets                                                 4m3s   KubeDB Ops-manager Operator  successfully reconciled the hazelcast with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0      3m53s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1      3m53s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
  Warning  running pod; ConditionStatus:False                                 3m53s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  3m43s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:False                                 3m43s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  3m33s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:False                                 3m33s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  3m23s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:False                                 3m23s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  3m13s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-1  3m13s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
  Normal   RestartNodes                                                       3m3s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           3m3s   KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hazelcast-quickstart
  Normal   Successful                                                         3m3s   KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hazelcast-quickstart for HazelcastOpsRequest: hzops-rotate-auth-generated
```
**Verify Auth is rotated**
````shell
$ kubectl get hz -n demo hazelcast-quickstart -ojson | jq .spec.authSecret.name
"hazelcast-quickstart-auth"
$ kubectl get secret -n demo hazelcast-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎          
$ kubectl get secret -n demo hazelcast-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
CYIpaMGLwfHmvA!h              
````
Now, you can exec into the pod `hazelcast-quickstart-0` and run a REST api using `username` and `password`
```bash
$ kubectl exec -it -n demo hazelcast-quickstart-0 -c hazelcast -- curl -u admin:'CYIpaMGLwfHmvA!h' http://localhost:5701/hazelcast/rest/cluster
{"members":[{"address":"[10.244.0.25]:5701","liteMember":false,"localMember":false,"uuid":"9490ac0d-6d0c-437d-898c-c6c6aa81402e","memberVersion":"5.5.2"},{"address":"[10.244.0.24]:5701","liteMember":false,"localMember":true,"uuid":"dc476cf0-74cd-4c8b-987c-c0bec27fbd26","memberVersion":"5.5.2"}],"connectionCount":1,"allConnectionCount":2}⏎  
```
If you can access the map and retrieve values using the REST API, it means the secrets are working correctly.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,
> Note: The `username` must be fixed as `admin`.
```shell
$ kubectl create secret generic hazelcast-quickstart-usergen-auth -n demo \
                                               --type=kubernetes.io/basic-auth \
                                               --from-literal=username=admin \
                                               --from-literal=password=test-password
secret/hazelcast-quickstart-usergen-auth created
```
Now create a `HazelcastOpsRequest` with `RotateAuth` type. Below is the YAML of the `HazelcastOpsRequest` that we are going to create,
```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-rotate-auth-user-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: hazelcast-quickstart
  authentication:
    secretRef:
      name: hazelcast-quickstart-usergen-auth
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `hazelcast-quickstart`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Hazelcast.
- `spec.authentication.secretRef.name` specifies that we want to use `hazelcast-quickstart-usergen-auth` for database authentication.


Let's create the `HazelcastOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{ .version }}/docs/examples/hazelcast/rotate-auth/rotate-auth-user-generated.yaml
hazelcastopsrequest.ops.kubedb.com/hzops-rotate-auth-user-generated created
```
Let's wait for `HazelcastOpsrequest` to be `Successful`. Run the following command to watch `HazelcastOpsrequest` CRO
```shell
$ kubectl get Hazelcastopsrequest -n demo
NAME                               TYPE         STATUS       AGE
hzops-rotate-auth-user-generated   RotateAuth   Successful   2m32s
```
If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Hazelcastopsrequest -n demo hzops-rotate-auth-user-generated
Name:         hzops-rotate-auth-user-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-18T07:12:08Z
  Generation:          1
  Resource Version:    370582
  UID:                 973d9598-0746-4434-9748-bc90cb6331ef
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  hazelcast-quickstart-usergen-auth
  Database Ref:
    Name:  hazelcast-quickstart
  Type:    RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-08-18T07:12:08Z
    Message:               hazelcast ops request has started to rotate auth for rmq nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-08-18T07:12:11Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-08-18T07:12:21Z
    Message:               successfully reconciled the hazelcast with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-18T07:13:21Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-18T07:12:31Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-quickstart-0
    Last Transition Time:  2025-08-18T07:12:31Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-quickstart-1
    Last Transition Time:  2025-08-18T07:12:31Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-18T07:12:41Z
    Message:               running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hazelcast-quickstart-0
    Last Transition Time:  2025-08-18T07:13:11Z
    Message:               running pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hazelcast-quickstart-1
    Last Transition Time:  2025-08-18T07:13:21Z
    Message:               Successfully completed reconfigure Ignite
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           5m43s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-rotate-auth-user-generated
  Normal   Starting                                                           5m43s  KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hazelcast-quickstart
  Normal   Successful                                                         5m43s  KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hazelcast-quickstart for HazelcastOpsRequest: hzops-rotate-auth-user-generated
  Normal   UpdateStatefulSets                                                 5m30s  KubeDB Ops-manager Operator  successfully reconciled the hazelcast with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0      5m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1      5m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
  Warning  running pod; ConditionStatus:False                                 5m20s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  5m10s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:False                                 5m10s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  5m     KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:False                                 5m     KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  4m50s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:False                                 4m50s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  4m40s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-1  4m40s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
  Normal   RestartNodes                                                       4m30s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           4m30s  KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hazelcast-quickstart
  Normal   Successful                                                         4m30s  KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hazelcast-quickstart for HazelcastOpsRequest: hzops-rotate-auth-user-generated
```
**Verify Auth is rotated**
````shell
$ kubectl get hz -n demo hazelcast-quickstart -ojson | jq .spec.authSecret.name
"hazelcast-quickstart-auth"
$ kubectl get secret -n demo hazelcast-quickstart-usergen-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎          
$ kubectl get secret -n demo hazelcast-quickstart-usergen-auth -o jsonpath='{.data.password}' | base64 -d
test-password⏎
````
Now, you can exec into the pod `hazelcast-quickstart-0` and
run a REST api using `username` and `password`
```bash
$ kubectl exec -it -n demo hazelcast-quickstart-0 -c hazelcast -- curl -u admin:'test-password' http://localhost:5701/hazelcast/rest/cluster
{"members":[{"address":"[10.244.0.26]:5701","liteMember":false,"localMember":true,"uuid":"dc476cf0-74cd-4c8b-987c-c0bec27fbd26","memberVersion":"5.5.2"},{"address":"[10.244.0.27]:5701","liteMember":false,"localMember":false,"uuid":"9490ac0d-6d0c-437d-898c-c6c6aa81402e","memberVersion":"5.5.2"}],"connectionCount":1,"allConnectionCount":2}⏎  
```
If you can access the map and retrieve values using the REST API, it means the secrets are
working correctly.

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo hazelcast-quickstart-usergen-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
CYIpaMGLwfHmvA!h⏎           
$ kubectl get secret -n demo hazelcast-quickstart-usergen-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
admin⏎                           
```
## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Hazelcastopsrequest hzops-rotate-auth-generated hzops-rotate-auth-user-generated -n demo
$ kubectl delete secret -n demo hazelcast-quickstart-usergen-auth
$ kubectl delete secret -n demo  hazelcast-quickstart-auth 
$ kubectl delete hz -n demo hazelcast-quickstart
$ kubectl delete ns demo
```
