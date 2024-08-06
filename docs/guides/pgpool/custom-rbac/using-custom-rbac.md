---
title: Run Pgpool with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: pp-custom-rbac-quickstart
    name: Custom RBAC
    parent: pp-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a Pgpool instance. This tutorial will show you how to use KubeDB to run Pgpool instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for Pgpool. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in Pgpool crd.   If this field is left empty, the KubeDB operator will use the default service account. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a Pgpool instance named `pgpool` to provide the bare minimum access permissions.

## Custom RBAC for Pgpool

At first, let's create a `Service Acoount` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

It should create a service account.

```yaml
$ kubectl get serviceaccount -n demo my-custom-serviceaccount -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2024-08-01T04:55:42Z"
  name: my-custom-serviceaccount
  namespace: demo
  resourceVersion: "269793"
  uid: 21a51f82-84b8-47ee-ab80-4404778bc5ee
```

Now, we need to create a role that has necessary access permissions for the Pgpool instance named `pgpool`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/custom-rbac/mg-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Below is the YAML for the Role we just created.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-custom-role
  namespace: demo
rules:
- apiGroups:
  - policy
  resourceNames:
  - pgpool
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for Pgpool pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```bash
$ kubectl create rolebinding my-custom-rolebinding --role=my-custom-role --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created

```

It should bind `my-custom-role` and `my-custom-serviceaccount` successfully.

```yaml
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2024-08-01T04:59:20Z"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "270018"
  uid: bd6b4fe3-5b2e-4cc4-a51e-2eba0a3af5e3
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-custom-role
subjects:
  - kind: ServiceAccount
    name: my-custom-serviceaccount
    namespace: demo
```

Now, create a Pgpool crd specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/custom-rbac/pp-custom.yaml
pgpool.kubedb.com/pgpool created
```

Below is the YAML for the Pgpool crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  deletionPolicy: WipeOut
```

Now, wait a few minutes. the KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we should see that a pod with the name `pgpool-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo pgpool-0
NAME       READY   STATUS    RESTARTS   AGE
pgpool-0   1/1     Running   0          50s
```

Check the pod's log to see if the pgpool is ready

```bash
$ kubectl logs -f -n demo pgpool-0
Configuring Pgpool-II...
Custom pgpool.conf file detected. Use custom configuration files.
Generating pool_passwd...
Generating pcp.conf...
Custom pool_hba.conf file detected. Use custom pool_hba.conf.
Starting Pgpool-II...
2024-08-01 05:03:29.081: main pid 61: LOG:  Backend status file /tmp/pgpool_status does not exist
2024-08-01 05:03:29.081: main pid 61: LOG:  health_check_stats_shared_memory_size: requested size: 12288
2024-08-01 05:03:29.081: main pid 61: LOG:  memory cache initialized
2024-08-01 05:03:29.081: main pid 61: DETAIL:  memcache blocks :64
2024-08-01 05:03:29.081: main pid 61: LOG:  allocating (135894880) bytes of shared memory segment
2024-08-01 05:03:29.081: main pid 61: LOG:  allocating shared memory segment of size: 135894880 
2024-08-01 05:03:30.129: main pid 61: LOG:  health_check_stats_shared_memory_size: requested size: 12288
2024-08-01 05:03:30.129: main pid 61: LOG:  health_check_stats_shared_memory_size: requested size: 12288
2024-08-01 05:03:30.129: main pid 61: LOG:  memory cache initialized
2024-08-01 05:03:30.129: main pid 61: DETAIL:  memcache blocks :64
2024-08-01 05:03:30.130: main pid 61: LOG:  pool_discard_oid_maps: discarded memqcache oid maps
2024-08-01 05:03:30.150: main pid 61: LOG:  create socket files[0]: /tmp/.s.PGSQL.9999
2024-08-01 05:03:30.150: main pid 61: LOG:  listen address[0]: *
2024-08-01 05:03:30.150: main pid 61: LOG:  Setting up socket for 0.0.0.0:9999
2024-08-01 05:03:30.150: main pid 61: LOG:  Setting up socket for :::9999
2024-08-01 05:03:30.151: main pid 61: LOG:  find_primary_node_repeatedly: waiting for finding a primary node
2024-08-01 05:03:30.151: main pid 61: LOG:  create socket files[0]: /var/run/pgpool/.s.PGSQL.9595
2024-08-01 05:03:30.151: main pid 61: LOG:  listen address[0]: *
2024-08-01 05:03:30.151: main pid 61: LOG:  Setting up socket for 0.0.0.0:9595
2024-08-01 05:03:30.151: main pid 61: LOG:  Setting up socket for :::9595
2024-08-01 05:03:30.151: pcp_main pid 68: LOG:  PCP process: 68 started
2024-08-01 05:03:30.152: sr_check_worker pid 69: LOG:  process started
2024-08-01 05:03:30.152: health_check pid 70: LOG:  process started
2024-08-01 05:03:30.152: health_check pid 71: LOG:  process started
2024-08-01 05:03:30.153: main pid 61: LOG:  pgpool-II successfully started. version 4.5.0 (hotooriboshi)
```

Once we see `pgpool-II successfully started` in the log, the pgpool is ready.

Also, if we want to verify that the pod is actually using our custom service account we can just describe the pod and see the `Service Accouunt Name`,
```bash
$ kubectl describe pp -n demo pgpool                                                        
Name:         pgpool
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Pgpool
Metadata:
  Creation Timestamp:  2024-08-01T05:03:26Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  271249
  UID:               53b75d96-4e5c-45ec-bccd-6c2ca5a363ec
Spec:
  Auth Secret:
    Name:            pgpool-auth
  Client Auth Mode:  md5
  Deletion Policy:   WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  pgpool
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     70
          Run As Non Root:  true
          Run As User:      70
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:            70
      Service Account Name:  my-custom-serviceaccount
  Postgres Ref:
    Name:       ha-postgres
    Namespace:  demo
  Replicas:     1
  Ssl Mode:     disable
  Version:      4.5.0
Status:
  Conditions:
    Last Transition Time:  2024-08-01T05:03:27Z
    Message:               The KubeDB operator has started the provisioning of Pgpool: demo/pgpool
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-08-01T05:17:50Z
    Message:               All replicas are ready for Pgpool demo/pgpool
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-08-01T05:03:39Z
    Message:               pgpool demo/pgpool is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-08-01T05:03:39Z
    Message:               pgpool demo/pgpool is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-08-01T05:03:39Z
    Message:               The Pgpool: demo/pgpool is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

## Reusing Service Account

An existing service account can be reused in another Pgpool instance. No new access permission is required to run the new Pgpool instance.

Now, create Pgpool crd `pgpool-new` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/custom-rbac/pgpool-new.yaml
pgpool.kubedb.com/quick-pgpool created
```

Below is the YAML for the Pgpool crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-new
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  deletionPolicy: WipeOut
```

Now, wait a few minutes. the KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we should see that a pod with the name `pgpool-new-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo pgpool-new-0
NAME           READY   STATUS    RESTARTS   AGE
pgpool-new-0   1/1     Running   0          55s
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo pgpool-new-0
Configuring Pgpool-II...
Custom pgpool.conf file detected. Use custom configuration files.
Generating pool_passwd...
Generating pcp.conf...
Custom pool_hba.conf file detected. Use custom pool_hba.conf.
Starting Pgpool-II...
2024-08-01 05:05:34.353: main pid 60: LOG:  Backend status file /tmp/pgpool_status does not exist
2024-08-01 05:05:34.353: main pid 60: LOG:  health_check_stats_shared_memory_size: requested size: 12288
2024-08-01 05:05:34.353: main pid 60: LOG:  memory cache initialized
2024-08-01 05:05:34.353: main pid 60: DETAIL:  memcache blocks :64
2024-08-01 05:05:34.353: main pid 60: LOG:  allocating (135894880) bytes of shared memory segment
2024-08-01 05:05:34.353: main pid 60: LOG:  allocating shared memory segment of size: 135894880 
2024-08-01 05:05:34.555: main pid 60: LOG:  health_check_stats_shared_memory_size: requested size: 12288
2024-08-01 05:05:34.555: main pid 60: LOG:  health_check_stats_shared_memory_size: requested size: 12288
2024-08-01 05:05:34.555: main pid 60: LOG:  memory cache initialized
2024-08-01 05:05:34.555: main pid 60: DETAIL:  memcache blocks :64
2024-08-01 05:05:34.556: main pid 60: LOG:  pool_discard_oid_maps: discarded memqcache oid maps
2024-08-01 05:05:34.567: main pid 60: LOG:  create socket files[0]: /tmp/.s.PGSQL.9999
2024-08-01 05:05:34.567: main pid 60: LOG:  listen address[0]: *
2024-08-01 05:05:34.567: main pid 60: LOG:  Setting up socket for 0.0.0.0:9999
2024-08-01 05:05:34.568: main pid 60: LOG:  Setting up socket for :::9999
2024-08-01 05:05:34.568: main pid 60: LOG:  find_primary_node_repeatedly: waiting for finding a primary node
2024-08-01 05:05:34.568: main pid 60: LOG:  create socket files[0]: /var/run/pgpool/.s.PGSQL.9595
2024-08-01 05:05:34.569: main pid 60: LOG:  listen address[0]: *
2024-08-01 05:05:34.569: main pid 60: LOG:  Setting up socket for 0.0.0.0:9595
2024-08-01 05:05:34.569: main pid 60: LOG:  Setting up socket for :::9595
2024-08-01 05:05:34.569: sr_check_worker pid 68: LOG:  process started
2024-08-01 05:05:34.570: health_check pid 69: LOG:  process started
2024-08-01 05:05:34.570: health_check pid 70: LOG:  process started
2024-08-01 05:05:34.570: pcp_main pid 67: LOG:  PCP process: 67 started
2024-08-01 05:05:34.570: main pid 60: LOG:  pgpool-II successfully started. version 4.5.0 (hotooriboshi)
```
`pgpool-II successfully started` in the log signifies that the pgpool is running successfully.

Also, if we want to verify that the pod is actually using our custom service account we can just describe the pod and see the `Service Accouunt Name`,
```bash
$ kubectl describe pp -n demo pgpool-new
Name:         pgpool-new
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Pgpool
Metadata:
  Creation Timestamp:  2024-08-01T05:05:32Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  271244
  UID:               e985525a-9479-4364-9c8f-192c476fd2dc
Spec:
  Auth Secret:
    Name:            pgpool-new-auth
  Client Auth Mode:  md5
  Deletion Policy:   WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  pgpool
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     70
          Run As Non Root:  true
          Run As User:      70
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:            70
      Service Account Name:  my-custom-serviceaccount
  Postgres Ref:
    Name:       ha-postgres
    Namespace:  demo
  Replicas:     1
  Ssl Mode:     disable
  Version:      4.5.0
Status:
  Conditions:
    Last Transition Time:  2024-08-01T05:05:33Z
    Message:               The KubeDB operator has started the provisioning of Pgpool: demo/pgpool-new
    Observed Generation:   2
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-08-01T05:17:49Z
    Message:               All replicas are ready for Pgpool demo/pgpool-new
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-08-01T05:05:45Z
    Message:               pgpool demo/pgpool-new is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-08-01T05:05:45Z
    Message:               pgpool demo/pgpool-new is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-08-01T05:05:45Z
    Message:               The Pgpool: demo/pgpool-new is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```
## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pp/pgpool
kubectl delete -n demo pp/pgpool-new
kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete sa -n demo my-custom-serviceaccount
kubectl delete -n demo pg/ha-postgres
kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart Pgpool](/docs/guides/pgpool/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Pgpool instance with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool instance with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

