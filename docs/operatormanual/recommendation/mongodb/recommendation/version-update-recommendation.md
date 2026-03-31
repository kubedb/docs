---
title: MongoDB Version Update Recommendation
menu:
  docs_{{ .version }}:
    identifier: mg-version-update-recommendation
    name: Version Update Recommendation
    parent: mg-recommendation-mongodb
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Version Update Recommendation

Database versions often need to be updated due to several reasons. Older database versions may have vulnerabilities that hackers can exploit. New versions often include optimizations for query execution, indexing, and storage mechanisms. Modern databases frequently introduce new features, such as better data types, improved indexing methods, or advanced analytics capabilities. Database vendors release patches and updates to address these issues and introduce new features.

`Recommendation` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative recommendation for KubeDB managed databases like [MongoDB](https://www.mongodb.com/) in a Kubernetes native way. KubeDB generates MongoDB Version Update recommendation regarding three particular cases. 

1. There's been an update in the current version image
2. There's a new major/minor version available
3. There's a version available with patch fix

Let's go through a demo to see version update recommendations being generated. First, get the available MongoDB versions provided by KubeDB.

```bash
$ kubectl get mongodbversion
NAME             VERSION   DISTRIBUTION   DB_IMAGE                                          DEPRECATED   AGE
4.4.26           4.4.26    Official       ghcr.io/appscode-images/mongo:4.4.26                           12d
5.0.31           5.0.31    Official       ghcr.io/appscode-images/mongo:5.0.31                           12d
6.0.24           6.0.24    Official       ghcr.io/appscode-images/mongo:6.0.24                           12d
7.0.21           7.0.21    Official       ghcr.io/appscode-images/mongo:7.0.21                           12d
7.0.28           7.0.28    Official       ghcr.io/appscode-images/mongo:7.0.28                           12d
8.0.10           8.0.10    Official       ghcr.io/appscode-images/mongo:8.0.10                           12d
8.0.10           8.0.10    Official       ghcr.io/appscode-images/mongo:8.0.10                           12d
percona-4.4.26   4.4.26    Percona        docker.io/percona/percona-server-mongodb:4.4.26                12d
percona-5.0.29   5.0.29    Percona        docker.io/percona/percona-server-mongodb:5.0.29                12d
percona-6.0.24   6.0.24    Percona        docker.io/percona/percona-server-mongodb:6.0.24                12d
percona-7.0.18   7.0.18    Percona        docker.io/percona/percona-server-mongodb:7.0.18                12d
percona-7.0.28   7.0.28    Percona        docker.io/percona/percona-server-mongodb:7.0.28                12d
percona-8.0.10   8.0.10    Percona        docker.io/percona/percona-server-mongodb:8.0.10                12d
percona-8.0.8    8.0.8     Percona        docker.io/percona/percona-server-mongodb:8.0.8                 12d
```

Let's deploy an MongoDB cluster with version `8.0.10`.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongo
  namespace: mg
spec:
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
        - name: mongodb
          resources:
            limits:
              cpu: 700m
              memory: 1Gi
            requests:
              cpu: 700m
              memory: 1Gi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsGroup: 0
            runAsNonRoot: true
            runAsUser: 999
            seccompProfile:
              type: RuntimeDefault
        - name: replication-mode-detector
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsGroup: 0
            runAsNonRoot: true
            runAsUser: 999
            seccompProfile:
              type: RuntimeDefault
      initContainers:
        - name: copy-config
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsGroup: 0
            runAsNonRoot: true
            runAsUser: 999
            seccompProfile:
              type: RuntimeDefault
      nodeSelector:
        kubernetes.io/os: linux
      securityContext:
        fsGroup: 999
  replicaSet:
    name: rs0
  replicas: 2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 4Gi
    storageClassName: local-path
  storageType: Durable
  version: 8.0.10
```

Wait for a while till mongodb cluster gets into `Ready` state. Required time depends on image pulling and node's physical specifications.

```bash
$ kubectl get mg mongo -n mg -w
NAME      VERSION        STATUS         AGE
mongo      8.0.10         Provisioning   98s
mongo      8.0.10         Provisioning   5m43s
mongo      8.0.10         Provisioning   8m7s
.
.
.
mongo      8.0.10         Ready          10m
mongo      8.0.10         Ready          10m
```

Once mongo instance is `Ready`, a `Recommendation` instance will be automatically generated by KubeDB `Ops-Manager` controller. Might take a few minutes to trigger an event for the database creation in the controller.

```bash
$ kubectl get recommendation -n mg
NAME                                              STATUS    OUTDATED   AGE
mongo-x-mongodb-x-update-version-uax2ot           Pending   false      10m
```

The `Recommendation` custom resource will be named as `<DB-name>-x-<DB type>-x-<Recommendation typer>-<random hash>`. Initially, the KubeDB `Supervisor` controller will mark the `Status` of this object to `Pending`. Let's check the complete Recommendation custom resource manifest:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: 8.0.10
  creationTimestamp: "2025-02-25T08:32:58Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: mongo
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
    kubedb.com/version-update-recommendation-type: major-minor
  name: mongo-x-mongodb-x-update-version-uax2ot
  namespace: mg
  resourceVersion: "76768"
  uid: 17722dda-992a-4755-a480-a6ef1f7149a7
spec:
  backoffLimit: 5
  description: Latest Major/Minor version is available. Recommending version Update
    from 8.0.10 to 8.0.17.
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: MongoDBOpsRequest
    metadata:
      name: update-version
      namespace: mg
    spec:
      databaseRef:
        name: mongo
      type: UpdateVersion
      updateVersion:
        targetVersion: 8.0.17
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
    kind: MongoDB
    name: mongo
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

In the generated Recommendation you will find a description, targeted db object, recommended operation or Ops-Request manifest, current status of the recommendation etc. Let's just focus on the recommendation description first.

```shell
$ kubectl get recommendation -n mg mongo-x-mongodb-x-update-version-uax2ot -o jsonpath='{.spec.description}'
Latest Major/Minor version is available. Recommending version Update from 8.0.10 to 8.0.17.
```

The recommendation says current version `8.0.10` should be upgraded to latest upgradable version `8.0.17`. You can also find the recommended operation which is a `MongoDBOpsRequest` of `UpdateVersion` type in this case.

```shell
$ kubectl get recommendation -n mg mongo-x-mongodb-x-update-version-uax2ot -o jsonpath='{.spec.operation}' | yq -y
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: update-version
  namespace: mg
spec:
  databaseRef:
    name: mongo
  type: UpdateVersion
  updateVersion:
    targetVersion: 8.0.17
status: {}
```

Note: For the above command to work you need to have YQ v3 installed.

Let's check the status part of this recommendation.

```bash
$ kubectl get recommendation -n mg mongo-x-mongodb-x-update-version-uax2ot -o jsonpath='{.status}' | yq -y
approvalStatus: Pending
failedAttempt: 0
outdated: false
parallelism: Namespace
phase: Pending
reason: WaitingForApproval
```

Now, This recommendation can be approved and operation can be executed immediately by setting `ApprovalStatus` to `Approved` and Setting `approvedWindow` to `Immediate`. You can approve this easily through Appscode UI or edit it manually. Also, You can use kubectl CLI for this - 

```bash
$ kubectl patch Recommendation mongo-x-mongodb-x-update-version-uax2ot \
                  -n mg \
                  --type merge \
                  --subresource='status' \
                  -p '{"status":{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"}}}'
recommendation.supervisor.appscode.com/mongo-x-mongodb-x-update-version-uax2ot patched
```

Now, check the status part again. You will find a condition have appeared which says `OpsRequest is successfully created`. 

```bash
$ kubectl get recommendation -n mg mongo-x-mongodb-x-update-version-uax2ot -o jsonpath='{.status}' | yq -y
approvalStatus: Approved
approvedWindow:
  window: Immediate
conditions:
  - lastTransitionTime: '2025-02-25T09:07:19Z'
    message: OpsRequest is successfully created
    reason: SuccessfullyCreatedOperation
    status: 'True'
    type: SuccessfullyCreatedOperation
createdOperationRef:
  name: mongo-1740474439-update-version-auto
failedAttempt: 0
outdated: false
parallelism: Namespace
phase: InProgress
reason: StartedExecutingOperation

```

You will find an `MongoDBOpsRequest` custom resource have been created and, it is updating the `mongo` cluster version to `8.0.17` with negligible downtime. Let's wait for it to reach `Successful` status.

```bash
$ kubectl get mongodbopsrequest -n mg mongo-1740474439-update-version-auto -w
NAME                                   TYPE            STATUS        AGE
mongo-1740474439-update-version-auto   UpdateVersion   Progressing   70s
mongo-1740474439-update-version-auto   UpdateVersion   Progressing   99s
.
.
mongo-1740474439-update-version-auto   UpdateVersion   Successful    2m15s
```

Let's recheck the recommendation for one last time. We should find that `.status.phase` has been marked as `Succeeded`. 

```bash
$ kubectl get recommendation -n mg mongo-x-mongodb-x-update-version-uax2ot
NAME                                              STATUS      OUTDATED   AGE
mongo-x-mongodb-x-update-version-uax2ot           Succeeded   false      78m
```

Finally, You can check `mongo` cluster version now, which should be upgraded to version `8.0.17`.

```bash
$ kubectl get mg mongo -n mg
NAME    VERSION   STATUS   AGE
mongo   8.0.17     Ready    40m
```

You may not want to do trigger recommended operations manually. Rather, trigger them autonomously in a preferred schedule when infrastructure is idle or traffic rate is at the lowest. For this purpose, You can create a `MaintenanceWindow` custom resource where you can set your desired schedule/period for triggering these recommended operations automatically. Here's a sample one:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: mongo-maintenance
  namespace: mg
spec:
  timezone: Asia/Dhaka
  days:
    Wednesday:
      - start: 5:40AM
        end: 7:00PM
  dates:
    - start: 2025-01-25T00:00:18Z
      end: 2025-01-25T23:41:18Z
```

You can now create a `ApprovalPolicy` custom resource to refer this `MaintenanceWindow` for particular DB type. Following is a sample `ApprovalPolicy` for any `MongoDB` custom resource deployed in `mg` namespace. This `ApprovalPolicy` custom resource is referring to the `mongo-maintenance` MaintenanceWindow created in the same namespace. You can also create `ClusterMaintenanceWindow` instead which is effective for cluster-wide operations and refer it here. The following ApprovalPolicy will trigger recommended operations when referred maintenance window timeframe is reached. 

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: mg-policy
  namespace: mg
maintenanceWindowRef:
  name: mongo-maintenance
targets:
  - group: kubedb.com
    kind: MongoDB
    operations:
      - group: ops.kubedb.com
        kind: MongoDBOpsRequest
```

Lastly, If you want to reject a recommendation, you can just set `ApprovalStatus` to `Rejected` in the recommendation status section. Here's how you can do it using kubectl cli.

```bash
$ kubectl patch Recommendation mongo-x-mongodb-x-update-version-uax2ot \
                  -n mg \
                  --type merge \
                  --subresource='status' \
                  -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/mongo-x-mongodb-x-update-version-uax2ot patched
```

## Next Steps

- Learn about [backup & restore](/docs/guides/mongodb/backup/stash/overview/index.md) MongoDB database using Stash.
- Learn how to configure [MongoDB Cluster](/docs/guides/mongodb/clustering/replicaset.md).
- Monitor your MongoDB database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
