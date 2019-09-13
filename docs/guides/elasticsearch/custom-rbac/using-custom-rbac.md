---
title: Run Elasticsearch with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: es-custom-rbac-quickstart
    name: Custom RBAC
    parent: es-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a Elasticsearch instance. This tutorial will show you how to use KubeDB to run Elasticsearch instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for Elasticsearch. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in Elasticsearch crd. If this field is left empty, the KubeDB operator will create a service account name matching Elasticsearch crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a Elasticsearch Database named `quick-elasticsearch` to provide the bare minimum access permissions.

## Custom RBAC for Elasticsearch

At first, let's create a `Service Acoount` in `demo` namespace.

```console
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

It should create a service account.

```yaml
$ kubectl get serviceaccount -n demo my-custom-serviceaccount -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2019-05-31T08:43:48Z"
  name: my-custom-serviceaccount
  namespace: demo
  resourceVersion: "12305"
  selfLink: /api/v1/namespaces/demo/serviceaccounts/my-custom-serviceaccount
  uid: 35550bca-8380-11e9-8efb-08002790c45e
secrets:
- name: my-custom-serviceaccount-token-6xj8k

```

Now, we need to create a role that has necessary access permissions for the Elasticsearch instance named `quick-elasticsearch`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/custom-rbac/es-custom-role.yaml
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
  - elasticsearch-db
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for Elasticsearch pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```console
$ kubectl create rolebinding my-custom-rolebinding --role=my-custom-role --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created

```

It should bind `my-custom-role` and `my-custom-serviceaccount` successfully.

```yaml
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "kubectl get rolebinding -n demo my-custom-rolebinding -o yaml"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "1405"
  selfLink: /apis/rbac.authorization.k8s.io/v1/namespaces/demo/rolebindings/my-custom-rolebinding
  uid: 123afc02-8297-11e9-8d10-080027a8b217
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-custom-role
subjects:
- kind: ServiceAccount
  name: my-custom-serviceaccount
  namespace: demo

```

Now, create a Elasticsearch crd specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/custom-rbac/es-custom-db.yaml
elasticsearch.kubedb.com/quick-elasticsearch created
```

Below is the YAML for the Elasticsearch crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: quick-elasticsearch
  namespace: demo
spec:
  version: "8.0-v2"
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-elasticsearch-0` has been created.

Check that the statefulset's pod is running

```console
$ kubectl get pod -n demo quick-elasticsearch-0
NAME                READY     STATUS    RESTARTS   AGE
quick-elasticsearch-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```console
$ kubectl logs -f -n demo quick-elasticsearch-0
+ set -o errexit
+ set -o pipefail
+ sync
...
Starting runit...
....................
..................
Search Guard Version: 6.3.0-23.1
Connected as CN=sgadmin,O=Elasticsearch Operator
Contacting elasticsearch cluster 'elasticsearch' and wait for YELLOW clusterstate ...
Clustername: quick-elasticsearch
Clusterstate: GREEN
Number of nodes: 3
Number of data nodes: 3
searchguard index does not exists, attempt to create it ... done (0-all replicas)
Populate config from /elasticsearch/plugins/search-guard-6/sgconfig/
Will update 'sg/config' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_config.yml
   SUCC: Configuration for 'config' created or updated
Will update 'sg/roles' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_roles.yml
   SUCC: Configuration for 'roles' created or updated
Will update 'sg/rolesmapping' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_roles_mapping.yml
   SUCC: Configuration for 'rolesmapping' created or updated
Will update 'sg/internalusers' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_internal_users.yml
   SUCC: Configuration for 'internalusers' created or updated
Will update 'sg/actiongroups' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_action_groups.yml
   SUCC: Configuration for 'actiongroups' created or updated
[2019-05-31T09:10:31,537][INFO ][c.f.s.c.IndexBaseConfigurationRepository] Search Guard License Info: No license needed because enterprise modules are not enabled
Done with success

```

Once we see `Done with success` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another Elasticsearch Database. No new access permission is required to run the new Elasticsearch Database.

Now, create Elasticsearch crd `minute-elasticsearch` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/custom-rbac/es-custom-db-two.yaml
elasticsearch.kubedb.com/quick-elasticsearch created
```

Below is the YAML for the Elasticsearch crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: minute-elasticsearch
  namespace: demo
spec:
  version: "8.0-v2"
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `minute-elasticsearch-0` has been created.

Check that the statefulset's pod is running

```console
$ kubectl get pod -n demo minute-elasticsearch-0
NAME                READY     STATUS    RESTARTS   AGE
minute-elasticsearch-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```console
$ kubectl logs -f -n demo minute-elasticsearch-0
+ set -o errexit
+ set -o pipefail
+ sync
...
Starting runit...
....................
..................
Search Guard Version: 6.3.0-23.1
Connected as CN=sgadmin,O=Elasticsearch Operator
Contacting elasticsearch cluster 'elasticsearch' and wait for YELLOW clusterstate ...
Clustername: quick-elasticsearch
Clusterstate: GREEN
Number of nodes: 3
Number of data nodes: 3
searchguard index does not exists, attempt to create it ... done (0-all replicas)
Populate config from /elasticsearch/plugins/search-guard-6/sgconfig/
Will update 'sg/config' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_config.yml
   SUCC: Configuration for 'config' created or updated
Will update 'sg/roles' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_roles.yml
   SUCC: Configuration for 'roles' created or updated
Will update 'sg/rolesmapping' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_roles_mapping.yml
   SUCC: Configuration for 'rolesmapping' created or updated
Will update 'sg/internalusers' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_internal_users.yml
   SUCC: Configuration for 'internalusers' created or updated
Will update 'sg/actiongroups' with /elasticsearch/plugins/search-guard-6/sgconfig/sg_action_groups.yml
   SUCC: Configuration for 'actiongroups' created or updated
[2019-05-31T09:10:31,537][INFO ][c.f.s.c.IndexBaseConfigurationRepository] Search Guard License Info: No license needed because enterprise modules are not enabled
Done with success

```

`Done with success` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo es/quick-elasticsearch -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo es/quick-elasticsearch

kubectl patch -n demo es/minute-elasticsearch -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo es/minute-elasticsearch

kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/uninstall.md).

## Next Steps

- [Quickstart Elasticsearch](/docs/guides/elasticsearch/quickstart/quickstart.md) with KubeDB Operator.
- [Snapshot and Restore](/docs/guides/elasticsearch/snapshot/backup-and-restore.md) process of Elasticsearch instances using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/elasticsearch/snapshot/scheduled-backup.md) of Elasticsearch instances using KubeDB.
- Initialize [Elasticsearch with Script](/docs/guides/elasticsearch/initialization/using-script.md).
- Initialize [Elasticsearch with Snapshot](/docs/guides/elasticsearch/initialization/using-snapshot.md).
- Monitor your Elasticsearch instance with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/elasticsearch/monitoring/using-coreos-prometheus-operator.md).
- Monitor your Elasticsearch instance with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Use [kubedb cli](/docs/guides/elasticsearch/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Elasticsearch object](/docs/concepts/databases/elasticsearch.md).
- Detail concepts of [Snapshot object](/docs/concepts/snapshot.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

