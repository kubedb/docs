---
title: Proxy Load To MySQL Group Replication With KubeDB Provisioned ProxySQL
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-quickstart-overview
    name: Concepts
    parent: guides-proxysql-quickstart
    weight: 20
menu_name: docs_{{ .version }}
---

/*
 formalities goes here...
*/

In this tutorial, we will learn how to setup a ProxySQL server/cluster with KubeDB, to proxy the incoming traffic for a MySQL Group Replication. We will learn all the features that KubeDB ProxySQL provides for the users thoughout this tutorial. The following steps should be followed : 
1. Ensure MySQL backend and its Appbinding.
2. Setup the ProxySQL.
3. Check traffic loads and track the routings. 
4. Cleanup.


## Ensure MySQL backend and its Appbinding for ProxySQL 

 

### The MySQL Backend

### Create Appbinding

We are assuming the mysql is from a external source. So we need to create an appbinding and mention its name under the `.spec.backend.name` field of the ProxySQL CRD to provide the neccessary connection information to the proxysql server. You can checkout the api documentation for Appbinding CRD [here](#edit). Also if you already have a KubeDB provisioned MySQL then you don't need to create another appbinding as KubeDB operator has already created one. Just mention the KubeDB made appbinding name in that place. The following one is what we have created for this tutorial. 

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mysql-backend-apb
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJekNDQWd1Z0F3SUJBZ0lVWUp4dVBqcW1EbjJPaVdkMGk5cUZ2MGdzdzQwd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0lURU9NQXdHQTFVRUF3d0ZiWGx6Y1d3eER6QU5CZ05WQkFvTUJtdDFZbVZrWWpBZUZ3MHlNakV3TVRBeApNalE1TlROYUZ3MHlNekV3TVRBeE1qUTVOVE5hTUNFeERqQU1CZ05WQkFNTUJXMTVjM0ZzTVE4d0RRWURWUVFLCkRBWnJkV0psWkdJd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUURkRTRMaEFabEQKQTh2aWk5QTlaT1VLOEVzeWU2ZDBWWnpRZmJUeS90VlF0N05ybkhjS1NkVHFmS0lkczl1bXhSd1Y2ak5WU2RtUQp2M3NUa0xwYTZBbFRwYklQazZ5S2UxRGs2YUhhbFZDSVc4SExLNW43YklzTEV3aEkyb3F4WmIrd0pydXVGSi95Clk4a2tyazVXaFBJZzRqM3VjV0FhcllqTVpxNXRQbU9sOFJXanhzY2k3WjJsN0lIcWplSjYrKzRENXlkeGx6L3gKdVhLNmxVM2J2Y2craWhUVno1UENNS2R4OHZNL052dTNuQ21Ja2hzQUkrNGVWZE4xenRIWG51UTFXRFhlUEFFYwpRQnJGanFuWk5pYmRUeU4zYTgrdmVUM2hLK3Fhc0ZFSU5aOFY4ZFVQSVV5cHFYSmk0anZCSW9FU0RvV2V1Z3QzCklhMjh6OE5XNk9WbkFnTUJBQUdqVXpCUk1CMEdBMVVkRGdRV0JCUUk3QU41RnNrT2lxb1pOVDNSc2ozQVBDVXcKbkRBZkJnTlZIU01FR0RBV2dCUUk3QU41RnNrT2lxb1pOVDNSc2ozQVBDVXduREFQQmdOVkhSTUJBZjhFQlRBRApBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCU2JDUkx1TitGd2loQ3ZiamlmcjFldGViNUEzeVRWWjN2CkN6MEhMK2RYekJEbVI5OGxPdWgvSDV4bEVyT3U1emdjSXJLK3BtM0toQVl1NXhZOHBFSklwRXBCRExnSEVQNTQKMkc1NllydXcva1J2cWJ4VWlwUjIzTWJGZXpyUW9BeXNKbmZGRmZrYXNVWlBRSjg0dE05UStoTEpUcnp0VkczZgphcnVCd3h0SDA3bklZMTFRUnhZbERPQWx6ck9icWdvUUtwODFXVTBrTzdDRVd2Q1ZOenphb2dWV08rZUtvdUw5Ci9aQjVYQ1FVRlRScFlEQjB1aFk1NTAwR1kxYnRpRUVKaVdZVTg0UVFzampINVZlRkFtN21ldWRkak9pTEM3dUMKSmFISkRLa0txZWtDSkZ6YzF0QmpHQVZUNHZaTGcrUldhUmJHa01Qdm1WZVFIOUtSMVN2aQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    service:
      name: mysql-server
      path: /
      port: 3306
      scheme: mysql
    url: tcp(mysql-server.demo.svc:3306)/
  secret:
    name: mysql-server-auth
  tlsSecret:
    name: mysql-server-client-cert 
  type: mysql
  version: 8.0.27
  ```
You can skip the `.spec.clientConfig.caBundle` and the `.spec.tlsSecret` sections if your mysql server is not TLS/SSL secured or if the information is redundant for the proxysql servers. 

Let's apply the appbinding in our cluster . 

```bash 
$ kubectl apply -f #edit
appbinding.appcatalog.appscode.com/mysql-backend-apb created
```

## Setup the ProxySQL

Now we have a mysql backend. And to connect with it we have an appbinding. For a high level reference we can assuem our proxysql as following,
* It will proxy for the `mysql-server`. Get the connection information from the `mysql-backend-apb` which will be mentioned in the `.spec.backend.name` section.
* The proxysql servers will have certain configuration which will be set from the `.spec.initConfig` and `.spec.configSecret` field.
* With all the similar configuration we will have a cluster of proxysql servers which are inter-connected and we will configure the numbers of cluster member from the `.spec.replicas` section.

