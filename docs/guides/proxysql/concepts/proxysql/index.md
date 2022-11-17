---
title: ProxySQL CRD
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-concepts-proxysql
    name: ProxySQL
    parent: guides-proxysql-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ProxySQL

## What is ProxySQL

`ProxySQL` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [ProxySQL](https://www.proxysql.com/) in a Kubernetes native way. You only need to describe the desired configurations in a ProxySQL object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## ProxySQL Spec

Like any official Kubernetes resource, a `ProxySQL` object has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections. Below is an example of the ProxySQL object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: demo-proxysql
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 1
  mode: GroupReplication
  backend:
    name: my-group
  authSecret:
    name: proxysql-cluster-auth
    externallyManaged: true
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  initConfig:
    mysqlUsers:
      - username: test
        active: 1
        default_hostgroup: 2
    adminVariables:
      restapi_enabled: true
      restapi_port: 6070
  configSecret:
    name: my-custom-config
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: proxy-issuer
    certificates:
    - alias: server
      subject:
        organizations:
          - kubedb:server
      dnsNames:
        - localhost
      ipAddresses:
        - "127.0.0.1"
  podTemplate:
    metadata:
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
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  6033
  terminationPolicy: WipeOut
  healthChecker:
    failureThreshold: 3
```

### .spec.version

`.spec.version` is a required field specifying the name of the [ProxySQLVersion](/docs/guides/proxysql/concepts/proxysql-version/index.md) CRD where the docker images are specified. Currently, when you install KubeDB, it creates the following `ProxySQLVersion` resources,

- `2.3.2-debian`
- `2.3.2-centos`
- `2.4.4-debian`
- `2.4.4-centos`

### .spec.backend

`.spec.backend` specifies the information about the appbinding of the backend MySQL/PerconaXtraDB/MariaDB. The appbinding should contain the basic informations like connections url, server type , ssl infos etc. To know more about what appbinding is, you can refer to the Appbinding page in the concept section. See the api [here](https://pkg.go.dev/kubedb.dev/apimachinery@v0.29.1/apis/kubedb/v1alpha2#:~:text=//%20Backend%20refers%20to%20the%20AppBinding%20of%20the%20backend%20MySQL/MariaDB/Percona%2DXtraDB%20server%0A%09Backend%20*core.LocalObjectReference%20%60json%3A%22backend%2Comitempty%22%60).

### .spec.authSecret

`.spec.authSecret` is an optional field that points to a secret used to hold credentials for `proxysql cluster admin` user. If not set, the KubeDB operator creates a new Secret `{proxysql-object-name}-auth` for storing the password for `proxysql cluster admin` user for each ProxySQL object. If you want to use an existing secret please specify that when creating the ProxySQL object using `.spec.authSecret`. Turn the `.spec.authSecret.extenallyManaged` field `true` in that case.

This secret contains a `username` key and a `password` key which contains the username and password respectively for `proxysql cluster admin` user. The password should always be alpha-numeric.  If no Secret is found, KubeDB sets the value of `username` key to  `"cluster"`. See the api [here](https://pkg.go.dev/kubedb.dev/apimachinery@v0.29.1/apis/kubedb/v1alpha2#:~:text=//%20ProxySQL%20secret%20containing%20username%20and%20password%20for%20root%20user%20and%20proxysql%20user%0A%09//%20%2Boptional%0A%09AuthSecret%20*SecretReference%20%60json%3A%22authSecret%2Comitempty%22%60).

> Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator (version 0.13.0 and higher).

Example:

```bash
$ kubectl create secret generic proxysql-cluster-auth -n demo \
--from-literal=username=cluster \
--from-literal=password=6q8u2jMOWOOZXk
secret "proxysql-cluster-auth" created
```

```yaml
apiVersion: v1
data:
  password: NnE4dTJqTU9XT09aWGs=
  username: Y2x1c3Rlcg==
kind: Secret
metadata:
  ...
  name: proxysql-cluster-auth
  namespace: demo
  ...
type: Opaque
```

### .spec.monitor

ProxySQL managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. In the `.spec.monitor` section you can configure neccessary settings regarding monitoring. See the api [here](https://pkg.go.dev/kubedb.dev/apimachinery@v0.29.1/apis/kubedb/v1alpha2#:~:text=//%20Monitor%20is%20used%20monitor%20proxysql%20instance%0A%09//%20%2Boptional%0A%09Monitor%20*mona.AgentSpec%20%60json%3A%22monitor%2Comitempty%22%60).

### .spec.InitConfig

`spec.initConfig` is the field where we can set the proxysql bootstrap configuration. In ProxySQL an initial configuration file is needed to bootstrap, named `proxysql.cnf`. In that file you should write down all the necessary configuration related to various proxysql tables and variables in a specific format.  In KubeDB ProxySQL we have eased this initial configuration setup with declarative yaml. All you need to do is to pass the configuration in the yaml in key-value format and the operator will turn that into a `proxysql.cnf` file with proper formatting . The `proxysql.cnf` file will be available in a secret with name `<proxysql-crd-name>-configuration` . When you change any configuration with the proxysqlOpsRequest , the secret will be auto updated with the new configuration. 

`.spec.initConfig` contains four subsections : `mysqlUsers`, `mysqlQueryRules`, `adminVariables`, `mysqlVariables`. The detailed description is given below. See the api [here](https://pkg.go.dev/kubedb.dev/apimachinery@v0.29.1/apis/kubedb/v1alpha2#:~:text=//%20%2Boptional%0A%09//%20InitConfiguration%20contains%20information%20with%20which%20the%20proxysql%20will%20bootstrap%20(only%204%20tables%20are%20configurable)%0A%09InitConfiguration%20*ProxySQLConfiguration%20%60json%3A%22initConfig%2Comitempty%22%60). 

`.spec.initConfig.mysqlUsers` section carries info for the `mysql_users` table. All the information provided through this field will eventually be used for setting up the `mysql_users` table inside the proxysql server. This section is an array field where each element of the array carries the necessary information for each individual users. An important note to be mentioned is that you don't need to fill up the password field for any user. The password will be automatically fetched by the KubeDB operator from the backend server.   

`.spec.initConfig.mysqlQueryRules`section carries info for the `mysql_query_rules` table. This section is also an array field and each element of the array should be a `query_rule` as per proxysql accepts.

`.spec.initConfig.mysqlVariables` section carries all the `mysql_variables` info that you want to set for the proxysql. You need to mention the variables you want to set with its value in a key-value format under this section and the KubeDB operator will bootstrap the proxysql with this.

`.spec.initConfig.adminVariables` section carries all the `admin_variables` info that you want to set for the proxysql. You need to mention the variables you want to set with its value in a key-value format under this section and the KubeDB operator will bootstrap the proxysql with this.

### .spec.configSecret

`.spec.configSecret` is another field to pass the bootstrap configuration for the proxysql. If you want to pass the configuration through a secret you can just mention the secret name under this field. The secret should look something like the following 

```bash
$ kubectl view-secret -n demo my-config-secret -a  
AdminVariables.cnf=admin_variables=
{
    checksum_mysql_query_rules: true
    refresh_interval: 2000
    connect_timeout_server: 3000
}
MySQLQueryRules.cnf=mysql_query_rules=
(
    {
        rule_id=1
        active=1
        match_pattern="^SELECT .* FOR UPDATE$"
        destination_hostgroup=2
        apply=1
    },
    {
        rule_id=2
        active=1
        match_pattern="^SELECT"
        destination_hostgroup=3
        apply=1
    }
)

MySQLUsers.cnf=mysql_users=
(
    {
        username = "user2"
        password = "pass2"
        default_hostgroup = 2
        active = 1
    },
    {
        username = "user3"
        password = "pass3"
        default_hostgroup = 2
        max_connections=1000      
        default_schema="test"
        active = 1
    },
    { username = "user4" , password = "pass4" , default_hostgroup = 0 , active = 1 ,comment = "hello all"}
)
MySQLVariables.cnf=mysql_variables=
{
    max_connections=1024
    default_schema="information_schema"
}
```

The secret should contain keys none other than `AdminVariables.cnf`, `MySQLVariables.cnf`, `MySQLUsers.cnf`, `MySQLVariables.cnf` . The key names define the contents of the values itself. Important info to add is that the value provided with the keys will be patched to the `proxysql.cnf` file exactly as it is. So be careful with the format when you are going to bootstrap proxysql in this way. 

### .spec.syncUsers

`spec.syncUsers` is a boolean field. While true, KubeDB Operator fetches all the users from the backend and puts them into the `mysql_users` table. Any update regarding a user in the backend will also reflect in the proxysql server. This field can be turned off by simply changing the value to false and applying the yaml. It is set false by default though.   


### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the ProxySQL frontend connections. See the api [here](https://pkg.go.dev/kmodules.xyz/client-go/api/v1#TLSConfig)

The following fields are configurable in the `spec.tls` section:

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.

  - `apiGroup` is the group name of the resource being referenced. The value for `Issuer` or   `ClusterIssuer` is "cert-manager.io"   (cert-manager v0.12.0 and later).
  - `kind` is the type of resource being referenced. KubeDB supports both `Issuer`   and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.

- `certificates` (optional) are a list of certificates used to configure the server and/or client certificate. It has the following fields:

  - `alias` represents the identifier of the certificate. It has the following possible value:
    - `server` is used for server certificate identification.
    - `client` is used for client certificate identification.
    - `metrics-exporter` is used for metrics exporter certificate identification.
  - `secretName` (optional) specifies the k8s secret name that holds the certificates.
    This field is optional. If the user does not specify this field, the default secret name will be created in the following format: `<database-name>-<cert-alias>-cert`.
  - `subject` (optional) specifies an `X.509` distinguished name. It has the following possible field,
    - `organizations` (optional) are the list of different organization names to be used on the Certificate.
    - `organizationalUnits` (optional) are the list of different organization unit name to be used on the Certificate.
    - `countries` (optional) are the list of country names to be used on the Certificate.
    - `localities` (optional) are the list of locality names to be used on the Certificate.
    - `provinces` (optional) are the list of province names to be used on the Certificate.
    - `streetAddresses` (optional) are the list of a street address to be used on the Certificate.
    - `postalCodes` (optional) are the list of postal code to be used on the Certificate.
    - `serialNumber` (optional) is a serial number to be used on the Certificate.
      You can found more details from [Here](https://golang.org/pkg/crypto/x509/pkix/#Name)

  - `duration` (optional) is the period during which the certificate is valid.
  - `renewBefore` (optional) is a specifiable time before expiration duration.
  - `dnsNames` (optional) is a list of subject alt names to be used in the Certificate.
  - `ipAddresses` (optional) is a list of IP addresses to be used in the Certificate.
  - `uriSANs` (optional) is a list of URI Subject Alternative Names to be set in the Certificate.
  - `emailSANs` (optional) is a list of email Subject Alternative Names to be set in the Certificate.


### .spec.podTemplate

KubeDB allows providing a template for proxysql pod through `.spec.podTemplate`. KubeDB operator will pass the information provided in `.spec.podTemplate` to the StatefulSet created for ProxySQL. See the api [here](https://pkg.go.dev/kmodules.xyz/offshoot-api/api/v1#PodTemplateSpec)

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

#### .spec.podTemplate.spec.imagePullSecrets

`KubeDB` provides the flexibility of deploying ProxySQL from a private Docker registry. `.spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker images if you are using a private docker registry.

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

## Next Steps

#TODO: edit the links
- Learn how to use KubeDB to load balance MySQL Group Replication 