---
title: Elasticsearch Topology Cluster TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: es-tls-topology
    name: Topology Cluster
    parent: es-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Elasticsearch with TLS/SSL (Transport Encryption)

KubeDB supports providing TLS/SSL encryption for Elasticsearch. This tutorial will show you how to use KubeDB to run an Elasticsearch topology cluster with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manager`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses the following crd fields to enable SSL/TLS encryption in Elasticsearch.

- `spec:`
    - `enableSSL`
    - `tls:`
        - `issuerRef`
        - `certificates`

Read about the fields in details in [elasticsearch concept](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

`tls` is applicable for all types of Elasticsearch (i.e., `combined` and `topology`).

Users must specify the `tls.issuerRef` field. KubeDB uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificates` to generate certificate secrets. These certificate secrets are then used to configure TLS for both the transport layer (node-to-node communication) and the HTTP layer (client-to-node communication), containing `ca.crt`, `tls.crt` and `tls.key`.

> Note: `tls.issuerRef` is optional. A user can deploy Elasticsearch without creating an `Issuer`/`ClusterIssuer` by just setting `enableSSL: true`. In that case, the KubeDB Elasticsearch operator automatically creates a self-signed CA and the necessary certificate secrets.

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Elasticsearch. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificate using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=elasticsearch/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls es-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/es-ca created
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: es-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: es-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/tls/es-issuer.yaml
issuer.cert-manager.io/es-ca-issuer created
```

## TLS/SSL encryption in Elasticsearch Topology Cluster

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-topology-tls
  namespace: demo
spec:
  version: xpack-8.19.9
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: es-ca-issuer
  topology:
    master:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    data:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    ingest:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

### Deploy Elasticsearch Topology Cluster with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/tls/es-topology-tls.yaml
elasticsearch.kubedb.com/es-topology-tls created
```

Now, wait until `es-topology-tls` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo -w
NAME              VERSION        STATUS         AGE
es-topology-tls   xpack-8.19.9   Provisioning   0s
es-topology-tls   xpack-8.19.9   Provisioning   18s
.
.
es-topology-tls   xpack-8.19.9   Ready          2m5s
```

### Verify TLS/SSL in Elasticsearch Topology Cluster

KubeDB creates a client certificate secret for Elasticsearch. Let's check it:

```bash
$ kubectl describe secret -n demo es-topology-tls-client-cert
Name:         es-topology-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=es-topology-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=elasticsearches.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.es-topology-tls-pods.demo.svc,*.es-topology-tls-pods.demo.svc.cluster.local,es-topology-tls,es-topology-tls.demo.svc,localhost
              cert-manager.io/certificate-name: es-topology-tls-client-cert
              cert-manager.io/common-name: es-topology-tls
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: es-ca-issuer
              cert-manager.io/subject-organizations: kubedb
              cert-manager.io/uri-sans: 

Type:  kubernetes.io/tls

Data
====
ca.crt:   1172 bytes
tls.crt:  1387 bytes
tls.key:  1704 bytes
```

Now, let's exec into the master node and verify the configuration that TLS is enabled for both transport and HTTP layers.

```bash
$ kubectl exec -n demo es-topology-tls-master-0 -c elasticsearch -- \
                                      cat /usr/share/elasticsearch/config/elasticsearch.yml | grep -A 2 -i xpack.security
xpack.security.enabled: true

xpack.security.transport.ssl.enabled: true
xpack.security.transport.ssl.verification_mode: certificate
xpack.security.transport.ssl.key: certs/transport/tls.key
xpack.security.transport.ssl.certificate: certs/transport/tls.crt 
xpack.security.transport.ssl.certificate_authorities: [ "certs/transport/ca.crt" ]

xpack.security.http.ssl.enabled: true
xpack.security.http.ssl.key:  certs/http/tls.key
xpack.security.http.ssl.certificate: certs/http/tls.crt
xpack.security.http.ssl.certificate_authorities: [ "certs/http/ca.crt" ]
```

We can see from the above output that both `xpack.security.transport.ssl.enabled: true` and `xpack.security.http.ssl.enabled: true` are set, which means TLS is enabled for both node-to-node and client-to-node communication across all topology node roles.

Now, let's exec into the master node and connect using HTTPS to confirm the topology cluster is accessible with TLS.

```bash
$ kubectl exec -it -n demo es-topology-tls-master-0 -c elasticsearch -- curl -k -XGET "https://localhost:9200/_cluster/health?pretty" --user "elastic:$ELASTIC_USER_PASSWORD"
{
  "cluster_name" : "es-topology-tls",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 4,
  "number_of_data_nodes" : 2,
  "active_primary_shards" : 1,
  "active_shards" : 2,
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

From the above output, we can see that we are able to connect to the Elasticsearch topology cluster using the TLS configuration. The cluster has 4 nodes total (1 master + 2 data + 1 ingest) and is reporting `green` status.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete es -n demo es-topology-tls
kubectl delete issuer -n demo es-ca-issuer
kubectl delete secret -n demo es-ca
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/topology-cluster/hot-warm-cold-cluster/index.md).
- Monitor your Elasticsearch database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
