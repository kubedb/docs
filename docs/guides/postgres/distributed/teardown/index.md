---
title: Tear Down a DC-DR Installation
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-teardown
    name: Teardown
    parent: guides-postgres-distributed
    weight: 80
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Tear Down a DC-DR Installation

This page removes, from smallest to largest: one DC-DR database, the hub add-ons, and finally the clusters.

## Before You Begin

- This page assumes the installation built in this guide, from [Build the stretched hub](/docs/guides/postgres/distributed/stretched-hub/index.md) to [Provision a DC-DR Postgres](/docs/guides/postgres/distributed/provision/index.md), with the database `pg-docs` in the namespace `demo`.
- Stop the applications that use the database, and take a backup of any data you want to keep.

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
```

## Delete a DC-DR database

What happens to the data depends on `spec.deletionPolicy`:

- `DoNotTerminate` blocks the deletion;
- `Halt` keeps the PVCs and the Secrets;
- `Delete` removes the PVCs and keeps the Secrets;
- `WipeOut` removes everything, the Secrets included.

See [spec.deletionPolicy](/docs/guides/postgres/concepts/postgres.md#specdeletionpolicy) for details.

To keep the data, change the policy first:

```bash
kubectl --kubeconfig $HUB -n demo patch pg pg-docs --type=merge -p '{"spec":{"deletionPolicy":"Halt"}}'
```

Delete the database, and wait until its per data center PetSets are gone **before** you delete its PlacementPolicy. The deletion waits on the per data center ManifestWorks, and the operator reads the PlacementPolicy while it tears them down; deleting the policy first can leave the deletion stuck.

```bash
$ kubectl --kubeconfig $HUB -n demo delete pg pg-docs
postgres.kubedb.com "pg-docs" deleted

$ kubectl --kubeconfig $HUB -n demo get petset | grep pg-docs     # wait until this prints nothing

$ kubectl --kubeconfig $HUB delete placementpolicy pg-docs
placementpolicy.apps.k8s.appscode.com "pg-docs" deleted

$ kubectl --kubeconfig $HUB delete failovergroup pg-docs
failovergroup.apps.k8s.appscode.com "pg-docs" deleted
```

The per data center PlacementPolicies (`pg-docs-spoke-1`, `pg-docs-spoke-2`) are removed with the database. With `WipeOut`, the pods, PVCs and volumes on both spokes are gone within about a minute.

### Remove the group's Lease and markers

Deleting a database does not remove its failover group's Lease and marker ConfigMaps. dr-controlplane `v0.1.0` never deletes them: the agent that holds the `primary-dc-<group>` Lease keeps renewing it, and every agent keeps refreshing its marker ConfigMap of the same name. A group created later with the same name would start from the old holder. They are:

| Object | Cluster | Namespace |
|---|---|---|
| Lease `primary-dc-<group>` | hub | `dc-failover` |
| ConfigMap `primary-dc-<group>` (marker) | every cluster with an agent: `hub`, `spoke-1`, `spoke-2` | `dc-failover` |
| ConfigMaps `primary-dc-<group>-standby-hold` and `primary-dc-<group>-override`, if you created them | the spoke you created them on | `dc-failover` |

Once no PlacementPolicy references the group any more, check that, then delete the Lease first and the markers after it. The agents stop refreshing a marker once its Lease is gone:

```bash
$ G=pg-docs
$ kubectl --kubeconfig $HUB get placementpolicy -o jsonpath='{range .items[*]}{.metadata.name} {.spec.clusterSpreadConstraint.failoverPolicy.failoverGroupRef.name}{"\n"}{end}' | awk -v g="$G" '$2 == g'
$ kubectl --kubeconfig $HUB -n dc-failover delete lease primary-dc-$G
lease.coordination.k8s.io "primary-dc-pg-docs" deleted
$ sleep 10
$ for k in $HUB $S1 $S2; do
    kubectl --kubeconfig $k -n dc-failover delete cm primary-dc-$G primary-dc-$G-standby-hold primary-dc-$G-override --ignore-not-found
  done
configmap "primary-dc-pg-docs" deleted
configmap "primary-dc-pg-docs" deleted
configmap "primary-dc-pg-docs" deleted
```

The `get placementpolicy` command must print nothing. For databases without a failover group, the shared scope is `primary-dc`; remove the Lease `primary-dc` and the ConfigMaps `primary-dc`, `primary-dc-standby-hold` and `primary-dc-override` the same way, but only after the last database without a group is gone.

> **Warning:** Never delete a Lease that a running database still follows: its markers go stale, and the database is fenced read-only.

## Remove the hub add-ons and keep the clusters

Delete every DC-DR database first, then uninstall in the reverse order of the installation. `helm uninstall` leaves some objects behind, so the commands after it are part of the procedure.

### dr-controlplane

```bash
helm --kubeconfig $HUB -n dc-failover uninstall dr --wait
kubectl --kubeconfig $HUB label managedcluster hub spoke-1 spoke-2 dr-controlplane.io/enabled-
```

The uninstall leaves the Leases on the hub, and the marker ConfigMaps and the agent credential on every cluster. A stale marker must never meet a new installation, so delete the namespace everywhere:

```bash
for k in $HUB $S1 $S2; do kubectl --kubeconfig $k delete ns dc-failover; done
```

### KubeDB and KubeStash

```bash
helm --kubeconfig $HUB -n kubedb uninstall kubedb --wait
```

If you installed KubeStash, uninstall it and remove the webhook configurations it leaves behind. They point at the deleted Service and make later API requests fail:

```bash
helm --kubeconfig $HUB -n stash uninstall kubestash --wait
kubectl --kubeconfig $HUB delete validatingwebhookconfiguration,mutatingwebhookconfiguration kubestash-kubestash-operator
kubectl --kubeconfig $HUB delete ns stash
```

To remove the CRDs too (this deletes every remaining object of those kinds):

```bash
kubectl --kubeconfig $HUB delete $(kubectl --kubeconfig $HUB get crd -o name | grep -E '\.(kubedb\.com|kubestash\.com|appscode\.com)$')
```

### KubeSlice and OCM

Follow the [KubeSlice documentation](https://kubeslice.io/documentation/open-source/) to offboard the namespaces, delete the SliceConfig, then the Cluster objects and the project, and finally uninstall the `kubeslice-controller` chart.

Then remove OCM: run `clusteradm unjoin --cluster-name <name>` against each managed cluster (`spoke-1`, `spoke-2` and `hub`), and `clusteradm clean` against the hub. See the [OCM documentation](https://open-cluster-management.io/).

## Return the VMs to plain Ubuntu

Remove the spokes first, then the hub Kubernetes, and etcd last: the hub API needs etcd until the end.

> **Warning:** Never run plain `kubeadm reset` on a hub VM. Its `remove-etcd-member` phase deletes the contents of `/var/lib/etcd`, kubeadm's default etcd directory, even for an etcd that kubeadm does not manage. Always pass `--skip-phases=remove-etcd-member`, and check the health of every etcd member after any recovery step.

### Spokes

On each spoke VM, run the uninstall script that the k3s installer created:

```bash
sudo /usr/local/bin/k3s-uninstall.sh
```

### Hub Kubernetes

On hub-b, then on hub-a:

```bash
sudo kubeadm reset -f --skip-phases=remove-etcd-member
sudo rm -rf /etc/cni/net.d
sudo apt-mark unhold kubelet kubeadm kubectl
sudo apt-get purge -y kubelet kubeadm kubectl cri-tools haproxy
sudo rm -f /etc/apt/sources.list.d/kubernetes.list /etc/apt/keyrings/kubernetes-apt-keyring.gpg
sudo sed -i -E '/[[:space:]](kube-hub-api|hub-a|hub-b|etcd-w)$/d' /etc/hosts
sudo rm -f /etc/sysctl.d/k8s.conf /etc/modules-load.d/k8s.conf
```

Remove containerd (`sudo apt-get purge -y containerd && sudo rm -rf /etc/containerd /var/lib/containerd`) only if nothing else on the VM uses it, and restore the swap line in `/etc/fstab` if the VM had swap.

### etcd

Only when the whole hub is going away: this destroys all hub state, including the failover Leases.

1. Take a snapshot from one member. `etcdctl snapshot save` needs exactly one endpoint, so override the multi-endpoint value from the environment file:

   ```bash
   sudo bash -c 'set -a; . /etc/etcd/etcdctl.env; ETCDCTL_ENDPOINTS=https://<this-vm-ip>:2379; set +a; etcdctl snapshot save /root/hub-etcd-$(date +%F).db'
   ```

2. On hub-a, hub-b and etcd-w:

   ```bash
   sudo systemctl disable --now etcd
   sudo rm -f /etc/systemd/system/etcd.service && sudo systemctl daemon-reload
   sudo rm -rf /etc/etcd /var/lib/etcd-hub /root/etcd-pki
   sudo rm -f /usr/local/bin/etcd /usr/local/bin/etcdctl /usr/local/bin/etcdutl
   ```

Keep the encryption-at-rest key (`encryption-config.yaml`) for as long as you might restore the etcd snapshot: without it the Secrets in the snapshot cannot be decrypted. Then delete the workstation's `~/dcdr` directory, which holds the hub admin kubeconfig.

## Next Steps

- Start again from [Concepts](/docs/guides/postgres/distributed/concepts/index.md).
