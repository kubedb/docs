---
title: Set Up the Spoke Clusters
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-spoke-clusters
    name: Spoke Clusters
    parent: guides-postgres-distributed
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Set Up the Spoke Clusters

This page creates one Kubernetes cluster per data center, registers the spokes and the hub with [Open Cluster Management (OCM)](https://open-cluster-management.io/), and connects all three clusters with a [KubeSlice](https://kubeslice.io/) network slice.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/distributed/spoke-clusters/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Before You Begin

- Build the hub as described in [Build the stretched hub](/docs/guides/postgres/distributed/stretched-hub/index.md), and keep its kubeconfig at `~/dcdr/hub.kubeconfig`.
- Install [`clusteradm`](https://open-cluster-management.io/docs/getting-started/installation/start-the-control-plane/) `v1.3.1` and `helm` v3 on the workstation:

  ```bash
  curl -fsSL https://github.com/open-cluster-management-io/clusteradm/releases/download/v1.3.1/clusteradm_linux_amd64.tar.gz \
    | tar -xz -C /usr/local/bin clusteradm
  ```

- Decide the data center names now. The spoke names become the OCM ManagedCluster names, which are the data center names everywhere in DC-DR. This guide uses `spoke-1` (DC A), `spoke-2` (DC B), and `hub` for the hub, which is the Arbiter data center.

Set these variables on the workstation:

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
export HUB_A_IP=10.10.1.11 SPOKE_1_IP=10.10.1.21 SPOKE_2_IP=10.10.2.21
```

## Step 1: Create the spoke clusters

Any conformant Kubernetes cluster can be a spoke. This guide uses a single-node [k3s](https://k3s.io/) cluster on each spoke VM, with flannel and the `local-path` storage class (both k3s defaults). On **spoke-1** (and likewise on **spoke-2** with its own name and IP):

```bash
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=v1.37.1+k3s1 sh -s - server \
  --disable traefik --disable metrics-server \
  --node-name spoke-1 --tls-san $SPOKE_1_IP
```

Copy each spoke's `/etc/rancher/k3s/k3s.yaml` to the workstation, replacing `127.0.0.1` with the spoke's IP:

```bash
ssh <user>@$SPOKE_1_IP sudo cat /etc/rancher/k3s/k3s.yaml | sed "s#https://127.0.0.1:6443#https://$SPOKE_1_IP:6443#" > $S1
ssh <user>@$SPOKE_2_IP sudo cat /etc/rancher/k3s/k3s.yaml | sed "s#https://127.0.0.1:6443#https://$SPOKE_2_IP:6443#" > $S2
chmod 600 $S1 $S2
```

```bash
$ kubectl --kubeconfig $S1 get nodes -o wide
NAME      STATUS   ROLES           AGE   VERSION        INTERNAL-IP   EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION              CONTAINER-RUNTIME
spoke-1   Ready    control-plane   17d   v1.37.1+k3s1   10.10.1.21    <none>        Ubuntu 24.04.1 LTS   6.8.0-147-generic (amd64)   containerd://2.3.4-k3s1

$ kubectl --kubeconfig $S2 get nodes -o wide
NAME      STATUS   ROLES           AGE   VERSION        INTERNAL-IP   EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION              CONTAINER-RUNTIME
spoke-2   Ready    control-plane   17d   v1.37.1+k3s1   10.10.2.21    <none>        Ubuntu 24.04.1 LTS   6.8.0-146-generic (amd64)   containerd://2.3.4-k3s1
```

## Step 2: Give every cluster a path to the hub API

The KubeSlice workers and the dr-controlplane agents in each cluster talk to the hub API. So that they do not depend on a single hub API server, every cluster (the hub included) runs a small HAProxy Deployment that balances over all hub API servers with an HTTPS `/healthz` check. It is reached in-cluster as `https://hub-api.hub-api-proxy.svc:6443`, a name the hub API server certificate already carries (see [kubeadm init](/docs/guides/postgres/distributed/stretched-hub/index.md#kubeadm-init)).

Download [hub-api-proxy.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/hub-api-proxy.yaml), change its two `server` lines to your hub API servers, and apply it to all three clusters:

```bash
$ curl -fsSLO https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/hub-api-proxy.yaml
$ # edit the "server hub-a ..." and "server hub-b ..." lines
$ for k in $HUB $S1 $S2; do kubectl --kubeconfig $k apply -f hub-api-proxy.yaml; done

$ for k in $HUB $S1 $S2; do kubectl --kubeconfig $k -n hub-api-proxy get deploy,svc; done
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/haproxy   2/2     2            2           17d

NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/hub-api   ClusterIP   10.97.17.102   <none>        6443/TCP   17d
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/haproxy   2/2     2            2           17d

NAME              TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/hub-api   ClusterIP   10.43.55.26   <none>        6443/TCP   17d
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/haproxy   2/2     2            2           17d

NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/hub-api   ClusterIP   10.43.84.225   <none>        6443/TCP   17d
```

Check from inside each cluster that the hub answers through the proxy and that its certificate verifies against the hub CA:

```bash
kubectl --kubeconfig $HUB config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d > ~/dcdr/hub-ca.crt
for k in $HUB $S1 $S2; do
  kubectl --kubeconfig $k -n hub-api-proxy create configmap hub-ca --from-file=ca.crt=$HOME/dcdr/hub-ca.crt \
    --dry-run=client -o yaml | kubectl --kubeconfig $k apply -f -
  kubectl --kubeconfig $k -n hub-api-proxy run probe --restart=Never --image=curlimages/curl:8.10.1 \
    --overrides='{"spec":{"volumes":[{"name":"ca","configMap":{"name":"hub-ca"}}],"containers":[{"name":"probe","image":"curlimages/curl:8.10.1","args":["-s","--max-time","10","--cacert","/ca/ca.crt","https://hub-api.hub-api-proxy.svc:6443/healthz"],"volumeMounts":[{"name":"ca","mountPath":"/ca"}]}]}}'
  kubectl --kubeconfig $k -n hub-api-proxy wait --for=jsonpath='{.status.phase}'=Succeeded pod/probe --timeout=90s
  kubectl --kubeconfig $k -n hub-api-proxy logs probe; echo
  kubectl --kubeconfig $k -n hub-api-proxy delete pod probe
done
```

Each cluster must print `ok`.

## Step 3: Register the clusters with OCM

### Initialize the OCM hub

```bash
$ KUBECONFIG=$HUB clusteradm init --wait --feature-gates=ManifestWorkReplicaSet=true

$ kubectl --kubeconfig $HUB -n open-cluster-management-hub get deploy
NAME                                       READY   UP-TO-DATE   AVAILABLE   AGE
cluster-manager-addon-manager-controller   1/1     1            1           17d
cluster-manager-addon-webhook              1/1     1            1           17d
cluster-manager-placement-controller       1/1     1            1           17d
cluster-manager-registration-controller    1/1     1            1           17d
cluster-manager-registration-webhook       1/1     1            1           17d
cluster-manager-work-controller            1/1     1            1           17d
cluster-manager-work-webhook               1/1     1            1           17d
```

### Join the spokes and the hub itself

Every cluster joins the OCM hub, **the hub included**: the hub is registered as the ManagedCluster `hub`, which is the Arbiter data center of every DC-DR database, and it needs the dr-controlplane agent like any other data center.

`clusteradm join` checks the hub API address from where it runs, so it must be an address the workstation can reach. This guide uses hub-a's IP. In production use a VIP, a load balancer or a DNS name that is in the hub API server certificate, so that the klusterlets survive the loss of hub-a.

```bash
TOKEN=$(KUBECONFIG=$HUB clusteradm get token | sed -n 's/^token=//p')
for c in spoke-1:$S1 spoke-2:$S2 hub:$HUB; do
  KUBECONFIG=${c#*:} clusteradm join --hub-token $TOKEN --hub-apiserver https://$HUB_A_IP:6443 \
    --ca-file ~/dcdr/hub-ca.crt --cluster-name ${c%%:*} \
    --feature-gates=RawFeedbackJsonString=true --wait
done
```

Accept the three clusters on the hub. It can take a few retries until all CSRs are present:

```bash
KUBECONFIG=$HUB clusteradm accept --clusters spoke-1,spoke-2,hub
```

KubeDB reads pod status back from the spokes through ManifestWork status feedback. Enable raw JSON feedback in every klusterlet's work agent, and raise its API rate limits:

```bash
PATCH='{"spec":{"workConfiguration":{"featureGates":[{"feature":"RawFeedbackJsonString","mode":"Enable"}],"hubKubeAPIBurst":100,"hubKubeAPIQPS":50,"kubeAPIBurst":100,"kubeAPIQPS":50}}}'
for k in $S1 $S2 $HUB; do kubectl --kubeconfig $k patch klusterlet klusterlet --type=merge -p "$PATCH"; done
```

Verify:

```bash
$ kubectl --kubeconfig $HUB get managedclusters
NAME      HUB ACCEPTED   MANAGED CLUSTER URLS        JOINED   AVAILABLE   AGE
hub       true           https://kube-hub-api:8443   True     True        17d
spoke-1   true                                       True     True        17d
spoke-2   true                                       True     True        17d
```

All three must be `JOINED` and `AVAILABLE`. OCM creates a namespace on the hub for each ManagedCluster (`hub`, `spoke-1`, `spoke-2`); the ManifestWorks for that cluster live there.

## Step 4: Connect the clusters with KubeSlice

### Install the KubeSlice controller on the hub

The controller endpoint is the in-cluster hub proxy from Step 2, which every KubeSlice worker uses to reach the controller. `ocm.enabled=true` lets OCM deliver the KubeSlice worker to each cluster as the `kubeslice` add-on.

```bash
helm --kubeconfig $HUB upgrade -i kubeslice-controller oci://ghcr.io/appscode-charts/kubeslice-controller \
  --version v2026.1.15 -n kubeslice-controller --create-namespace \
  --set kubeslice.controller.endpoint=https://hub-api.hub-api-proxy.svc:6443 \
  --set ocm.enabled=true \
  --wait --timeout 10m
```

### Create a project

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/project.yaml
project.controller.kubeslice.io/demo-distributed-postgres created
```

KubeSlice creates the project namespace `kubeslice-demo-distributed-postgres` on the hub.

### Register every cluster

For each cluster in the slice (`hub`, `spoke-1`, `spoke-2`):

1. Label the node(s) that run the slice gateway. On the hub, label both hub nodes:

   ```bash
   kubectl --kubeconfig $HUB label node hub-a hub-b kubeslice.io/node-type=gateway --overwrite
   kubectl --kubeconfig $S1 label node spoke-1 kubeslice.io/node-type=gateway --overwrite
   kubectl --kubeconfig $S2 label node spoke-2 kubeslice.io/node-type=gateway --overwrite
   ```

2. Create a KubeSlice `Cluster` and a `kubeslice` ManagedClusterAddOn for it on the hub. [kubeslice-cluster.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/kubeslice-cluster.yaml) shows them for `spoke-1`; create the same pair for `spoke-2` and `hub`, changing the names and the ManagedClusterAddOn namespace. Set `networkInterface` to the NIC of the gateway node's default route (`ip route get 8.8.8.8` prints it after `dev`):

   ```yaml
   apiVersion: controller.kubeslice.io/v1alpha1
   kind: Cluster
   metadata:
     name: spoke-1
     namespace: kubeslice-demo-distributed-postgres
   spec:
     networkInterface: enp1s0
     clusterProperty: {}
   ---
   apiVersion: addon.open-cluster-management.io/v1alpha1
   kind: ManagedClusterAddOn
   metadata:
     name: kubeslice
     namespace: spoke-1
   spec:
     installNamespace: kubeslice-system
     configs:
     - group: controller.kubeslice.io
       resource: clusters
       name: spoke-1
       namespace: kubeslice-demo-distributed-postgres
   ```

3. Create the application namespaces in every cluster. The database namespace (`demo`) and the KubeDB namespace (`kubedb`) both join the slice: the KubeDB operator on the hub connects to each primary at `<db>.<namespace>.svc.slice.local`.

   ```bash
   for k in $HUB $S1 $S2; do
     for ns in demo kubedb; do kubectl --kubeconfig $k create ns $ns --dry-run=client -o yaml | kubectl --kubeconfig $k apply -f -; done
   done
   ```

Wait until the KubeSlice worker (`kubeslice-operator`) is running in every cluster:

```bash
for k in $HUB $S1 $S2; do kubectl --kubeconfig $k -n kubeslice-system rollout status deploy/kubeslice-operator --timeout=600s; done
```

Check that each worker registered under its own name:

```bash
for k in $HUB $S1 $S2; do
  kubectl --kubeconfig $k -n kubeslice-system get deploy kubeslice-operator \
    -o jsonpath='{.spec.template.spec.containers[?(@.name=="manager")].env[?(@.name=="CLUSTER_NAME")].value}{"\n"}'
done
```

> **Troubleshooting:** On the first installation, the add-on's first render can carry an empty cluster name. The corrected render then never applies, because a Deployment selector is immutable, and the worker cannot register. If a cluster prints an empty name above, delete that cluster's `kubeslice-system/kubeslice-operator` Deployment and touch the add-on's ManifestWork on the hub so that it is re-applied right away, for example `kubectl --kubeconfig $HUB -n spoke-1 annotate manifestwork addon-kubeslice-deploy-0 resync="$(date +%s)" --overwrite`.

Every KubeSlice `Cluster` must report its CNI subnets before you create the slice; the SliceConfig webhook rejects the slice otherwise:

```bash
$ kubectl --kubeconfig $HUB -n kubeslice-demo-distributed-postgres get clusters.controller.kubeslice.io \
    -o custom-columns=NAME:.metadata.name,REGISTRATION:.status.registrationStatus,HEALTH:.status.clusterHealth.clusterHealthStatus,NETWORK:.status.networkPresent
NAME      REGISTRATION   HEALTH   NETWORK
hub       Registered     Normal   true
spoke-1   Registered     Normal   true
spoke-2   Registered     Normal   true

```

Also check that `status.cniSubnet` is set on each of them (`kubectl get clusters.controller.kubeslice.io <name> -o yaml`).

### Create the slice

[sliceconfig.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/sliceconfig.yaml) creates the slice `demo-slice` over all three clusters with a WireGuard gateway, and onboards the `demo` and `kubedb` namespaces:

```yaml
apiVersion: controller.kubeslice.io/v1alpha1
kind: SliceConfig
metadata:
  name: demo-slice
  namespace: kubeslice-demo-distributed-postgres
spec:
  sliceSubnet: 10.1.0.0/16
  maxClusters: 16
  sliceType: Application
  sliceGatewayProvider:
    sliceGatewayType: Wireguard
    sliceCaType: Local
  sliceIpamType: Local
  rotationInterval: 60
  vpnConfig:
    cipher: AES-128-CBC
  clusters:
  - hub
  - spoke-1
  - spoke-2
  qosProfileDetails:
    queueType: HTB
    priority: 1
    tcType: BANDWIDTH_CONTROL
    bandwidthCeilingKbps: 5120
    bandwidthGuaranteedKbps: 2560
    dscpClass: AF11
  namespaceIsolationProfile:
    applicationNamespaces:
    - namespace: demo
      clusters:
      - '*'
    - namespace: kubedb
      clusters:
      - '*'
    isolationEnabled: false
    allowedNamespaces:
    - namespace: kube-system
      clusters:
      - '*'
```

`sliceSubnet` must not overlap your VM networks, pod CIDRs or service CIDRs. Every namespace whose pods must reach a database over the slice, including your application namespaces, must be listed under `applicationNamespaces`. Add yours now, before you deploy those applications.

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/sliceconfig.yaml
sliceconfig.controller.kubeslice.io/demo-slice created
```

### Forward `slice.local` DNS in every cluster

Pods resolve exported services as `<name>.<namespace>.svc.slice.local` through the cluster's CoreDNS, which must forward the `slice.local` zone to the KubeSlice DNS service of that cluster. Get the DNS service IP in each cluster:

```bash
kubectl --kubeconfig <cluster-kubeconfig> -n kubeslice-system get svc -l app=kubeslice-dns -o jsonpath='{.items[0].spec.clusterIP}'
```

The block to add, with that cluster's IP:

```
slice.local:53 {
    errors
    cache 30
    loadbalance
    forward . <kubeslice-dns-cluster-ip>
}
```

- **On the hub (kubeadm CoreDNS):** add the block at the top of the `Corefile` key of the `kube-system/coredns` ConfigMap (`kubectl --kubeconfig $HUB -n kube-system edit cm coredns`).
- **On a k3s spoke:** k3s rewrites its CoreDNS ConfigMap on restart, so put the block in the `coredns-custom` ConfigMap, which k3s CoreDNS imports:

  ```bash
  DNSIP=$(kubectl --kubeconfig $S1 -n kubeslice-system get svc -l app=kubeslice-dns -o jsonpath='{.items[0].spec.clusterIP}')
  kubectl --kubeconfig $S1 -n kube-system create cm coredns-custom \
    --from-literal=slice.local.server="slice.local:53 {
      errors
      cache 30
      loadbalance
      forward . $DNSIP
  }" --dry-run=client -o yaml | kubectl --kubeconfig $S1 apply -f -
  ```

  Repeat for spoke-2 with `$S2`.

Then restart CoreDNS in every cluster:

```bash
for k in $HUB $S1 $S2; do kubectl --kubeconfig $k -n kube-system rollout restart deploy/coredns; done
```

> **Important:** Configure `slice.local` DNS in every cluster before you install KubeDB. The Postgres pods find each other, and the hub operator finds each primary, through `slice.local` names.

### Verify the slice

```bash
$ kubectl --kubeconfig $HUB -n kubeslice-demo-distributed-postgres get sliceconfig,workersliceconfigs,workerslicegateways
NAME                                             AGE
sliceconfig.controller.kubeslice.io/demo-slice   17d

NAME                                                       AGE
workersliceconfig.worker.kubeslice.io/demo-slice-hub       17d
workersliceconfig.worker.kubeslice.io/demo-slice-spoke-1   17d
workersliceconfig.worker.kubeslice.io/demo-slice-spoke-2   17d

NAME                                                                AGE
workerslicegateway.worker.kubeslice.io/demo-slice-hub-spoke-1       17d
workerslicegateway.worker.kubeslice.io/demo-slice-hub-spoke-2       17d
workerslicegateway.worker.kubeslice.io/demo-slice-spoke-1-hub       17d
workerslicegateway.worker.kubeslice.io/demo-slice-spoke-1-spoke-2   17d
workerslicegateway.worker.kubeslice.io/demo-slice-spoke-2-hub       17d
workerslicegateway.worker.kubeslice.io/demo-slice-spoke-2-spoke-1   17d
```

There is a WorkerSliceGateway for every ordered pair of clusters. In each cluster, the slice router and the gateway pods must be running:

```bash
$ kubectl --kubeconfig $S1 -n kubeslice-system get pods
NAME                                              READY   STATUS      RESTARTS        AGE
demo-slice-spoke-1-hub-0-0-56c99df48-qgm7x        3/3     Running     15 (7d ago)     17d
demo-slice-spoke-1-hub-1-1-867548b5d8-7njqc       3/3     Running     3 (7d ago)      7d17h
demo-slice-spoke-1-spoke-2-0-0-c87b8597d-zf5pr    3/3     Running     15 (7d ago)     17d
demo-slice-spoke-1-spoke-2-1-0-7fcf449fc7-5xj4z   3/3     Running     15 (7d ago)     17d
forwarder-kernel-crmcm                            1/1     Running     5 (7d ago)      17d
kubeslice-dns-765c9ff9f9-4qqbb                    1/1     Running     5 (7d ago)      17d
kubeslice-install-crds-9dvbp                      0/1     Completed   0               17d
kubeslice-netop-vvh64                             1/1     Running     5 (7d ago)      17d
kubeslice-operator-5f78f9dd9-jgvjr                2/2     Running     19 (6d1h ago)   17d
kubeslice-postdelete-job-5bfk9                    0/1     Completed   0               17d
nsc-grpc-server-h9d8n                             1/1     Running     5 (7d ago)      17d
nsm-install-crds-bhxm8                            0/1     Completed   0               17d
nsmgr-fg64j                                       2/2     Running     10 (7d ago)     17d
registry-k8s-949cb847f-22xxj                      1/1     Running     5 (7d ago)      17d
vl3-slice-router-demo-slice-676fdf8dd8-8qgw2      2/2     Running     10 (7d ago)     17d
```

Finally, prove that traffic flows. Export an echo server from spoke-2 with [slice-check.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/slice-check.yaml), and call it by its `slice.local` name from the hub and from spoke-1. A pod in a slice namespace reaches the slice only after the injected KubeSlice sidecar has brought up its `nsm0` interface, so use a long-running client pod and wait for `nsm0` before calling:

```bash
kubectl --kubeconfig $S2 apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/slice-check.yaml
for k in $HUB $S1; do
  kubectl --kubeconfig $k -n demo run slice-client --image=busybox:1.36 --restart=Never -- sleep 900
  kubectl --kubeconfig $k -n demo wait --for=condition=Ready pod/slice-client --timeout=180s
  until kubectl --kubeconfig $k -n demo exec slice-client -c slice-client -- ip -4 addr show nsm0 >/dev/null 2>&1; do sleep 5; done
  kubectl --kubeconfig $k -n demo exec slice-client -c slice-client -- wget -qO- -T 5 http://slice-check.demo.svc.slice.local:8080/
  kubectl --kubeconfig $k -n demo delete pod slice-client
done
```

Both calls must print `hello-from-spoke-2`. Then remove the test objects:

```bash
kubectl --kubeconfig $S2 delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/spoke-clusters/yamls/slice-check.yaml
```

## Next Steps

- [Install KubeDB and dr-controlplane](/docs/guides/postgres/distributed/install/index.md).
