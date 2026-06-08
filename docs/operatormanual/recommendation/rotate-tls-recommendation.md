---
title: Rotate TLS Recommendation
menu:
  docs_{{ .version }}:
    identifier: rotate-tls-recommendation
    name: Rotate TLS
    parent: recommendation
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate TLS Recommendation

TLS certificate rotation in databases is essential for maintaining security, ensuring compliance, and preventing service disruptions. Regular rotation mitigates risks like certificate expiry and key compromise, adapts to evolving cryptographic standards, and maintains trust relationships with Certificate Authorities. It also enhances operational resilience by testing renewal processes and ensures smooth auditing and monitoring. To minimize risks and streamline the process, KubeDB provides ReconfigureTLS OpsRequest support. KubeDB Ops-manager generates Recommendation to rotate TLS certificates via this OpsRequest when their expiry is near.

> Note: We provide support for `Recommendation` across most database systems. Below is an example demonstrating how recommendations are applied for the `Elasticsearch` database.

`Recommendation` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative recommendation for KubeDB managed databases like [Elasticsearch](https://www.elastic.co/products/elasticsearch) and [OpenSearch](https://opensearch.org/) in a Kubernetes native way. KubeDB generates Elasticsearch/Opensearch Rotate TLS recommendation regarding if:

- At least one of its certificate’s lifespan is more than one month and less than one month remaining till expiry

- At least one of its certificates has one-third of its lifespan remaining till expiry.


## Prerequisite
- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

#### Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in MongoDB. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mongo/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
$ kubectl create secret tls mongo-ca \
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
## How TLS Rotation Works

Let's go through a demo to see `RotateTLS` recommendations being generated. First, get the available Elasticsearch versions provided by KubeDB.

```bash
$ kubectl get elasticsearchversions | grep xpack
xpack-6.8.23        6.8.23    ElasticStack   ghcr.io/appscode-images/elastic:6.8.23                                  12d
xpack-7.17.15       7.17.15   ElasticStack   ghcr.io/appscode-images/elastic:7.17.15                                 12d
xpack-7.17.28       7.17.28   ElasticStack   ghcr.io/appscode-images/elastic:7.17.28                                 12d
xpack-8.17.10       8.17.10   ElasticStack   ghcr.io/appscode-images/elastic:8.17.10                                 12d
xpack-8.17.6        8.17.6    ElasticStack   ghcr.io/appscode-images/elastic:8.17.6                                  12d
xpack-8.18.2        8.18.2    ElasticStack   ghcr.io/appscode-images/elastic:8.18.2                                  12d
xpack-8.18.8        8.18.8    ElasticStack   ghcr.io/appscode-images/elastic:8.18.8                                  12d
xpack-8.19.9        8.19.9    ElasticStack   ghcr.io/appscode-images/elastic:8.19.9                                  12d
xpack-8.2.3         8.2.3     ElasticStack   ghcr.io/appscode-images/elastic:8.2.3                                   12d
xpack-8.5.3         8.5.3     ElasticStack   ghcr.io/appscode-images/elastic:8.5.3                                   12d
xpack-9.0.2         9.0.2     ElasticStack   ghcr.io/appscode-images/elastic:9.0.2                                   12d
xpack-9.0.8         9.0.8     ElasticStack   ghcr.io/appscode-images/elastic:9.0.8                                   12d
xpack-9.1.4         9.1.4     ElasticStack   ghcr.io/appscode-images/elastic:9.1.4                                   12d
xpack-9.1.9         9.1.9     ElasticStack   ghcr.io/appscode-images/elastic:9.1.9                                   12d
xpack-9.2.3         9.2.3     ElasticStack   ghcr.io/appscode-images/elastic:9.2.3                                   12d
```

Let's deploy an Elasticsearch cluster with version `xpack-9.1.9`. We are going to create a cluster topology with 2 master nodes, 3 data nodes and 2 ingest node. We also have to provide an available storageclass for each of the node types. Make sure to have an issuer/clusterIssuer to refer in the manifest. Though KubeDB managed elasticsearch supports TLS in both cert-manager provisioned and Operator provisioned ways, rotate tls only works when certificates are provisioned via cert-manager.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-tls
  namespace: demo
spec:
  deletionPolicy: WipeOut
  version: xpack-9.1.9 
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  enableSSL: true
  tls:
    certificates:
      - alias: client
        duration: 1h20m
      - alias: http
        duration: 2h10m
```

Wait for a while till elasticsearch cluster gets into `Ready` state. Required time depends on image pulling and node's physical specifications.

```bash
$ kubectl get elasticsearch,pods -n demo
NAME                              VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-tls   xpack-9.1.9   Ready    4m39s

NAME           READY   STATUS    RESTARTS   AGE
pod/es-tls-0   1/1     Running   0          4m34s
pod/es-tls-1   1/1     Running   0          4m28s
```

Once elastic instance is `Ready`, a `Recommendation` instance will be automatically generated by KubeDB `Ops-Manager` controller. Might take a few minutes to trigger an event for the database creation in the controller.

```bash
$ kubectl get es,pods,elasticsearchopsrequest -n demo
NAME                              VERSION       STATUS   AGE
elasticsearch.kubedb.com/es-tls   xpack-9.1.9   Ready    118m

NAME           READY   STATUS    RESTARTS   AGE
pod/es-tls-0   1/1     Running   0          8m20s
pod/es-tls-1   1/1     Running   0          7m42s

NAME                                                                       TYPE             STATUS       AGE
elasticsearchopsrequest.ops.kubedb.com/es-tls-1780909429-rotate-tls-auto   ReconfigureTLS   Successful   63m
elasticsearchopsrequest.ops.kubedb.com/es-tls-1780912735-rotate-tls-auto   ReconfigureTLS   Successful   8m50s


$ kubectl get recommendation -n demo
NAME                                             STATUS      OUTDATED   AGE
es-tls-x-elasticsearch-x-rotate-tls-hn5hf8       Succeeded   false      22m
es-tls-x-elasticsearch-x-rotate-tls-ys1s5q       Succeeded   true       77m
es-tls-x-elasticsearch-x-update-version-jnyey9   Pending     false      116m

```
The `Recommendation` custom resource will be named as:

```
<DB-name>-x-<DB-type>-x-<Recommendation-type>-<random-suffix>
```

---

# RotateAuth Recommendation Example (MongoDB)

Let's check the complete Recommendation custom resource manifest:

```bash
$ kubectl get recommendation -n demo mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 -o yaml
```

### Understanding the Recommendation

In the `spec.operation` field, the recommendation suggests rotating the **authentication secret** of the `mg-rarecommendation` cluster. The recommended operation is a `MongoDBOpsRequest` of type `RotateAuth`.

This indicates that KubeDB has detected that the authentication secret is nearing expiry and needs to be rotated.

---

### Automatic Approval Behavior

RotateAuth recommendations typically do not require explicit approval. This means KubeDB Supervisor automatically approves and executes the operation when the configured `deadline` is reached.

In this example:

* Recommendation created at: `2026-06-08T05:49:38Z`
* Deadline set at: `2026-06-08T05:59:34Z`
* Secret expiry at: `2026-06-08T06:09:34Z`

Since the deadline occurs before the secret expiry, Supervisor ensures timely rotation.

Once the deadline is reached, Supervisor:

* Sets `approvalStatus: Approved`
* Sets `approvedWindow: Immediate`
* Creates and executes the corresponding `MongoDBOpsRequest`

---

### Checking Recommendation Status

```bash
$ kubectl get recommendation mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 \
  -n demo -o json | jq '.status'
```

---

### Understanding Status Conditions

From the `status.conditions`:

* `SuccessfullyCreatedOperation` → The `MongoDBOpsRequest` was created successfully
* `SuccessfullyExecutedOperation` → The authentication rotation completed successfully

The executed operation can be referenced here:

```yaml
status:
  createdOperationRef:
    name: mg-rarecommendation-1780898378-rotate-auth-auto
```

---

### Resulting Operation

```bash
$ kubectl get mongodbopsrequest -n demo mg-rarecommendation-1780898378-rotate-auth-auto
```

This operation rotates the authentication secret with minimal or no downtime, ensuring:

* No expired credentials
* Improved security posture
* Fully automated secret lifecycle management

---

# RotateTLS Recommendation Example (Elasticsearch)

Let's check the complete Recommendation custom resource manifest:

```bash
$ kubectl get recommendation -n demo es-tls-x-elasticsearch-x-rotate-tls-hn5hf8 -o yaml
```

### Understanding the Recommendation

In the `spec.operation` field, the recommendation suggests rotating the **TLS certificates** of the `es-tls` cluster. The recommended operation is an `ElasticsearchOpsRequest` of type `ReconfigureTLS`, where `tls.rotateCertificates: true` is set.

This indicates that KubeDB has detected that TLS certificates are nearing expiry.

---

### Automatic Approval Behavior

RotateTLS recommendations typically do not require explicit approval. This means KubeDB Supervisor automatically approves and executes the operation when the configured deadline is reached.

In this example:

* Recommendation created at: `2026-06-08T09:43:55Z`
* Deadline set at: `2026-06-08T09:58:51Z`
* Certificate expiry at: `2026-06-08T10:03:51Z`

Since the deadline occurs shortly before expiry, Supervisor ensures rotation happens in time.

Once the deadline is reached, Supervisor:

* Sets `approvalStatus: Approved`
* Sets `approvedWindow: Immediate`
* Creates and executes the corresponding `ElasticsearchOpsRequest`

---

### Checking Recommendation Status

```bash
$ kubectl get recommendation es-tls-x-elasticsearch-x-rotate-tls-hn5hf8 \
  -n demo -o json | jq '.status'
```

---

### Understanding Status Conditions

From the `status.conditions`:

* `SuccessfullyCreatedOperation` → The `ElasticsearchOpsRequest` was created successfully
* `SuccessfullyExecutedOperation` → The TLS rotation completed successfully

The executed operation can be referenced here:

```yaml
status:
  createdOperationRef:
    name: es-tls-1780912735-rotate-tls-auto
```
The `Recommendation` custom resource will be named as:

```
<DB-name>-x-<DB-type>-x-<Recommendation-type>-<random-suffix>
```

---

# RotateAuth Recommendation Example (MongoDB)

Let's check the complete Recommendation custom resource manifest:

```bash
$ kubectl get recommendation -n demo mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 -o yaml
```

### Understanding the Recommendation

In the `spec.operation` field, the recommendation suggests rotating the **authentication secret** of the `mg-rarecommendation` cluster. The recommended operation is a `MongoDBOpsRequest` of type `RotateAuth`.

This indicates that KubeDB has detected that the authentication secret is nearing expiry and needs to be rotated.

---

### Automatic Approval Behavior

RotateAuth recommendations typically do not require explicit approval. This means KubeDB Supervisor automatically approves and executes the operation when the configured `deadline` is reached.

In this example:

* Recommendation created at: `2026-06-08T05:49:38Z`
* Deadline set at: `2026-06-08T05:59:34Z`
* Secret expiry at: `2026-06-08T06:09:34Z`

Since the deadline occurs before the secret expiry, Supervisor ensures timely rotation.

Once the deadline is reached, Supervisor:

* Sets `approvalStatus: Approved`
* Sets `approvedWindow: Immediate`
* Creates and executes the corresponding `MongoDBOpsRequest`

---

### Checking Recommendation Status

```bash
$ kubectl get recommendation mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 \
  -n demo -o json | jq '.status'
```

---

### Understanding Status Conditions

From the `status.conditions`:

* `SuccessfullyCreatedOperation` → The `MongoDBOpsRequest` was created successfully
* `SuccessfullyExecutedOperation` → The authentication rotation completed successfully

The executed operation can be referenced here:

```yaml
status:
  createdOperationRef:
    name: mg-rarecommendation-1780898378-rotate-auth-auto
```

---

### Resulting Operation

```bash
$ kubectl get mongodbopsrequest -n demo mg-rarecommendation-1780898378-rotate-auth-auto
```

This operation rotates the authentication secret with minimal or no downtime, ensuring:

* No expired credentials
* Improved security posture
* Fully automated secret lifecycle management

---

# RotateTLS Recommendation Example (Elasticsearch)

Let's check the complete Recommendation custom resource manifest:

```bash
$ kubectl get recommendation -n demo es-tls-x-elasticsearch-x-rotate-tls-hn5hf8 -o yaml
```

### Understanding the Recommendation

In the `spec.operation` field, the recommendation suggests rotating the **TLS certificates** of the `es-tls` cluster. The recommended operation is an `ElasticsearchOpsRequest` of type `ReconfigureTLS`, where `tls.rotateCertificates: true` is set.

This indicates that KubeDB has detected that TLS certificates are nearing expiry.

---

### Automatic Approval Behavior

RotateTLS recommendations typically do not require explicit approval. This means KubeDB Supervisor automatically approves and executes the operation when the configured deadline is reached.

In this example:

* Recommendation created at: `2026-06-08T09:43:55Z`
* Deadline set at: `2026-06-08T09:58:51Z`
* Certificate expiry at: `2026-06-08T10:03:51Z`

Since the deadline occurs shortly before expiry, Supervisor ensures rotation happens in time.

Once the deadline is reached, Supervisor:

* Sets `approvalStatus: Approved`
* Sets `approvedWindow: Immediate`
* Creates and executes the corresponding `ElasticsearchOpsRequest`

Let's check `opsrequest` status
```bash
$ kubectl get esops -n demo es-tls-1780909429-rotate-tls-auto 
NAME                                TYPE             STATUS       AGE
es-tls-1780909429-rotate-tls-auto   ReconfigureTLS   Successful   151m
```


You may not want to trigger recommended operations manually. Rather, trigger them autonomously in a preferred schedule when infrastructure is idle or traffic rate is at the lowest. For this purpose, You can create a `MaintenanceWindow` custom resource where you can set your desired schedule/period for triggering these recommended operations automatically. See [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for detailed documentation. Here's a sample one:


You can now create a `ApprovalPolicy` custom resource to refer this `MaintenanceWindow` for particular DB type. See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for detailed documentation. Following is a sample `ApprovalPolicy` for any `MongoDB` custom resource deployed in `demo` namespace. This `ApprovalPolicy` custom resource is referring to the `elastic-maintenance` MaintenanceWindow created in the same namespace. You can also create `ClusterMaintenanceWindow` instead (see [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md)) which is effective for cluster-wide operations and refer it here. The following ApprovalPolicy will trigger recommended operations when referred maintenance window timeframe is reached.

Lastly, If you want to reject a recommendation, you can just set `ApprovalStatus` to `Rejected` in the recommendation status section. Here's how you can do it using kubectl cli.

```bash
$ kubectl patch recommendation es-tls-x-elasticsearch-x-rotate-tls-hn5hf8 \
                                    -n demo \
                                    --type merge \
                                    --subresource='status' \
                                    -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/es-tls-x-elasticsearch-x-rotate-tls-hn5hf8 patched
```

For complete reference on all Recommendation fields, phases, and status conditions, see [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md).
