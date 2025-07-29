---
title: Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: es-rotateauth
    name: Rotate Authentication Guide
    parent: es-rotateauth-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Rotate Authentication of Elasticsearch

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Elasticsearch` user's authentication credentials using a `ElasticsearchOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential, and does not provide the secret details directly to the user.
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `username` and `password`, and then reference this Secret in the `ElasticsearchOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Create a Elasticsearch database
The KubeDB operator implements an Elasticsearch CRD to define the specification of an Elasticsearch database.

The Elasticsearch instance used for this tutorial:

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: sample-es
  namespace: demo
spec:
  version: xpack-8.11.1
  storageType: Durable
  topology:
    master:
      suffix: master
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      suffix: data
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      suffix: client
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi

```
Let's create the above `Elasticsearch` object,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/stash/kubedb/examples/elasticsearch/sample_es.yaml
elasticsearch.kubedb.com/sample-es created
```


Now, wait until `sample-es` has status Ready. i.e,

```shell
$ kubectl get Elasticsearch -n demo -w
NAME        VERSION        STATUS   AGE
sample-es   xpack-8.11.1   Ready    3m12s
```
## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get es -n demo sample-es -ojson | jq .spec.authSecret.name
"sample-es-auth"
$ kubectl get secret -n demo sample-es-auth -o jsonpath='{.data.username}' | base64 -d
elastic⏎                                                                                                
$ kubectl get secret -n demo sample-es-auth -o jsonpath='{.data.password}' | base64 -d
l;)1knmenzgH0c2M⏎                                                                                                                                               
````
Now, you can exec into the pod `sample-es` and connect to database using `username` and `password`.

**Port-forward the Service**
At first, let’s port-forward the `sample-es` Service. Run the following command into a separate terminal.
```shell
$ kubectl port-forward -n demo service/sample-es 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

**Insert dat**

```shell
$ curl -XPOST --user "elastic:l;)1knmenzgH0c2M" "http://localhost:9200/products/_doc?pretty" -H 'Content-Type: application/json' -d'
                              {
                                  "name": "KubeDB",
                                  "vendor": "AppsCode Inc.",
                                  "description": "Database Operator for Kubernetes"
                              }
                              '
```
You'll see:
```shell
{
  "_index" : "products",
  "_id" : "1tCNEpgBr20vA8aRT9BX",
  "_version" : 1,
  "result" : "created",
  "_shards" : {
    "total" : 2,
    "successful" : 1,
    "failed" : 0
  },
  "_seq_no" : 0,
  "_primary_term" : 1
}
```
If you can access the data table and run queries, it means the secrets are working correctly.
## Create RotateAuth ElasticsearchOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Postgres using operator generated, we have to create a `ElasticsearchOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `ElasticsearchOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: essops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-es
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `sample-es` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Elasticsearch.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/rotate-auth/yamls/rotate-auth-generated.yaml
 Elasticsearchopsrequest.ops.kubedb.com/essops-rotate-auth-generated created
```
Let's wait for `ElasticsearchOpsrequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsrequest` CRO
```shell
$ kubectl get esops -n demo
NAME                           TYPE         STATUS       AGE
essops-rotate-auth-generated   RotateAuth   Successful   7m12s

```
If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Elasticsearchopsrequest -n demo essops-rotate-auth-generated
Name:         essops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-07-16T09:35:47Z
  Generation:          1
  Resource Version:    797065
  UID:                 5c642055-67d9-47d8-ad55-119a33720723
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   sample-es
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-16T09:35:47Z
    Message:               Elasticsearch ops request is updating database version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2025-07-16T09:35:50Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-16T09:36:02Z
    Message:               Successfully updated petSets
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-16T09:36:07Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-client-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-client-0
    Last Transition Time:  2025-07-16T09:36:07Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-client-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-client-0
    Last Transition Time:  2025-07-16T09:36:07Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-client-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-client-0
    Last Transition Time:  2025-07-16T09:39:47Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-07-16T09:36:57Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-client-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-client-1
    Last Transition Time:  2025-07-16T09:36:57Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-client-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-client-1
    Last Transition Time:  2025-07-16T09:36:57Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-client-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-client-1
    Last Transition Time:  2025-07-16T09:37:47Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-data-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-data-0
    Last Transition Time:  2025-07-16T09:37:47Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-data-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-data-0
    Last Transition Time:  2025-07-16T09:37:47Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-data-0
    Last Transition Time:  2025-07-16T09:38:37Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-data-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-data-1
    Last Transition Time:  2025-07-16T09:38:37Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-data-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-data-1
    Last Transition Time:  2025-07-16T09:38:37Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-data-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-data-1
    Last Transition Time:  2025-07-16T09:39:02Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-master-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-master-0
    Last Transition Time:  2025-07-16T09:39:02Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-master-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-master-0
    Last Transition Time:  2025-07-16T09:39:02Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-master-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-master-0
    Last Transition Time:  2025-07-16T09:39:52Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-16T09:39:52Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                              Age    From                         Message
  ----     ------                                                              ----   ----                         -------
  Normal   PauseDatabase                                                       9m19s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/sample-es
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-client-0        8m59s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-client-0
  Warning  create es client; ConditionStatus:True; PodName:sample-es-client-0  8m59s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-client-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-client-0         8m59s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-client-0
  Warning  create es client; ConditionStatus:False                             8m54s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              8m14s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-client-1        8m9s   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-client-1
  Warning  create es client; ConditionStatus:True; PodName:sample-es-client-1  8m9s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-client-1
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-client-1         8m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-client-1
  Warning  create es client; ConditionStatus:False                             8m4s   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              7m24s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-data-0          7m19s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-data-0
  Warning  create es client; ConditionStatus:True; PodName:sample-es-data-0    7m19s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-data-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-data-0           7m19s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-data-0
  Warning  create es client; ConditionStatus:False                             7m14s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              6m34s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-data-1          6m29s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-data-1
  Warning  create es client; ConditionStatus:True; PodName:sample-es-data-1    6m29s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-data-1
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-data-1           6m29s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-data-1
  Warning  create es client; ConditionStatus:False                             6m24s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              6m9s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-master-0        6m4s   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-master-0
  Warning  create es client; ConditionStatus:True; PodName:sample-es-master-0  6m4s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-master-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-master-0         6m4s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-master-0
  Warning  create es client; ConditionStatus:False                             5m59s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              5m19s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   RestartNodes                                                        5m14s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   ResumeDatabase                                                      5m14s  KubeDB Ops-manager Operator  Resuming Elasticsearch
  Normal   ResumeDatabase                                                      5m14s  KubeDB Ops-manager Operator  Resuming Elasticsearch demo/sample-es
  Normal   ResumeDatabase                                                      5m14s  KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/sample-es
  Normal   Successful                                                          5m14s  KubeDB Ops-manager Operator  Successfully updated authsecret.

```
**Verify Auth is rotated**
```shell
````shell
$ kubectl get es -n demo sample-es -ojson | jq .spec.authSecret.name
"sample-es-auth"
$ kubectl get secret -n demo sample-es-auth -o jsonpath='{.data.username}' | base64 -d
elastic⏎                                                                            
$ kubectl get secret -n demo sample-es-auth -o jsonpath='{.data.password}' | base64 -d
k3pcRRtJi8iMhlVy⏎                     
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo sample-es-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
elastic⏎                                                                                                          
$ kubectl get secret -n demo sample-es-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
l;)1knmenzgH0c2M⏎                        
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

> Note: `username` must be `elastic`.

```shell
$ kubectl create secret generic sample-es-auth-user -n demo \
   --type=kubernetes.io/basic-auth \
   --from-literal=username=elastic \
   --from-literal=password=testpassword
secret/sample-es-auth-user created
```
Now create a `ElasticsearchOpsRequest` with `RotateAuth` type. Below is the YAML of the `ElasticsearchOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: sample-es
  authentication:
    secretRef:
      name: sample-es-auth-user
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `sample-es`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on postgres.
- `spec.authentication.secretRef.name` specifies that we are using `sample-es-auth-user` as `spec.authSecret.name` for authentication.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/rotate-auth/yamls/rotate-auth-user.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-rotate-auth-user created
```
Let’s wait for `ElasticsearchOpsRequest` to be Successful. Run the following command to watch `ElasticsearchOpsRequest` CRO:

```shell
$ kubectl get Elasticsearchopsrequest -n demo
NAME                          TYPE         STATUS       AGE
essops-rotate-auth-generated  RotateAuth   Successful   100s
esops-rotate-auth-user        RotateAuth   Successful   62s
```
We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed.
```shell
$kubectl describe Elasticsearchopsrequest -n demo esops-rotate-auth-user
Name:         esops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-07-16T10:50:03Z
  Generation:          1
  Resource Version:    804361
  UID:                 a7510ebe-b026-4070-ae29-e309be7781ad
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  sample-es-auth-user
  Database Ref:
    Name:   sample-es
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-16T10:50:03Z
    Message:               Elasticsearch ops request is updating database version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2025-07-16T10:50:03Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-16T10:50:15Z
    Message:               Successfully updated petSets
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-16T10:50:20Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-client-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-client-0
    Last Transition Time:  2025-07-16T10:50:20Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-client-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-client-0
    Last Transition Time:  2025-07-16T10:50:20Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-client-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-client-0
    Last Transition Time:  2025-07-16T10:54:20Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-07-16T10:51:05Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-client-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-client-1
    Last Transition Time:  2025-07-16T10:51:05Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-client-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-client-1
    Last Transition Time:  2025-07-16T10:51:05Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-client-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-client-1
    Last Transition Time:  2025-07-16T10:52:00Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-data-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-data-0
    Last Transition Time:  2025-07-16T10:52:00Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-data-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-data-0
    Last Transition Time:  2025-07-16T10:52:00Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-data-0
    Last Transition Time:  2025-07-16T10:52:50Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-data-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-data-1
    Last Transition Time:  2025-07-16T10:52:50Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-data-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-data-1
    Last Transition Time:  2025-07-16T10:52:50Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-data-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-data-1
    Last Transition Time:  2025-07-16T10:53:40Z
    Message:               pod exists; ConditionStatus:True; PodName:sample-es-master-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--sample-es-master-0
    Last Transition Time:  2025-07-16T10:53:40Z
    Message:               create es client; ConditionStatus:True; PodName:sample-es-master-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--sample-es-master-0
    Last Transition Time:  2025-07-16T10:53:40Z
    Message:               evict pod; ConditionStatus:True; PodName:sample-es-master-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--sample-es-master-0
    Last Transition Time:  2025-07-16T10:54:25Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-16T10:54:25Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                              Age    From                         Message
  ----     ------                                                              ----   ----                         -------
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-client-0        6m12s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-client-0
  Warning  create es client; ConditionStatus:True; PodName:sample-es-client-0  6m12s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-client-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-client-0         6m12s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-client-0
  Warning  create es client; ConditionStatus:False                             6m7s   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              5m32s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-client-1        5m27s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-client-1
  Warning  create es client; ConditionStatus:True; PodName:sample-es-client-1  5m27s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-client-1
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-client-1         5m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-client-1
  Warning  create es client; ConditionStatus:False                             5m22s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              4m37s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-data-0          4m32s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-data-0
  Warning  create es client; ConditionStatus:True; PodName:sample-es-data-0    4m32s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-data-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-data-0           4m32s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-data-0
  Warning  create es client; ConditionStatus:False                             4m27s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              3m47s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-data-1          3m42s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-data-1
  Warning  create es client; ConditionStatus:True; PodName:sample-es-data-1    3m42s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-data-1
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-data-1           3m42s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-data-1
  Warning  create es client; ConditionStatus:False                             3m37s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              2m57s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:sample-es-master-0        2m52s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:sample-es-master-0
  Warning  create es client; ConditionStatus:True; PodName:sample-es-master-0  2m52s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:sample-es-master-0
  Warning  evict pod; ConditionStatus:True; PodName:sample-es-master-0         2m52s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:sample-es-master-0
  Warning  create es client; ConditionStatus:False                             2m47s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                              2m12s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   RestartNodes                                                        2m7s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   ResumeDatabase                                                      2m7s   KubeDB Ops-manager Operator  Resuming Elasticsearch
  Normal   ResumeDatabase                                                      2m7s   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/sample-es
  Normal   ResumeDatabase                                                      2m7s   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/sample-es
  Normal   Successful                                                          2m7s   KubeDB Ops-manager Operator  Successfully updated authsecret.
```
**Verify auth is rotate**
```shell
$ kubectl get Elasticsearch -n demo sample-es -ojson | jq .spec.authSecret.name
"sample-es-auth-user"
$ kubectl get secret -n demo sample-es-auth-user -o jsonpath='{.data.username}' | base64 -d
elastic⏎ 
$ kubectl get secret -n demo sample-es-auth-user -o jsonpath='{.data.password}' | base64 -d
testpassword⏎                                                                                                        
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo sample-es-auth-user -o go-template='{{ index .data "username.prev" }}' | base64 -d
elastic                                                        
$ kubectl get secret -n demo quick-mg-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
k3pcRRtJi8iMhlVy⏎                                             
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Elasticsearchopsrequest essps-rotate-auth-generated esops-rotate-auth-user -n demo
Elasticsearchopsrequest.ops.kubedb.com "essops-rotate-auth-generated" deleted
Elasticsearchopsrequest.ops.kubedb.com "esops-rotate-auth-user" deleted
$ kubectl delete secret -n demo  sample-es-auth-user
secret "sample-es-auth-user" deleted
$ kubectl delete secret -n demo  sample-es-auth
secret "sample-es-auth" deleted
```
## Next Steps

- [Quickstart Kibana](/docs/guides/elasticsearch/elasticsearch-dashboard/kibana/index.md) with KubeDB Operator.
- Learn how to configure [Elasticsearch Topology Cluster](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md).
- Learn about [backup & restore](/docs/guides/elasticsearch/backup/stash/overview/index.md) Elasticsearch database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
