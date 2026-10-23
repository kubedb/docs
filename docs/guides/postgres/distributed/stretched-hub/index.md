---
title: Build the Stretched Hub
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-stretched-hub
    name: Stretched Hub
    parent: guides-postgres-distributed
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Build the Stretched Hub

The hub is one Kubernetes cluster whose control plane spans two data centers, backed by an external etcd cluster with one member in each of three sites. It holds the failover Leases, so it must survive the loss of any single site. Read [Concepts](/docs/guides/postgres/distributed/concepts/index.md) first.

This page builds the hub with `kubeadm` from plain Ubuntu VMs. If you already run a Kubernetes cluster whose etcd has members in three sites, you can use it as the hub and continue with [Set up the spoke clusters](/docs/guides/postgres/distributed/spoke-clusters/index.md).

> **Note:** YAML and configuration files used in this tutorial are stored in [docs/guides/postgres/distributed/stretched-hub/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/stretched-hub/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Before You Begin

### Machines

This guide uses five Ubuntu 24.04 (amd64) VMs with internet access. The names and addresses below are the example values used throughout this section; replace them with yours.

| VM | Site | Role | Example IP | Size used for this guide |
|---|---|---|---|---|
| `hub-a` | DC A | hub control plane, worker and etcd member | `10.2.1.161` | 16 vCPU, 32 GiB |
| `hub-b` | DC B | hub control plane, worker and etcd member | `10.2.1.162` | 16 vCPU, 32 GiB |
| `etcd-w` | DC W | etcd member only, never a Kubernetes node | `10.2.0.239` | 16 vCPU, 16 GiB |
| `spoke-1` | DC A | database cluster (next page) | `10.2.0.233` | 16 vCPU, 24 GiB |
| `spoke-2` | DC B | database cluster (next page) | `10.2.0.236` | 16 vCPU, 24 GiB |

These sizes are the tested configuration, not measured minimums. For more resilience, add hub VMs: keep the etcd member count odd and spread over the three sites.

Every VM needs:

- passwordless `sudo` for your SSH user;
- an NTP-synchronized clock (`timedatectl show -p NTPSynchronized --value` prints `yes`); etcd and TLS depend on it;
- cgroup v2 (`stat -fc %T /sys/fs/cgroup` prints `cgroup2fs`);
- TCP ports `2379`, `2380`, `6443` and `10250` reachable from every other VM;
- open traffic between the spoke VMs and the hub VMs for the KubeSlice gateways (the tested VMs had no firewall between them).

The pod CIDR, the service CIDR and the KubeSlice slice subnet must not overlap each other or your VM networks. This guide uses `10.244.0.0/16`, `10.96.0.0/12` and `10.1.0.0/16`.

### etcd timing

The etcd settings below (`--heartbeat-interval=100`, `--election-timeout=1000`, in milliseconds) assume less than 10 ms round trip between the sites. For higher latency, set the heartbeat interval to about the round-trip time and the election timeout to ten times the heartbeat.

### Workstation

Run the `kubectl`, `helm` and `clusteradm` commands of this section from one workstation that reaches all VMs. It needs `kubectl` 1.37, `helm` v3, `openssl` and `ssh`/`scp`.

### Common variables

Run this on every VM and on the workstation before the commands below, with your values:

```bash
export HUB_A_IP=10.2.1.161 HUB_B_IP=10.2.1.162 ETCD_W_IP=10.2.0.239
export K8S_VERSION=1.37.1 K8S_APT_CHANNEL=v1.37 ETCD_VERSION=v3.7.2
export POD_CIDR=10.244.0.0/16 SERVICE_CIDR=10.96.0.0/12
```

## Step 1: Prepare the hub VMs

On **hub-a** and **hub-b**, load the kernel modules, turn swap off, and install containerd and the Kubernetes packages (from the [kubeadm installation guide](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)):

```bash
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF
sudo modprobe overlay
sudo modprobe br_netfilter
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sudo sysctl --system

sudo swapoff -a
sudo sed -ri '/\sswap\s/s/^#?/#/' /etc/fstab

sudo apt-get update
sudo apt-get install -y containerd apt-transport-https ca-certificates curl gpg
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml >/dev/null
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
sudo systemctl enable --now containerd
sudo systemctl restart containerd

sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/$K8S_APT_CHANNEL/deb/Release.key \
  | sudo gpg --dearmor --yes -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/$K8S_APT_CHANNEL/deb/ /" \
  | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
PKGVER=$(apt-cache madison kubeadm | awk -v v="$K8S_VERSION" '$3 ~ "^"v"-" {print $3}' | head -1)
sudo apt-get install -y kubelet=$PKGVER kubeadm=$PKGVER kubectl=$PKGVER cri-tools
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl enable kubelet
```

## Step 2: Create the etcd certificates

The etcd cluster gets its own CA, in the layout `kubeadm` uses. Create the CA and the certificates for all three members and for the API servers on **hub-a**, using the `kubeadm` certificate phases (as in [Set up a High Availability etcd cluster with kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/setup-ha-etcd-with-kubeadm/)):

```bash
sudo -E bash <<'EOF'
D=/root/etcd-pki
mkdir -p $D && chmod 700 $D
[ -f $D/ca/etcd/ca.key ] || kubeadm init phase certs etcd-ca --cert-dir $D/ca

for e in hub-a=$HUB_A_IP hub-b=$HUB_B_IP etcd-w=$ETCD_W_IP; do
  n=${e%%=*}; ip=${e#*=}
  rm -rf $D/$n && mkdir -p $D/$n/etcd
  cp $D/ca/etcd/ca.crt $D/ca/etcd/ca.key $D/$n/etcd/
  cat >$D/$n.yaml <<CFG
apiVersion: kubeadm.k8s.io/v1beta4
kind: InitConfiguration
nodeRegistration:
  name: $n
localAPIEndpoint:
  advertiseAddress: $ip
---
apiVersion: kubeadm.k8s.io/v1beta4
kind: ClusterConfiguration
certificatesDir: $D/$n
etcd:
  local:
    serverCertSANs: ["$ip", "$n"]
    peerCertSANs: ["$ip", "$n"]
CFG
  for c in etcd-server etcd-peer etcd-healthcheck-client; do
    kubeadm init phase certs $c --config $D/$n.yaml
  done
  rm $D/$n/etcd/ca.key
  tar -C $D/$n -czf $D/bundle-$n.tgz etcd
done

rm -rf $D/apiserver && mkdir -p $D/apiserver/etcd
cp $D/ca/etcd/ca.crt $D/ca/etcd/ca.key $D/apiserver/etcd/
cat >$D/apiserver.yaml <<CFG
apiVersion: kubeadm.k8s.io/v1beta4
kind: ClusterConfiguration
certificatesDir: $D/apiserver
CFG
kubeadm init phase certs apiserver-etcd-client --config $D/apiserver.yaml
rm $D/apiserver/etcd/ca.key
tar -C $D/apiserver -czf $D/bundle-apiserver.tgz apiserver-etcd-client.crt apiserver-etcd-client.key etcd/ca.crt
ls -1 $D/bundle-*.tgz
EOF
```

Copy each member's bundle to that VM as `~/etcd-bundle.tgz` (hub-a's own included), and keep `bundle-apiserver.tgz` on hub-a. Back up `/root/etcd-pki/ca`: you need the CA key to add or replace an etcd member later.

## Step 3: Run etcd on the three sites

On **hub-a**, **hub-b** and **etcd-w**, install etcd as a systemd service. Set `NODE` and `IP` to this VM's name and address:

```bash
export NODE=hub-a IP=$HUB_A_IP     # hub-b: NODE=hub-b IP=$HUB_B_IP; etcd-w: NODE=etcd-w IP=$ETCD_W_IP

tmp=$(mktemp -d)
curl -fsSL https://github.com/etcd-io/etcd/releases/download/$ETCD_VERSION/etcd-$ETCD_VERSION-linux-amd64.tar.gz \
  | tar -xz -C $tmp --strip-components=1
sudo install -m 755 $tmp/etcd $tmp/etcdctl $tmp/etcdutl /usr/local/bin/

sudo mkdir -p /etc/etcd/pki /var/lib/etcd-hub
sudo tar -C /etc/etcd/pki --strip-components=1 -xzf ~/etcd-bundle.tgz
sudo chmod 600 /etc/etcd/pki/*.key
sudo chmod 700 /var/lib/etcd-hub

cat <<EOF | sudo tee /etc/systemd/system/etcd.service
[Unit]
Description=etcd (DC-DR hub)
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/etcd \\
  --name=$NODE \\
  --data-dir=/var/lib/etcd-hub \\
  --listen-client-urls=https://127.0.0.1:2379,https://$IP:2379 \\
  --advertise-client-urls=https://$IP:2379 \\
  --listen-peer-urls=https://$IP:2380 \\
  --initial-advertise-peer-urls=https://$IP:2380 \\
  --initial-cluster=hub-a=https://$HUB_A_IP:2380,hub-b=https://$HUB_B_IP:2380,etcd-w=https://$ETCD_W_IP:2380 \\
  --initial-cluster-state=new \\
  --initial-cluster-token=dcdr-hub-etcd \\
  --listen-metrics-urls=http://127.0.0.1:2381 \\
  --client-cert-auth=true \\
  --trusted-ca-file=/etc/etcd/pki/ca.crt \\
  --cert-file=/etc/etcd/pki/server.crt \\
  --key-file=/etc/etcd/pki/server.key \\
  --peer-client-cert-auth=true \\
  --peer-trusted-ca-file=/etc/etcd/pki/ca.crt \\
  --peer-cert-file=/etc/etcd/pki/peer.crt \\
  --peer-key-file=/etc/etcd/pki/peer.key \\
  --heartbeat-interval=100 \\
  --election-timeout=1000 \\
  --pre-vote=true
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

cat <<EOF | sudo tee /etc/etcd/etcdctl.env
ETCDCTL_ENDPOINTS=https://$HUB_A_IP:2379,https://$HUB_B_IP:2379,https://$ETCD_W_IP:2379
ETCDCTL_CACERT=/etc/etcd/pki/ca.crt
ETCDCTL_CERT=/etc/etcd/pki/healthcheck-client.crt
ETCDCTL_KEY=/etc/etcd/pki/healthcheck-client.key
EOF

sudo systemctl daemon-reload
sudo systemctl enable etcd
```

The etcd members protect the three-site quorum with two Raft mechanisms, Pre-Vote and CheckQuorum:

- **Pre-Vote** (`--pre-vote=true`). Before a member disrupts the cluster with a real election, it runs a Pre-Vote phase: it asks the other members whether they would vote for it, without raising the Raft term. A member that still sees the leader refuses. So a member returning from a partition, for example the member of an isolated site, cannot depose a healthy leader when it reconnects. Pre-Vote has been the default since etcd 3.4. The unit sets it explicitly so it cannot be turned off by accident.
- **CheckQuorum** (always on in the etcd server; there is no flag). A leader that cannot hear from a majority of members within an election timeout steps down, so a leader stranded on the minority side of a partition stops acting as leader. While a member has heard from the leader within the election timeout, its leader lease is still valid, and it ignores vote and pre-vote requests.

Together they keep the etcd quorum, and with it the failover Leases stored in it, stable through partitions.

> **Important:** The data directory is `/var/lib/etcd-hub`, not kubeadm's default `/var/lib/etcd`. A plain `kubeadm reset` deletes the contents of `/var/lib/etcd` even for an etcd that kubeadm does not manage. Never run plain `kubeadm reset` on a hub VM; see [Teardown](/docs/guides/postgres/distributed/teardown/index.md).

A member waits for quorum before it reports ready, so start all three within about a minute of each other. On each of the three VMs:

```bash
sudo systemctl start --no-block etcd
```

Then check the cluster from any member:

```bash
$ sudo bash -c 'set -a; . /etc/etcd/etcdctl.env; set +a; etcdctl member list -w table; etcdctl endpoint health -w table'
┌──────────────────┬─────────┬────────┬─────────────────────────┬─────────────────────────┬────────────┐
│        ID        │ STATUS  │  NAME  │       PEER ADDRS        │      CLIENT ADDRS       │ IS LEARNER │
├──────────────────┼─────────┼────────┼─────────────────────────┼─────────────────────────┼────────────┤
│ 47d89c878be9b774 │ started │  hub-a │ https://10.2.1.161:2380 │ https://10.2.1.161:2379 │      false │
│ 50a8c7f04bb4b032 │ started │  hub-b │ https://10.2.1.162:2380 │ https://10.2.1.162:2379 │      false │
│ 81ab7f316ba60954 │ started │ etcd-w │ https://10.2.0.239:2380 │ https://10.2.0.239:2379 │      false │
└──────────────────┴─────────┴────────┴─────────────────────────┴─────────────────────────┴────────────┘
┌─────────────────────────┬────────┬─────────────┬───────┐
│        ENDPOINT         │ HEALTH │    TOOK     │ ERROR │
├─────────────────────────┼────────┼─────────────┼───────┤
│ https://10.2.0.239:2379 │   true │  8.189269ms │       │
│ https://10.2.1.162:2379 │   true │  9.534072ms │       │
│ https://10.2.1.161:2379 │   true │ 11.744404ms │       │
└─────────────────────────┴────────┴─────────────┴───────┘
```

All three members must be `started` and healthy. Judge health by the exit code of `etcdctl endpoint health`. `member list` alone is not enough, because the list is served by the other members; also check `systemctl is-active etcd` on each VM.

## Step 4: Load-balance the hub API on the hub VMs

Each hub VM runs HAProxy on `127.0.0.1:8443`, balanced over every hub API server with an HTTPS `/healthz` check (the [kubeadm HA guide](https://github.com/kubernetes/kubeadm/blob/main/docs/ha-considerations.md) layout). The name `kube-hub-api` resolves to `127.0.0.1` on each hub VM, so the kubelet and the control plane survive the loss of either API server.

On **hub-a** and **hub-b**:

```bash
sudo apt-get install -y haproxy
cat <<EOF | sudo tee /etc/haproxy/haproxy.cfg
global
    log /dev/log local0
    daemon
    maxconn 4000

defaults
    mode                    tcp
    log                     global
    option                  tcplog
    timeout connect         5s
    timeout client          1h
    timeout server          1h

frontend hub-api
    bind 127.0.0.1:8443
    default_backend hub-apiservers

backend hub-apiservers
    option httpchk
    http-check connect ssl
    http-check send meth GET uri /healthz
    http-check expect status 200
    balance roundrobin
    server hub-a $HUB_A_IP:6443 check verify none inter 3s fall 3 rise 2
    server hub-b $HUB_B_IP:6443 check verify none inter 3s fall 3 rise 2
EOF
sudo haproxy -c -f /etc/haproxy/haproxy.cfg
sudo systemctl enable haproxy
sudo systemctl restart haproxy

echo "$HUB_A_IP hub-a"       | sudo tee -a /etc/hosts
echo "$HUB_B_IP hub-b"       | sudo tee -a /etc/hosts
echo "$ETCD_W_IP etcd-w"     | sudo tee -a /etc/hosts
echo "127.0.0.1 kube-hub-api" | sudo tee -a /etc/hosts
```

## Step 5: Initialize the hub on hub-a

### Encryption at rest

Secrets in the hub's etcd are encrypted with an `aescbc` key. The key lives only on hub-a and hub-b, never on etcd-w, so the third site never sees plaintext Secrets. On **hub-a**:

```bash
umask 077
cat > ~/encryption-config.yaml <<EOF
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
      - secrets
    providers:
      - aescbc:
          keys:
            - name: key1
              secret: $(head -c 32 /dev/urandom | base64)
      - identity: {}
EOF
sudo mkdir -p /etc/kubernetes/pki /etc/kubernetes/enc
sudo install -m 600 ~/encryption-config.yaml /etc/kubernetes/enc/encryption-config.yaml
sudo tar -C /etc/kubernetes/pki -xzf /root/etcd-pki/bundle-apiserver.tgz
```

Keep a safe copy of `encryption-config.yaml`. Without it, Secrets in an etcd backup cannot be decrypted.

### kubeadm init

The API server certificate must carry the in-cluster name `hub-api.hub-api-proxy.svc`. The spokes, the KubeSlice workers and the dr-controlplane agents reach the hub through an HAProxy Service of that name (set up on the [next page](/docs/guides/postgres/distributed/spoke-clusters/index.md#step-2-give-every-cluster-a-path-to-the-hub-api)). If you plan to put a VIP, a load balancer or a DNS name in front of the hub, add it to `certSANs` now.

On **hub-a**, write the kubeadm configuration (an example rendered for this guide's addresses is in [yamls/kubeadm-config.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/stretched-hub/yamls/kubeadm-config.yaml)) and run `kubeadm init`:

```bash
cat <<EOF | sudo tee /root/kubeadm-init.yaml
apiVersion: kubeadm.k8s.io/v1beta4
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: $HUB_A_IP
  bindPort: 6443
nodeRegistration:
  name: hub-a
  criSocket: unix:///var/run/containerd/containerd.sock
  taints: []
---
apiVersion: kubeadm.k8s.io/v1beta4
kind: ClusterConfiguration
kubernetesVersion: v$K8S_VERSION
controlPlaneEndpoint: kube-hub-api:8443
networking:
  podSubnet: $POD_CIDR
  serviceSubnet: $SERVICE_CIDR
etcd:
  external:
    endpoints:
    - https://$HUB_A_IP:2379
    - https://$HUB_B_IP:2379
    - https://$ETCD_W_IP:2379
    caFile: /etc/kubernetes/pki/etcd/ca.crt
    certFile: /etc/kubernetes/pki/apiserver-etcd-client.crt
    keyFile: /etc/kubernetes/pki/apiserver-etcd-client.key
apiServer:
  certSANs:
  - kube-hub-api
  - 127.0.0.1
  - $HUB_A_IP
  - hub-a
  - $HUB_B_IP
  - hub-b
  - hub-api.hub-api-proxy.svc
  - hub-api.hub-api-proxy.svc.cluster.local
  extraArgs:
  - name: encryption-provider-config
    value: /etc/kubernetes/enc/encryption-config.yaml
  extraVolumes:
  - name: encryption-config
    hostPath: /etc/kubernetes/enc
    mountPath: /etc/kubernetes/enc
    readOnly: true
    pathType: DirectoryOrCreate
EOF
sudo kubeadm init --config /root/kubeadm-init.yaml --upload-certs
```

`taints: []` keeps the control-plane nodes schedulable: the hub VMs also run the hub workloads.

Create the join material for hub-b. Pass `--config` to `upload-certs`: without it kubeadm does not know that etcd is external and leaves the etcd client certificate out of the uploaded Secret, and the join fails.

```bash
export KUBECONFIG=/etc/kubernetes/admin.conf
sudo -E kubeadm token create --print-join-command
sudo -E kubeadm init phase upload-certs --upload-certs --config /root/kubeadm-init.yaml
```

Note the endpoint, the token and the `--discovery-token-ca-cert-hash` from the first command, and the certificate key printed on the last line of the second. The certificate key expires after two hours.

## Step 6: Join hub-b as a second control plane

Copy `encryption-config.yaml` from hub-a to **hub-b**, then on hub-b:

```bash
sudo mkdir -p /etc/kubernetes/enc
sudo install -m 600 ~/encryption-config.yaml /etc/kubernetes/enc/encryption-config.yaml

cat <<EOF | sudo tee /root/kubeadm-join.yaml
apiVersion: kubeadm.k8s.io/v1beta4
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: kube-hub-api:8443
    token: <token>
    caCertHashes:
    - <sha256:ca-cert-hash>
nodeRegistration:
  name: hub-b
  criSocket: unix:///var/run/containerd/containerd.sock
  taints: []
controlPlane:
  localAPIEndpoint:
    advertiseAddress: $HUB_B_IP
    bindPort: 6443
  certificateKey: <certificate-key>
EOF
sudo kubeadm join --config /root/kubeadm-join.yaml
```

## Step 7: Install the pod network

On **hub-a**, install [flannel](https://github.com/flannel-io/flannel) with your pod CIDR:

```bash
curl -fsSL https://github.com/flannel-io/flannel/releases/download/v0.28.9/kube-flannel.yml \
  | sed "s#10.244.0.0/16#$POD_CIDR#" \
  | sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f -
sudo kubectl --kubeconfig /etc/kubernetes/admin.conf wait --for=condition=Ready node --all --timeout=300s
```

## Step 8: Get a kubeconfig for the workstation

Copy hub-a's `/etc/kubernetes/admin.conf` to the workstation as `~/dcdr/hub.kubeconfig`, and set its `server:` to `https://<hub-a-ip>:6443` (the name `kube-hub-api` resolves only on the hub VMs). Keep this file private.

```bash
mkdir -p ~/dcdr && chmod 700 ~/dcdr
ssh <user>@$HUB_A_IP sudo cat /etc/kubernetes/admin.conf \
  | sed -E "s#server: https://[^ ]+#server: https://$HUB_A_IP:6443#" > ~/dcdr/hub.kubeconfig
chmod 600 ~/dcdr/hub.kubeconfig
export HUB=~/dcdr/hub.kubeconfig
```

## Verify the hub

```bash
$ kubectl --kubeconfig $HUB get nodes -o wide
NAME    STATUS   ROLES           AGE   VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION              CONTAINER-RUNTIME
hub-a   Ready    control-plane   17d   v1.37.1   10.2.1.161    <none>        Ubuntu 24.04.1 LTS   6.8.0-142-generic (amd64)   containerd://2.2.1
hub-b   Ready    control-plane   17d   v1.37.1   10.2.1.162    <none>        Ubuntu 24.04.1 LTS   6.8.0-147-generic (amd64)   containerd://2.2.1

$ kubectl --kubeconfig $HUB describe nodes | grep -i taints
Taints:             <none>
Taints:             <none>

$ kubectl --kubeconfig $HUB -n kube-system get pods
NAME                            READY   STATUS    RESTARTS        AGE
coredns-7db744bcd9-gw72k        1/1     Running   0               8d
coredns-7db744bcd9-lpvk9        1/1     Running   0               8d
kube-apiserver-hub-a            1/1     Running   4 (7d23h ago)   17d
kube-apiserver-hub-b            1/1     Running   2 (7d17h ago)   17d
kube-controller-manager-hub-a   1/1     Running   23 (6d1h ago)   17d
kube-controller-manager-hub-b   1/1     Running   16 (6d1h ago)   17d
kube-proxy-l8tw9                1/1     Running   0               17d
kube-proxy-nr4zk                1/1     Running   2 (7d17h ago)   17d
kube-scheduler-hub-a            1/1     Running   22 (6d1h ago)   17d
kube-scheduler-hub-b            1/1     Running   17 (6d1h ago)   17d
```

Both nodes are `Ready` control planes with no taint, and every control-plane component runs twice.

Prove that Secrets are encrypted in etcd. Create a Secret, then read its raw value from **etcd-w**, the site that must never see plaintext:

```bash
kubectl --kubeconfig $HUB -n default create secret generic enc-check --from-literal=k=plaintext-marker
ssh <user>@$ETCD_W_IP "sudo bash -c 'set -a; . /etc/etcd/etcdctl.env; set +a; etcdctl get /registry/secrets/default/enc-check --print-value-only'" | head -c 40 | od -c | head -3
kubectl --kubeconfig $HUB -n default delete secret enc-check
```

The value must start with `k8s:enc:aescbc:v1:key1:`, and `plaintext-marker` must not appear in it.

## Next Steps

- [Set up the spoke clusters](/docs/guides/postgres/distributed/spoke-clusters/index.md).
