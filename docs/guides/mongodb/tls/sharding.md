---
title: MongoDB Shard TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: mg-tls-shard
    name: Sharding
    parent: mg-tls
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Run MongoDB with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption (via, `sslMode` and `clusterAuthMode`) for MongoDB. This tutorial will show you how to use KubeDB to run a MongoDB database with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Mongodb.

- `spec:`
  - `sslMode`
  - `tls:`
    - `issuerRef`
    - `certificate`
  - `clusterAuthMode`

Read about the fields in details in [mongodb concept](/docs/guides/mongodb/concepts/mongodb.md),

`sslMode`, and `tls` is applicable for all types of MongoDB (i.e., `standalone`, `replicaset` and `sharding`), while `clusterAuthMode` provides [ClusterAuthMode](https://docs.mongodb.com/manual/reference/program/mongod/#cmdoption-mongod-clusterauthmode) for MongoDB clusters (i.e., `replicaset` and `sharding`).

When, SSLMode is anything other than `disabled`, users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `mongo.pem` and `client.pem`.

The subject of `client.pem` certificate is added as `root` user in `$external` mongodb database. So, user can use this client certificate for `MONGODB-X509` `authenticationMechanism`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in MongoDB. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mongo/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls mongo-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mongo-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: mongo-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/tls/issuer.yaml
issuer.cert-manager.io/mongo-ca-issuer created
```

## TLS/SSL encryption in MongoDB Sharding

Below is the YAML for MongoDB Sharding. Here, [`spec.sslMode`](/docs/guides/mongodb/concepts/mongodb.md#specsslMode) specifies `sslMode` for `sharding` and [`spec.clusterAuthMode`](/docs/guides/mongodb/concepts/mongodb.md#specclusterAuthMode) provides `clusterAuthMode` for sharding servers.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mongo-sh-tls
  namespace: demo
spec:
  version: "4.1.13-v1"
  sslMode: requireSSL
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: mongo-ca-issuer
  clusterAuthMode: x509
  shardTopology:
    configServer:
      replicas: 2
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 2
      shards: 2
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  terminationPolicy: WipeOut
```

### Deploy MongoDB Sharding

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/tls/mg-shard-ssl.yaml
mongodb.kubedb.com/mongo-sh-tls created
```

Now, wait until `mongo-sh-tls created` has status `Ready`. ie,

```bash
$ watch kubectl get mg -n demo
Every 2.0s: kubectl get mongodb -n demo
NAME           VERSION     STATUS     AGE
mongo-sh-tls   4.1.13-v1   Ready      4m24s
```

### Verify TLS/SSL in MongoDB Sharding

Now, connect to `mongos` component of this database through [mongo-shell](https://docs.mongodb.com/v4.0/mongo/) and verify if `SSLMode` and `ClusterAuthMode` has been set up as intended.

```bash
$ kubectl describe secret -n demo mongo-sh-tls-client-cert
Name:         mongo-sh-tls-client-cert
Namespace:    demo
Labels:       <none>
Annotations:  cert-manager.io/alt-names:
              cert-manager.io/certificate-name: mongo-sh-tls-client-cert
              cert-manager.io/common-name: root
              cert-manager.io/ip-sans:
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: mongo-ca-issuer
              cert-manager.io/uri-sans:

Type:  kubernetes.io/tls

Data
====
ca.crt:   1147 bytes
tls.crt:  1172 bytes
tls.key:  1679 bytes
```

Now, Let's exec into a mongodb container and find out the username to connect in a mongo shell,

```bash
$ kubectl exec -it mongo-sh-tls-mongos-0 -n demo bash
root@mongo-sh-tls-mongos-0:/$ ls /var/run/mongodb/tls
ca.crt  client.pem  mongo.pem
mongodb@mgo-sh-tls-mongos-0:/$ openssl x509 -in /var/run/mongodb/tls/client.pem -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=root,O=kubedb
```

Now, we can connect using `CN=root,O=kubedb` as root to connect to the mongo shell,

```bash
root@mongo-sh-tls-mongos-0:/# mongo --tls --tlsCAFile /var/run/mongodb/tls/ca.crt --tlsCertificateKeyFile /var/run/mongodb/tls/client.pem admin --host localhost --authenticationMechanism MONGODB-X509 --authenticationDatabase='$external' -u "CN=root,O=kubedb" --quiet
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	http://docs.mongodb.org/
Questions? Try the support group
	http://groups.google.com/group/mongodb-user
mongos>
```

We are connected to the mongo shell. Let's run some command to verify the sslMode and the user,

```bash
mongos> db.adminCommand({ getParameter:1, sslMode:1 })
{
	"sslMode" : "requireSSL",
	"ok" : 1,
	"operationTime" : Timestamp(1599491398, 1),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1599491398, 1),
		"signature" : {
			"hash" : BinData(0,"cn2Mhfy2blonon3jPz6Daen0nnc="),
			"keyId" : NumberLong("6869760899591176209")
		}
	}
}
mongos> use $external
switched to db $external
mongos> show users
{
	"_id" : "$external.CN=root,O=kubedb",
	"userId" : UUID("4865dda6-5e31-4b79-a085-7d6fea51c9be"),
	"user" : "CN=root,O=kubedb",
	"db" : "$external",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	],
	"mechanisms" : [
		"external"
	]
}
> exit
bye
```

You can see here that, `sslMode` is set to `requireSSL` and `clusterAuthMode` is set to `x509` and also an user is created in `$external` with name `"CN=root,O=kubedb"`.

## Changing the SSLMode & ClusterAuthMode

User can update `sslMode` & `ClusterAuthMode` if needed. Some changes may be invalid from mongodb end, like using `sslMode: disabled` with `clusterAuthMode: x509`.

The good thing is, **KubeDB operator will throw error for invalid SSL specs while creating/updating the MongoDB object.** i.e.,

```bash
$ kubectl patch -n demo mg/mgo-sh-tls -p '{"spec":{"sslMode": "disabled","clusterAuthMode": "x509"}}' --type="merge"
Error from server (Forbidden): admission webhook "mongodb.validators.kubedb.com" denied the request: can't have disabled set to mongodb.spec.sslMode when mongodb.spec.clusterAuthMode is set to x509
```

To **update from Keyfile Authentication to x.509 Authentication**, change the `sslMode` and `clusterAuthMode` in recommended sequence as suggested in [official documentation](https://docs.mongodb.com/manual/tutorial/update-keyfile-to-x509/). Each time after changing the specs, follow the procedure that is described above to verify the changes of `sslMode` and `clusterAuthMode` inside the database.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mongodb -n demo mongo-sh-tls
kubectl delete issuer -n demo mongo-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md)  MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
