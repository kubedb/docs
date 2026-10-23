---
title: Install KubeDB and dr-controlplane for DC-DR
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-install
    name: Install
    parent: guides-postgres-distributed
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Install KubeDB and dr-controlplane for DC-DR

This page installs, on the hub:

- **KubeDB** with the PetSet OCM integration and the DC-DR orchestrator turned on;
- the **PetSet, PlacementPolicy and FailoverGroup CRDs** at the version DC-DR needs;
- **dr-controlplane**, whose agents are delivered to every cluster as an OCM add-on.

The hub's own three-site etcd stores the failover Leases, so the hub is also the *coordination plane*: neither KubeDB nor dr-controlplane needs a separate coordination cluster.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/distributed/install/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/install/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Before You Begin

- Complete [Build the stretched hub](/docs/guides/postgres/distributed/stretched-hub/index.md) and [Set up the spoke clusters](/docs/guides/postgres/distributed/spoke-clusters/index.md). The `kubedb` namespace must already be on the slice, and `slice.local` DNS must work in every cluster.
- This guide was tested with these releases:

  | Component | Version |
  |---|---|
  | KubeDB | `v2026.10.12` (provisioner `v0.68.0`, `pg-coordinator` `v0.52.0`) |
  | PetSet CRDs | `v0.2.0` |
  | dr-controlplane | `v0.1.0` |

- Each spoke needs a default StorageClass for the database volumes. k3s ships `local-path`, which this guide uses. You can install another provisioner, such as [Longhorn](https://longhorn.io/), and make it the default instead.
- You need a KubeDB Enterprise license issued for the **hub** cluster. Get the hub's cluster ID:

  ```bash
  $ kubectl --kubeconfig $HUB get ns kube-system -o jsonpath='{.metadata.uid}'
  ```

  Request a license for that ID from the [AppsCode License Server](https://appscode.com/issue-license?p=kubedb) and save it, for example at `/path/to/license.txt`. Only the hub needs the license: KubeDB runs only on the hub.

Use the kubeconfig variables from the previous pages:

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
```

## Step 1: Install KubeDB on the hub

The DC-DR settings are in [kubedb-values.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/install/yamls/kubedb-values.yaml):

```yaml
petset:
  features:
    ocm:
      enabled: true      # place DC-DR pods on the spokes through OCM ManifestWorks
kubedb-provisioner:
  dcDR:
    enabled: true        # run the DC-DR failover orchestrator
    localDC: hub         # the hub's own OCM ManagedCluster name
    # coordKubeconfigSecret is left unset: the hub itself stores the
    # primary-DC Leases, so the operator reads them in-cluster.
```

- `petset.features.ocm.enabled` makes the PetSet controller place pods on the spokes through ManifestWorks, according to the PlacementPolicy.
- `kubedb-provisioner.dcDR.enabled` starts the DC-DR orchestrator in the provisioner (`--dc-dr-enabled`). It also grants the provisioner the extra RBAC it needs to read ManagedClusters and the Leases.
- `kubedb-provisioner.dcDR.localDC` is the name of the cluster the operator runs in, `hub`.
- Leave `kubedb-provisioner.dcDR.coordKubeconfigSecret` unset on a stretched hub.

Install the chart:

```bash
$ helm --kubeconfig $HUB upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
    --version v2026.10.12 \
    --namespace kubedb --create-namespace \
    --values https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/install/yamls/kubedb-values.yaml \
    --set-file global.license=/path/to/license.txt \
    --wait --timeout 15m --burst-limit=10000
```

Every Helm install or upgrade regenerates the webhook serving certificates. A webhook pod that restarts at the same moment can load the previous certificate, and the PetSet webhook then rejects every PetSet with `x509: certificate signed by unknown authority`. After every install or upgrade, restart the two webhook-serving Deployments:

```bash
for d in kubedb-petset kubedb-kubedb-webhook-server; do
  kubectl --kubeconfig $HUB -n kubedb rollout restart deploy/$d
  kubectl --kubeconfig $HUB -n kubedb rollout status deploy/$d --timeout=300s
done
```

## Step 2: Apply the PetSet v0.2.0 CRDs

The `v2026.10.12` KubeDB chart runs the PetSet controller `v0.2.0`, but the PetSet chart it bundles (`v2026.9.18`) can install an older PlacementPolicy CRD (without `failoverPolicy.failoverGroupRef`) and no FailoverGroup CRD. Apply the PetSet `v0.2.0` CRDs over them. Helm never updates CRDs on upgrade, so also run this once on an existing KubeDB installation that you enable DC-DR on:

```bash
$ kubectl --kubeconfig $HUB apply --server-side --force-conflicts \
    -f https://raw.githubusercontent.com/kubeops/petset/v0.2.0/crds/apps.k8s.appscode.com_petsets.yaml \
    -f https://raw.githubusercontent.com/kubeops/petset/v0.2.0/crds/apps.k8s.appscode.com_placementpolicies.yaml \
    -f https://raw.githubusercontent.com/kubeops/petset/v0.2.0/crds/apps.k8s.appscode.com_failovergroups.yaml

$ kubectl --kubeconfig $HUB get crd failovergroups.apps.k8s.appscode.com placementpolicies.apps.k8s.appscode.com petsets.apps.k8s.appscode.com
NAME                                      SCOPE        VERSIONS      CREATED AT
failovergroups.apps.k8s.appscode.com      Cluster      v1(storage)   2026-10-13T04:22:09Z
placementpolicies.apps.k8s.appscode.com   Cluster      v1(storage)   2026-10-13T03:51:56Z
petsets.apps.k8s.appscode.com             Namespaced   v1(storage)   2026-10-13T03:51:55Z
```

## Step 3: Verify KubeDB

```bash
$ kubectl --kubeconfig $HUB -n kubedb get pods
NAME                                            READY   STATUS    RESTARTS   AGE
kubedb-kubedb-autoscaler-0                      2/2     Running   0          7d17h
kubedb-kubedb-ops-manager-0                     2/2     Running   0          16h
kubedb-kubedb-provisioner-0                     2/2     Running   0          7d15h
kubedb-kubedb-webhook-server-6bd8f6bb57-vclc7   2/2     Running   0          16h
kubedb-petset-5f4b9ff9c4-86jlf                  2/2     Running   0          16h
kubedb-sidekick-55b79b87c7-jmlsf                2/2     Running   0          16h
```

Each pod has two containers: the second one is the KubeSlice sidecar, injected because `kubedb` is a slice namespace.

The provisioner must run with the DC-DR flags and log that the orchestrator started:

```bash
$ kubectl --kubeconfig $HUB -n kubedb get sts kubedb-kubedb-provisioner \
    -o jsonpath='{.spec.template.spec.containers[?(@.name=="operator")].args}' | tr ',' '\n' | grep dc-dr
"--dc-dr-enabled"
"--dc-dr-local-dc=hub"]

$ kubectl --kubeconfig $HUB -n kubedb logs kubedb-kubedb-provisioner-0 -c operator | grep 'DC-DR orchestrator started'
```

The last command must print one line, with `localDC="hub"`.

> **Troubleshooting:** If the provisioner exits with `unknown flag: --dc-dr-enabled`, the node runs a cached provisioner image without the DC-DR orchestrator. Remove the cached image on that node (`sudo crictl rmi ghcr.io/kubedb/kubedb-provisioner:v0.68.0`), then delete the pod `kubedb/kubedb-kubedb-provisioner-0` so that the image is pulled again.

Check that the provisioner has the RBAC that DC-DR needs. Every line must end in `yes`; a missing permission freezes the DR status:

```bash
SA=system:serviceaccount:kubedb:kubedb-kubedb-provisioner
for r in postgresopsrequests.ops.kubedb.com managedclusters.cluster.open-cluster-management.io \
         leases.coordination.k8s.io manifestworks.work.open-cluster-management.io \
         placementpolicies.apps.k8s.appscode.com; do
  echo "$r list: $(kubectl --kubeconfig $HUB auth can-i list $r --as=$SA -n demo)"
done
```

### Optional: KubeStash

DC-DR does not need KubeStash, and backup and restore of a DC-DR database is not covered by this guide. To install KubeStash anyway, the same Enterprise license covers it:

```bash
helm --kubeconfig $HUB upgrade -i kubestash oci://ghcr.io/appscode-charts/kubestash \
  --version v2026.10.12 --namespace stash --create-namespace \
  --set-file global.license=/path/to/license.txt \
  --wait --timeout 15m --burst-limit=10000
```

## Step 4: Install dr-controlplane

dr-controlplane runs in *hub mode*: no separate etcd or API server is deployed (`etcd.deploy=false`, `controlplane.enabled=false`). On the hub it runs:

- the **topology controller**, which creates one `primary-dc*` Lease per failover scope from the PlacementPolicies, and the FailoverGroup controller;
- the **addon manager**, which installs an **agent** on every ManagedCluster labelled `dr-controlplane.io/enabled=true`.

An agent must run in **every** data center of the PlacementPolicy, the Arbiter (`hub`) included. KubeDB counts a data center as healthy only while its `dc-health-<dc>` Lease is fresh; without an agent on the hub, the Arbiter data center reads unhealthy and the database stays `Degraded`.

### Create the namespace

```bash
for k in $HUB $S1 $S2; do
  kubectl --kubeconfig $k create ns dc-failover --dry-run=client -o yaml | kubectl --kubeconfig $k apply -f -
done
```

### Create the agents' hub credential

The agents need a hub credential that may manage Leases in `dc-failover` and nothing else. [dr-agent-rbac.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/install/yamls/dr-agent-rbac.yaml) creates a ServiceAccount, a Role and RoleBinding limited to Leases, and a token Secret for it:

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/install/yamls/dr-agent-rbac.yaml
serviceaccount/dr-controlplane-agent created
role.rbac.authorization.k8s.io/dr-controlplane-agent created
rolebinding.rbac.authorization.k8s.io/dr-controlplane-agent created
secret/dr-controlplane-agent-token created
```

Build a kubeconfig from that token and store it in the Secret `dc-failover/hub-agent-kubeconfig` on the hub. Its server is the in-cluster hub proxy, which the agents in every cluster can reach:

```bash
TOKEN=$(kubectl --kubeconfig $HUB -n dc-failover get secret dr-controlplane-agent-token -o jsonpath='{.data.token}' | base64 -d)
CA=$(kubectl --kubeconfig $HUB config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')
cat > /tmp/hub-agent-kubeconfig <<EOF
apiVersion: v1
kind: Config
clusters:
- name: hub
  cluster:
    server: https://hub-api.hub-api-proxy.svc:6443
    certificate-authority-data: $CA
users:
- name: dr-controlplane-agent
  user:
    token: $TOKEN
contexts:
- name: hub
  context:
    cluster: hub
    user: dr-controlplane-agent
    namespace: dc-failover
current-context: hub
EOF
kubectl --kubeconfig $HUB -n dc-failover create secret generic hub-agent-kubeconfig \
  --from-file=kubeconfig=/tmp/hub-agent-kubeconfig
rm /tmp/hub-agent-kubeconfig
```

The addon manager copies this credential to every agent's cluster as the Secret `dc-failover/coord-kubeconfig`, with the server rewritten to `addon.agent.coordExternalEndpoint`.

### Select the clusters that get an agent

```bash
$ kubectl --kubeconfig $HUB label managedcluster hub spoke-1 spoke-2 dr-controlplane.io/enabled=true --overwrite
managedcluster.cluster.open-cluster-management.io/hub labeled
managedcluster.cluster.open-cluster-management.io/spoke-1 labeled
managedcluster.cluster.open-cluster-management.io/spoke-2 labeled
```

### Install the chart

The dr-controlplane chart is published in the AppsCode Helm repository. [dr-controlplane-values.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/install/yamls/dr-controlplane-values.yaml) configures hub mode:

```yaml
namespace: dc-failover
createNamespace: false
etcd:
  deploy: false
controlplane:
  enabled: false
agent:
  enabled: false          # agents arrive on every cluster as an OCM add-on instead
topology:
  enabled: true
  requireSpread: false
addon:
  autoInstall:
    enabled: true
  coordKubeconfigMirrorNamespaces: []
  placement:
    labelSelector:
      matchLabels:
        dr-controlplane.io/enabled: "true"
  agent:
    installNamespace: dc-failover
    createNamespace: false
    coordKubeconfigSourceSecret: hub-agent-kubeconfig
    coordExternalEndpoint: https://hub-api.hub-api-proxy.svc:6443
```

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update appscode
$ helm --kubeconfig $HUB upgrade -i dr appscode/dr-controlplane --version v0.1.0 \
    --namespace dc-failover \
    --values https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/install/yamls/dr-controlplane-values.yaml \
    --wait --timeout 10m
```

The agents use the chart's default timings: health Lease 15 s renewed every 5 s, and primary-DC Lease duration 45 s, renew deadline 40 s, retry period 5 s.

### Verify dr-controlplane

On the hub, the topology controller, the addon manager and the hub's own agent run:

```bash
$ kubectl --kubeconfig $HUB -n dc-failover get pods
NAME                                             READY   STATUS    RESTARTS        AGE
dr-controlplane-addon-manager-864dd7cd47-mhs2c   1/1     Running   0               8d
dr-controlplane-agent-hub-65dd56884f-vc8gh       1/1     Running   21 (6d1h ago)   9d
dr-controlplane-topology-56444f996f-smxml        1/1     Running   0               9d
```

Each spoke runs its agent:

```bash
$ kubectl --kubeconfig $S1 -n dc-failover get deploy
NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
dr-controlplane-agent-spoke-1   1/1     1            1           9d
```

The add-on is `Available` on all three clusters. Availability comes from ManifestWork status feedback, so it can take about a minute after the install:

```bash
$ kubectl --kubeconfig $HUB get managedclusteraddon -A | grep -E 'NAME|dr-controlplane'
NAMESPACE   NAME              AVAILABLE   DEGRADED   PROGRESSING
hub         dr-controlplane   True                   False
spoke-1     dr-controlplane   True                   False
spoke-2     dr-controlplane   True                   False
```

Every data center renews its health Lease, each renewed within the last 15 s:

```bash
$ kubectl --kubeconfig $HUB -n dc-failover get lease -o custom-columns=NAME:.metadata.name,HOLDER:.spec.holderIdentity,RENEWED:.spec.renewTime
NAME                  HOLDER                                                                            RENEWED
dc-health-hub         hub                                                                               2026-10-23T04:01:02.520552Z
dc-health-spoke-1     spoke-1                                                                           2026-10-23T04:00:59.679736Z
dc-health-spoke-2     spoke-2                                                                           2026-10-23T04:00:58.775953Z
dr-agent-writer-hub   dr-controlplane-agent-hub-65dd56884f-vc8gh_85c84f50-b127-479e-a15d-9b18de52eafd   2026-10-23T04:01:01.390473Z
primary-dc            spoke-2                                                                           2026-10-23T04:00:58.661546Z
primary-dc-pg-docs    spoke-2                                                                           2026-10-23T04:00:59.620976Z
```

This output is from an installation that already runs databases. On a fresh installation you see only the `dc-health-*` Leases and the agent's own `dr-agent-writer-*` election Lease. The topology controller creates a `primary-dc` or `primary-dc-<group>` Lease when the first PlacementPolicy with a `failoverPolicy` for that scope appears, on the [next page](/docs/guides/postgres/distributed/provision/index.md).

> **Important:** Do not edit or delete the Leases in `dc-failover` by hand while databases use them. Use the Postgres annotations described in [Switchover and failover](/docs/guides/postgres/distributed/switchover-failover/index.md).

## Next Steps

- [Provision a DC-DR Postgres](/docs/guides/postgres/distributed/provision/index.md).
