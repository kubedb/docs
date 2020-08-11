---
title: X-Pack Certificate
menu:
  docs_{{ .version }}:
    identifier: es-issue-certificate-x-pack
    name: Issue Certificate
    parent: es-x-pack
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Issue TLS Certificates

X-Pack requires certificates to enable TLS. KubeDB creates necessary certificates automatically. However, if you want to use your own certificates, you can provide them through `spec.certificateSecret` field of Elasticsearch object.

This tutorial will show you how to generate certificates for X-Pack and use them with Elasticsearch database.

In KubeDB Elasticsearch, keystore and truststore files in JKS format are used instead of certificates and private keys in PEM format.

KubeDB applies same **truststore**  for both transport layer TLS and REST layer TLS.

But, KubeDB distinguishes between the following types of keystore for security purpose.

- **transport layer keystore** are used to identify and secure traffic between Elasticsearch nodes on the transport layer
- **http layer keystore** are used to identify Elasticsearch clients on the REST and transport layer.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

You also need to have [*OpenSSL*](https://www.openssl.org/source/) and Java *keytool* for generating all required artifacts.

In order to find out if you have OpenSSL installed, open a terminal and type

```console
$ openssl version
OpenSSL 1.0.2g  1 Mar 2016
```

Make sure it’s version 1.0.1k or higher

And check *keytool* by calling

```console
keytool
```

If already installed, it will print a list of available commands.

To keep generated files separated, open a new terminal and create a directory `/tmp/kubedb/certs`

```console
mkdir -p /tmp/kubedb/certs
cd /tmp/kubedb/certs
```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Generate truststore

First, we need root certificate to sign other server & client certificates. And also this certificate is imported as *truststore*.

You need to follow these steps

1. Get root certificate configuration file

    ```console
    $ wget https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/x-pack/openssl-config/openssl-ca.ini
    ```

    ```ini
    [ ca ]
    default_ca = CA_default

    [ CA_default ]
    private_key     = root-key.pem
    default_days    = 1000        # how long to certify for
    default_md      = sha256      # use public key default MD
    copy_extensions = copy        # Required to copy SANs from CSR to cert

    [ req ]
    prompt             = no
    default_bits       = 4096
    distinguished_name = ca_distinguished_name

    [ ca_distinguished_name ]
    O  = Elasticsearch Operator
    CN = KubeDB Com. Root CA
    ```

2. Set a password of your keystore and truststore files

    ```console
    $ export KEY_PASS=secret
    ```

    > Note: You need to provide this KEY_PASS in your Secret as `key_pass`

3. Generate private key and certificate

    ```console
    $ openssl req -x509 -config openssl-ca.ini -newkey rsa:4096 -sha256 -nodes -out root.pem -keyout root-key.pem -batch -passin "pass:$KEY_PASS"
    ```

    Here,

    - `root-key.pem` holds Private Key
    - `root.pem`holds CA Certificate

4. Finally, import certificate as keystore

    ```console
    $ keytool -import -file root.pem -keystore root.jks -storepass $KEY_PASS -srcstoretype pkcs12 -noprompt
    ```

    Here,

    - `root.jks` is truststore for Elasticsearch

## Generate keystore

Here are the steps for generating certificate and keystore for Elasticsearch:

1. Get certificate configuration file
2. Generate private key and certificate signing request (CSR)
3. Sign certificate using root certificate
4. Generate PKCS12 file using root certificate
5. Import PKCS12 as keystore

You need to follow these steps to generate three keystore.

To sign certificate, we need another configuration file.

```console
$ wget https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/info.version" >}}/docs/examples/elasticsearch/x-pack/openssl-config/openssl-sign.ini
```

```ini
[ ca ]
default_ca = CA_default

[ CA_default ]
base_dir      = .
certificate   = $base_dir/root.pem          # The CA certifcate
private_key   = $base_dir/root-key.pem      # The CA private key
new_certs_dir = $base_dir                   # Location for new certs after signing
database      = $base_dir/index.txt         # Database index file
serial        = $base_dir/serial.txt        # The current serial number
unique_subject = no                         # Set to 'no' to allow creation of several certificates with same subject.

default_days    = 1000        # how long to certify for
default_md      = sha256      # use public key default MD
email_in_dn     = no
copy_extensions = copy        # Required to copy SANs from CSR to cert

[ req ]
default_bits       = 4096
default_keyfile    = root-key.pem
distinguished_name = ca_distinguished_name

[ ca_distinguished_name ]
O  = Elasticsearch Operator
CN = KubeDB Com. Root CA

[ signing_req ]
keyUsage               = digitalSignature, keyEncipherment

[ signing_policy ]
organizationName       = optional
commonName             = supplied
```

Here,

- `certificate` denotes CA certificate path
- `private_key` denotes CA key path

Also, you need to create a `index.txt` file and `serial.txt` file with value `01`

```console
touch index.txt
echo '01' > serial.txt
```

### Node

Following configuration is used to generate CSR for node certificate.

```ini
[ req ]
prompt             = no
default_bits       = 4096
distinguished_name = node_distinguished_name
req_extensions     = node_req_extensions

[ node_distinguished_name ]
O  = Elasticsearch Operator
CN = custom-certificate-es-ssl

[ node_req_extensions ]
keyUsage            = digitalSignature, keyEncipherment
extendedKeyUsage    = serverAuth, clientAuth
subjectAltName      = @alternate_names

[ alternate_names ]
DNS.1 = localhost
RID.1 = 1.2.3.4.5.5
```

Here,

- `RID.1=1.2.3.4.5.5` is used in node certificate. All certificates with registeredID `1.2.3.4.5.5` is considered as valid certificate for transport layer.

Now run following commands

```console
$ wget https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/x-pack/openssl-config/openssl-node.ini
$ openssl req -config openssl-node.ini -newkey rsa:4096 -sha256 -nodes -out node-csr.pem -keyout node-key.pem
$ openssl ca -config openssl-sign.ini -batch -policy signing_policy -extensions signing_req -out node.pem -infiles node-csr.pem
$ openssl pkcs12 -export -certfile root.pem -inkey node-key.pem -in node.pem -password "pass:$KEY_PASS" -out node.pkcs12
$ keytool -importkeystore -srckeystore node.pkcs12  -storepass $KEY_PASS  -srcstoretype pkcs12 -srcstorepass $KEY_PASS  -destkeystore node.jks -deststoretype pkcs12
```

Generated `node.jks` will be used as keystore for transport layer TLS.

### Client

Following configuration is used to generate CSR for client certificate.

```ini
[ req ]
prompt             = no
default_bits       = 4096
distinguished_name = client_distinguished_name
req_extensions     = client_req_extensions

[ client_distinguished_name ]
O  = Elasticsearch Operator
CN = custom-certificate-es-ssl

[ client_req_extensions ]
keyUsage            = digitalSignature, keyEncipherment
extendedKeyUsage    = serverAuth, clientAuth
subjectAltName      = @alternate_names

[ alternate_names ]
DNS.1 = localhost
DNS.2 = custom-certificate-es-ssl.demo.svc
```

Here,

- `custom-certificate-es-ssl` is used as a Common Name so that host `custom-certificate-es-ssl` is verified as valid Client.

Now run following commands

```console
$ wget https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/x-pack/openssl-config/openssl-client.ini
$ openssl req -config openssl-client.ini -newkey rsa:4096 -sha256 -nodes -out client-csr.pem -keyout client-key.pem
$ openssl ca -config openssl-sign.ini -batch -policy signing_policy -extensions signing_req -out client.pem -infiles client-csr.pem
$ openssl pkcs12 -export -certfile root.pem -inkey client-key.pem -in client.pem -password "pass:$KEY_PASS" -out client.pkcs12
$ keytool -importkeystore -srckeystore client.pkcs12  -storepass $KEY_PASS  -srcstoretype pkcs12 -srcstorepass $KEY_PASS  -destkeystore client.jks -deststoretype pkcs12
```

Generated `client.jks` will be used as keystore for http layer TLS.

## Create Secret

Now create a Secret with these certificates to use in your Elasticsearch object.

```console
$ kubectl create secret -n demo generic custom-certificate-es-ssl-cert \
                --from-file=root.pem \
                --from-file=root.jks \
                --from-file=node.jks \
                --from-file=client.jks \
                --from-literal=key_pass=$KEY_PASS

secret/custom-certificate-es-ssl-cert created
```

> Note: `root.pem` is added in Secret so that user can use these to connect Elasticsearch

Use this Secret `custom-certificate-es-ssl-cert` in your Elasticsearch object.

## Create an Elasticsearch database

Below is the Elasticsearch object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: custom-certificate-es-ssl
  namespace: demo
spec:
  version: 7.3.2
  enableSSL: true
  certificateSecret:
    secretName: custom-certificate-es-ssl-cert
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Here,

- `spec.certificateSecret` specifies Secret with certificates those will be used in Elasticsearch database.

Create example above with following command

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/x-pack/custom-certificate-es-ssl.yaml
elasticsearch.kubedb.com/custom-certificate-es-ssl created
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created.

```console
$ kubectl get es -n demo custom-certificate-es-ssl -o wide
NAME               VERSION   STATUS    AGE
custom-certificate-es-ssl   7.3.2    Running   1m
```

## Connect to Elasticsearch Database

We need to provide `root.pem` to connect to elasticsearch nodes.

Let's forward port 9200 of `custom-certificate-es-ssl-0` pod. Run following command in a separate terminal,

```console
$ kubectl port-forward -n demo custom-certificate-es-ssl-0 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, we can connect with the database at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username: Run following command to get *username*

  ```console
  $ kubectl get secrets -n demo custom-certificate-es-ssl-auth -o jsonpath='{.data.\ADMIN_USERNAME}' | base64 -d
  elastic
  ```

- Password: Run following command to get *password*

  ```console
  $ kubectl get secrets -n demo custom-certificate-es-ssl-auth -o jsonpath='{.data.\ADMIN_PASSWORD}' | base64 -d
  uft73z6j
  ```

- Root CA: Run following command to get `root.pem` file

  ```console
  $ kubectl get secrets -n demo custom-certificate-es-ssl-cert -o jsonpath='{.data.\root\.pem}' | base64 --decode > root.pem
  ```

Now, let's check health of our Elasticsearch database.

```console
$ curl --user "elastic:uft73z6j" "https://localhost:9200/_cluster/health?pretty" --cacert root.pem
```

```json
{
  "cluster_name" : "custom-certificate-es-ssl",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 1,
  "number_of_data_nodes" : 1,
  "active_primary_shards" : 0,
  "active_shards" : 0,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

Additionally, to query the settings about xpack,

```json
$ curl --user "elastic:uft73z6j" "https://localhost:9200/_nodes/_all/settings?pretty" --cacert root.pem
{
  "_nodes" : {
    "total" : 1,
    "successful" : 1,
    "failed" : 0
  },
  "cluster_name" : "custom-certificate-es-ssl",
  "nodes" : {
    "L75i6kmaRRWqy7-IqnDbbA" : {
      "name" : "custom-certificate-es-ssl-0",
      "transport_address" : "10.4.0.166:9300",
      "host" : "10.4.0.166",
      "ip" : "10.4.0.166",
      "version" : "7.3.2",
      "build_flavor" : "default",
      "build_type" : "docker",
      "build_hash" : "508c38a",
      "roles" : [
        "master",
        "data",
        "ingest"
      ],
      "attributes" : {
        "ml.machine_memory" : "7841263616",
        "xpack.installed" : "true",
        "ml.max_open_jobs" : "20"
      },
      "settings" : {
        "cluster" : {
          "initial_master_nodes" : "custom-certificate-es-ssl-0",
          "name" : "custom-certificate-es-ssl"
        },
        "node" : {
          "name" : "custom-certificate-es-ssl-0",
          "attr" : {
            "xpack" : {
              "installed" : "true"
            },
            "ml" : {
              "machine_memory" : "7841263616",
              "max_open_jobs" : "20"
            }
          },
          "data" : "true",
          "ingest" : "true",
          "master" : "true"
        },
        "path" : {
          "logs" : "/usr/share/elasticsearch/logs",
          "home" : "/usr/share/elasticsearch"
        },
        "discovery" : {
          "seed_hosts" : "custom-certificate-es-ssl-master"
        },
        "client" : {
          "type" : "node"
        },
        "http" : {
          "compression" : "false",
          "type" : "security4",
          "type.default" : "netty4"
        },
        "transport" : {
          "type" : "security4",
          "features" : {
            "x-pack" : "true"
          },
          "type.default" : "netty4"
        },
        "xpack" : {
          "security" : {
            "http" : {
              "ssl" : {
                "enabled" : "true"
              }
            },
            "enabled" : "true",
            "transport" : {
              "ssl" : {
                "enabled" : "true"
              }
            }
          }
        },
        "network" : {
          "host" : "0.0.0.0"
        }
      }
    }
  }
}
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo es/custom-certificate-es-ssl -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo es/custom-certificate-es-ssl

kubectl delete ns demo
```

## Next Steps

- Learn how to use TLS certificates to connect Elasticsearch from [here](/docs/guides/elasticsearch/x-pack/use-tls.md).
- Learn how to generate [x-pack configuration](/docs/guides/elasticsearch/x-pack/configuration.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
