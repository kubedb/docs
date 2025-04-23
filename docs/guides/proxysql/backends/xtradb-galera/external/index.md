---
title: Proxy External Percona XtraDB Galera With KubeDB ProxySQL
menu:
  docs_{{ .version }}:
    identifier: external-percona-xtradb-galera
    name: External
    parent: percona-xtradb-backend
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB ProxySQL with Percona XtraDB Galera Cluster

This guide will show you how to use `KubeDB` operator to set up `ProxySQL` for externally managed Percona XtraDB cluster.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/index.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Percona XtraDB Backend 

In this tutorial we are going to test set up a ProxySQL server with KubeDB operator for a Percona XtraDB cluster. We have a Percona XtraDB cluster running in our K8s cluster which is TLS secured. We need to prepare an appbinding for this cluster so that our operator can get enough information to set up a ProxySQL server for this specific Percona XtraDB cluster. 

**What we have**

We have a 3 node cluster, a service for this cluster, the root-auth secret, a secret which contains the client certificates for TLS secured connections and some more other secrets and services. Let's see the resources first. 

```bash
~ $ kubectl get pods -n demo | grep xtradb
xtradb-galera-0           2/2     Running   0          31m
xtradb-galera-1           2/2     Running   0          31m
xtradb-galera-2           2/2     Running   0          31m

~ $ kubectl get svc -n demo | grep xtradb
xtradb-galera        ClusterIP   10.96.11.201    <none>        3306/TCP            31m
... ... ...

~ $ kubectl get secret -n demo | grep  
xtradb-galera-auth                    kubernetes.io/basic-auth              2      32m
xtradb-galera-client-cert             kubernetes.io/tls                     3      32m
... ... ... 

~ $ kubectl view-secret -n demo xtradb-galera-auth -a                                                                             
password=0cPVJdA*jfPs.C(L
username=root

~ $ kubectl view-secret -n demo xtradb-galera-client-cert -a                                                                      
ca.crt=-----BEGIN CERTIFICATE-----
MIIDIzCCAgugAwIBAgIUYJxuPjqmDn2OiWd0i9qFv0gsw40wDQYJKoZIhvcNAQEL
BQAwITEOMAwGA1UEAwwFbXlzcWwxDzANBgNVBAoMBmt1YmVkYjAeFw0yMjEwMTAx
MjQ5NTNaFw0yMzEwMTAxMjQ5NTNaMCExDjAMBgNVBAMMBW15c3FsMQ8wDQYDVQQK
DAZrdWJlZGIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDdE4LhAZlD
A8vii9A9ZOUK8Esye6d0VZzQfbTy/tVQt7NrnHcKSdTqfKIds9umxRwV6jNVSdmQ
v3sTkLpa6AlTpbIPk6yKe1Dk6aHalVCIW8HLK5n7bIsLEwhI2oqxZb+wJruuFJ/y
Y8kkrk5WhPIg4j3ucWAarYjMZq5tPmOl8RWjxsci7Z2l7IHqjeJ6++4D5ydxlz/x
uXK6lU3bvcg+ihTVz5PCMKdx8vM/Nvu3nCmIkhsAI+4eVdN1ztHXnuQ1WDXePAEc
QBrFjqnZNibdTyN3a8+veT3hK+qasFEINZ8V8dUPIUypqXJi4jvBIoESDoWeugt3
Ia28z8NW6OVnAgMBAAGjUzBRMB0GA1UdDgQWBBQI7AN5FskOiqoZNT3Rsj3APCUw
nDAfBgNVHSMEGDAWgBQI7AN5FskOiqoZNT3Rsj3APCUwnDAPBgNVHRMBAf8EBTAD
AQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBSbCRLuN+FwihCvbjifr1eteb5A3yTVZ3v
Cz0HL+dXzBDmR98lOuh/H5xlErOu5zgcIrK+pm3KhAYu5xY8pEJIpEpBDLgHEP54
2G56Yruw/kRvqbxUipR23MbFezrQoAysJnfFFfkasUZPQJ84tM9Q+hLJTrztVG3f
aruBwxtH07nIY11QRxYlDOAlzrObqgoQKp81WU0kO7CEWvCVNzzaogVWO+eKouL9
/ZB5XCQUFTRpYDB0uhY5500GY1btiEEJiWYU84QQsjjH5VeFAm7meuddjOiLC7uC
JaHJDKkKqekCJFzc1tBjGAVT4vZLg+RWaRbGkMPvmVeQH9KR1Svi
-----END CERTIFICATE-----

tls.crt=-----BEGIN CERTIFICATE-----
MIIDEDCCAfigAwIBAgIQAciDaLH+9Oh4QWxmu+fMFjANBgkqhkiG9w0BAQsFADAh
MQ4wDAYDVQQDDAVteXNxbDEPMA0GA1UECgwGa3ViZWRiMB4XDTIyMTIwMjA0MjM0
NloXDTIzMDMwMjA0MjM0NlowDzENMAsGA1UEAxMEcm9vdDCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBALZ+Fd5lbGg7tIoxNsaKmOsEZNnLiWo5u/lQ5eaI
JufmvxTpaZmqw68yIX4yLZd7iXmSPrydEy6uJYq4HPghyapV20eg7dpHfWkjpmpx
OgudXBHeETyD2P4fR8KQjgyn8qF5pwwq210M46Olq/AatJFAEW/4+7wAPLugLl6Y
V0vFhbAcDmLXAxfz6HyiafF1czPDsaqi4sOV0WC5hnD2NnAcxpR7LfGVPSLosz2x
hs/aEnBdW9+AWhyDjJjslGslyWC8vge6F7dvJrkJcROM0ndk/IEOnNz0KP7dae/T
4XDj8/D2nwbxg421N7BOfby65ZQFMbDLJ0vsM9QdYa6faDECAwEAAaNWMFQwDgYD
VR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMCMAwGA1UdEwEB/wQCMAAw
HwYDVR0jBBgwFoAUCOwDeRbJDoqqGTU90bI9wDwlMJwwDQYJKoZIhvcNAQELBQAD
ggEBAHj/QRv9seYBuA7NUTPQmxEom/1A9XW6IW330mHKCMs7C4c4bRPXSh1hj8bz
CUoI/v4ItNBzcGFJc2LJSZNHVRuNZddDOebxepYngm/2u5wQot8ER2a+TkNBZtSs
kQI9O10awelzbhLoV9is6X3LsTnxk5AOm/fiShfISAdxAbejBOchTjF1g5CrlvD7
k4rOFJRXVDsQH0ken8JH9sKDcJwVM3Mjm+lO68Cv5kR7JOY1mrShvMVPCjEKC2kA
0xb+SNYBgjBsso8CkgJfqCiBi6S/zn/f83Qn60n7PJoHtZrDHhJSt07mTPuk3Ro6
NoEwUfZKavW5HRTH64qUizJFSb0=
-----END CERTIFICATE-----

tls.key=-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAtn4V3mVsaDu0ijE2xoqY6wRk2cuJajm7+VDl5ogm5+a/FOlp
marDrzIhfjItl3uJeZI+vJ0TLq4lirgc+CHJqlXbR6Dt2kd9aSOmanE6C51cEd4R
PIPY/h9HwpCODKfyoXmnDCrbXQzjo6Wr8Bq0kUARb/j7vAA8u6AuXphXS8WFsBwO
YtcDF/PofKJp8XVzM8OxqqLiw5XRYLmGcPY2cBzGlHst8ZU9IuizPbGGz9oScF1b
34BaHIOMmOyUayXJYLy+B7oXt28muQlxE4zSd2T8gQ6c3PQo/t1p79PhcOPz8Paf
BvGDjbU3sE59vLrllAUxsMsnS+wz1B1hrp9oMQIDAQABAoIBACwS/40aybfS06Oc
hzIkPxJjmUfQlHuHPhLUqvGmaF8Rp4yRYuOuDly9qsEjtUckmus1mtlKxls7y+1Y
0gZLgr0Ux0ThZRCWu38tEQAcIHy1oIrgKyGGZl3ZiCdBak08Mqk1DFcv8pLijgfz
9zah/IIoCw4UABhDpmdaJFjMSikOPrIHOgRO6UmREkjjcN8T/qLAY34+13oM0zY2
AyUyuD2hVxBYDu9dR8IN+PngALnUBDuAmnhPf9DwVyz2gkxRhcFEzIZbbSYrBL60
LuclP07gmggWyvM2UirwovE/jyTrbqhlYk7S2uo+5zpPhpdzCTwuQQRstoK7tVM5
Ty0OVtECgYEA6Fl3UYDPfqVggtCPkrqvvRGeT0DiKVuYDK0ju2AzeVewVMeYWGr9
mCrc87twFWksty8SU1ilCe/RItGXKiTE+jk34s6/Wi3Pj3+grT6vNQYA0mbgAYUj
xBKAQFov0xAh6bLYHMwabYVtpYDvlMVqak1HDkUMqimrBN72XsldKT8CgYEAyRFz
9Oqu/PeuUBjMfPzbGRKzGI2ObZGV6WZBFbygXLGIQJrC4tzDZIsBirhCsd9znWGx
J9MZzpUc5mz91FRrg95OnM4acvuMpwv6XlXNJIZrM5nxOfGjqta11Fmgr4bSajBW
nuL3BHtoeinTvEcv3Sxa8Nmyy8/9o/G+4KIlIo8CgYEAok0UZu9Wgb3dq6MqFzGm
3qg28F9/W6pqjLhI1HN/oUxalO4TgffCiw+t5edRhPNB0/fikivCpS1K5kqHkF28
5pkfa6RF0CVd7nwVbc7yrlQyMMbBxO4OrMDLq6gT7hg/yDIwefUspMJmdAybzk0U
Z4rxjos3LIoMt0tTx6RbGhsCgYEArE/MtAO7WwdX10SpWiPIEEC6Qzxs5vFxK8h5
1osEUuvB/LukcI8I1E1cUOmAHreEeUeTbrG22Bdp4P9euGxwh14ouLDYcdmpvC7D
rbySRc78aAhxdlrjDDFdOlJlJofAI0ixsxCG6MxpyOe3kQ7gsgalGOs4Evp4P9uY
3SGX+XkCgYBPNmR7nodCjljuSSS5uvcU0j4W6VHUj+uwAbuZR9lBCdCdhwgG9Zg4
oJQ2E75DXW2QieEIgBysXlIHf1LyvF9re6xIJIbl2p7m+/U0cPsGJhq+/CEyehJp
I30CEBNnaJM4N3pqrBvjWEcmuhvmiHc31vmf2aqnKY++SuAkfJpuAw==
-----END RSA PRIVATE KEY-----

~ $ kubectl get services -n demo xtradb-galera -oyaml                                                                            
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: ... ... ...
  labels:
    ... ... ...
  name: xtradb-galera
  namespace: demo
  ownerReferences:
  - ... ... ...
  resourceVersion: ... ... ...
  uid: ... ... ...
spec:
  clusterIP: ... ... ...
  clusterIPs:
  - ... ... ...
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: primary
    port: 3306
    protocol: TCP
    targetPort: db
  selector:
    ... ... ...
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}

```

We have shown the resources in detail. You should create similar thing with similar field names and key names. 

Now, based on these above information we are going to create our appbinding. The following is the appbinding yaml. 

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: xtradb-galera-appbinding
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJekNDQWd1Z0F3SUJBZ0lVWUp4dVBqcW1EbjJPaVdkMGk5cUZ2MGdzdzQwd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0lURU9NQXdHQTFVRUF3d0ZiWGx6Y1d3eER6QU5CZ05WQkFvTUJtdDFZbVZrWWpBZUZ3MHlNakV3TVRBeApNalE1TlROYUZ3MHlNekV3TVRBeE1qUTVOVE5hTUNFeERqQU1CZ05WQkFNTUJXMTVjM0ZzTVE4d0RRWURWUVFLCkRBWnJkV0psWkdJd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUURkRTRMaEFabEQKQTh2aWk5QTlaT1VLOEVzeWU2ZDBWWnpRZmJUeS90VlF0N05ybkhjS1NkVHFmS0lkczl1bXhSd1Y2ak5WU2RtUQp2M3NUa0xwYTZBbFRwYklQazZ5S2UxRGs2YUhhbFZDSVc4SExLNW43YklzTEV3aEkyb3F4WmIrd0pydXVGSi95Clk4a2tyazVXaFBJZzRqM3VjV0FhcllqTVpxNXRQbU9sOFJXanhzY2k3WjJsN0lIcWplSjYrKzRENXlkeGx6L3gKdVhLNmxVM2J2Y2craWhUVno1UENNS2R4OHZNL052dTNuQ21Ja2hzQUkrNGVWZE4xenRIWG51UTFXRFhlUEFFYwpRQnJGanFuWk5pYmRUeU4zYTgrdmVUM2hLK3Fhc0ZFSU5aOFY4ZFVQSVV5cHFYSmk0anZCSW9FU0RvV2V1Z3QzCklhMjh6OE5XNk9WbkFnTUJBQUdqVXpCUk1CMEdBMVVkRGdRV0JCUUk3QU41RnNrT2lxb1pOVDNSc2ozQVBDVXcKbkRBZkJnTlZIU01FR0RBV2dCUUk3QU41RnNrT2lxb1pOVDNSc2ozQVBDVXduREFQQmdOVkhSTUJBZjhFQlRBRApBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCU2JDUkx1TitGd2loQ3ZiamlmcjFldGViNUEzeVRWWjN2CkN6MEhMK2RYekJEbVI5OGxPdWgvSDV4bEVyT3U1emdjSXJLK3BtM0toQVl1NXhZOHBFSklwRXBCRExnSEVQNTQKMkc1NllydXcva1J2cWJ4VWlwUjIzTWJGZXpyUW9BeXNKbmZGRmZrYXNVWlBRSjg0dE05UStoTEpUcnp0VkczZgphcnVCd3h0SDA3bklZMTFRUnhZbERPQWx6ck9icWdvUUtwODFXVTBrTzdDRVd2Q1ZOenphb2dWV08rZUtvdUw5Ci9aQjVYQ1FVRlRScFlEQjB1aFk1NTAwR1kxYnRpRUVKaVdZVTg0UVFzampINVZlRkFtN21ldWRkak9pTEM3dUMKSmFISkRLa0txZWtDSkZ6YzF0QmpHQVZUNHZaTGcrUldhUmJHa01Qdm1WZVFIOUtSMVN2aQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    service:
      name: xtradb-galera
      path: /
      port: 3306
      scheme: mysql
    url: tcp(xtradb-galera.demo.svc:3306)/
  secret:
    name: xtradb-galera-auth
  type: perconaxtradb
  tlsSecret:
    name: xtradb-galera-client-cert
  version: 8.0.26
```

Now we will see how we have filled out the appbinding for each fields. 

`spec.clientConfig.caBundle` : We got the `ca.crt` field value from the `xtradb-galera-client-cert` secret, encoded with base64 and placed it here.

`spec.clientConfig.service.name` : The service name which we created to communicate with the xtradb cluster. 

`spec.clientConfig.service.port` : Took the value as the primary service port.

`spec.clientConfig.service.shceme` : This will always be mysql, as Percona XtraDB is a mysql fork.

`spec.clientConfig.url` : Just followed the Kubernetes convention to hit a service in a specific namespace to a specific port and path. 

`spec.secret.name` : This is the root secret name. You can replace it with some other user credential rather than root. In that case make sure the user has got proper privileges on the sys.* and mysql.user tables. 

`spec.type` : This is set to `perconaxtradb` as our operator knows it that way. 

`spec.tlsSecret.name` : This is the secret reference which carries the cient certs for tls secured connections. The secret should contain `ca.crt`,`tls.crt` and `tls.key` keys and corresponding values.  

`spec.version` : The XtraDB version is mentioned here.

These are enough information to set up a ProxySQL server/cluster for the Percona XtraDB cluster. Now we will apply this to our cluster and refer the appbinding name in the ProxySQL yaml. 

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/quickstart/xtradbext/examples/appbinding.yaml
appbinding.appcatalog.appscode.com/xtradb-galera-appbinding created
```

We are ready with our backend appbinding. But before we proceed to the ProxySQL server, lets first create some test user and database so that we can use them for testing.

Let's first create a user in the backend xtradb server and a database to test the proxy traffic .

```bash
$ kubectl exec -it -n demo xtradb-galera-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1602
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create user `test`@'%' identified by 'pass';
Query OK, 0 rows affected (0.00 sec)

mysql> create database test;
Query OK, 1 row affected (0.01 sec)

mysql> use test;
Database changed

mysql> show tables;
Empty set (0.00 sec)

mysql> create table testtb(name varchar(103), primary key(name));
Query OK, 0 rows affected (0.01 sec)

mysql> grant all privileges on test.* to 'test'@'%';
Query OK, 0 rows affected (0.00 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.00 sec)

mysql> exit
Bye
```

We are now ready with our backend. In the next section we will set up our ProxySQL for this backend.

## Deploy ProxySQL Server 

With the following yaml we are going to create our desired ProxySQL server.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.4.4-debian"
  replicas: 1
  syncUsers: true
  backend:
    name: xtradb-galera-appbinding
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/quickstart/xtradbext/examples/sample-proxysql-v1.yaml
  proxysql.kubedb.com/proxysql-server created
```


```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.4.4-debian"
  replicas: 1
  syncUsers: true
  backend:
    name: xtradb-galera-appbinding
  terminationPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/quickstart/xtradbext/examples/sample-proxysql-v1alpha2.yaml
  proxysql.kubedb.com/proxysql-server created
```

This is the simplest version of a KubeDB ProxySQL server. Here in the `.spec.version` field we are saying that we want a ProxySQL-2.4.4 with base image of debian. In the `.spec.replicas` section we have written 1, so the operator will create a single node ProxySQL. The `spec.syncUser` field is set to  true, which means all the users in the backend MySQL server will be fetched to the ProxySQL server. 

Let's wait for the ProxySQL to be Ready. 

```bash
$ kubectl get proxysql -n demo
NAME           VERSION        STATUS   AGE
proxy-server   2.4.4-debian   Ready    4m
```

Let's check the pod.

```bash
$ kubectl get pods -n demo | grep proxy
proxy-server-0   1/1     Running   0          4m
```

### Check Associated Kubernetes Objects

KubeDB operator will create some services and secrets for the ProxySQL object. Let's check. 

```bash
$ kubectl get svc,secret -n demo | grep proxy
service/proxy-server          ClusterIP   10.96.181.182   <none>        6033/TCP            4m
service/proxy-server-pods     ClusterIP   None            <none>        6032/TCP,6033/TCP   4m
secret/proxy-server-auth             kubernetes.io/basic-auth              2      4m
secret/proxy-server-configuration    Opaque                                1      4m
secret/proxy-server-monitor          kubernetes.io/basic-auth              2      4m
```

You can find the description of the associated objects here. 

### Check Internal Configuration 

Let's exec into the ProxySQL server pod and get into the admin panel. 

```bash
$ kubectl exec -it -n demo proxy-mysql-0 -- bash                                                  11:20
root@proxy-mysql-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > " 
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 1204
Server version: 8.0.35 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > 
```

Let's check the mysql_servers table first. We didn't set it from the yaml. The KubeDB operator will do that for us. 

```bash
ProxySQLAdmin > select * from mysql_servers;
+--------------+------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname               | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | xtradb-galera.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 1       | 0              |         |
+--------------+------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+

1 rows in set (0.000 sec)
```

Let's check the mysql_users table. 

```bash
ProxySQLAdmin > select username from mysql_users;
+----------+
| username |
+----------+
| root     |
| monitor  |
| test     |
+----------+
2 rows in set (0.000 sec)
```

So we are now ready to test our traffic proxy. In the next section we are going to have some demo's. 

### Check Traffic Proxy

To test the traffic routing through the ProxySQL server let's first create a pod with ubuntu base image in it. We will use the following yaml. 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: ubuntu
  name: ubuntu
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ubuntu
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: ubuntu
    spec:
      containers:
        - image: ubuntu
          imagePullPolicy: IfNotPresent
          name: ubuntu
          command: ["/bin/sleep", "3650d"]
          resources: {}
```

Let's apply the yaml. 

```yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/quickstart/xtradbext/examples/ubuntu.yaml
deployment.apps/ubuntu created
```

Let's exec into the pod and install mysql-client. 

```bash
$ kubectl exec -it -n demo ubuntu-867d4588d8-tl7hh -- bash                12:00
root@ubuntu-867d4588d8-tl7hh:/# apt update
... ... ..
root@ubuntu-867d4588d8-tl7hh:/# apt install mysql-client -y
Reading package lists... Done
... .. ...
root@ubuntu-867d4588d8-tl7hh:/#
```

Now let's try to connect with the ProxySQL server through the `proxy-server` service as the `test` user. 

```bash
root@ubuntu-867d4588d8-tl7hh:/# mysql -utest -ppass -hproxy-server.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1881
Server version: 8.0.35 (ProxySQL)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> 
```

We are successfully connected as the `test` user. Let's run some read/write query on this connection.

```bash
mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| test               |
+--------------------+
2 rows in set (0.00 sec)

mysql> use test;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
mysql> show tables;
+----------------+
| Tables_in_test |
+----------------+
| testtb         |
+----------------+
1 row in set (0.00 sec)

mysql> insert into testtb(name) values("Kim Torres");
Query OK, 1 row affected (0.01 sec)

mysql> insert into testtb(name) values("Tony SoFua");
Query OK, 1 row affected (0.01 sec)

mysql> select * from testtb;
+------------+
| name       |
+------------+
| Kim Torres |
| Tony SoFua |
+------------+
2 rows in set (0.00 sec)

mysql> 
```

We can see the queries are successfully executed through the ProxySQL server. 

We can see that the read-write queries are successfully executed in the ProxySQL server. So the ProxySQL server is ready to use.

## Conclusion 

In this tutorial we have seen some very basic version of KubeDB ProxySQL. KubeDB provides many more for ProxySQL. In this site we have discussed on lot's of other features like `TLS Secured ProxySQL` , `Declarative Configuration` , `MariaDB and Percona-XtraDB Backend` , `Reconfigure` and much more. Checkout out other docs to learn more. 