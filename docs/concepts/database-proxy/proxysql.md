---
title: ProxySQL
menu:
  docs_{{ .version }}:
    identifier: proxysql
    name: ProxySQL
    parent: database-proxy
    weight: 35
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# ProxySQL

## What is ProxySQL

`ProxySQL` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [ProxySQL](https://www.proxysql.com/) in a Kubernetes native way. You only need to describe the desired configurations in a ProxySQL object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## ProxySQL Spec

Like any official Kubernetes resource, a `ProxySQL` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections. Below is an example of the ProxySQL object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: demo-proxysql-for-mysql
  namespace: demo
spec:
  version: "2.0.4"
  replicas: 1
  mode: GroupReplication
  backend:
    ref:
      apiGroup: "kubedb.com"
      kind: MySQL
      name: my-group
    replicas: 3
  proxysqlSecret:
    secretName: demo-proxysql-for-mysql-auth
  monitor:
    agent: prometheus.io/coreos-operator
    prometheus:
      namespace: demo
      labels:
        app: kubedb
      interval: 10s
  configSource:
    configMap:
      name: my-custom-config
  podTemplate:
    annotations:
      passMe: ToProxySQLPod
    controller:
      annotations:
        passMe: ToStatefulSet
    spec:
      serviceAccountName: my-service-account
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
      - name: myregistrykey
      args:
      - --reload
      env:
      - name: LOAD_BALANCE_MODE
        value: GroupReplication
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
        limits:
          memory: "128Mi"
          cpu: "500m"
  serviceTemplate:
    annotations:
      passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  6033
        targetPort: http
  updateStrategy:
    type: RollingUpdate
```

### .spec.version

`.spec.version` is a required field specifying the name of the [ProxySQLVersion](/docs/concepts/catalog/proxysql.md) CRD where the docker images are specified. Currently, when you install KubeDB, it creates the following `ProxySQLVersion` resources,

- `2.0.4`

### .spec.backend

`.spec.backend` specifies the information about backend MySQL/PerconaXtraDB/MariaDB.

You can specify the following fields in `.spec.backend` field,

- `.spec.backend.ref` lets one locate the typed referenced object. In this case, it is the MySQL/PerconaXtraDB/MariaDB object in the same namespace as the ProxySQL object.

  - `apiGroup` is the group for the resource being referenced. Here it is `"kubedb.com"`.
  - `kind` is the type of resource being referenced. Here it is `"MySQL"`.
  - `name` specifies the name of the resource being referenced. Here it is `"my-group"`.

### .spec.proxysqlSecret

`.spec.proxysqlSecret` is an optional field that points to a Secret used to hold credentials for `proxysql` user. If not set, the KubeDB operator creates a new Secret `{proxysql-object-name}-auth` for storing the password for `proxysql` user for each ProxySQL object. If you want to use an existing secret please specify that when creating the ProxySQL object using `.spec.proxysqlSecret.secretName`.

This secret contains a `proxysqluser` key and a `proxysqlpass` key which contains the username and password respectively for `proxysql` user. If no Secret is found, KubeDB sets the value of `proxysqluser` key to be `proxysql`.

Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

Example:

```console
$ kubectl create secret generic demo-proxysql-for-mysql-auth -n demo \
--from-literal=proxysqluser=proxysql \
--from-literal=proxysqlpass=6q8u_2jMOW-OOZXk
secret "demo-proxysql-for-mysql-auth" created
```

```yaml
apiVersion: v1
data:
  proxysqlpass: NnE4dV8yak1PVy1PT1pYaw==
  proxysqluser: cHJveHlzcWw=
kind: Secret
metadata:
  ...
  name: demo-proxysql-for-mysql-auth
  namespace: demo
  ...
type: Opaque
```

### .spec.monitor

ProxySQL managed by KubeDB can be monitored with builtin-Prometheus and CoreOS-Prometheus operator out-of-the-box. To learn more,

- [Monitor ProxySQL with builtin Prometheus](/docs/guides/proxysql/monitoring/using-builtin-prometheus.md)
- [Monitor ProxySQL with CoreOS Prometheus operator](/docs/guides/proxysql/monitoring/using-coreos-prometheus-operator.md)

### .spec.configSource

`.spec.configSource` is an optional field that allows users to provide custom configuration for ProxySQL. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/proxysql/configuration/using-custom-config.md).

### .spec.podTemplate

KubeDB allows providing a template for proxysql pod through `.spec.podTemplate`. KubeDB operator will pass the information provided in `.spec.podTemplate` to the StatefulSet created for ProxySQL.

KubeDB accept following fields to set in `.spec.podTemplate`:

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (statefulset's annotation)
- spec:
  - args
  - env
  - resources
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - livenessProbe
  - readinessProbe
  - lifecycle

Usage of some field of `.spec.podTemplate` is described below,

#### .spec.podTemplate.spec.args

`.spec.podTemplate.spec.args` is an optional field. This can be used to provide additional arguments to proxysql installation.

#### .spec.podTemplate.spec.env

`.spec.podTemplate.spec.env` is an optional field that specifies the environment variables to pass to the ProxySQL docker image. Here is a list of currently supported environment variables to the ProxySQL image:

- `MYSQL_ROOT_PASSWORD`
- `MYSQL_PROXY_USER`
- `MYSQL_PROXY_PASSWORD`
- `PEERS`
- `LOAD_BALANCE_MODE`

Note that, KubeDB does not allow the following environment variables to set in `.spec.env`.

- `MYSQL_ROOT_PASSWORD`
- `MYSQL_PROXY_USER`
- `MYSQL_PROXY_PASSWORD`

If you try to set any of the forbidden environment variables i.e. `MYSQL_ROOT_PASSWORD` in ProxySQL object, KubeDB operator will reject the request with the following error,

```ini
Error from server (Forbidden): error when creating "./proxysql.yaml": admission webhook "proxysql.validators.kubedb.com" denied the request: environment variable MYSQL_ROOT_PASSWORD is forbidden to use in ProxySQL spec
```

Also note that KubeDB does not allow to update the environment variables as updating them does not have any effect once the ProxySQL is created.  If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./proxysql.yaml": admission webhook "proxysql.validators.kubedb.com" denied the request: precondition failed for:
...At least one of the following was changed:
    apiVersion
    kind
    name
    namespace
    spec.proxysqlSecret
    spec.podTemplate.spec.nodeSelector
    spec.podTemplate.spec.env
```

#### .spec.podTemplate.spec.imagePullSecrets

`KubeDB` provides the flexibility of deploying ProxySQL from a private Docker registry. `.spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker images if you are using a private docker registry. To learn how to deploy SQL from a private registry, please visit [here](/docs/guides/proxysql/private-registry/using-private-registry.md).

#### .spec.podTemplate.spec.nodeSelector

`.spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### .spec.podTemplate.spec.serviceAccountName

 `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine-tune role-based access control.

 If this field is left empty, the KubeDB operator will create a service account name matching the ProxySQL object name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

 If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

 If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

#### .spec.podTemplate.spec.resources

`.spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the ProxySQL Pod. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### .spec.serviceTemplate

You can also provide a template for the services created by KubeDB operator for the ProxySQL through `.spec.serviceTemplate`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `.spec.serviceTemplate`:

- metadata:
- annotations
- spec:
- type
- ports
- clusterIP
- externalIPs
- loadBalancerIP
- loadBalancerSourceRanges
- externalTrafficPolicy
- healthCheckNodePort
- sessionAffinityConfig

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163) to understand these fields in detail.

### .spec.updateStrategy

You can specify [update strategy](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#update-strategies) of StatefulSet created by KubeDB for ProxySQL thorough `.spec.updateStrategy` field. The default value of this field is `RollingUpdate`. In the future, we will use this field to determine how automatic migration from an old ProxySQL version to a new one should behave.

## Next Steps

- Learn how to use KubeDB to load balance MySQL Group Replication [here](/docs/guides/proxysql/quickstart/load-balance-mysql-group-replication.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
