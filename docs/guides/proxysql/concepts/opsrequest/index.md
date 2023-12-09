---
title: ProxySQLOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-concepts-proxysqlopsrequest
    name: ProxySQLOpsRequest
    parent: guides-proxysql-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ProxySQLOpsRequest

## What is ProxySQLOpsRequest

`ProxySQLOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [ProxySQL](https://www.proxysql.com/) administrative operations like database version updating, horizontal scaling, vertical scaling,reconfiguration etc. in a Kubernetes native way.

## ProxySQLOpsRequest CRD Specifications

Like any official Kubernetes resource, a `ProxySQLOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `ProxySQLOpsRequest` CRs for different administrative operations is given below:

**Sample ProxySQLOpsRequest for updating database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: proxyops-update
  namespace: demo
spec:
  type: UpdateVersion
  proxyRef:
    name: proxy-server
  updateVersion:
    targetVersion: "2.4.4-debian"
```

**Sample ProxySQLOpsRequest Objects for Horizontal Scaling of proxysql cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  proxyRef:
    name: proxy-server
  horizontalScaling:
    member: 5

```

**Sample ProxySQLOpsRequest Objects for Vertical Scaling of the proxysql cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: proxyops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  proxyRef:
    name: proxy-server
  verticalScaling:
    proxysql:
      requests:
        memory: "1.2Gi"
        cpu: "0.6"
      limits:
        memory: "1.2Gi"
        cpu: "0.6"
```

**Sample ProxySQLOpsRequest Objects for Reconfiguring ProxySQL cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: reconfigure-vars
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    adminVariables:
      refresh_interval: 2055
      cluster_check_interval_ms: 205
    mysqlVariables:
      max_transaction_time: 1540000
      max_stmts_per_connection: 19
```

**Sample ProxySQLOpsRequest Objects for Reconfiguring TLS of the ProxySQL:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: recon-tls-update
  namespace: demo
spec:
  type: ReconfigureTLS
  proxyRef:
    name: proxy-server
  tls:
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
      emailAddresses:
      - "mikebaker@gmail.com"
      certificates:
    - alias: client
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
      emailAddresses:
      - "mikebaker@gmail.com"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: recon-tls-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  proxyRef:
    name: proxy-server
  tls:
    rotateCertificates: true
```


Here, we are going to describe the various sections of a `ProxySQLOpsRequest` crd.

A `ProxySQLOpsRequest` object has the following fields in the `spec` section.

### spec.proxyRef

`spec.proxyRef` is a required field that point to the [ProxySQL](/docs/guides/proxysql/concepts/proxysql) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.proxyRef.name :** specifies the name of the [ProxySQL](/docs/guides/proxysql/concepts/proxysql) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `ProxySQLOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `ProxySQLOpsRequest` CR. For example, if you want to update your proxysql and scale up its replica then you have to create two separate `ProxySQLOpsRequest`. At first, you have to create a `ProxySQLOpsRequest` for updating. Once it is completed, then you can create another `ProxySQLOpsRequest` for scaling. You should not create two `ProxySQLOpsRequest` simultaneously.

### spec.updateVersion

If you want to update your ProxySQL version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [ProxySQLVersion](/docs/guides/proxysql/concepts/proxysql-version/index.md) CR that contains the ProxySQL version information where you want to update.

> You can only update between ProxySQL versions. KubeDB does not support downgrade for ProxySQL.

### spec.horizontalScaling

If you want to scale-up or scale-down your ProxySQL cluster or different components of it, you have to specify `spec.horizontalScaling` section. `spec.horizontalScaling.member` indicates the desired number of nodes for ProxySQL cluster after scaling. For example, if your cluster currently has 4 nodes, and you want to add additional 2 nodes then you have to specify 6 in `spec.horizontalScaling.member` field. Similarly, if you want to remove one node from the cluster, you have to specify 3 in `spec.horizontalScaling.` field.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `ProxySQL` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-field:

- `spec.verticalScaling.proxysql` indicates the desired resources for ProxySQL standalone or cluster after scaling.
- `spec.verticalScaling.exporter` indicates the desired resources for the `exporter` container.
- `spec.verticalScaling.coordinator` indicates the desired resources for the `coordinator` container.


All of them has the below structure:

```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for the container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

### spec.configuration

If you want to reconfigure your Running ProxySQL cluster with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-fields:
- `mysqlUsers` : To reconfigure the `mysql_users` table, you need to provide the desired user infos under the `spec.configuration.mysqlUsers.users` section. Set the `.spec.configuration.mysqlUsers.reqType` to either `add`, `update` or `delete` based on the operation you want to do.
- `mysqlQueryRules` : To reconfigure the `mysql_query_rules` table, you need to provide the desired rule infos under the `spec.configuration.mysqlQueryRules.rules` section. Set the `.spec.configuration.mysqlQueryRules.reqType` to either `add`, `update` or `delete` based on the operation you want to do.
- `mysqlVariables` : You can reconfigure mysql variables for the proxysql server using this field. You can reconfigure almost all the mysql variables except `mysql-interfaces`, `mysql-monitor_username`, `mysql-monitor_password`, `mysql-ssl_p2s_cert`, `mysql-ssl_p2s_key`, `mysql-ssl_p2s_ca`.
- `adminVariables` : You can reconfigure admin variables for the proxysql server using this field. You can reconfigure almost all the admin variables except `admin-admin_credentials` and `admin-mysql_interface`. 

### spec.tls

If you want to reconfigure the TLS configuration of your database i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/proxysql/concepts/proxysql/index.md/#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this proxysql.
- `spec.tls.remove` specifies that we want to remove tls from this proxysql.


### ProxySQLOpsRequest Status

`.status` describes the current state and progress of a `ProxySQLOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `ProxySQLOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                             |
| ---------- | ----------------------------------------------------------------------------------- |
| Successful | KubeDB has successfully performed the operation requested in the ProxySQLOpsRequest |
| Failed     | KubeDB has failed the operation requested in the ProxySQLOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the ProxySQLOpsRequest                 |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `ProxySQLOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `ProxySQLOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. ProxySQLOpsRequest has the following types of conditions:

| Type                          | Meaning                                                                   |
| ----------------------------- | ------------------------------------------------------------------------- |
| `Progressing`                 | Specifies that the operation is now in the progressing state              |
| `Successful`                  | Specifies such a state that the operation on the database was successful. |
| `Failed`                      | Specifies such a state that the operation on the database failed.         |
| `ScaleDownCluster`            | Specifies such a state that the scale down operation of replicaset        |
| `ScaleUpCluster`              | Specifies such a state that the scale up operation of replicaset          |
| `Reconfigure`                 | Specifies such a state that the reconfiguration of replicaset nodes       |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
  - `status` will be `True` if the current transition succeeded.
  - `status` will be `False` if the current transition failed.
  - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.