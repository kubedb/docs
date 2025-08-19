---
title: Restart Hazelcast
menu:
  docs_{{ .version }}:
    identifier: hz-restart-details
    name: Restart Hazelcast
    parent: hz-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Hazelcast

KubeDB supports restarting the Hazelcast database via a HazelcastOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Hazelcast

Before deploying hazelcast we need to create license secret since we are running enterprise version of hazelcast.


```bash
kubectl create secret generic hz-license-key -n demo --from-literal=licenseKey=TrialLicense#10Nodes#eyJhbGxvd2VkTmF0aXZlTWVtb3J5U2l6ZSI6MTAwLCJhbGxvd2VkTnVtYmVyT2ZOb2RlcyI6MTAsImFsbG93ZWRUaWVyZWRTdG9yZVNpemUiOjAsImFsbG93ZWRUcGNDb3JlcyI6MCwiY3JlYXRpb25EYXRlIjoxNzQ4ODQwNDc3LjYzOTQ0NzgxNiwiZXhwaXJ5RGF0ZSI6MTc1MTQxNDM5OS45OTk5OTk5OTksImZlYXR1cmVzIjpbMCwyLDMsNCw1LDYsNyw4LDEwLDExLDEzLDE0LDE1LDE3LDIxLDIyXSwiZ3JhY2VQZXJpb2QiOjAsImhhemVsY2FzdFZlcnNpb24iOjk5LCJvZW0iOmZhbHNlLCJ0cmlhbCI6dHJ1ZSwidmVyc2lvbiI6IlY3In0=.6PYD6i-hejrJ5Czgc3nYsmnwF7mAI-78E8LFEuYp-lnzXh_QLvvsYx4ECD0EimqcdeG2J5sqUI06okLD502mCA==
secret/hz-license-key created
```

In this section, we are going to deploy a Hazelcast database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hazelcast-quickstart
  namespace: demo
spec:
  deletionPolicy: WipeOut
  licenseSecret:
    name: hz-license-key
  replicas: 3
  version: 5.5.2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```


Let's create the `Hazelcast` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/restart/hazelcast.yaml
hazelcast.kubedb.com/hazelcast-quickstart created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-restart
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: hazelcast-quickstart
  type: Restart
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Hazelcast CR. It should be available in the same namespace as the opsRequest
- The meaning of `spec.apply` fields will be found [here](/docs/guides/hazelcast/concepts/hazelcastopsrequest.md#spectimeout)

> Note: The method of restarting the combined node is exactly same as above. All you need, is to specify the corresponding Hazelcast name in `spec.databaseRef.name` section.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/restart/ops.yaml
hazelcastopsrequest.ops.kubedb.com/hazelcast-restart created
```

Now the Ops-manager operator will restart the hazelcast members as per the request.

```shell
~/k/hazelcast $ kubectl get hzops -n demo
NAME                TYPE      STATUS       AGE
hazelcast-restart   Restart   Successful   4m

~/k/hazelcast $ kubectl describe Hazelcastopsrequest -n demo hazelcast-restart 
Name:         hazelcast-restart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T04:04:14Z
  Generation:          1
  Resource Version:    5414702
  UID:                 65bced3e-3155-4875-a1ac-5bab3348bc60
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hazelcast-quickstart
  Type:    Restart
Status:
  Conditions:
    Last Transition Time:  2025-08-19T04:04:14Z
    Message:               Hazelcast ops-request has started to restart hazelcast nodes
    Observed Generation:   1
    Reason:                Restart
    Status:                True
    Type:                  Restart
    Last Transition Time:  2025-08-19T04:07:07Z
    Message:               Successfully Restarted Hazelcast nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-19T04:04:27Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-quickstart-0
    Last Transition Time:  2025-08-19T04:04:27Z
    Message:               evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hazelcast-quickstart-0
    Last Transition Time:  2025-08-19T04:04:37Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-19T04:05:27Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-quickstart-1
    Last Transition Time:  2025-08-19T04:05:27Z
    Message:               evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hazelcast-quickstart-1
    Last Transition Time:  2025-08-19T04:06:17Z
    Message:               get pod; ConditionStatus:True; PodName:hazelcast-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hazelcast-quickstart-2
    Last Transition Time:  2025-08-19T04:06:17Z
    Message:               evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hazelcast-quickstart-2
    Last Transition Time:  2025-08-19T04:07:07Z
    Message:               Controller has successfully restart the Hazelcast replicas
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                           Age    From                         Message
  ----     ------                                                           ----   ----                         -------
  Normal   Starting                                                         5m23s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hazelcast-restart
  Normal   Starting                                                         5m23s  KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hazelcast-quickstart
  Normal   Successful                                                       5m23s  KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hazelcast-quickstart for HazelcastOpsRequest: hazelcast-restart
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0    5m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-0  5m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-0
  Warning  running pod; ConditionStatus:False                               5m     KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1    4m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-1  4m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-2    3m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hazelcast-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-2  3m20s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hazelcast-quickstart-2
  Normal   RestartNodes                                                     2m30s  KubeDB Ops-manager Operator  Successfully Restarted Hazelcast nodes
  Normal   Starting                                                         2m30s  KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hazelcast-quickstart
  Normal   Successful                                                       2m30s  KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hazelcast-quickstart for HazelcastOpsRequest: hazelcast-restart

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcastopsrequest -n demo restart
kubectl delete hazelcast -n demo hazelcast-qickstart
kubectl delete ns demo
```

## Next Steps
- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).

Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
