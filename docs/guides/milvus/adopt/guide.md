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

Adoption is a short, ordered handover. `kubectl dba adopt milvus` runs it for you: it reads the running cluster, checks it, creates the Milvus object and moves the roles one at a time, and it can give the cluster back. The rest of this guide is the contract that command follows, and the way to do the same by hand.

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

## Adopt with kubectl dba

The [`kubectl dba`](/docs/setup/README.md) plugin adopts a cluster in three commands. `RELEASE` names the running cluster: its workloads are those with `app.kubernetes.io/instance=RELEASE`, and the `component` label names each workload's role, as the upstream Milvus Helm chart sets them. For other charts, select the workloads with `--selector` and `--role-label`, or map each one with `--workload role[/group]=Kind/name`.

### 1. Check: `plan`

```bash
$ kubectl dba adopt milvus plan my-release -n demo
Pre-flight checks for my-release in namespace demo:
  [OK   ] proxy         Deployment my-release-milvus-proxy, 3 replicas
  [OK   ] mixcoord      Deployment my-release-milvus-mixcoord, 2 replicas
  ...
  [OK   ] config read from ConfigMap my-release-milvus key default.yaml as default.yaml; ConfigMap my-release-milvus key user.yaml as user.yaml
  [OK   ] client Service my-release-milvus (10.43.248.56) is taken over: same name, IP and ports
  [OK   ] write-ahead log: woodpecker
  [OK   ] etcd http://my-release-etcd-0...:2379,..., root path "by-dev" (kept)
  [OK   ] object storage minio.storage.svc, bucket milvus-data, root path "prod" (kept)
  [OK   ] image milvusdb/milvus:v2.6.15 -> MilvusVersion 2.6.15
  [OK   ] root credentials from common.security.defaultRootPassword log in to my-release-milvus-proxy-...
  [OK   ] Service my-release-milvus-mixcoord is not headless; run deletes it (KubeDB recreates it headless; Milvus does not use it)
```

`plan` changes nothing. It finds the workloads, the effective `milvus.yaml` (the chart's ConfigMap files, merged in the order Milvus reads them), the image and the Service clients use, and checks every row of [the contract](#the-contract). It logs in to a running proxy with the root credentials, because Milvus applies a root password only on its first start: a `defaultRootPassword` that was changed since is rejected here, and you give the current one with `--auth-secret` (an existing basic-auth Secret) or `--root-password-file`. An `ERROR` blocks the adoption; a `WARN` (an init container, environment variables, affinity or secondary networks that are not carried over) needs a look before you continue. With `--output-dir` it writes the objects it would create, so you can review or edit them and `kubectl apply` them yourself; `run` then uses what exists.

### 2. Adopt: `run`

```bash
$ kubectl dba adopt milvus run my-release -n demo --yes
```

`run` repeats the checks, then:

1. Creates the Secrets (`<name>-auth`, `<name>-object-storage`, `<name>-adopt-config` with the carried `milvus.yaml`), a ConfigMap `<name>-adopt` that records the adoption, and the Milvus with every role at 0 replicas. It deletes a chart Service that sits on a name KubeDB needs headless, and records it.
2. Waits for the operator to render the config and create the PetSets, and compares the rendered config with the running one on every setting that decides where data is read from: object storage, etcd root path, write-ahead log, security and ports. On any difference it stops before a pod moves.
3. Hands the roles over one at a time, waiting for each pod to be Ready, and records each role in the ConfigMap. If the command stops, run it again and it continues where it was.

Two methods:

| `--method` | How | Client impact | Capacity |
|---|---|---|---|
| `overlap` (default) | Proxy, MixCoord, DataNode, StreamingNode, QueryNode. Start KubeDB's pods of a role, wait until they are Ready, then stop the chart's. QueryNodes move one at a time: one starts, serves for `--querynode-delay` (default 1m; raise it for large collections), then one old QueryNode stops | No role stops serving. On a test cluster (single-replica collections) clients saw one 3 s insert timeout while the StreamingNodes moved, and one 30 to 60 s stall of searches and inserts at the first or last QueryNode swap, while chart and KubeDB QueryNodes ran side by side. No acknowledged write was lost | One extra set of pods of a role at a time (one extra QueryNode) |
| `stop-start` | MixCoord, DataNode, QueryNode, StreamingNode, Proxy. Stop the chart's pods of a role, then start KubeDB's | Each role is briefly unavailable (searches while QueryNodes reload, writes while StreamingNodes restart) | None |

`kubectl dba adopt milvus status my-release -n demo` shows where each role runs.

### 3. Give it back, if needed: `rollback`

```bash
$ kubectl dba adopt milvus rollback my-release -n demo --yes --cleanup
```

`rollback` deletes the Milvus with `kubedb.com/milvus-orphan-on-delete`, so its pods keep serving, then moves each role back to the chart's workloads, starting the chart's pods before it stops KubeDB's. Once the chart's proxies are Ready, the client Service gets back its original selector, labels and ports. `--cleanup` then deletes the PetSets (at 0 replicas), the Services and the rendered config KubeDB left, and recreates the chart Services the adoption removed. PVCs and the credential Secrets are kept. It works on a partly adopted cluster too, and for a Milvus adopted by hand (pass `--name`). Expect a short pause in writes while the StreamingNodes move back: the chart's StreamingNodes take write-ahead log channels as soon as they start, before their readiness probe passes (about 30 s on the test cluster).

## The contract

Each row is a way an adoption can go wrong silently. Gather the current values from the running cluster first: its effective `milvus.yaml`, its Services, PVCs and pod specs.

| Field | Set it to | What happens otherwise |
|---|---|---|
| `spec.metaStorage.externallyManaged` | `true` | KubeDB creates a new, empty etcd |
| `spec.metaStorage.endpoints` | The current etcd client endpoints. One Service address is enough | The operator cannot reach the metadata |
| `spec.metaStorage.rootPath` | The current `etcd.rootPath` (the Milvus default is `by-dev`) | KubeDB defaults it to `<name>-etcd`. Milvus starts on an empty metadata tree and looks healthy but empty |
| `spec.metaStorage.authSecret`, `clientTLSSecret` | Only if the etcd uses auth or TLS | Connection fails |
| `spec.objectStorage.configSecret` | A Secret holding every non-default key of the current `minio:` section: `address`, `port`, `useSSL`, `bucketName`, `rootPath`, `useIAM`, `cloudProvider`, `region` and the credentials | KubeDB defaults `rootPath` to `files`, `useSSL` and `useIAM` to `false`. Segments are read from the wrong place |
| `spec.configuration.secretName` | A Secret whose `milvus.yaml` key holds the current effective config, without the `etcd` section (the Milvus object sets the etcd endpoints and root path). Keep the `woodpecker` section as it is, with `storage.type: minio` | KubeDB's own defaults apply, for example `quotaAndLimits.enabled: true`. KubeDB's WAL defaults replace the whole `woodpecker` section, so dropping it loses settings such as `woodpecker.storage.rootPath` and can move where the write-ahead log is read from |
| `proxy.http.port` in the carried `milvus.yaml` | `""` if clients use the RESTful API on the gRPC port (the Milvus default) | KubeDB defaults the REST API to port 8080. A running cluster serves it on 19530, so REST clients would break |
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
- Its selector moves to the KubeDB proxies only once at least one of them is Ready. Until then it keeps serving from the old proxies, so new connections are never left without endpoints. It gets no owner reference, so it is not deleted with the Milvus.
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

## Before you create the Milvus

- Check the write-ahead log the cluster really uses: `mq.type` (or `messageQueue`) in its effective `milvus.yaml` must be `woodpecker`, and the StreamingNode logs must show `walName=woodpecker`. Some charts default cluster mode to Pulsar even when `woodpecker.storage` settings are present; the upstream Milvus Helm chart needs `woodpecker.enabled: true`.
- KubeDB creates headless Services named `<name>-mixcoord`, `<name>-datanode`, `<name>-querynode` and `<name>-streamingnode`. A Service of one of these names left by the chart that is **not** headless cannot be converted (a Service's cluster IP is immutable), and KubeDB reports it. Milvus components reach each other through their etcd sessions, not through these Services, so delete such a Service before you create the Milvus.
- If you rehearse with a second, temporary Milvus object, give it its own copy of the auth Secret. KubeDB records on an auth Secret which Milvus uses it, and a second Milvus is then refused that Secret.

## Prepare the Secrets

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/adopt/yamls/object-storage-secret.yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/adopt/yamls/auth-secret.yaml
$ kubectl create secret generic -n demo milvus-current-config --from-file=milvus.yaml=./current-milvus.yaml
```

## Hand over the workloads by hand

This is what `kubectl dba adopt milvus run` does; follow it only if you cannot use the command.

The pods of one role, old and new, can run side by side for the length of a handover: every Milvus 2.6 component registers its own session in etcd, MixCoord runs active-standby, and the other roles share work across their pods. So a role can move either by starting the KubeDB pods before stopping the old ones (overlap, no client impact, needs room for the extra pods), or by stopping the old ones first (stop-start, a short gap per role).

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

3. Overlap: for each role, in the order Proxy, MixCoord, DataNode, StreamingNode, QueryNode, start the KubeDB pods, wait until they are Ready, then stop the old ones. Move QueryNodes one at a time (KubeDB to 1, wait Ready and let it serve a while, old to N-1, and so on), so at most one extra QueryNode runs:

   ```bash
   $ kubectl patch milvus -n demo milvus --type=merge -p '{"spec":{"topology":{"distributed":{"mixcoord":{"replicas":2}}}}}'
   $ kubectl wait -n demo --for=condition=Ready pod -l app.kubernetes.io/instance=milvus,kubedb.com/role=mixcoord --timeout=15m
   # then stop the old workload of the role; its PVCs are kept
   $ kubectl scale -n demo deployment/<old-mixcoord> --replicas=0
   $ kubectl wait -n demo --for=delete pod -l <old-mixcoord-selector> --timeout=10m
   ```

   Stop-start: for each role, in the order MixCoord, DataNode, QueryNode, StreamingNode, Proxy, stop the old workload first, wait until its pods are gone, then raise the KubeDB role.

   Move to the next role once the new pods are Ready. While the KubeDB proxies are at 0 the Milvus reports `NotReady`; that is expected.

### Single window

Scale every old workload to 0, then apply the Milvus with the final replicas. This is simpler, and the whole cluster is unavailable until the QueryNodes have loaded the collections.

## Validate

```bash
$ kubectl get milvus -n demo milvus
$ kubectl get milvus -n demo milvus -o jsonpath='{range .status.conditions[*]}{.type}={.status}{"\n"}{end}'
```

Then, from a client, check the collection count against the number taken before the handover, run known queries with known results, and run a smoke test from each application.

## Roll back

The data never moved, so a rollback only changes who runs the pods. `kubectl dba adopt milvus rollback` does it; by hand:

1. Make sure the annotation `kubedb.com/milvus-orphan-on-delete: "true"` is set, set `spec.deletionPolicy` to `Halt`, and delete the Milvus. Its pods keep serving under their PetSets, and the operator no longer moves the client Service.
2. For each role, in the order QueryNode, StreamingNode, DataNode, MixCoord, Proxy: scale the old workload back up, wait until its pods are Ready, then scale the KubeDB PetSet to 0.
3. Before the proxy PetSet goes to 0, point the client Service back at the old proxies: its original selector, and ports whose `targetPort` the old proxy pods define (KubeDB's ports target the port names `proxy`, `metrics` and `http` of its own pods).

## Hand the cluster back

With the annotation `kubedb.com/milvus-orphan-on-delete: "true"` set, deleting the Milvus no longer deletes the cluster. Before it lets the object go, the operator removes the Milvus owner reference from the PetSets, Services, config Secret, ServiceAccount, Role and RoleBinding (and from a KubeDB-managed meta etcd, if one is used), and keeps PVCs and Secrets as with `deletionPolicy: Halt`, whatever the policy says. The pods keep serving under the PetSet controller.

```bash
$ kubectl patch milvus -n demo milvus --type=merge -p '{"spec":{"deletionPolicy":"Halt"}}'
$ kubectl delete milvus -n demo milvus
$ kubectl get petset,pods,svc -n demo -l app.kubernetes.io/instance=milvus
```

`DoNotTerminate` still blocks deletion, which is why the policy is changed first. Keep the PetSet controller installed while the pods are run by PetSets. To move the cluster back to a chart, start its workloads role by role while scaling the PetSets down, in the same order as the handover.
