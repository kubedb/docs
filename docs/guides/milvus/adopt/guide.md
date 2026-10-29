---
title: Adopt an Existing Milvus Cluster
menu:
  docs_{{ .version }}:
    identifier: milvus-adopt-guide
    name: Guide
    parent: milvus-adopt
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Adopt an Existing Milvus Cluster

This guide moves a Distributed Milvus that already runs in Kubernetes (for example from a Helm chart) under KubeDB management **in place**. The metadata stays in the existing etcd, the segments, indexes and write-ahead log stay in the existing bucket, and the credentials stay the same. Only the workloads change hands: the chart's Deployments and StatefulSets are replaced by KubeDB PetSets that run the same Milvus version against the same data.

There is no automatic takeover. Adoption is a short, ordered handover that you run, and this guide gives the contract the Milvus object must meet so that the new pods see the existing data.

## Before You Begin

- You should be familiar with the [Milvus](/docs/guides/milvus/concepts/milvus.md) CRD and a [Distributed](/docs/guides/milvus/quickstart/distributed.md) deployment.
- Rehearse the whole procedure on a copy of the setup first, with test data. Measure how long QueryNodes take to load your collections; that is most of the handover time.
- Take a backup you have restored at least once (see [Backup and Restore](/docs/guides/milvus/backup/logical/guide.md)).

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/adopt/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/adopt/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## What must already be true

| Requirement | Why |
|---|---|
| Milvus 2.6 with a matching KubeDB `MilvusVersion` (use the `-gpu` entry for a GPU build) | Adoption keeps the version; it is not an upgrade |
| Write-ahead log is Woodpecker on object storage (`mq.type: woodpecker`, `woodpecker.storage.type: minio`) | KubeDB always runs Woodpecker on object storage and rejects any other WAL. Changing the WAL on a live cluster is a migration, not an adoption |
| An etcd used only by this Milvus | KubeDB connects to it as an externally managed meta store and never changes it |

## The contract

Each row is a way an adoption can go wrong silently. Gather the current values from the running cluster first: its effective `milvus.yaml`, its Services, PVCs and pod specs.

| Field | Set it to | What happens otherwise |
|---|---|---|
| `spec.metaStorage.externallyManaged` | `true` | KubeDB creates a new, empty etcd |
| `spec.metaStorage.endpoints` | The current etcd client endpoints. One Service address is enough | The operator cannot reach the metadata |
| `spec.metaStorage.rootPath` | The current `etcd.rootPath` (the Milvus default is `by-dev`) | KubeDB defaults it to `<name>-etcd`. Milvus starts on an empty metadata tree and looks healthy but empty |
| `spec.metaStorage.authSecret`, `clientTLSSecret` | Only if the etcd uses auth or TLS | Connection fails |
| `spec.objectStorage.configSecret` | A Secret holding every non-default key of the current `minio:` section: `address`, `port`, `useSSL`, `bucketName`, `rootPath`, `useIAM`, `cloudProvider`, `region` and the credentials | KubeDB defaults `rootPath` to `files`, `useSSL` and `useIAM` to `false`. Segments are read from the wrong place |
| `spec.configuration.secretName` | A Secret whose `milvus.yaml` key holds the current effective config, without the `mq`, `woodpecker` and `etcd` sections | KubeDB's own defaults apply, for example `quotaAndLimits.enabled: true` |
| `spec.authSecret` | A `kubernetes.io/basic-auth` Secret with the **current** root user and password | Milvus applies a root password only on first start. A new password does not reach the running cluster, and the health checker cannot log in |
| `spec.disableSecurity` | `true` only if the cluster runs without authentication | Clients that connect without credentials break |
| `metadata.name` | The name of the Service clients use today, if possible | The proxy Service is named after the Milvus. If a Service of that name exists, KubeDB takes it over and points it at the new proxies, so clients keep their address |
| `spec.healthChecker.disableWriteCheck` | `true` | The health checker creates a `health_check_test` collection and writes to it every period |
| `spec.autoOps.disabled` | `true`, if every change must go through your own change control | The ops manager can propose version updates and credential rotations for the database |
| `spec.deletionPolicy` | `DoNotTerminate` | Deleting the object would delete the workloads |
| `kubedb.com/milvus-orphan-on-delete` annotation | `"true"` | Without it there is no clean way to hand the cluster back. See [Hand the cluster back](#hand-the-cluster-back) |

Also check:

- The StreamingNode PVC. KubeDB names it `data-<name>-streamingnode-<ordinal>`. With Woodpecker on object storage it holds no data that must survive, but confirm that before the handover. If it does, pre-create PVCs with those names from the existing volumes.
- The pod user. KubeDB images run as UID 1000 (`spec.restrictedSCC: true` lets OpenShift assign it). Volumes written as root by the old pods must be writable by the new user.
- No `MilvusArchiver` selects the namespace. The archiver needs a KubeDB-managed etcd and is rejected for an externally managed one.
- GPU and SR-IOV settings, if used: see [GPU and SR-IOV](/docs/guides/milvus/gpu-sriov/guide.md). Remove any advertise-IP init container of your own; KubeDB adds its own.

## Keep the client endpoint

Name the Milvus after the Service your clients use. KubeDB then takes that Service over instead of creating a new one:

- The Service keeps its name and cluster IP, so DNS and IP-based clients are unaffected.
- Its selector moves to the KubeDB proxies, and it gets no owner reference, so it is not deleted with the Milvus.
- KubeDB sets three ports: `proxy` (19530), `metrics` (9091) and `http` (8080). A port that shares a name, or a number and protocol, with one of these is replaced. Every other port is kept as it is.
- Labels, annotations and the Service type come from KubeDB and `spec.serviceTemplates`.

If clients use another port number, or the Service is a `LoadBalancer` or `NodePort`, describe it in the `primary` service template. Ports in the template are matched by name:

```yaml
spec:
  serviceTemplates:
    - alias: primary
      metadata:
        annotations:
          # copy the annotations the Service carries today, e.g. for the load balancer
      spec:
        type: LoadBalancer
        ports:
          - name: proxy
            port: 19531        # the port clients use today
            nodePort: 31953    # only for NodePort/LoadBalancer, to keep the node port
```

Inside the pods the proxy always listens on 19530; the template only changes the Service side. So do not carry a different `proxy.port` in the `milvus.yaml` you give KubeDB.

When you later retire the chart, do not run `helm uninstall`: Helm deletes every object in its release, including the Service KubeDB now uses. Delete the chart's other objects one by one, then remove the release record with `kubectl delete secret -l owner=helm,name=<release>`.

## Prepare the Secrets

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/adopt/yamls/object-storage-secret.yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/adopt/yamls/auth-secret.yaml
$ kubectl create secret generic -n demo milvus-current-config --from-file=milvus.yaml=./current-milvus.yaml
```

## Hand over the workloads

Two copies of a role must never run at the same time: both would register in the same etcd. The handover therefore always stops the old workload of a role before KubeDB starts its own.

### Staged, one role at a time

1. Create the Milvus with every role at `replicas: 0`. KubeDB validates the metadata store, renders the config and creates empty PetSets and the Services. Nothing runs yet:

   ```bash
   $ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/adopt/yamls/milvus-adopt.yaml
   ```

2. Compare the rendered config with the running one before any pod moves:

   ```bash
   # the rendered config Secret is named <name>-<last six characters of the Milvus UID>
   $ MV_UID=$(kubectl get milvus -n demo milvus -o jsonpath='{.metadata.uid}')
   $ kubectl get secret -n demo milvus-${MV_UID: -6} -o jsonpath='{.data.milvus\.yaml}' | base64 -d > rendered.yaml
   $ diff <(yq -P 'sort_keys(..)' current-milvus.yaml) <(yq -P 'sort_keys(..)' rendered.yaml)
   ```

   Every difference in `etcd`, `minio`, `common.security` and the role sections must be explained before you continue.

3. For each role, in the order MixCoord, DataNode, QueryNode, StreamingNode, Proxy:

   ```bash
   # stop the old workload of the role; its PVCs are kept
   $ kubectl scale -n demo deployment/<old-mixcoord> --replicas=0
   $ kubectl wait -n demo --for=delete pod -l <old-mixcoord-selector> --timeout=10m
   # start the KubeDB one
   $ kubectl patch milvus -n demo milvus --type=merge -p '{"spec":{"topology":{"distributed":{"mixcoord":{"replicas":2}}}}}'
   $ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus,kubedb.com/role=mixcoord -w
   ```

   Move to the next role once the new pods are Ready. While the proxies are at 0 the Milvus reports `NotReady`; that is expected.

### Single window

Scale every old workload to 0, then apply the Milvus with the final replicas. This is simpler, and the whole cluster is unavailable until the QueryNodes have loaded the collections.

## Validate

```bash
$ kubectl get milvus -n demo milvus
$ kubectl get milvus -n demo milvus -o jsonpath='{range .status.conditions[*]}{.type}={.status}{"\n"}{end}'
```

Then, from a client, check the collection count against the number taken before the handover, run known queries with known results, and run a smoke test from each application.

## Roll back

The data never moved, so a rollback only changes who runs the pods:

1. Make sure the annotation `kubedb.com/milvus-orphan-on-delete: "true"` is set, and set `spec.deletionPolicy` to `Halt`.
2. Scale the KubeDB roles to 0, in the reverse order.
3. Scale the old workloads back up.
4. Delete the Milvus once the old workloads serve.

## Hand the cluster back

With the annotation `kubedb.com/milvus-orphan-on-delete: "true"` set, deleting the Milvus no longer deletes the cluster. Before it lets the object go, the operator removes the Milvus owner reference from the PetSets, Services, config Secret, ServiceAccount, Role and RoleBinding (and from a KubeDB-managed meta etcd, if one is used), and keeps PVCs and Secrets as with `deletionPolicy: Halt`, whatever the policy says. The pods keep serving under the PetSet controller.

```bash
$ kubectl patch milvus -n demo milvus --type=merge -p '{"spec":{"deletionPolicy":"Halt"}}'
$ kubectl delete milvus -n demo milvus
$ kubectl get petset,pods,svc -n demo -l app.kubernetes.io/instance=milvus
```

`DoNotTerminate` still blocks deletion, which is why the policy is changed first. Keep the PetSet controller installed while the pods are run by PetSets. To move the cluster back to a chart, start its workloads role by role while scaling the PetSets down, in the same order as the handover.
