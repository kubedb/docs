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
| `overlap` (default) | Proxy, MixCoord, DataNode, StreamingNode, QueryNode. Start KubeDB's pods of a role, wait until they are Ready, then stop the chart's. QueryNodes move one at a time: one starts, serves for `--querynode-delay` (default 1m; raise it for large collections), then one old QueryNode stops | No role stops serving. On a test cluster (single-replica collections) clients saw one 3 s insert timeout while the StreamingNodes moved, and, in some runs, a 30 to 60 s stall of searches and inserts at the first or last QueryNode swap, while chart and KubeDB QueryNodes ran side by side. No acknowledged write was lost | One extra set of pods of a role at a time (one extra QueryNode) |
| `stop-start` | MixCoord, DataNode, QueryNode, StreamingNode, Proxy. Stop the chart's pods of a role, then start KubeDB's | Each role is briefly unavailable (searches while QueryNodes reload, writes while StreamingNodes restart) | None |

`kubectl dba adopt milvus status my-release -n demo` shows where each role runs.

After the adoption, leave the Helm release alone: `helm upgrade` would start the chart's workloads again next to KubeDB's, and `helm uninstall` would delete the client Service KubeDB now uses. To retire the chart, see [Keep the client endpoint](#keep-the-client-endpoint).

### 3. Give it back, if needed: `rollback`

```bash
$ kubectl dba adopt milvus rollback my-release -n demo --yes --cleanup
```

`rollback` deletes the Milvus with `kubedb.com/milvus-orphan-on-delete`, so its pods keep serving, then moves each role back to the chart's workloads, starting the chart's pods before it stops KubeDB's. Once the chart's proxies are Ready, the client Service gets back its original selector, labels and ports. `--cleanup` then deletes the PetSets (at 0 replicas), the Services and the rendered config KubeDB left, and recreates the chart Services the adoption removed. PVCs and the credential Secrets are kept. It works on a partly adopted cluster too, and for a Milvus adopted by hand (pass `--name`). Expect a short pause in writes while the StreamingNodes move back: the chart's StreamingNodes take write-ahead log channels as soon as they start, before their readiness probe passes (about 30 s on the test cluster). A rollback takes longer than the adoption because each role waits for the chart's pods to be Ready and settle; on the test cluster with the default delays it took 15 to 25 minutes, mostly in the QueryNode step. `--querynode-delay` and `--settle` shorten it, at the price of a shorter wait between steps.

### Options in a file

Instead of repeating flags on every command, put the options of an adoption in one file and pass it with `-f`. `plan --output-dir DIR` writes one, `DIR/adopt.yaml`, with everything it found; review it, adjust it, and use it for the rest:

```bash
$ kubectl dba adopt milvus plan my-release -n demo --output-dir ./adopt
$ kubectl dba adopt milvus run      -f ./adopt/adopt.yaml --yes
$ kubectl dba adopt milvus status   -f ./adopt/adopt.yaml
$ kubectl dba adopt milvus rollback -f ./adopt/adopt.yaml --yes
```

`--yes` is a flag only; it is not a field of the file.

A complete file, for a cluster whose chart does not set the upstream labels and runs QueryNodes on two GPU pools:

```yaml
apiVersion: cli.kubedb.com/v1alpha1
kind: MilvusAdoption
release: milvus-prod              # replaces the RELEASE argument
namespace: milvus                 # used unless -n is given
name: ""                          # default: the name of the client Service
version: ""                       # default: the MilvusVersion matching the running image
source:
  workloads:                      # or selector / roleLabel for labelled workloads
  - {role: proxy,         kind: Deployment, name: milvus-proxy}
  - {role: mixcoord,      kind: Deployment, name: milvus-mixcoord}
  - {role: datanode,      kind: Deployment, name: milvus-datanode}
  - {role: querynode, group: l40s, kind: Deployment, name: milvus-querynode-l40s}
  - {role: querynode, group: h200, kind: Deployment, name: milvus-querynode-h200}
  - {role: streamingnode, kind: Deployment, name: milvus-streamingnode}
  configFile: ""                  # the effective milvus.yaml, only if it is not mounted
auth:
  authSecret: milvus-root         # basic-auth Secret with the current root password
  rootPasswordFile: ""            # or a file with the password
  etcdAuthSecret: ""              # only if the etcd uses authentication
  skipCredentialCheck: false
tls:                              # see "Keep the TLS of the cluster"
  caSecret: ""                    # Secret with the CA and its private key; KubeDB takes over TLS
  issuer: ""                      # or a cert-manager issuer on that CA
  clientSecret: ""                # client certificate for KubeDB, for mTLS with neither of the above
restrictedSCC: false              # true on OpenShift
handover:
  method: overlap                 # or stop-start
  querynodeDelay: 10m             # about the time one QueryNode takes to load its data
  settle: 20s
  timeout: 30m
  readyTimeout: 30m
  skipConfigCheck: false
rollback:
  cleanup: true
```

A flag given on the command line wins over the same option in the file. An unknown key is an error, so a misspelt option is never ignored. `--yes` is never read from the file: a file alone does not start a handover.

### What you must give

| Option (flag / file) | Needed |
|---|---|
| `RELEASE` / `release` | Always |
| `-n` / `namespace` | Unless the kubeconfig context already points at the namespace |
| `--yes` | For `run` and `rollback` to act; without it they only show what they would do |
| `--auth-secret` or `--root-password-file` / `auth` | When the root password is not the chart's `defaultRootPassword`; `plan` tells you |
| `--etcd-auth-secret` / `auth.etcdAuthSecret` | When the etcd uses authentication |
| `--tls-ca-secret`, `--tls-issuer` or `--tls-client-secret` / `tls` | For KubeDB to take over TLS, or when client mTLS is on; see [Keep the TLS of the cluster](#keep-the-tls-of-the-cluster) |
| `--object-storage-ca-secret` or `--object-storage-ca-configmap`, and `--object-storage-address` and `--object-storage-port` / `objectStorage` | When the Milvus must reach the object storage over TLS with a private CA after the adoption; see [Move the object storage to TLS](#move-the-object-storage-to-tls) |
| `--workload` / `source.workloads` (or `--selector`, `--role-label`) | When the workloads do not carry `app.kubernetes.io/instance=RELEASE` and a `component` label |
| `--config-file` / `source.configFile` | When the Milvus config is not mounted from a ConfigMap or Secret |

Everything else has a default. The version is not needed: it is read from the running image (`milvusdb/milvus:v2.6.15-gpu` selects MilvusVersion `2.6.15-gpu`); give `--version` only for an image pinned by digest or with a custom tag. Neither is the Helm chart version: the command reads the workloads, ConfigMap and Service the chart created, not Helm.

`kubectl dba` is a kubectl plugin that runs on your machine, with your kubeconfig; installing KubeDB in the cluster does not install it. See [Setup](/docs/setup/README.md).

## Keep the TLS of the cluster

Clients keep working through an adoption: they trust the same CA and connect to the same name. When KubeDB can issue from that CA, it takes over TLS completely, including renewal and `ReconfigureTLS`. `plan` reads the running config and picks one of three cases:

| The cluster has | What the adoption does |
|---|---|
| A cert-manager issuer on the CA clients trust | Sets `spec.tls.issuerRef` to it. KubeDB issues and renews the server and client certificates. `plan` finds the issuer by itself when the running certificate came from cert-manager; otherwise give `--tls-issuer` |
| The CA Secret, with its private key, and no issuer | Give `--tls-ca-secret`. `run` creates the Issuer `<name>-ca` from that Secret, then as above. `rollback --cleanup` removes the Issuer and leaves the Secret |
| Neither | Uses the running certificate as it is, through `spec.tls.existingSecrets.server`. If its Secret does not have the keys `tls.crt`, `tls.key` and `ca.crt`, it is copied to `<name>-tls-server`, which `rollback --cleanup` removes. Renewing it stays with you, and a `ReconfigureTLS` rotation is refused |

TLS that is off on a connection stays off: switching it on during a handover would split the old pods from the new ones. Do that afterwards with a `ReconfigureTLS` request. An etcd that uses TLS keeps its client certificate, through `spec.metaStorage.clientTLSSecret`, because KubeDB does not manage that etcd.

```bash
# Create the Issuer from the CA Secret
$ kubectl dba adopt milvus plan my-release -n demo --tls-ca-secret milvus-ca --output-dir ./adopt
# Or use an issuer that exists
$ kubectl dba adopt milvus plan my-release -n demo --tls-issuer ClusterIssuer/corporate-ca --output-dir ./adopt
```

KubeDB's certificates cover the client Service name. Any other name in the running certificate, such as an ingress host, is carried over to `spec.tls.certificates`, and `plan` lists them.

The new certificate must chain to the CA clients already trust, so this is checked twice:

- `plan` stops if a CA issuer or `--tls-ca-secret` is not the CA that signed the running certificate. It cannot read the CA of a `ClusterIssuer` or of an issuer that is not a CA issuer, so it warns.
- `run` waits for the certificates and checks them **before any pod moves**. It stops if they do not verify against the running CA, or if the issuer does not fill `ca.crt`, which KubeDB needs to reach Milvus. An ACME issuer does not, so use the third case there.

When KubeDB does not issue the certificates, `plan` also stops if:

- the running certificate does not cover the client Service name in all three forms (`<name>`, `<name>.<namespace>.svc` and `<name>.<namespace>.svc.cluster.local`), which KubeDB connects to,
- client and internal TLS use different certificates, because KubeDB mounts one for both,
- client mTLS is on and no client certificate is given with `--tls-client-secret`.

In every case `plan` stops if a certificate and key do not form a pair, a certificate does not verify against its CA, or a certificate file does not come from a Secret or ConfigMap, for example a `hostPath`. It warns when a certificate expires within 30 days.

Three things to know:

- A client that pins the server certificate itself, rather than trusting the CA, breaks when KubeDB issues a new one. Use the third case for it.
- Milvus refuses to serve the RESTful API on the gRPC port when TLS is on, and the proxy exits with `tls mode must be 0`. A cluster with TLS therefore already sets `proxy.http.port`, and the adoption keeps it.
- With client mTLS, `plan` has no client certificate to log in with, so it cannot check the root credentials and warns instead. Make sure of them yourself.

Without `kubectl dba`, set the same fields on the Milvus object:

```yaml
spec:
  tls:
    issuerRef:                  # or existingSecrets.server.name, with the keys tls.crt, tls.key, ca.crt
      apiGroup: cert-manager.io
      kind: Issuer
      name: milvus-ca
    external:
      mode: TLS                 # the running common.security.tlsMode: 1 is TLS, 2 is mTLS
    internal:
      mode: TLS                 # only if common.security.internaltlsEnabled is true
```

Remove `tls`, `internaltls`, `common.security.tlsMode` and `common.security.internaltlsEnabled` from the carried `milvus.yaml`: it wins over what the operator renders, so the old paths would point at files the new pods do not have.

## Move the object storage to TLS

The running cluster may reach its object storage over plain HTTP, and the new Milvus may have to use TLS with a private CA. This is independent of the TLS above, which is between clients and Milvus. Give the CA with `--object-storage-ca-secret NAME[/KEY]` or `--object-storage-ca-configmap NAME[/KEY]` (the key defaults to `ca.crt`; the Secret or ConfigMap must be in the namespace of the Milvus, because the pods mount it), and the TLS endpoint with `--object-storage-address` and `--object-storage-port` when it differs from the running one. An address or port without a CA is refused.

```bash
$ kubectl dba adopt milvus plan my-release -n demo \
    --object-storage-ca-secret s3-ca \
    --object-storage-address s3.example.com --object-storage-port 443 --output-dir ./adopt
```

With a CA flag the adoption:

- sets `spec.objectStorage.tls.caSecret` (or `caConfigMap`), so KubeDB mounts the CA, sets `minio.useSSL: true` and points `minio.ssl.tlsCACert` at it,
- writes `useSSL: "true"` and the new address and port to the Secret `<name>-object-storage`,
- leaves `minio.useSSL`, `minio.ssl` and, when you give a new endpoint, `minio.address` and `minio.port` out of the carried `milvus.yaml`, because the carried config wins over what the operator renders and would keep the new pods on plain HTTP,
- expects the rendered config to differ from the running one in exactly those settings, and stops on anything else.

`plan` checks that the CA exists, has the key, parses as PEM and is not expired. It cannot reach the object storage, so `run` tests it before it creates anything: a short-lived Pod runs the Milvus image, which has `curl`, with the CA mounted, and opens a TLS connection to the endpoint. A wrong CA, a host the certificate does not name, or an endpoint the cluster cannot reach stops `run` with `nothing was created or moved`, and the Pod is removed. Pass `--skip-object-storage-check` to skip it. The bucket and root path stay as they are, so the old pods and the new pods read the same data while both run.

If `rollback` finds that the KubeDB pods of a role are not all Ready, for example because they crash on a bad setting, it does not wait for the PetSet to scale down in order, which it never would. It starts the chart's workload first and then deletes the PetSet.

Without the flags nothing changes. To switch an adopted Milvus to TLS later, set `spec.objectStorage.tls` and the new address, and remove `minio.useSSL` and `minio.address` from the Secret `<name>-adopt-config`.

## GPU and SR-IOV settings

`plan` maps the secondary networks of each chart workload to `spec.topology.distributed.<role>.network.sriov`. From the `k8s.v1.cni.cncf.io/networks` annotation it takes the network (`attachmentRef`) and the interface, and a second entry becomes the `gds` network. It finds the device resource of each network in the `k8s.v1.cni.cncf.io/resourceName` annotation of the NetworkAttachmentDefinition, or, when the pod has one network and one device, takes that device. `rdma/*` resources become `rdmaResourceName`. The GPU count, resource name, node selector and tolerations become `gpu`. KubeDB then generates its own advertise-IP init container, which writes the address of `net1` for Milvus.

`plan` does not guess. If a network cannot be tied to one device, or the network is in another namespace, it prints what it found and leaves `network` out; set it by hand, as in the [GPU and SR-IOV](/docs/guides/milvus/gpu-sriov/guide.md) guide. KubeDB expects `network.sriov` on all five roles when it is on one, so check that `plan` mapped every role that has a secondary network. The init container of the chart, the security context and any other pod setting are not carried over.

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
| `spec.tls` | Only if the cluster uses client or internal TLS; see [Keep the TLS of the cluster](#keep-the-tls-of-the-cluster) | KubeDB pods serve plain text and clients that use TLS fail |
| `spec.healthChecker.disableWriteCheck` | `true` | The health checker creates a `health_check_test` collection and writes to it every period |
| `spec.autoOps.disabled` | `true`, if every change must go through your own change control | The ops manager can propose version updates and credential rotations for the database |
| `spec.deletionPolicy` | `DoNotTerminate` | Deleting the object would delete the workloads |
| `kubedb.com/milvus-orphan-on-delete` annotation | `"true"` | Without it there is no clean way to hand the cluster back. See [Hand the cluster back](#hand-the-cluster-back) |

Also check:

- The StreamingNode PVC. KubeDB names it `data-<name>-streamingnode-<ordinal>`. With Woodpecker on object storage it holds no data that must survive, but confirm that before the handover. If it does, pre-create PVCs with those names from the existing volumes.
- The pod user. KubeDB images run as UID 1000 (`spec.restrictedSCC: true` lets OpenShift assign it). Volumes written as root by the old pods must be writable by the new user.
- No `MilvusArchiver` selects the namespace. The archiver needs a KubeDB-managed etcd and is rejected for an externally managed one.
- GPU and SR-IOV settings, if used: `plan` maps them (see [GPU and SR-IOV settings](#gpu-and-sr-iov-settings)); check the result against [GPU and SR-IOV](/docs/guides/milvus/gpu-sriov/guide.md). Remove any advertise-IP init container of your own; KubeDB adds its own.

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
- Remove leftover pods in phase `Failed` or `Succeeded` (for example `Error`, `Completed` or `ContainerStatusUnknown` pods from an earlier run) that match the chart's selectors before `run` or `rollback`. The command waits for "at most 0 pods" of a role and otherwise runs into its 15-minute timeout.
- If you rehearse with a second, temporary Milvus object, give it its own copy of the auth Secret. KubeDB records on an auth Secret which Milvus uses it, and a second Milvus on the same Secret is not supported: it can stay `Provisioning` without an event.

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
   $ MV_UID=$(kubectl get milvuses.kubedb.com -n demo milvus -o jsonpath='{.metadata.uid}')
   $ kubectl get secret -n demo milvus-${MV_UID: -6} -o jsonpath='{.data.milvus\.yaml}' | base64 -d > rendered.yaml
   $ diff <(yq -P 'sort_keys(..)' current-milvus.yaml) <(yq -P 'sort_keys(..)' rendered.yaml)
   ```

   Every difference in `etcd`, `minio`, `common.security` and the role sections must be explained before you continue.

3. Overlap: for each role, in the order Proxy, MixCoord, DataNode, StreamingNode, QueryNode, start the KubeDB pods, wait until they are Ready, then stop the old ones. Move QueryNodes one at a time (KubeDB to 1, wait Ready and let it serve a while, old to N-1, and so on), so at most one extra QueryNode runs:

   ```bash
   $ kubectl patch milvuses.kubedb.com -n demo milvus --type=merge -p '{"spec":{"topology":{"distributed":{"mixcoord":{"replicas":2}}}}}'
   # the pods appear a few seconds after the patch; kubectl wait fails if none exist yet
   $ until kubectl get pod -n demo -l app.kubernetes.io/instance=milvus,kubedb.com/role=mixcoord -o name | grep -q .; do sleep 2; done
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
$ kubectl get milvuses.kubedb.com -n demo milvus
$ kubectl get milvuses.kubedb.com -n demo milvus -o jsonpath='{range .status.conditions[*]}{.type}={.status}{"\n"}{end}'
```

Then, from a client, check the collection count against the number taken before the handover, run known queries with known results, and run a smoke test from each application.

## Roll back

The data never moved, so a rollback only changes who runs the pods. `kubectl dba adopt milvus rollback` does it; by hand:

1. Make sure the annotation `kubedb.com/milvus-orphan-on-delete: "true"` is set, set `spec.deletionPolicy` to `Halt`, and delete the Milvus. Its pods keep serving under their PetSets, and the operator no longer moves the client Service.
2. For each role, in the order QueryNode, StreamingNode, DataNode, MixCoord, Proxy: scale the old workload back up, wait until its pods are Ready, then scale the KubeDB PetSet to 0.
3. Before the proxy PetSet goes to 0, point the client Service back at the old proxies: its original selector, and ports whose `targetPort` the old proxy pods define (KubeDB's ports target the port names `proxy`, `metrics` and `http` of its own pods). Replace the selector instead of merging into it, because a merge patch keeps the KubeDB keys and the Service then selects no old proxy:

   ```bash
   $ kubectl patch svc -n demo milvus --type=json -p '[{"op":"replace","path":"/spec/selector","value":{"app.kubernetes.io/instance":"RELEASE","app.kubernetes.io/name":"milvus","component":"proxy"}}]'
   ```

   Use the selector of your chart's proxy pods in place of the example value.

## Hand the cluster back

With the annotation `kubedb.com/milvus-orphan-on-delete: "true"` set, deleting the Milvus no longer deletes the cluster. Before it lets the object go, the operator removes the Milvus owner reference from the PetSets, Services, config Secret, ServiceAccount, Role and RoleBinding (and from a KubeDB-managed meta etcd, if one is used), and keeps PVCs and Secrets as with `deletionPolicy: Halt`, whatever the policy says. The pods keep serving under the PetSet controller.

```bash
$ kubectl patch milvuses.kubedb.com -n demo milvus --type=merge -p '{"spec":{"deletionPolicy":"Halt"}}'
$ kubectl delete milvus -n demo milvus
$ kubectl get petset,pods,svc -n demo -l app.kubernetes.io/instance=milvus
```

`DoNotTerminate` still blocks deletion, which is why the policy is changed first. Keep the PetSet controller installed while the pods are run by PetSets. To move the cluster back to a chart, start its workloads role by role while scaling the PetSets down, in the same order as the handover.
