---
title: Cassandra Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: cas-rotate-auth-cassandra
    name: Cassandra RotateAuth Guide
    parent: cas-rotate-auth
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Cassandra Authentication

KubeDB supports rotating Authentication for existing Cassandra via a CassandraOpsRequest. There are two ways to do that.
1. **Operator Generated**: User will not provide any secret. KubeDB operator will generate a random password and update the existing secret with that password.
2. **User Defined**: User can create a `kubernetes.io/basic-auth` type secret with `username` and  `password` and refers this to `CassandraOpsRequest`.

This tutorial will show you how to use KubeDB to rotate authentication credentials.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Create Cassandra with Enabling Authentication

In this section, we are going to deploy a Cassandra topology cluster with authentication enabled. In the next few sections we will rotate the authentication using `CassandraOpsRequest` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 2Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/rotate-auth/cassandra-prod.yaml
cassandra.kubedb.com/cassandra-prod created
```

Now, wait until `cassandra-prod` has status `Ready`. i.e,

```bash
$ kubectl get cas -n demo -w
NAME              TYPE            VERSION   STATUS         AGE
cassandra-prod    kubedb.com/v1   5.0.3     Provisioning   3s
cassandra-prod    kubedb.com/v1   5.0.3     Provisioning   10s
.
.
cassandra-prod    kubedb.com/v1   5.0.3     Ready          2m13s
```

We can verify from the above output that authentication is enabled for this cluster. By default, KubeDB operator create default credentials for the Cassandra cluster. The default credentials are stored in a secret named `<cassandra-name>-auth` in the same namespace as the Cassandra cluster. You can find the secret by running the following command:

```bash
$ kubectl get cas -n demo cassandra-prod -ojson | jq .spec.authSecret.name
"cassandra-prod-auth"

$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.username}' | base64 -d
admin

$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.password}' | base64 -d
UajtzLlDwiizuHoV
```

### Create RotateAuth CassandraOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the cassandra using operator generated, we have to create a `CassandraOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `CassandraOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: cassandra-prod
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on cassandra.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/rotate-auth/cassandra-rotate-auth-generated.yaml
cassandraopsrequest.ops.kubedb.com/casops-rotate-auth-generated created
```

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CRO,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                          TYPE         STATUS       AGE
casops-rotate-auth-generated   RotateAuth   Successful   3m18s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe cassandraopsrequest -n demo casops-rotate-auth-generated 
Name:         casops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-21T05:30:38Z
  Generation:          1
  Resource Version:    76198
  UID:                 d20026ed-ca6c-442f-add8-10122a3b317b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-21T05:30:38Z
    Message:               Cassandra ops-request has started to rotate auth for cassandra nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-21T05:30:41Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-21T05:30:46Z
    Message:               successfully reconciled the Cassandra with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-21T05:33:36Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-21T05:30:51Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-21T05:30:51Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-21T05:30:56Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-21T05:31:36Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-21T05:31:36Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-21T05:33:36Z
    Message:               Successfully completed reconfigure cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           6m11s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-rotate-auth-generated
  Normal   Starting                                                           6m11s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         6m11s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate-auth-generated
  Normal   UpdatePetSets                                                      6m3s   KubeDB Ops-manager Operator  successfully reconciled the Cassandra with updated version
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    5m58s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  5m58s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 5m53s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    5m13s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  5m13s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    4m33s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  4m33s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    3m53s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  3m53s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartNodes                                                       3m13s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           3m13s  KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         3m13s  KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate-auth-generated
```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
$ kubectl get cas -n demo cassandra-prod -ojson | jq .spec.authSecret.name
"cassandra-prod-auth"

$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.username}' | base64 -d
admin

$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.password}' | base64 -d
t0jL7;5CFWhqn~3o
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath="{.data.username\.prev}" | base64 -d
admin
$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath="{.data.password\.prev}" | base64 -d
UajtzLlDwiizuHoV
```

The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.

#### 2. Using user created credentials

At first, we need to create a secret with `kubernetes.io/basic-auth` type using custom `username` and `password`. Below is the command to create a secret with `kubernetes.io/basic-auth` type,

```bash
$ kubectl create secret generic cassandra-user-auth -n demo \
          --type=kubernetes.io/basic-auth \
          --from-literal=username=cassandra \
          --from-literal=password=cassandra-secret
secret/cassandra-user-auth created
```

Now create a Cassandra Ops Request with `RotateAuth` type. Below is the YAML of the `CassandraOpsRequest` that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: cassandra-prod
  authentication:
    secretRef:
      kind: Secret
      name: cassandra-user-auth
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on cassandra.
- `spec.authentication.secretRef.name` specifies that we are using `cassandra-user-auth` secret for authentication.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/rotate-auth/cassandra-rotate-auth-user.yaml
cassandraopsrequest.ops.kubedb.com/casops-rotate-auth-user created
```

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CRO,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                          TYPE         STATUS       AGE
casops-rotate-auth-generated   RotateAuth   Successful   53m
casops-rotate-auth-user        RotateAuth   Successful   2m58s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe cassandraopsrequest -n demo casops-rotate-auth-user 
Name:         casops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-21T05:47:56Z
  Generation:          1
  Resource Version:    78421
  UID:                 63a3baf4-f0da-4b74-a883-0d374168bf92
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  cassandra-user-auth
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-21T05:49:26Z
    Message:               Cassandra ops-request has started to rotate auth for cassandra nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-21T05:49:29Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-21T05:49:34Z
    Message:               successfully reconciled the Cassandra with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-21T05:52:19Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-21T05:49:39Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-21T05:49:39Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-21T05:49:44Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-21T05:50:19Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-21T05:50:19Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-21T05:52:19Z
    Message:               Successfully completed reconfigure cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
sabbir@sabbir-pc ~/g/s/k/docs (cas)> kubectl get cassandraopsrequest
No resources found in default namespace.
sabbir@sabbir-pc ~/g/s/k/docs (cas)> kubectl get cassandraopsrequest -n demo
NAME                      TYPE         STATUS       AGE
casops-rotate-auth-user   RotateAuth   Successful   6h5m
sabbir@sabbir-pc ~/g/s/k/docs (cas)> kubectl describe cassandraopsrequest -n demo casops-rotate-auth-user 
Name:         casops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-21T05:47:56Z
  Generation:          1
  Resource Version:    78421
  UID:                 63a3baf4-f0da-4b74-a883-0d374168bf92
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  cassandra-user-auth
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-07-21T05:49:26Z
    Message:               Cassandra ops-request has started to rotate auth for cassandra nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-07-21T05:49:29Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-07-21T05:49:34Z
    Message:               successfully reconciled the Cassandra with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-21T05:52:19Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-21T05:49:39Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-21T05:49:39Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-21T05:49:44Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-21T05:50:19Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-21T05:50:19Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-21T05:52:19Z
    Message:               Successfully completed reconfigure cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
$ kubectl get cas -n demo cassandra-prod -ojson | jq .spec.authSecret.name
"cassandra-user-auth"

$ kubectl get secret -n demo cassandra-user-auth -o=jsonpath='{.data.username}' | base64 -d
cassandra

$ kubectl get secret -n demo cassandra-user-auth -o=jsonpath='{.data.password}' | base64 -d
cassandra-secret
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo cassandra-user-auth -o=jsonpath="{.data.username\.prev}" | base64 -d
admin
$ kubectl get secret -n demo cassandra-user-auth -o=jsonpath="{.data.password\.prev}" | base64 -d
rM4OJfqoTzvKMAx8
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cassandraopsrequest -n demo casops-rotate-auth-generated casops-rotate-auth-user
kubectl delete cassandra -n demo cassandra-prod
kubectl delete secret -n demo cassandra-user-auth
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

