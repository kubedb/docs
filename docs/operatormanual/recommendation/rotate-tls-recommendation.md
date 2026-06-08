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
$ kubectl get mongodb,mongodbopsrequest,pods -n demo
NAME                          VERSION   STATUS   AGE
mongodb.kubedb.com/mg-alone   8.0.10    Ready    50m

NAME                                                                       TYPE            STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-alone-1780896035-update-version-auto   UpdateVersion   Successful   39m
mongodbopsrequest.ops.kubedb.com/mg-alone-1780898378-rotate-auth-auto      RotateAuth      Successful   38s

NAME             READY   STATUS    RESTARTS   AGE
pod/es-tls-0     1/1     Running   0          64m
pod/es-tls-1     1/1     Running   0          64m

$ kubectl get recommendation -n demo 
NAME                                             STATUS    OUTDATED   AGE
es-tls-x-elasticsearch-x-update-version-35l3km   Pending   false      2m43s
```

The `Recommendation` custom resource will be named as `<DB-name>-x-<DB type>-x-<Recommendation type>-<random hash>`. Initially, the KubeDB `Supervisor` controller will mark the `Status` of this object to `Pending`. Let's check the complete Recommendation custom resource manifest:

```yaml
$ kubectl get recommendation -n demo es-tls-x-elasticsearch-x-update-version-35l3km -oyaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: xpack-9.1.9
  creationTimestamp: "2026-06-08T04:47:53Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: es-tls
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
    kubedb.com/version-update-recommendation-type: major-minor
  name: es-tls-x-elasticsearch-x-update-version-35l3km
  namespace: demo
  resourceVersion: "88176"
  uid: 1162309e-ba19-47dc-bbb6-45cd2d898a45
spec:
  backoffLimit: 10
  description: Latest Major/Minor version is available. Recommending version Update
    from xpack-9.1.9 to xpack-9.2.3.
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: ElasticsearchOpsRequest
    metadata:
      name: update-version
      namespace: demo
    spec:
      databaseRef:
        name: es-tls
      type: UpdateVersion
      updateVersion:
        targetVersion: xpack-9.2.3
    status: {}
  recommender:
    name: kubedb-ops-manager
  requireExplicitApproval: true
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase ==
      'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-tls
  vulnerabilityReport:
    message: no matches for kind "ImageScanReport" in version "scanner.appscode.com/v1alpha1"
    status: Failure
status:
  approvalStatus: Pending
  failedAttempt: 0
  outdated: false
  parallelism: Namespace
  phase: Pending
  reason: WaitingForApproval
```


In the `spec.operation` field the recommendation says current version `xpack-9.1.9` should be upgraded to latest upgradable version `xpack-9.2.3`. You can also find the recommended operation which is a `ElasticsearchOpsRequest` of `UpdateVersion` type in this case. Additionally `spec.operation.updateVersion.targetVersion` indicate the update version `xpack-9.2.3`.


Now, the `status` part of this recommendation says can be approved and operation can be executed immediately by setting `ApprovalStatus` to `Approved` and Setting `approvedWindow` to `Immediate`. You can approve this easily through Appscode UI or edit it manually. Also, You can use kubectl CLI for this - 

```shell
$ kubectl patch recommendation -n demo es-tls-x-elasticsearch-x-update-version-35l3km \
                                  -n demo \
                                                                                               --type merge \
                                                                                               --subresource='status' \
                                                                                               -p '{"status":{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"}}}'
recommendation.supervisor.appscode.com/es-tls-x-elasticsearch-x-update-version-35l3km patched
```

Now, check the status part again. You will find a condition have appeared which says `OpsRequest is successfully created`. 

```bash
$ kubectl get recommendation -n demo es-tls-x-elasticsearch-x-update-version-35l3km -o jsonpath='{.status}'
{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"},"conditions":[{"lastTransitionTime":"2026-06-08T04:54:59Z","message":"OpsRequest is successfully created","reason":"SuccessfullyCreatedOperation","status":"True","type":"SuccessfullyCreatedOperation"}],"createdOperationRef":{"name":"es-tls-1780894499-update-version-auto"},"failedAttempt":0,"outdated":false,"parallelism":"Namespace","phase":"InProgress","reason":"StartedExecutingOperation"}
```

You will find an `ElasticsearchOpsRequest` custom resource have been created and, it is updating the `es-vurecommendation` cluster version to `xpack-9.2.3` with negligible downtime. Let's wait for it to reach `Successful` status.

```bash
$ kubectl get elasticsearchopsrequest -n demo 
NAME                                                 TYPE            STATUS       AGE
es-vurecommendation-1780635038-update-version-auto   UpdateVersion   Failed       9m
es-vurecommendation-1780635179-update-version-auto   UpdateVersion   Failed       6m39s
es-vurecommendation-1780635299-update-version-auto   UpdateVersion   Failed       4m39s
es-vurecommendation-1780635419-update-version-auto   UpdateVersion   Successful   2m39s
```

Let's recheck the recommendation for one last time. We should find that `.status.phase` has been marked as `Succeeded`. 

```bash
$ kubectl get recommendation -n demo es-vurecommendation-x-elasticsearch-x-update-version-22erfx
NAME                                                          STATUS      OUTDATED   AGE
es-vurecommendation-x-elasticsearch-x-update-version-22erfx   Succeeded   false      21m
```

Finally, You can check `es-vurecommendation` cluster version now, which should be upgraded to version `xpack-9.2.3`.

```bash
$ kubectl get es es-vurecommendation -n demo
NAME                  VERSION       STATUS   AGE
es-vurecommendation   xpack-9.2.3   Ready    25m
```

You may not want to do trigger recommended operations manually. Rather, trigger them autonomously in a preferred schedule when infrastructure is idle or traffic rate is at the lowest. For this purpose, You can create a `MaintenanceWindow` custom resource where you can set your desired schedule/period for triggering these recommended operations automatically. See [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for detailed documentation. 


You can now create a `ApprovalPolicy` custom resource to refer this `MaintenanceWindow` for particular DB type. See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for detailed documentation. Following is a sample `ApprovalPolicy` for any `Elasticsearch` custom resource deployed in `es` namespace. This `ApprovalPolicy` custom resource is referring to the `elastic-maintenance` MaintenanceWindow created in the same namespace. You can also create `ClusterMaintenanceWindow` instead (see [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md)) which is effective for cluster-wide operations and refer it here. The following ApprovalPolicy will trigger recommended operations when referred maintenance window timeframe is reached. 



Lastly, If you want to reject a recommendation, you can just set `ApprovalStatus` to `Rejected` in the recommendation status section. Here's how you can do it using kubectl cli.

```bash
$ kubectl patch Recommendation es-vurecommendation-x-elasticsearch-x-update-version-22erfx  \
                  -n demo \
                  --type merge \
                  --subresource='status' \
                  -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/es-vurecommendation-x-elasticsearch-x-update-version-22erfx  patched
```

For complete reference on all Recommendation fields, phases, and status conditions, see [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md).
