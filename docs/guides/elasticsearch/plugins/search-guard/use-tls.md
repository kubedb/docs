---
title: Run TLS Secured Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-use-tls-search-guard
    name: Use TLS
    parent: es-search-guard-elasticsearch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run TLS Secured Elasticsearch

Search Guard provides facility to secure your Elasticsearch cluster with TLS. By default, KubeDB does not enable TLS security. You have to enable it by setting `spec.enableSSL: true`. If TLS is enabled, only HTTPS calls are allowed to database server.

This tutorial will show you how to connect with Elasticsearch cluster using certificate when TLS is enabled.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create Elasticsearch

In order to enable TLS, we have to set `spec.enableSSL` field of Elasticsearch object to `true`. Below is the YAML of Elasticsearch object that will be created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: ssl-elasticsearch
  namespace: demo
spec:
  version: searchguard-7.9.3
  replicas: 2
  enableSSL: true
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the Elasticsearch object we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/search-guard/ssl-elasticsearch.yaml
elasticsearch.kubedb.com/ssl-elasticsearch created
```

```bash
$ kubectl get es -n demo ssl-elasticsearch
NAME                STATUS    AGE
ssl-elasticsearch   Running   17m
```

## Connect to Elasticsearch Database

As we have enabled TLS for our Elasticsearch cluster, only HTTPS calls are allowed to Elasticsearch server. So, we need to provide certificate to connect with Elasticsearch. If you do not provide certificate manually through `spec.certificateSecret` field of Elasticsearch object, KubeDB will create a secret `{elasticsearch name}-cert` with necessary certificates.

Let's check the certificates that has been created for Elasticsearch `ssl-elasticsearch` by KubeDB operator.

```bash
$ kubectl get secret -n demo ssl-elasticsearch-cert -o yaml
```

```yaml
apiVersion: v1
data:
  client.jks: TFMwdExTMUNSVWRKVGlCLi4uLi49PQ==
  node.jks: TFMwdExTMUNSVWRKVGlCLi4uLi49PQ==
  root.jks: TFMwdExTMUNSVWRKVGlCLi4uLi49PQ==
  root.pem: TFMwdExTMUNSVWRKVGlCLi4uLi49PQ==
  sgadmin.jks: TFMwdExTMUNSVWRKVGlCLi4uLi49PQ==
  key_pass: TFMwdExTMUNSVWRKVGlCLi4uLi49PQ==
kind: Secret
metadata:
  creationTimestamp: 2018-02-19T09:51:45Z
  labels:
    app.kubernetes.io/name: elasticsearches.kubedb.com
    app.kubernetes.io/instance: ssl-elasticsearch
  name: ssl-elasticsearch-cert
  namespace: demo
  resourceVersion: "754"
  selfLink: /api/v1/namespaces/demo/secrets/ssl-elasticsearch-cert
  uid: 7efdaf31-155a-11e8-a001-42010a8000d5
type: Opaque
```

Here, `root.pem` file is the root CA in `.pem` format. We will require to provide this file while sending REST request to the Elasticsearch server.

Let's forward port 9200 of `ssl-elasticsearch-0` pod. Run following command in a separate terminal,

```bash
$ kubectl port-forward -n demo ssl-elasticsearch-0 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, we can connect with the database at `localhost:9200`.

**Connection information:**

- Address: `localhost:9200`
- Username: Run following command to get *username*

  ```bash
  $ kubectl get secrets -n demo ssl-elasticsearch-auth -o jsonpath='{.data.\ADMIN_USERNAME}' | base64 -d
  elastic
  ```

- Password: Run following command to get *password*

  ```bash
  $ kubectl get secrets -n demo ssl-elasticsearch-auth -o jsonpath='{.data.\ADMIN_PASSWORD}' | base64 -d
  uv2io5au
  ```

- Root CA: Run following command to get `root.pem` file
  
  ```bash
  $ kubectl get secrets -n demo ssl-elasticsearch-cert -o jsonpath='{.data.\root\.pem}' | base64 --decode > root.pem
  ```

Now, let's check health of our Elasticsearch database.

```bash
$ curl --user "elastic:uv2io5au" "https://localhost:9200/_cluster/health?pretty" --cacert root.pem
```

```json
{
  "cluster_name" : "ssl-elasticsearch",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 2,
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

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo es/ssl-elasticsearch -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo es/ssl-elasticsearch

$ kubectl delete ns demo
```

## Next Steps

- Learn how to [create TLS certificates](/docs/guides/elasticsearch/plugins/search-guard/issue-certificate.md).
- Learn how to generate [search-guard configuration](/docs/guides/elasticsearch/plugins/search-guard/configuration.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
