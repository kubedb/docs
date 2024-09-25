---
title: Encryption with Vault KMIP
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-integration-with-vault-kmip
    name: Encryption with Vault KMIP
    parent: guides-mongodb-integration-with-vault
    weight: 20
menu_name: docs_{{ .version }}
---

# Encrypt data in KubeDB MongoDB with Hashicorp Vault KMIP Secret Engine.

To demonstrate how to configure KubeDB MongoDB with [HashiCorp Vault KMIP secret engine](https://developer.hashicorp.com/vault/docs/secrets/kmip) for encryption, you can follow this step-by-step example. This documentation will guide you through setting up Vault, configuring the KMIP secret engine, and then configuring KubeDB to use it for MongoDB data encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).

- HashiCorp Vault instance with the KMIP secret engine enabled.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

### Setup Hashicorp Vault KMIP secret engine

User can setup Vault KMIP secret engine with [Vault Enterprise](https://developer.hashicorp.com/vault/tutorials/adp/kmip-engine?variants=vault-deploy%3Aenterprise) or [HCP Vault Dedicated](https://developer.hashicorp.com/vault/tutorials/adp/kmip-engine?variants=vault-deploy%3Ahcp).
For this demo we will use [Hashicorp Cloud Provider(HCP)](https://portal.cloud.hashicorp.com/) Vault Dedicated.

So First we created a `Vault Plus` cluster in HCP. Then we need to configure Vault KMIP according to [this](https://developer.hashicorp.com/vault/tutorials/adp/kmip-engine?variants=vault-deploy%3Ahcp) documentation step by step.

```bash
$ export VAULT_ADDR=<Public_Cluster_URL>
$ export VAULT_TOKEN=<Generated_Vault_Token>
$ export VAULT_NAMESPACE=admin

$ vault secrets enable kmip
Success! Enabled the kmip secrets engine at: kmip/

$ vault write kmip/config \
     listen_addrs=0.0.0.0:5696 \
     server_hostnames=$(echo ${VAULT_ADDR:8} | rev | cut -c6- | rev)
Success! Data written to: kmip/config

$ vault write -f kmip/scope/finance
Success! Data written to: kmip/scope/finance

$ vault write kmip/scope/finance/role/accounting operation_all=true
Success! Data written to: kmip/scope/finance/role/accounting

$ vault read kmip/ca -format=json | jq -r '.data | .ca_pem' >> vault-ca.pem

$ vault write -format=json \
    kmip/scope/finance/role/accounting/credential/generate \
    format=pem > credential.json

$ jq -r .data.certificate < credential.json > cert.pem

$ jq -r .data.certificate < credential.json > cert.pem

$ cat cert.pem key.pem > client.pem
```
We will use this `client.pem` and `vault-ca.pem` files to configure KMIP in MongoDB.

### Create MongoDB configuration with KMIP

Now we need to make a `mongod.conf` file to use it as configuration folder for our `MongoDB`.

```bash
$ cat mongod.conf
security:
  enableEncryption: true
  kmip:
    serverName: <kmip address or public cluster address for HCP Vault Cluster without port>
    port: 5696
    clientCertificateFile: /etc/certs/client.pem
    serverCAFile: /etc/certs/ca.pem
```
Here `/etc/certs/client.pem` and `/etc/certs/ca.pem` will be mounted by secret in KubeDB MongoDB main `mongodb` container.

Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo mg-configuration --from-file=./mongod.conf
secret/mg-configuration created
```

Verify the secret has the configuration file.
```bash
$ kubectl get secret -n demo mg-configuration -o yaml
apiVersion: v1
data:
  mongod.conf: c2VjdXJpdHk6CiAgZW5hYmxlRW5jcnlwdGlvbjogdHJ1ZQogIGttaXA6CiAgICBzZXJ2ZXJOYW1lOiB2YXVsdC1jbHVzdGVyLWRvYy1wdWJsaWMtdmF1bHQtYTMzYmI3NjEuMzcxMzFkZDEuejEuaGFzaGljb3JwLmNsb3VkCiAgICBwb3J0OiA1Njk2CiAgICBjbGllbnRDZXJ0aWZpY2F0ZUZpbGU6IC9ldGMvY2VydHMvY2xpZW50LnBlbQogICAgc2VydmVyQ0FGaWxlOiAvZXRjL2NlcnRzL2NhLnBlbQ==
kind: Secret
metadata:
  creationTimestamp: "2024-09-24T09:10:55Z"
  name: mg-configuration
  namespace: demo
  resourceVersion: "322831"
  uid: 005f0cac-6bbb-4fb6-a728-87b0ca55785a
type: Opaque

$ echo c2VjdXJpdHk6CiAgZW5hYmxlRW5jcnlwdGlvbjogdHJ1ZQogIGttaXA6CiAgICBzZXJ2ZXJOYW1lOiB2YXVsdC1jbHVzdGVyLWRvYy1wdWJsaWMtdmF1bHQtYTMzYmI3NjEuMzcxMzFkZDEuejEuaGFzaGljb3JwLmNsb3VkCiAgICBwb3J0OiA1Njk2CiAgICBjbGllbnRDZXJ0aWZpY2F0ZUZpbGU6IC9ldGMvY2VydHMvY2xpZW50LnBlbQogICAgc2VydmVyQ0FGaWxlOiAvZXRjL2NlcnRzL2NhLnBlbQ== | base64 -d
security:
  enableEncryption: true
  kmip:
    serverName: vault-cluster-doc-public-vault-a33bb761.37131dd1.z1.hashicorp.cloud
    port: 5696
    clientCertificateFile: /etc/certs/client.pem
    serverCAFile: /etc/certs/ca.pem
```

### Create MongoDB

Before creating `MongoDB`, we need to create a secret with `client.pem` and `vault-ca.pem` to use as volume for our `MongoDB`
```bash
$ kubectl create secret generic vault-tls-secret -n demo \
        --from-file=client.pem=client.pem \
        --from-file=ca.pem=vault-ca.pem
secret/vault-tls-secret created
```

Now lets create KubeDB MongoDB. We will use mongodb version `percona-5.0.23` for our demo purpose.
```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-kmip
  namespace: demo
spec:
  podTemplate:
    spec:
      containers:
        - name: "mongodb"
          volumeMounts:
            - name: certs
              mountPath: /etc/certs
      volumes:
      - name: certs
        secret:
          secretName: vault-tls-secret
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  storageType: Durable
  deletionPolicy: WipeOut
  version: "percona-5.0.23"
  configSecret:
    name: mg-configuration
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guids/mongodb/vault-integration/kmip-enryption/examples/mg.yaml
mongodb.kubedb.com/mg-kmip created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `mg-kmip-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo mg-kmip-0
NAME                  READY     STATUS    RESTARTS   AGE
mg-kmip-0             1/1       Running   0          1m
```

Now, we will check if the database has started with the custom configuration we have provided.

To make sure that this `mg-kmip` MongoDB is KMIP encrypted, we can check the log of this `mg-kmip-0` pod

```bash
kubectl logs -f --all-containers -n demo mg-kmip-0
```
We should see these logs which confirm that this `MongoDB` is setup with KMIP
```log
{"t":{"$date":"2024-09-24T09:26:13.551+00:00"},"s":"I",  "c":"STORAGE",  "id":29116,   "ctx":"initandlisten","msg":"Master encryption key has been created on the key management facility","attr":{"keyManagementFacilityType":"KMIP server","keyIdentifier":{"kmipKeyIdentifier":"73ORm3aFQxGKZtJQ3196VXV5NmfT3AlG"}}}
{"t":{"$date":"2024-09-24T09:26:13.551+00:00"},"s":"I",  "c":"STORAGE",  "id":29037,   "ctx":"initandlisten","msg":"Initializing KeyDB with wiredtiger_open config: {cfg}","attr":{"cfg":"create,config_base=false,extensions=[local=(entry=percona_encryption_extension_init,early_load=true,config=(cipher=AES256-CBC,rotation=false))],encryption=(name=percona,keyid=\"\"),log=(enabled,file_max=5MB),transaction_sync=(enabled=true,method=fsync),"}}
{"t":{"$date":"2024-09-24T09:26:13.799+00:00"},"s":"I",  "c":"STORAGE",  "id":29039,   "ctx":"initandlisten","msg":"Encryption keys DB is initialized successfully"}
```


Now, we can connect to this database through [mongo-shell](https://docs.mongodb.com/v4.2/mongo/). In this tutorial, we are connecting to the MongoDB server from inside the pod.

```bash
$ kubectl get secrets -n demo mg-kmip-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mg-kmip-auth -o jsonpath='{.data.\password}' | base64 -d
bJI!1H!)V7!2U.wJ

$ kubectl exec -it mg-kmip-0 -n demo -- bash

> mongo admin

> db.auth("root","bJI!1H!)V7!2U.wJ")
1

> db._adminCommand( {getCmdLineOpts: 1})
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--port=27017",
		"--ipv6",
		"--bind_ip=::,0.0.0.0",
		"--tlsMode=disabled",
		"-f",
		"/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "::,0.0.0.0",
			"ipv6" : true,
			"port" : 27017,
			"tls" : {
				"mode" : "disabled"
			}
		},
		"security" : {
			"authorization" : "enabled",
			"enableEncryption" : true,
			"kmip" : {
				"clientCertificateFile" : "/etc/certs/client.pem",
				"port" : 5696,
				"serverCAFile" : "/etc/certs/ca.pem",
				"serverName" : "vault-cluster-doc-public-vault-a33bb761.37131dd1.z1.hashicorp.cloud"
			}
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1
}
> exit
bye
```

We can see that in `parsed.security` field, encryption is enabled.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo mg/mg-kmip

kubectl delete -n demo secret mg-configuration
kubectl delete -n demo secret vault-tls-secret

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/kubestash/overview/index.md) MongoDB databases using KubeStash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).



