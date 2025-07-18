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
bT1qvxvpXWgnmDzu
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

> **Note:** For combined cassandra, you just need to refer cassandra combined object in `databaseRef` field. To learn more about combined cassandra, please visit [here](/docs/guides/cassandra/clustering/combined-cluster/index.md).

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
  Creation Timestamp:  2025-05-15T11:11:04Z
  Generation:          1
  Resource Version:    290550
  UID:                 71ff7cec-f895-424c-b14f-9b957ccf9ccd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-05-15T11:11:04Z
    Message:               Cassandra ops-request has started to rotate auth for cassandra nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-05-15T11:11:07Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-05-15T11:11:15Z
    Message:               successfully reconciled the Cassandra with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-05-15T11:11:20Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-controller-0
    Last Transition Time:  2025-05-15T11:11:20Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-controller-0
    Last Transition Time:  2025-05-15T11:11:55Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-controller-0
    Last Transition Time:  2025-05-15T11:12:00Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-controller-1
    Last Transition Time:  2025-05-15T11:12:00Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-controller-1
    Last Transition Time:  2025-05-15T11:12:35Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-controller-1
    Last Transition Time:  2025-05-15T11:12:40Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-broker-0
    Last Transition Time:  2025-05-15T11:12:40Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-broker-0
    Last Transition Time:  2025-05-15T11:13:15Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-broker-0
    Last Transition Time:  2025-05-15T11:13:20Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-broker-1
    Last Transition Time:  2025-05-15T11:13:20Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-broker-1
    Last Transition Time:  2025-05-15T11:13:55Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-broker-1
    Last Transition Time:  2025-05-15T11:14:00Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-05-15T11:14:00Z
    Message:               Successfully completed reconfigure cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   3m51s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-rotate-auth-generated
  Normal   Starting                                                                   3m51s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                                 3m51s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate-auth-generated
  Normal   UpdatePetSets                                                              3m40s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-0             3m35s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-0           3m35s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-0  3m30s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-0   3m     KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-1             2m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-1           2m55s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-1  2m50s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-1   2m20s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-0                 2m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-0               2m15s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-0      2m10s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-0       100s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-1                 95s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-1               95s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-1      90s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-1       60s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-1
  Normal   RestartNodes                                                               55s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   55s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                                 55s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate-auth-generated
```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
$ kubectl get cas -n demo cassandra-prod -ojson | jq .spec.authSecret.name
"cassandra-prod-auth"

$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.username}' | base64 -d
admin

$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.password}' | base64 -d
al9jY2xvYW5pbmc=
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.username.prev}' | base64 -d
admin
$ kubectl get secret -n demo cassandra-prod-auth -o=jsonpath='{.data.password.prev}' | base64 -d
zvrFXkStB~9A!NTC
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
      name: cassandra-user-auth
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on cassandra.
- `spec.authentication.secretRef.name` specifies that we are using `cassandra-user-auth` secret for authentication.

Let's create the `CassandraOpsRequest` CR we have shown above,

> **Note:** For combined cassandra, you just need to refer cassandra combined object in `databaseRef` field. To learn more about combined cassandra, please visit [here](/docs/guides/cassandra/clustering/combined-cluster/index.md).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/rotate-auth/cassandra-rotate-auth-user.yaml
cassandraopsrequest.ops.kubedb.com/casops-rotate-auth-user created
```

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CRO,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                          TYPE         STATUS       AGE
casops-rotate-auth-generated   RotateAuth   Successful   83m
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
  Creation Timestamp:  2025-05-15T12:31:13Z
  Generation:          1
  Resource Version:    310786
  UID:                 13513a65-ac25-4667-8a11-80e356500c53
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
    Last Transition Time:  2025-05-15T12:31:13Z
    Message:               Cassandra ops-request has started to rotate auth for cassandra nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-05-15T12:31:16Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-05-15T12:31:24Z
    Message:               successfully reconciled the Cassandra with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-05-15T12:31:29Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-controller-0
    Last Transition Time:  2025-05-15T12:31:29Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-controller-0
    Last Transition Time:  2025-05-15T12:32:04Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-controller-0
    Last Transition Time:  2025-05-15T12:32:09Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-controller-1
    Last Transition Time:  2025-05-15T12:32:09Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-controller-1
    Last Transition Time:  2025-05-15T12:32:44Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-controller-1
    Last Transition Time:  2025-05-15T12:32:49Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-broker-0
    Last Transition Time:  2025-05-15T12:32:49Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-broker-0
    Last Transition Time:  2025-05-15T12:33:24Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-broker-0
    Last Transition Time:  2025-05-15T12:33:29Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-broker-1
    Last Transition Time:  2025-05-15T12:33:29Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-broker-1
    Last Transition Time:  2025-05-15T12:34:04Z
    Message:               check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--cassandra-prod-broker-1
    Last Transition Time:  2025-05-15T12:34:09Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-05-15T12:34:09Z
    Message:               Successfully completed reconfigure cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   3m17s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-rotate-auth-user
  Normal   Starting                                                                   3m17s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                                 3m17s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate-auth-user
  Normal   UpdatePetSets                                                              3m6s   KubeDB Ops-manager Operator  successfully reconciled the Cassandra with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-0             3m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-0           3m1s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-0  2m56s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-0   2m26s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-1             2m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-1           2m21s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-1  2m16s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-1   106s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-0                 101s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-0               101s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-0      96s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-0       66s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-1                 61s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-1               61s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-1      56s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:cassandra-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-1       26s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:cassandra-prod-broker-1
  Normal   RestartNodes                                                               21s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   21s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                                 21s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate-auth-user
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
$ kubectl get secret -n demo cassandra-user-auth -o=jsonpath='{.data.username.prev}' | base64 -d
admin
$ kubectl get secret -n demo cassandra-user-auth -o=jsonpath='{.data.password.prev}' | base64 -d
al9jY2xvYW5pbmc=
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
- Different Cassandra topology clustering modes [here](/docs/guides/cassandra/clustering/_index.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).
- Cassandra ConnectCluster with KubeDB [here](/docs/guides/cassandra/connectcluster/quickstart.md).
- Cassandra Schema Registry with KubeDB [here](/docs/guides/cassandra/schemaregistry/overview.md).
- Cassandra RestProxy with KubeDB [here](/docs/guides/cassandra/restproxy/overview.md).
- Cassandra Migration with KubeDB [here](/docs/guides/cassandra/migration/overview.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

