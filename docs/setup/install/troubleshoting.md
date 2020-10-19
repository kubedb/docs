---
title: Troubleshooting KubeDB Installation
description: Troubleshooting guide for KubeDB installation
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-troubleshoot
    name: Troubleshooting
    parent: installation-guide
    weight: 40
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

## Installing in GKE Cluster

If you are installing KubeDB on a GKE cluster, you will need cluster admin permissions to install KubeDB operator. Run the following command to grant admin permision to the cluster.

```bash
$ kubectl create clusterrolebinding "cluster-admin-$(whoami)" \
  --clusterrole=cluster-admin                                 \
  --user="$(gcloud config get-value core/account)"
```

In addition, if your GKE cluster is a [private cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/private-clusters), you will need to either add an additional firewall rule that allows master nodes access port `8443/tcp` on worker nodes, or change the existing rule that allows access to ports `443/tcp` and `10250/tcp` to also allow access to port `8443/tcp`. The procedure to add or modify firewall rules is described in the official GKE documentation for private clusters mentioned before.

## Configuring Network Volume Accessor

For network volume such as NFS, KubeDB needs to deploy a helper deployment in the same namespace as the Repository that uses the NFS as backend to provide Snapshot listing facility. We call this helper deployment network volume accessor. You can configure its resources, user id, privileged permission etc. during installation as below,

```bash
$ helm install kubedb-enterprise appscode/kubedb-enterprise \
    -n kube-system                                        \
    --set netVolAccessor.cpu=200m                         \
    --set netVolAccessor.memory=128Mi                     \
    --set netVolAccessor.runAsUser=0                      \
    --set netVolAccessor.privileged=true
```

## Detect KubeDB version

To detect KubeDB version, exec into the operator pod and run `kubedb version` command.

```bash
$ POD_NAMESPACE=kube-system
$ POD_NAME=$(kubectl get pods -n $POD_NAMESPACE -l app.kubernetes.io/name=kubedb -o jsonpath={.items[0].metadata.name})
$ kubectl exec $POD_NAME -c operator -n $POD_NAMESPACE -- /kubedb version

Version = {{< param "info.version" >}}
VersionStrategy = tag
Os = alpine
Arch = amd64
CommitHash = 85b0f16ab1b915633e968aac0ee23f877808ef49
GitBranch = release-0.5
GitTag = {{< param "info.version" >}}
CommitTimestamp = 2020-08-10T05:24:23

$ kubectl exec -it $POD_NAME -c operator -n $POD_NAMESPACE restic version
restic 0.9.6
compiled with go1.9 on linux/amd64
```
