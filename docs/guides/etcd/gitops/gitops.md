---
title: GitOps Etcd
menu:
  docs_{{ .version }}:
    identifier: etcd-using-gitops
    name: GitOps Etcd
    parent: etcd-gitops
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Etcd using the KubeDB GitOps Operator

This guide shows how to use the KubeDB GitOps operator to create an `Etcd` cluster and manage updates
to it through a GitOps workflow.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured to talk to it. If you do not already have
  one, you can create one with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the `KubeDB` operator in your cluster following the steps
  [here](/docs/setup/README.md). Two extra Helm values are needed for this guide:
  - `--set kubedb-crd-manager.installGitOpsCRDs=true` to install the GitOps CRDs and enable the
    GitOps operator.
  - `--set featureGates.Etcd=true` — etcd support is an alpha feature gate and is off by default.

- Install a GitOps tool such as `ArgoCD` or `FluxCD` and point it at a Git repository. This guide
  uses `ArgoCD`; you can install it by following the steps
  [here](https://argo-cd.readthedocs.io/en/stable/getting_started/), and the `argocd` CLI
  [here](https://argo-cd.readthedocs.io/en/stable/cli_installation/).

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: the YAML files used in this tutorial are stored in
> [docs/examples/etcd/gitops](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/gitops)
> of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Register the repository with ArgoCD

### Public repository

```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace demo
```

### Private repository over HTTPS

```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace demo --username <username> --password <github-token>
```

### Private repository over SSH

```bash
argocd repo add <ssh-repo-url> \
  --ssh-private-key-path <path-to-private-key>

argocd app create <application-name> \
  --repo <repository-url> \
  --path <repository-path> \
  --dest-server <kubernetes-api-server> \
  --dest-namespace <target-namespace>
```

## Create an Etcd cluster using GitOps

Commit the following manifest to your repository:

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: Etcd
metadata:
  name: ha-etcd
  namespace: demo
spec:
  version: "3.5.21"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      containers:
        - name: etcd
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
  deletionPolicy: WipeOut
```

Lay the repository out like this:

```bash
$ tree .
├── kubedb
    └── etcd.yaml

1 directory, 1 file
```

Commit and push. ArgoCD syncs the repository and creates the `gitops.kubedb.com/v1alpha1` `Etcd`
object in the cluster. The gitops operator validates it and creates the actual
`kubedb.com/v1alpha2` `Etcd` database:

```bash
$ kubectl get etcds.gitops.kubedb.com,etcds.kubedb.com -n demo
NAME                             AGE
etcd.gitops.kubedb.com/ha-etcd   2m11s

NAME                      VERSION   STATUS   AGE
etcd.kubedb.com/ha-etcd   3.5.21    Ready    2m11s
```

And the KubeDB provisioner creates the usual offshoot resources for the database. Note there is a
single container per pod — etcd exposes its own Prometheus metrics on port `2381`, so KubeDB does not
add an exporter sidecar:

```bash
$ kubectl get petset,pod,secret,service,appbinding -n demo -l 'app.kubernetes.io/instance=ha-etcd'
NAME                                   AGE
petset.apps.k8s.appscode.com/ha-etcd   3m26s

NAME            READY   STATUS    RESTARTS   AGE
pod/ha-etcd-0   1/1     Running   0          3m26s
pod/ha-etcd-1   1/1     Running   0          3m04s
pod/ha-etcd-2   1/1     Running   0          2m42s

NAME                  TYPE                       DATA   AGE
secret/ha-etcd-auth   kubernetes.io/basic-auth   2      3m29s

NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/ha-etcd        ClusterIP   10.43.169.122   <none>        2379/TCP                     3m29s
service/ha-etcd-pods   ClusterIP   None            <none>        2379/TCP,2380/TCP   3m29s

NAME                                         TYPE              VERSION   AGE
appbinding.appcatalog.appscode.com/ha-etcd   kubedb.com/etcd   3.5.21    3m26s
```

## Update the cluster using GitOps

Every change below follows the same loop: edit `kubedb/etcd.yaml`, commit, push, let ArgoCD sync, and
watch the gitops operator turn the drift into an `EtcdOpsRequest`. You never create an OpsRequest
yourself.

### Scale resources vertically

Change the container resources in `etcd.yaml`:

```yaml
  podTemplate:
    spec:
      containers:
        - name: etcd
          resources:
            limits:
              memory: 2Gi
            requests:
              cpu: 700m
              memory: 2Gi
```

Commit and push. The gitops operator detects the resource change and creates a `VerticalScaling`
`EtcdOpsRequest`:

```bash
$ kubectl get etcds.gitops.kubedb.com,etcds.kubedb.com,etcdopsrequest -n demo
NAME                             AGE
etcd.gitops.kubedb.com/ha-etcd   13m

NAME                      VERSION   STATUS   AGE
etcd.kubedb.com/ha-etcd   3.5.21    Ready    13m

NAME                                                           TYPE              STATUS        AGE
etcdopsrequest.ops.kubedb.com/ha-etcd-verticalscaling-i0kr1l    VerticalScaling   Progressing   2s
```

Once the OpsRequest is `Successful`, verify the change on a pod:

```bash
$ kubectl get pod -n demo ha-etcd-0 -o json | jq '.spec.containers[0].resources'
```

> If an `EtcdAutoscaler` with a `spec.compute` section already targets `ha-etcd`, the gitops operator
> deliberately leaves vertical scaling alone so the two controllers do not fight over the same field.

### Scale the number of members

Change `spec.replicas` from `3` to `5`:

```yaml
spec:
  replicas: 5
```

Commit and push. The gitops operator creates a `HorizontalScaling` `EtcdOpsRequest`:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                             TYPE                STATUS        AGE
ha-etcd-horizontalscaling-wvxu5x HorizontalScaling   Progressing   6s
ha-etcd-verticalscaling-i0kr1l   VerticalScaling     Successful    7m54s
```

The OpsRequest itself just patches the replica count; the provisioner does the careful part. Each new
ordinal is added to the etcd membership as a **learner** first, catches up with the leader, and is
only then promoted to a voting member — one membership change per reconcile pass, so quorum is never
put at risk. That is why members appear one at a time rather than all at once:

```bash
$ kubectl get pod -n demo -l 'app.kubernetes.io/instance=ha-etcd'
NAME        READY   STATUS    RESTARTS   AGE
ha-etcd-0   1/1     Running   0          9m4s
ha-etcd-1   1/1     Running   0          10m
ha-etcd-2   1/1     Running   0          9m44s
ha-etcd-3   1/1     Running   0          2m58s
ha-etcd-4   1/1     Running   0          2m23s
```

Scaling down works the same way, in reverse: the highest-ordinal member is removed from the etcd
membership before the PetSet shrinks.

> Keep the member count odd. etcd tolerates `(N-1)/2` failures, so a 4-member cluster tolerates
> exactly as many failures as a 3-member one while costing more. The webhook does not reject an even
> count, but it is not what you want.

### Expand the volume

Increase the storage request:

```yaml
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

Commit and push. The gitops operator creates a `VolumeExpansion` `EtcdOpsRequest`:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                            TYPE              STATUS        AGE
ha-etcd-volumeexpansion-9cx1la  VolumeExpansion   Progressing   4s
```

Verify once it succeeds:

```bash
$ kubectl get pvc -n demo -l 'app.kubernetes.io/instance=ha-etcd'
NAME                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   AGE
data-ha-etcd-0      Bound    pvc-1a2b3c4d-0000-4a0a-9a3f-6d2b0c8a1e44   2Gi        RWO            41m
data-ha-etcd-1      Bound    pvc-2b3c4d5e-0000-4a0a-9a3f-6d2b0c8a1e45   2Gi        RWO            40m
data-ha-etcd-2      Bound    pvc-3c4d5e6f-0000-4a0a-9a3f-6d2b0c8a1e46   2Gi        RWO            40m
```

> Shrinking the volume is not supported. If the committed request is smaller than the live one, the
> gitops operator refuses the change with a `downward scaling of volume is not supported` error
> instead of generating an OpsRequest.

### Reconfigure etcd tuning knobs

Custom configuration for etcd in KubeDB means the typed `spec.configuration.tuning` block, and
nothing else. Add it:

```yaml
spec:
  configuration:
    tuning:
      quotaBackendBytes: 8589934592
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
      snapshotCount: 10000
```

Commit and push. The gitops operator creates a `Reconfigure` `EtcdOpsRequest` carrying the new tuning
block:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                         TYPE          STATUS        AGE
ha-etcd-reconfigure-i4r23j   Reconfigure   Progressing   3s
```

etcd has no live-reload path for these flags, so the `Reconfigure` OpsRequest restarts the members —
one at a time, moving raft leadership off the leader first, waiting for each pod to become Ready and
for quorum to be healthy before touching the next.

Removing the `tuning` block from Git generates another `Reconfigure` OpsRequest that clears the knobs
again. KubeDB has no defaults of its own for these: an unset knob simply renders no flag, so etcd's
own built-in default applies.

> `spec.configuration.applyConfig` is **not** supported for etcd — the `Reconfigure` webhook rejects
> it outright — and `spec.configuration.configSecret` is accepted but never mounted. etcd's
> `--config-file` is mutually exclusive with the individual command line flags KubeDB must set for
> cluster bootstrap, so only the typed tuning knobs can be reconciled.

### Update the version

Change `spec.version` to another version present in your `EtcdVersion` catalog — for example, from
`3.5.21` to `3.6.4`:

```yaml
spec:
  version: "3.6.4"
```

Commit and push. The gitops operator creates an `UpdateVersion` `EtcdOpsRequest`:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                           TYPE            STATUS        AGE
ha-etcd-versionupdate-1wxgt9   UpdateVersion   Progressing   5s
```

Version updates are constrained to the same major version, no downgrade, and at most one minor step
at a time. Those rules are enforced by the ops-manager when the generated `UpdateVersion` request
starts executing, not at admission — so a target outside that window produces an `EtcdOpsRequest`
that is created and then reported `Failed`.

### Rotate credentials and TLS

Changing `spec.authSecret` generates a `RotateAuth` `EtcdOpsRequest`, and changing `spec.tls`
generates a `ReconfigureTLS` one:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                             TYPE             STATUS       AGE
ha-etcd-rotate-auth-zot83x       RotateAuth       Successful   170m
ha-etcd-reconfiguretls-91fseg    ReconfigureTLS   Successful   16m
```

`RotateAuth` does **not** restart pods — etcd applies password changes live through its own RBAC
API. `ReconfigureTLS` does require the leader-aware rolling restart, since the certificates are
mounted files.

### Migrate to another StorageClass

Changing `spec.storage.storageClassName` generates a `StorageMigration` `EtcdOpsRequest`:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                              TYPE               STATUS        AGE
ha-etcd-storagemigration-p2ku7z   StorageMigration   Progressing   8s
```

### Fields that trigger a Restart

Two spec sections have no dedicated OpsRequest type and are applied through a `Restart` instead:

- `spec.monitor`
- `spec.archiver`

```bash
$ kubectl get etcdopsrequest -n demo
NAME                      TYPE      STATUS        AGE
ha-etcd-restart-nhjk9u    Restart   Progressing   2s
```

The `Restart` OpsRequest is leader aware: if the pod being restarted currently holds raft leadership,
the operator hands leadership off to another healthy, caught-up member first, then recreates pods one
at a time.

### Fields applied without an OpsRequest

A few fields are cheap and safe to patch straight onto the live `Etcd` object, so no OpsRequest is
generated for them: `spec.autoOps`, `spec.serviceTemplates`, `spec.halted`, `spec.healthChecker`,
`spec.deletionPolicy` and `spec.allowedSchemas`.

## Inspect what GitOps did

The wrapper object records its own history:

```bash
$ kubectl get etcds.gitops.kubedb.com -n demo ha-etcd -o jsonpath='{.status.gitops}' | jq
{
  "gitOpsInfo": [
    {
      "observedGeneration": 2,
      "operations": [
        {
          "name": "ha-etcd-verticalscaling-i0kr1l",
          "status": "Successful"
        }
      ],
      "changeRequestStatus": "Current"
    }
  ]
}
```

`changeRequestStatus` is one of `Pending`, `InProgress`, `Current` or `Failed`.

## Cleanup

```bash
$ kubectl delete etcds.gitops.kubedb.com -n demo ha-etcd
$ kubectl delete ns demo
```

Deleting the wrapper deletes the live `Etcd` object it owns; what happens to the data then depends on
`spec.deletionPolicy`.

## Next Steps

- Read the [Etcd GitOps Overview](/docs/guides/etcd/gitops/overview.md).
- Back up and restore your etcd cluster with
  [KubeStash](/docs/guides/etcd/backup/kubestash/overview/index.md).
- Detail concepts of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object.
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
