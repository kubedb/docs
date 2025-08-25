---
title: Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: sl-rotate-auth-details
    name: Guide
    parent: sl-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Rotate Authentication of Solr

**Rotate Authentication** is a feature of the KubeDB Ops-Manager that allows you to rotate a `Solr` user's authentication credentials using a `SolrOpsRequest`. There are two ways to perform this rotation.

1. **Operator Generated:** The KubeDB operator automatically generates a random credential, updates the existing secret with the new credential The KubeDB operator automatically generates a random credential and updates the existing secret with the new credential..
2. **User Defined:** The user can create their own credentials by defining a Secret of type `kubernetes.io/basic-auth` containing the desired `username` and `password`, and then reference this Secret in the `SolrOpsRequest`.


> Note: YAML files used in this tutorial are stored in [docs/guides/solr/quickstart/overview/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/solr/quickstart/overview/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> This tutorial demonstrates how to rotate authentication credentials for Solr managed by KubeDB. Before you begin, ensure that the Solr CRD is installed and running. If not, follow [this guide](/docs/guides/solr/quickstart/overview/index.md) to set it up.

## Verify authentication
The user can verify whether they are authorized by executing a query directly in the database. To do this, the user needs `username` and `password` in order to connect to the database using the `kubectl exec` command. Below is an example showing how to retrieve the credentials from the Secret.

````shell
$ kubectl get solr -n demo solr-combined -ojson | jq .spec.authSecret.name
"solr-combined-auth"
$ kubectl get secret -n demo solr-combined-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎           
$ kubectl get secret -n demo solr-combined-auth -o jsonpath='{.data.password}' | base64 -d
QtnsJluRRjaaWWec⏎             
````

## Create RotateAuth SolrOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the Solr using operator generated, we have to create a `SolrOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `SolrOpsRequest` CRO that we are going to create,
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: solrops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: solr-combined
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `solr-combined` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Solr.

Let's create the `SolrOpsRequest` CR we have shown above,
```shell
 $ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/rotate-auth/rotate-auth.yaml
solropsrequest.ops.kubedb.com/solrops-rotate-auth-generated created
```
Let's wait for `SolrOpsrequest` to be `Successful`. Run the following command to watch `SolrOpsrequest` CRO
```shell
 $ kubectl get Solropsrequest -n demo
 NAME                          TYPE         STATUS       AGE
solrops-rotate-auth-generated   RotateAuth   Successful    2m3s
```
If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Solropsrequest -n demo solrops-rotate-auth-generated 
Name:         solrops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2025-07-21T05:45:10Z
  Generation:          1
  Resource Version:    151277
  UID:                 eb1695d2-2354-4f12-9dc2-cf14f55031e9
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   solr-combined
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-21T05:45:10Z
    Message:               Solr ops-request has started to rotate auth for solr nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-21T05:45:13Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-21T05:45:20Z
    Message:               successfully reconciled the Solr with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-21T05:47:13Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-21T05:45:25Z
    Message:               get pod; ConditionStatus:True; PodName:solr-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-combined-0
    Last Transition Time:  2025-07-21T05:45:25Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-combined-0
    Last Transition Time:  2025-07-21T05:45:30Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-21T05:46:10Z
    Message:               get pod; ConditionStatus:True; PodName:solr-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-combined-1
    Last Transition Time:  2025-07-21T05:46:10Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-combined-1
    Last Transition Time:  2025-07-21T05:47:13Z
    Message:               Successfully completed reconfigure solr
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  7m23s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/solrops-rotate-auth-generated
  Normal   Starting                                                  7m23s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-combined
  Normal   Successful                                                7m23s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-combined for SolrOpsRequest: solrops-rotate-auth-generated
  Normal   UpdatePetSets                                             7m13s  KubeDB Ops-manager Operator  successfully reconciled the Solr with updated version
  Warning  get pod; ConditionStatus:True; PodName:solr-combined-0    7m8s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-combined-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-combined-0  7m8s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-combined-0
  Warning  running pod; ConditionStatus:False                        7m3s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-combined-1    6m23s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-combined-1
  Warning  evict pod; ConditionStatus:True; PodName:solr-combined-1  6m23s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-combined-1
  Normal   RestartNodes                                              5m20s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   RestartNodes                                              5m20s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                  5m20s  KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-combined
  Normal   Successful                                                5m20s  KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-combined for SolrOpsRequest: solrops-rotate-auth-generated

```
**Verify Auth is rotated**
```shell
$  kubectl get solr -n demo solr-combined -ojson | jq .spec.authSecret.name
"solr-combined-auth"
$ kubectl get secret -n demo solr-combined-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎ 
$ kubectl get secret -n demo solr-combined-auth -o jsonpath='{.data.password}' | base64 -d
dt(MVdBeBDlEy~Cp⏎                                    
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```shell
$ kubectl get secret -n demo solr-combined-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
admin⏎           
$ kubectl get secret -n demo solr-combined-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
QtnsJluRRjaaWWec⏎              
```
The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.
#### 2. Using user created credentials

At first, we need to create a secret with kubernetes.io/basic-auth type using custom username and password. Below is the command to create a secret with kubernetes.io/basic-auth type,

```shell
$ kubectl create secret generic solr-combined-user-auth -n demo \
                                                --type=kubernetes.io/basic-auth \
                                                --from-literal=username=admin \
                                                --from-literal=password=Solr-secret
 secret/solr-combined-user-auth created
```
Now create a `SolrOpsRequest` with `RotateAuth` type. Below is the YAML of the `SolrOpsRequest` that we are going to create,

```shell
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: solrops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: solr-combined
  authentication:
    secretRef:
      name: solr-combined-user-auth
  timeout: 5m
  apply: IfReady
```
Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `solr-combined`cluster.
- `spec.type` specifies that we are performing `RotateAuth` on Solr.
- `spec.authentication.secretRef.name` specifies that we are using `solr-combined-user-auth` as `spec.authSecret.name` for authentication.

Let's create the `SolrOpsRequest` CR we have shown above,

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/rotate-auth/rotateauthuser.yaml
solropsrequest.ops.kubedb.com/solrops-rotate-auth-user created
```
Let’s wait for `SolrOpsRequest` to be Successful. Run the following command to watch `SolrOpsRequest` CRO:

```shell
$ kubectl get Solropsrequest -n demo
NAME                          TYPE         STATUS       AGE
solrops-rotate-auth-generated   RotateAuth   Successful    13m
solrops-rotate-auth-user        RotateAuth   Successful    2m3s
```
We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed.
```shell
$ kubectl describe Solropsrequest -n demo solrops-rotate-auth-user
Name:         solrops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2025-07-21T05:57:25Z
  Generation:          1
  Resource Version:    152942
  UID:                 35345a38-15f1-40d1-ae7f-155e4f9663d3
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  solr-combined-user-auth
  Database Ref:
    Name:   solr-combined
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-21T05:57:25Z
    Message:               Solr ops-request has started to rotate auth for solr nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-21T05:57:28Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-21T05:57:36Z
    Message:               successfully reconciled the Solr with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-21T05:59:27Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-21T05:57:41Z
    Message:               get pod; ConditionStatus:True; PodName:solr-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-combined-0
    Last Transition Time:  2025-07-21T05:57:41Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-combined-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-combined-0
    Last Transition Time:  2025-07-21T05:57:46Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-21T05:58:26Z
    Message:               get pod; ConditionStatus:True; PodName:solr-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-combined-1
    Last Transition Time:  2025-07-21T05:58:26Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-combined-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-combined-1
    Last Transition Time:  2025-07-21T05:59:28Z
    Message:               Successfully completed reconfigure solr
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age   From                         Message
  ----     ------                                                    ----  ----                         -------
  Normal   Starting                                                  13m   KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/solrops-rotate-auth-user
  Normal   Starting                                                  13m   KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-combined
  Normal   Successful                                                13m   KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-combined for SolrOpsRequest: solrops-rotate-auth-user
  Normal   UpdatePetSets                                             12m   KubeDB Ops-manager Operator  successfully reconciled the Solr with updated version
  Warning  get pod; ConditionStatus:True; PodName:solr-combined-0    12m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-combined-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-combined-0  12m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-combined-0
  Warning  running pod; ConditionStatus:False                        12m   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-combined-1    12m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-combined-1
  Warning  evict pod; ConditionStatus:True; PodName:solr-combined-1  12m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-combined-1
  Normal   RestartNodes                                              11m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                  11m   KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-combined
  Normal   Successful                                                11m   KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-combined for SolrOpsRequest: solrops-rotate-auth-user

```
**Verify auth is rotate**
```shell
$  kubectl get solr -n demo solr-combined -ojson | jq .spec.authSecret.name
"solr-combined-user-auth"
$ kubectl get secret -n demo solr-combined-user-auth -o jsonpath='{.data.username}' | base64 -d
solr⏎      
$ kubectl get secret -n demo solr-combined-user-auth -o jsonpath='{.data.password}' | base64 -d
Solr-secret⏎                                                                                     
```
Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:
```shell
$ kubectl get secret -n demo solr-combined-user-auth -o go-template='{{ index .data "username.prev" }}' | base64 -d
Solr                                                                                    
$ kubectl get secret -n demo solr-combined-user-auth -o go-template='{{ index .data "password.prev" }}' | base64 -d
dt(MVdBeBDlEy~Cp⏎   
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources you can delete the CRD or namespace.
Or, you can delete one by one resource by their name by this tutorial, run:

```shell
$ kubectl delete Solropsrequest solrops-rotate-auth-generated solrops-rotate-auth-user -n demo
Solropsrequest.ops.kubedb.com "solrops-rotate-auth-generated" "solrops-rotate-auth-user" deleted
$ kubectl delete secret -n demo  solr-combined-user-auth
secret "solr-combined-user-auth" deleted
$ kubectl delete secret -n demo  solr-combined-auth
secret "solr-combined-auth" deleted
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).
- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
