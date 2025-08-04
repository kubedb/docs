---
title: Cassandra Topology TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: cas-tls-topology
    name: Topology Cluster
    parent: cas-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Cassandra with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for Cassandra. This tutorial will show you how to use KubeDB to run a Cassandra cluster with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses following crd fields to enable SSL/TLS encryption in Cassandra.

- `spec:`
    - `tls:`
        - `issuerRef`
        - `certificate`

Read about the fields in details in [cassandra concept](/docs/guides/cassandra/concepts/cassandra.md),

`tls` is applicable for all types of Cassandra (i.e., `standalone` and `topology`).

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets. These certificate secrets are then used to generate required certificates including `ca.crt`, `tls.crt`, `tls.key`, `keystore.jks` and `truststore.jks`.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Cassandra. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=cassandra/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls cassandra-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: cassandra-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: cassandra-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/tls/cas-issuer.yaml
issuer.cert-manager.io/cassandra-ca-issuer created
```

## TLS/SSL encryption in Cassandra Topology Cluster

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod-tls
  namespace: demo
spec:
  version: 5.0.3
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: cassandra-ca-issuer
  topology:
    rack:
      - name: r0
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 2Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
```

### Deploy Cassandra Topology Cluster with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/tls/cassandra-prod-tls.yaml
cassandra.kubedb.com/cassandra-prod-tls created
```

Now, wait until `cassandra-prod-tls created` has status `Ready`. i.e,

```bash
$ kubectl get cassandra -n demo -w
NAME                 TYPE                  VERSION   STATUS         AGE
cassandra-prod-tls   kubedb.com/v1alpha2   5.0.3     Provisioning   20s
cassandra-prod-tls   kubedb.com/v1alpha2   5.0.3     Provisioning   81s
.
.
cassandra-prod-tls   kubedb.com/v1alpha2   5.0.3     Ready          104s
```

### Verify TLS/SSL in Cassandra Topology Cluster

```bash
$  kubectl describe secret cassandra-prod-tls-client-cert -n demo
Name:         cassandra-prod-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=cassandra-prod-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=cassandras.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.cassandra-prod-tls-rack-r0-pods.demo.svc.cluster.local,cassandra-prod-tls,cassandra-prod-tls-rack-r0-pods,cassandra-prod-tls-rack-r0-pod...
              cert-manager.io/certificate-name: cassandra-prod-tls-client-cert
              cert-manager.io/common-name: cassandra-prod-tls.demo.svc
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: cassandra-ca-issuer
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1159 bytes
tls.crt:  1578 bytes
tls.key:  1704 bytes
```

Now, Let's exec into a cassandra pod and verify the configuration that the TLS is enabled.

```bash
$  kubectl exec -it -n demo cassandra-prod-tls-rack-r0-0 -- cqlsh -u admin -p qAbFK0B8gtUgj3Gp
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connection error: ('Unable to connect to any servers', {'127.0.0.1:9042': ConnectionShutdown('Connection to 127.0.0.1:9042 was closed')})
command terminated with exit code 1
sabbir@sabbir-pc ~/g/s/s/t/C/tls (main) [1]> kubectl exec -it -n demo cassandra-prod-tls-rack-r0-0 -- cqlsh -u admin -p qAbFK0B8gtUgj3Gp --ssl
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connected to Test Cluster at 127.0.0.1:9042
[cqlsh 6.2.0 | Cassandra 5.0.3 | CQL spec 3.4.7 | Native protocol v5]
Use HELP for help.
admin@cqlsh> 
```

We can see from the above output that, cqlsh can only be accessed through --ssl flag  which means that TLS is enabled.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cassandra -n demo cassandra-prod-tls
kubectl delete issuer -n demo cassandra-ca-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Monitor your Cassandra cluster with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).
- Monitor your Cassandra cluster with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/cassandra/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
