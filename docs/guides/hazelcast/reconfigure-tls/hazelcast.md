---
title: Reconfigure Hazelcast TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: hz-reconfigure-tls-hazelcast
    name: Reconfigure Hazelcast TLS/SSL Encryption
    parent: hz-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Hazelcast TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Hazelcast database via a HazelcastOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

Before deploying hazelcast we need to create license secret since we are running enterprise version of hazelcast.


```bash
kubectl create secret generic hz-license-key -n demo --from-literal=licenseKey='your hazelcast license key'
secret/hz-license-key created
```

## Add TLS to a Hazelcast database

Here, We are going to create a Hazelcast without TLS and then reconfigure the database to use TLS.

### Deploy Hazelcast without TLS

In this section, we are going to deploy a Hazelcast topology cluster without TLS. In the next few sections we will reconfigure TLS using `HazelcastOpsRequest` CRD. Below is the YAML of the `Hazelcast` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hz-prod
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/reconfigure-tls/hazelcast.yaml
hazelcast.kubedb.com/hz-prod created
```

Now, wait until `hz-prod` has status `Ready`. i.e,

```bash
$ kubectl get hz -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
hz-prod    kubedb.com/v1   5.2.2     Provisioning   0s
hz-prod    kubedb.com/v1   5.2.2     Provisioning   9s
.
.
hz-prod    kubedb.com/v1   5.2.2     Ready          2m10s
```

Now, we can exec one hazelcast pod and verify configuration that the TLS is disabled.

```bash
kubectl exec -n demo hz-prod-0 -- cat /data/hazelcast/hazelcast.yaml | grep -A 1 -i ssl

Defaulted container "hazelcast" out of: hazelcast, hazelcast-init (init)
```

We can verify from the above output that TLS is disabled for this cluster.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Hazelcast. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls hz-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/hz-ca created
```

Now, Let's create an `Issuer` using the `hz-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: hz-issuer
  namespace: demo
spec:
  ca:
    secretName: hz-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/reconfigure-tls/hazelcast-issuer.yaml
issuer.cert-manager.io/hz-issuer created
```

### Create HazelcastOpsRequest

In order to add TLS to the hazelcast, we have to create a `HazelcastOpsRequest` CRO with our created issuer. Below is the YAML of the `HazelcastOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hz-prod
  tls:
    issuerRef:
      name: hz-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - hazelcast
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `hz-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on hazelcast.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/hazelcast/concepts/hazelcast.md#spectls).

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/reconfigure-tls/hazelcast-add-tls.yaml
hazelcastopsrequest.ops.kubedb.com/hzops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `HazelcastOpsRequest` to be `Successful`.  Run the following command to watch `HazelcastOpsRequest` CRO,

```bash
$ kubectl get hazelcastopsrequest -n demo
NAME            TYPE             STATUS       AGE
hzops-add-tls   ReconfigureTLS   Successful   4m36s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe hazelcastopsrequest -n demo hzops-add-tls 
Name:         hzops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T05:39:00Z
  Generation:          1
  Resource Version:    5429257
  UID:                 0919d423-147f-4abb-b421-d8da43e65448
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   hz-prod
  Timeout:  5m
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          hazelcast
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       hz-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-08-19T05:39:00Z
    Message:               Hazelcast ops-request has started to reconfigure tls for Hazelcast nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-08-19T05:39:24Z
    Message:               Successfully synced TLS certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-08-19T05:39:14Z
    Message:               get certificate retries; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificateRetries
    Last Transition Time:  2025-08-19T05:39:14Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2025-08-19T05:39:14Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-08-19T05:39:34Z
    Message:               Successfully updated pet sets
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-19T05:42:04Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-19T05:39:44Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-0
    Last Transition Time:  2025-08-19T05:39:44Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-1
    Last Transition Time:  2025-08-19T05:39:44Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-2
    Last Transition Time:  2025-08-19T05:39:44Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-19T05:39:54Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-0
    Last Transition Time:  2025-08-19T05:41:14Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-1
    Last Transition Time:  2025-08-19T05:41:54Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-2
    Last Transition Time:  2025-08-19T05:42:04Z
    Message:               Successfully completed reconfigureTLS for Hazelcast.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                Age    From                         Message
  ----     ------                                                ----   ----                         -------
  Normal   Starting                                              4m25s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-add-tls
  Normal   Starting                                              4m25s  KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hz-prod
  Normal   Successful                                            4m25s  KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-add-tls
  Warning  get certificate retries; ConditionStatus:True         4m11s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           4m11s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               4m11s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate retries; ConditionStatus:True         4m11s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           4m11s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               4m11s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                     4m11s  KubeDB Ops-manager Operator  Successfully synced TLS certificates
  Warning  get certificate retries; ConditionStatus:True         4m1s   KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           4m1s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               4m1s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate retries; ConditionStatus:True         4m1s   KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           4m1s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               4m1s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                     4m1s   KubeDB Ops-manager Operator  Successfully synced TLS certificates
  Normal   UpdateStatefulSets                                    3m51s  KubeDB Ops-manager Operator  Successfully updated pet sets
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-0      3m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-1      3m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-2      3m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-2
  Warning  running pod; ConditionStatus:False                    3m41s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m21s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m21s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m51s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m51s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m41s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m41s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m21s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m21s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  2m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    2m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  2m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    2m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  111s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  111s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    111s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  101s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  101s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    101s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  91s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  91s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-2  91s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-2
  Normal   RestartNodes                                          81s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                              81s    KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hz-prod
  Normal   Successful                                            81s    KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-add-tls

```

Now, Let's exec into a hazelcast pod and verify the configuration that the TLS is enabled.

```bash
kubectl exec -n demo hz-prod-0 -- cat /data/hazelcast/hazelcast.yaml | grep -A 1 -i ssl

Defaulted container "hazelcast" out of: hazelcast, hazelcast-init (init)
    ssl:
      enabled: true
```

We can see from the above output that, TLS is enabled.

## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
kubectl exec -n demo hz-prod-0 -- /bin/sh -c '\
                    openssl s_client -connect localhost:5701 -showcerts < /dev/null 2>/dev/null | \
                    sed -ne "/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p" > /tmp/server.crt && \
                    openssl x509 -in /tmp/server.crt -noout -enddate'

Defaulted container "hazelcast" out of: hazelcast, hazelcast-init (init)
notAfter=Nov 17 05:39:04 2025 GMT

```
### Create HazelcastOpsRequest

Now we are going to increase it using a HazelcastOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hz-prod
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `hz-prod`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this hazelcast cluster.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/reconfigure-tls/hzops-rotate.yaml
hazelcastopsrequest.ops.kubedb.com/hzops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `HazelcastOpsRequest` to be `Successful`.  Run the following command to watch `HazelcastOpsRequest` CRO,

```bash
$ kubectl get hazelcastopsrequests -n demo hzops-rotate
NAME            TYPE             STATUS       AGE
hzops-rotate    ReconfigureTLS   Successful   4m4s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe hazelcastopsrequest -n demo hzops-rotate
Name:         hzops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T06:09:53Z
  Generation:          1
  Resource Version:    5434588
  UID:                 b496d26c-4941-4433-9d8b-8dd353ade6d0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hz-prod
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-08-19T06:09:53Z
    Message:               Hazelcast ops-request has started to reconfigure tls for Hazelcast nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-08-19T06:10:16Z
    Message:               Successfully synced TLS certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-08-19T06:10:06Z
    Message:               get certificate retries; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificateRetries
    Last Transition Time:  2025-08-19T06:10:06Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2025-08-19T06:10:06Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-08-19T06:10:27Z
    Message:               Successfully updated pet sets
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-19T06:13:06Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-19T06:10:37Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-0
    Last Transition Time:  2025-08-19T06:10:37Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-1
    Last Transition Time:  2025-08-19T06:10:37Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-2
    Last Transition Time:  2025-08-19T06:10:37Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-19T06:10:57Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-0
    Last Transition Time:  2025-08-19T06:11:37Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-1
    Last Transition Time:  2025-08-19T06:12:56Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-2
    Last Transition Time:  2025-08-19T06:13:06Z
    Message:               Successfully completed reconfigureTLS for Hazelcast.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                Age    From                         Message
  ----     ------                                                ----   ----                         -------
  Normal   Starting                                              4m15s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-rotate
  Normal   Starting                                              4m15s  KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hz-prod
  Normal   Successful                                            4m15s  KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-rotate
  Warning  get certificate retries; ConditionStatus:True         4m2s   KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           4m2s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               4m2s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate retries; ConditionStatus:True         4m2s   KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           4m2s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               4m2s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                     4m2s   KubeDB Ops-manager Operator  Successfully synced TLS certificates
  Warning  get certificate retries; ConditionStatus:True         3m52s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           3m52s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               3m52s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate retries; ConditionStatus:True         3m52s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           3m52s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               3m52s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                     3m52s  KubeDB Ops-manager Operator  Successfully synced TLS certificates
  Normal   UpdateStatefulSets                                    3m41s  KubeDB Ops-manager Operator  Successfully updated pet sets
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-0      3m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-1      3m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-2      3m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-2
  Warning  running pod; ConditionStatus:False                    3m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m51s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m51s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m41s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m41s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  2m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    2m31s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  2m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    2m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    2m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  112s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  112s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    112s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  102s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  102s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    102s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  92s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  92s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    92s    KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  82s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  82s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    82s    KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  72s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  72s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-2  72s    KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-2
  Normal   RestartNodes                                          62s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                              62s    KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hz-prod
  Normal   Successful                                            62s    KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -n demo hz-prod-0 -- /bin/sh -c '\
                                   openssl s_client -connect localhost:5701 -showcerts < /dev/null 2>/dev/null | \
                                   sed -ne "/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p" > /tmp/server.crt && \
                                   openssl x509 -in /tmp/server.crt -noout -enddate'
Defaulted container "hazelcast" out of: hazelcast, hazelcast-init (init)
notAfter=Nov 17 06:10:38 2025 GMT
```

As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubedb-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls hz-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/hz-new-ca created
```

Now, Let's create a new `Issuer` using the `hz-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: hz-new-issuer
  namespace: demo
spec:
  ca:
    secretName: hz-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/reconfigure-tls/hazelcast-new-issuer.yaml
issuer.cert-manager.io/hz-new-issuer created
```

### Create HazelcastOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `HazelcastOpsRequest` CRO with the newly created issuer. Below is the YAML of the `HazelcastOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hz-prod
  tls:
    issuerRef:
      name: hz-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `hz-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our hazelcast.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/reconfigure-tls/hazelcast-update-tls-issuer.yaml
Hazelcastopsrequest.ops.kubedb.com/hzops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `HazelcastOpsRequest` to be `Successful`.  Run the following command to watch `HazelcastOpsRequest` CRO,

```bash
$ kubectl get hazelcastopsrequests -n demo hzops-update-issuer
NAME                  TYPE             STATUS       AGE
hzops-update-issuer   ReconfigureTLS   Successful   8m6s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe hazelcastopsrequest -n demo hzops-update-issuer
Name:         hzops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T06:22:15Z
  Generation:          1
  Resource Version:    5436918
  UID:                 a5592739-6968-44be-9d73-800c719853d5
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hz-prod
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       hz-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-08-19T06:22:15Z
    Message:               Hazelcast ops-request has started to reconfigure tls for Hazelcast nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-08-19T06:22:38Z
    Message:               Successfully synced TLS certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-08-19T06:22:28Z
    Message:               get certificate retries; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificateRetries
    Last Transition Time:  2025-08-19T06:22:28Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2025-08-19T06:22:28Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-08-19T06:22:48Z
    Message:               Successfully updated pet sets
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-19T06:25:48Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-19T06:22:58Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-0
    Last Transition Time:  2025-08-19T06:22:58Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-1
    Last Transition Time:  2025-08-19T06:22:58Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-2
    Last Transition Time:  2025-08-19T06:22:58Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-19T06:23:08Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-0
    Last Transition Time:  2025-08-19T06:24:58Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-1
    Last Transition Time:  2025-08-19T06:25:38Z
    Message:               running pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  RunningPod--hz-prod-2
    Last Transition Time:  2025-08-19T06:25:48Z
    Message:               Successfully completed reconfigureTLS for Hazelcast.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                Age    From                         Message
  ----     ------                                                ----   ----                         -------
  Normal   Starting                                              6m5s   KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-update-issuer
  Normal   Starting                                              6m5s   KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hz-prod
  Normal   Successful                                            6m5s   KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-update-issuer
  Warning  get certificate retries; ConditionStatus:True         5m52s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           5m52s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               5m52s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate retries; ConditionStatus:True         5m52s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           5m52s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               5m52s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                     5m52s  KubeDB Ops-manager Operator  Successfully synced TLS certificates
  Warning  get certificate retries; ConditionStatus:True         5m42s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           5m42s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               5m42s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate retries; ConditionStatus:True         5m42s  KubeDB Ops-manager Operator  get certificate retries; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True           5m42s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True               5m42s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                     5m42s  KubeDB Ops-manager Operator  Successfully synced TLS certificates
  Normal   UpdateStatefulSets                                    5m32s  KubeDB Ops-manager Operator  Successfully updated pet sets
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-0      5m22s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-1      5m22s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-2      5m22s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-2
  Warning  running pod; ConditionStatus:False                    5m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  5m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    5m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  5m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    5m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  4m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    4m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  4m42s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    4m42s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  4m32s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    4m32s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  4m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    4m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  4m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    4m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  4m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    4m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m42s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m42s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m32s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                    3m32s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  3m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    3m22s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  3m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    3m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  3m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  3m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    3m2s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  2m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:False                    2m52s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-0  2m42s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-1  2m42s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  running pod; ConditionStatus:True; PodName:hz-prod-2  2m42s  KubeDB Ops-manager Operator  running pod; ConditionStatus:True; PodName:hz-prod-2
  Normal   RestartNodes                                          2m32s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                              2m32s  KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hz-prod
  Normal   Successful                                            2m32s  KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-update-issuer
```

Now, Let's exec into a hazelcast server pod and find out the ca subject to see if it matches the one we have provided.

```bash
kubectl exec -n demo hz-prod-0 -- /bin/sh -c '\
                 openssl s_client -connect localhost:5701 < /dev/null 2>/dev/null | \
                 grep -A8 "Certificate chain"'

Defaulted container "hazelcast" out of: hazelcast, hazelcast-init (init)
Certificate chain
 0 s:CN=hz-prod
   i:CN=ca-updated, O=kubedb-updated
   a:PKEY: rsaEncryption, 2048 (bit); sigalg: RSA-SHA256
   v:NotBefore: Aug 19 06:22:18 2025 GMT; NotAfter: Nov 17 06:22:18 2025 GMT
 1 s:CN=ca-updated, O=kubedb-updated
   i:CN=ca-updated, O=kubedb-updated
   a:PKEY: rsaEncryption, 2048 (bit); sigalg: RSA-SHA256
   v:NotBefore: Aug 19 06:17:34 2025 GMT; NotAfter: Aug 19 06:17:34 2026 GMT
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a HazelcastOpsRequest.

### Create HazelcastOpsRequest

Below is the YAML of the `HazelcastOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hz-prod
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `hz-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Hazelcast.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/reconfigure-tls/hzops-remove.yaml
hazelcastopsrequest.ops.kubedb.com/hzops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `HazelcastOpsRequest` to be `Successful`.  Run the following command to watch `HazelcastOpsRequest` CRO,

```bash
$ kubectl get hazelcastopsrequest -n demo hzops-remove
NAME           TYPE             STATUS        AGE
hzops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe hazelcastopsrequest -n demo hzops-remove

```

Now, Let's exec into one of the broker node and find out that TLS is disabled or not.

```bash
kubectl exec -n demo hz-prod-0 -- cat /data/hazelcast/hazelcast.yaml | grep -A 1 -i ssl

Defaulted container "hazelcast" out of: hazelcast, hazelcast-init (init)
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete opsrequest hzops-add-tls hzops-remove hzops-rotate hzops-update-issuer
kubectl delete hazelcast -n demo hz-prod
kubectl delete issuer -n demo hz-issuer hz-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).

Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
