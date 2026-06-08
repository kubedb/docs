---
title: Rotate Auth Recommendation
menu:
  docs_{{ .version }}:
    identifier: rotate-auth-recommendation
    name: Rotate Auth
    parent: recommendation
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Authentication Rotate Recommendation

Rotating authentication secrets in database management is vital to mitigate security risks, such as credential leakage or unauthorized access, and to comply with regulatory requirements. Regular rotation limits the exposure of compromised credentials, reduces the risk of insider threats, and enforces updated security policies like stronger passwords or algorithms. It also ensures operational resilience by testing the rotation process and revoking stale or unused credentials. KubeDB provides `RotateAuth` which reduces manual errors, and strengthens database security with minimal effort. KubeDB Ops-manager generates Recommendation for rotating authentication secrets via this OpsRequest.

> Note: We provide support for `Recommendation` across most database systems. Below is an example demonstrating how recommendations are applied for the [MongoDB](/docs/guides/mongodb) database.

`Recommendation` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative recommendation for KubeDB managed databases  in a Kubernetes native way. The recommendation will only be created if `.spec.authSecret.rotateAfter` is set. KubeDB generates MongoDB/Opensearch Rotate Auth recommendation regarding two particular cases.

1. AuthSecret lifespan is more than one month and, less than one month remaining till expiry
2. AuthSecret lifespan is less than one month and, less than one third of lifespan remaining till expiry

Let's go through a demo to see `RotateAuth` recommendations being generated. 
```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-rarecommendation
  namespace: demo
spec:
  version: "8.0.10"
  authSecret:
    rotateAfter: 1h
  storage:
    resources:
      requests:
        storage: 500Mi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

Wait for a while till MongoDB cluster gets into `Ready` state. Required time depends on image pulling and node's physical specifications.

```bash
$ kubectl get mongodb,pods -n demo
NAME                          VERSION   STATUS   AGE
mongodb.kubedb.com/mg-rarecommendation   8.0.10    Ready    50m

NAME             READY   STATUS    RESTARTS   AGE
pod/mg-rarecommendation-0   1/1     Running   0          19s

```

Since, `.spec.authSecret.rotateAfter` is set as `1h`, it is expected that the recommendation engine will generate a rotate-auth recommendation at least after 40 minutes (two-third of lifespan) of the authsecret creation. Once generated you will get a similar recommendation as follows.

```bash
$ kubectl get mongodb,mongodbopsrequest,pods -n demo
NAME                          VERSION   STATUS   AGE
mongodb.kubedb.com/mg-rarecommendation   8.0.10    Ready    50m

NAME                                                                       TYPE            STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mg-rarecommendation-1780896035-update-version-auto   UpdateVersion   Successful   39m
mongodbopsrequest.ops.kubedb.com/mg-rarecommendation-1780898378-rotate-auth-auto      RotateAuth      Successful   38s

NAME             READY   STATUS    RESTARTS   AGE
pod/mg-rarecommendation-0   1/1     Running   0          19s

$ kubectl get recommendation -n demo 
NAME                                              STATUS      OUTDATED   AGE
mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3           Succeeded   false      11m
mg-rarecommendation-x-mongodb-x-update-same-version-6tbbbc   Pending     false      32m
mg-rarecommendation-x-mongodb-x-update-version-c2pemb        Succeeded   false      49m
```

The `Recommendation` custom resource will be named as `<DB-name>-x-<DB type>-x-<Recommendation type>-<random hash>`.  Let's check the complete Recommendation custom resource manifest:

```yaml
$ kubectl get recommendation -n demo mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 -oyaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  creationTimestamp: "2026-06-08T05:49:38Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: mg-rarecommendation
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: rotate-auth
  name: mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3
  namespace: demo
  resourceVersion: "90645"
  uid: e31ea7fa-ac14-47d3-b30d-366b685a148d
spec:
  backoffLimit: 10
  deadline: "2026-06-08T05:59:34Z"
  description: Recommending AuthSecret rotation,mg-rarecommendation-auth AuthSecret needs to
    be rotated before 2026-06-08 06:09:34 +0000 UTC
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: MongoDBOpsRequest
    metadata:
      name: rotate-auth
      namespace: demo
    spec:
      databaseRef:
        name: mg-rarecommendation
      type: RotateAuth
    status: {}
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase ==
      'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: MongoDB
    name: mg-rarecommendation
status:
  approvalStatus: Approved
  approvedWindow:
    window: Immediate
  conditions:
  - lastTransitionTime: "2026-06-08T05:59:38Z"
    message: OpsRequest is successfully created
    reason: SuccessfullyCreatedOperation
    status: "True"
    type: SuccessfullyCreatedOperation
  - lastTransitionTime: "2026-06-08T06:00:38Z"
    message: OpsRequest is successfully executed
    reason: SuccessfullyExecutedOperation
    status: "True"
    type: SuccessfullyExecutedOperation
  createdOperationRef:
    name: mg-rarecommendation-1780898378-rotate-auth-auto
  failedAttempt: 0
  observedGeneration: 1
  outdated: false
  parallelism: Namespace
  phase: Succeeded
  reason: SuccessfullyExecutedOperation
```

In the `spec.operation` field, the recommendation suggests rotating the authentication secret of `mg-rarecommendation`. The recommended operation is an `ElasticsearchOpsRequest` of `RotateAuth` type.

Notice that  rotate-auth recommendations do not set `requireExplicitApproval`. This means KubeDB Supervisor automatically approves and executes the operation when the `deadline` is reached, ensuring authentication secrets are always rotated on time without requiring manual intervention.

In this case, the `deadline` (`"2026-06-05T09:54:05Z"`) had already passed by the time the recommendation was created (`"2026-06-05T10:23:51Z"`), so Supervisor immediately set `approvalStatus: Approved` with `approvedWindow: Immediate` and triggered the OpsRequest automatically.


Now, check the status part again. You will find a condition have appeared which says `OpsRequest is successfully created`.

```bash
$ kubectl get recommendation mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 \
                                    -n demo -o json | jq '.status'
{
  "approvalStatus": "Approved",
  "approvedWindow": {
    "window": "Immediate"
  },
  "conditions": [
    {
      "lastTransitionTime": "2026-06-08T05:59:38Z",
      "message": "OpsRequest is successfully created",
      "reason": "SuccessfullyCreatedOperation",
      "status": "True",
      "type": "SuccessfullyCreatedOperation"
    },
    {
      "lastTransitionTime": "2026-06-08T06:00:38Z",
      "message": "OpsRequest is successfully executed",
      "reason": "SuccessfullyExecutedOperation",
      "status": "True",
      "type": "SuccessfullyExecutedOperation"
    }
  ],
  "createdOperationRef": {
    "name": "mg-rarecommendation-1780898378-rotate-auth-auto"
  },
  "failedAttempt": 0,
  "observedGeneration": 1,
  "outdated": false,
  "parallelism": "Namespace",
  "phase": "Succeeded",
  "reason": "SuccessfullyExecutedOperation"
}
```

You will find an `ElasticsearchOpsRequest` custom resource has been created and it is rotating the auth secret of `mg-rarecommendation` cluster with negligible downtime. Let's wait for it to reach `Successful` status.

```bash
$ kubectl get mongodbopsrequest -n demo mg-rarecommendation-1780898378-rotate-auth-auto 
NAME                                   TYPE         STATUS       AGE
mg-rarecommendation-1780898378-rotate-auth-auto   RotateAuth   Successful   10m
```


You may not want to trigger recommended operations manually. Rather, trigger them autonomously in a preferred schedule when infrastructure is idle or traffic rate is at the lowest. For this purpose, You can create a `MaintenanceWindow` custom resource where you can set your desired schedule/period for triggering these recommended operations automatically. See [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for detailed documentation. Here's a sample one:


You can now create a `ApprovalPolicy` custom resource to refer this `MaintenanceWindow` for particular DB type. See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for detailed documentation. Following is a sample `ApprovalPolicy` for any `MongoDB` custom resource deployed in `demo` namespace. This `ApprovalPolicy` custom resource is referring to the `elastic-maintenance` MaintenanceWindow created in the same namespace. You can also create `ClusterMaintenanceWindow` instead (see [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md)) which is effective for cluster-wide operations and refer it here. The following ApprovalPolicy will trigger recommended operations when referred maintenance window timeframe is reached.

Lastly, If you want to reject a recommendation, you can just set `ApprovalStatus` to `Rejected` in the recommendation status section. Here's how you can do it using kubectl cli.

```bash
$ kubectl patch recommendation mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 \
                                    -n demo \
                                    --type merge \
                                    --subresource='status' \
                                    -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/mg-rarecommendation-x-mongodb-x-rotate-auth-1fwvy3 patched
```

For complete reference on all Recommendation fields, phases, and status conditions, see [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md).
